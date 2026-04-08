'use client';

import { useState, useRef, useEffect, useCallback, useSyncExternalStore } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import {
  createAgentClient,
  AGENT_ID,
  streamAgent,
  approveAgentToolCall,
  declineAgentToolCall,
  type StreamChunk
} from '@/packages/lib/agent-client';
import { type ChatContext, formatContextsForAgent } from './chat-context';
import { v4 as uuid } from 'uuid';
import { chatStreamStore } from './chat-stream-store';

export type MessagePart =
  | { type: 'text'; content: string }
  | {
      type: 'tool-call';
      toolName: string;
      toolCallId: string;
      args?: unknown;
      status: 'running' | 'done';
    };

export interface TokenUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  costUsd?: number;
  durationMs?: number;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  parts?: MessagePart[];
  timestamp: Date;
  contexts?: ChatContext[];
  kind?: 'status';
  usage?: TokenUsage;
}

interface UseAgentChatOptions {
  threadId: string | null;
  resourceId?: string;
  contexts?: ChatContext[];
  autoRunTools?: boolean;
  model?: string;
  onFirstMessage?: (content: string) => void;
  waitForThread?: (id: string) => Promise<void>;
}

async function getAuthHeaders(
  token: string | null,
  organizationId: string | undefined | null
): Promise<Record<string, string>> {
  const headers: Record<string, string> = {};

  let authToken = token;
  if (!authToken) {
    try {
      const session = await authClient.getSession();
      authToken = session?.data?.session?.token ?? null;
    } catch {
      // ignore
    }
  }

  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`;
  }
  if (organizationId) {
    headers['X-Organization-Id'] = organizationId;
  }

  return headers;
}

function extractText(content: unknown): string {
  if (typeof content === 'string') return content;
  if (Array.isArray(content)) {
    return content
      .filter((p: { type?: string; text?: string }) => p.type === 'text' && p.text)
      .map((p: { text: string }) => p.text)
      .join('');
  }
  if (content && typeof content === 'object') {
    const obj = content as { content?: string; parts?: { type?: string; text?: string }[] };
    if (typeof obj.content === 'string') return obj.content;
    if (Array.isArray(obj.parts)) return extractText(obj.parts);
  }
  return '';
}

function extractMessageParts(content: unknown): { text: string; parts: MessagePart[] } {
  if (typeof content === 'string') return { text: content, parts: [] };
  if (!content || typeof content !== 'object') return { text: '', parts: [] };

  const obj = content as {
    format?: number;
    content?: string;
    parts?: Array<Record<string, unknown>>;
    toolInvocations?: Array<{
      toolCallId?: string;
      toolName?: string;
      args?: unknown;
      state?: string;
    }>;
  };

  const parts: MessagePart[] = [];
  let text = '';

  if (Array.isArray(obj.parts)) {
    for (const part of obj.parts) {
      if (part.type === 'text') {
        const t = (part.text as string) || (part.content as string) || '';
        if (t) {
          text += t;
          parts.push({ type: 'text', content: t });
        }
      } else if (part.type === 'tool-invocation') {
        const inv = part.toolInvocation as
          | {
              toolCallId?: string;
              toolName?: string;
              args?: unknown;
              state?: string;
            }
          | undefined;
        if (inv) {
          parts.push({
            type: 'tool-call',
            toolName: inv.toolName || 'tool',
            toolCallId: inv.toolCallId || '',
            args: inv.args,
            status: 'done'
          });
        }
      }
    }
  }

  if (parts.every((p) => p.type === 'tool-call') && Array.isArray(obj.toolInvocations)) {
    for (const inv of obj.toolInvocations) {
      if (!parts.some((p) => p.type === 'tool-call' && p.toolCallId === inv.toolCallId)) {
        parts.push({
          type: 'tool-call',
          toolName: inv.toolName || 'tool',
          toolCallId: inv.toolCallId || '',
          args: inv.args,
          status: 'done'
        });
      }
    }
  }

  if (!text && typeof obj.content === 'string') {
    text = obj.content;
  }

  return { text, parts };
}

export interface AgentQuestionFieldOption {
  label: string;
  value: string;
}

export interface AgentQuestionField {
  name: string;
  label: string;
  type: 'text' | 'password' | 'select' | 'toggle' | 'textarea';
  required?: boolean;
  placeholder?: string;
  defaultValue?: string;
  options?: AgentQuestionFieldOption[];
}

export interface AgentQuestion {
  title: string;
  description?: string;
  fields: AgentQuestionField[];
}

export interface PendingToolApproval {
  runId: string;
  toolCallId: string;
  toolName: string;
  args: unknown;
}

export interface OmStatus {
  messages: { tokens: number; threshold: number };
  observations: { tokens: number; threshold: number };
  isObserving: boolean;
  observationsText: string | null;
}

export function useAgentChat({
  threadId,
  resourceId,
  contexts = [],
  autoRunTools = false,
  model,
  onFirstMessage,
  waitForThread
}: UseAgentChatOptions) {
  const subscribe = useCallback(
    (cb: () => void) => chatStreamStore.subscribe(threadId, cb),
    [threadId]
  );
  const snapshot = useSyncExternalStore(
    subscribe,
    () => chatStreamStore.getSnapshot(threadId),
    chatStreamStore.getEmptySnapshot
  );

  const { messages, isStreaming, pendingToolApproval, omStatus } = snapshot;

  const [inputValue, setInputValue] = useState('');
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [activeQuestion, setActiveQuestion] = useState<AgentQuestion | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const lastAutoApprovedRef = useRef<string | null>(null);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const organizationId = activeOrg?.id;

  const scrollToBottom = useCallback(() => {
    if (scrollRef.current) {
      const el = scrollRef.current;
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight;
      });
    }
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  useEffect(() => {
    if (!threadId) return;

    if (chatStreamStore.hasActiveStream(threadId)) return;

    let cancelled = false;
    setIsLoadingHistory(true);

    (async () => {
      try {
        if (waitForThread) await waitForThread(threadId);
        if (cancelled) return;

        const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
        const client = createAgentClient(headers);
        const thread = client.getMemoryThread({ threadId, agentId: AGENT_ID });
        const result = await thread.listMessages({
          resourceId: resourceId || threadId
        });

        if (cancelled) return;
        if (chatStreamStore.hasActiveStream(threadId)) return;

        const msgs: ChatMessage[] = [];
        const rawMessages = result?.messages ?? [];
        for (const msg of rawMessages) {
          const role = msg.role === 'user' ? 'user' : msg.role === 'assistant' ? 'assistant' : null;
          if (!role) continue;

          if (role === 'assistant') {
            const { text, parts } = extractMessageParts(msg.content);
            if (!text && parts.length === 0) continue;
            msgs.push({
              id: msg.id,
              role,
              content: text,
              ...(parts.length > 0 ? { parts } : {}),
              timestamp: msg.createdAt ? new Date(msg.createdAt) : new Date()
            });
          } else {
            const text = extractText(msg.content);
            if (!text) continue;
            msgs.push({
              id: msg.id,
              role,
              content: text,
              timestamp: msg.createdAt ? new Date(msg.createdAt) : new Date()
            });
          }
        }
        chatStreamStore.setMessages(threadId, msgs);
      } catch {
        // thread may not exist on server yet
      } finally {
        if (!cancelled) setIsLoadingHistory(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [threadId, resourceId, token, organizationId, waitForThread]);

  const streamResponse = useCallback(
    async (userContent: string) => {
      if (!threadId) return;

      const abortController = new AbortController();
      const assistantMessageId = uuid();
      chatStreamStore.beginStream(threadId, assistantMessageId, abortController);

      try {
        const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
        const contextPrefix = formatContextsForAgent(contexts);
        const stream = streamAgent(
          contextPrefix + userContent,
          threadId,
          resourceId || threadId,
          headers,
          abortController.signal,
          model,
          !autoRunTools
        );

        for await (const chunk of stream) {
          chatStreamStore.handleChunk(threadId, chunk);
        }
      } catch (err: unknown) {
        if (err instanceof Error && err.name === 'AbortError') return;
        const errorMessage =
          err instanceof Error ? err.message : 'Failed to get response from AI agent';
        chatStreamStore.finishStream(threadId, errorMessage);
        return;
      }
      chatStreamStore.finishStream(threadId);
    },
    [threadId, resourceId, token, organizationId, contexts, model, autoRunTools]
  );

  const handleApproveToolCall = useCallback(async () => {
    if (!threadId) return;
    const snap = chatStreamStore.getSnapshot(threadId);
    const pending = snap.pendingToolApproval;
    if (!pending) return;

    const assistantMessageId = chatStreamStore.prepareApprovalStream(threadId);
    if (!assistantMessageId) {
      chatStreamStore.stopStream(threadId);
      return;
    }

    const abortController = new AbortController();
    chatStreamStore.startApprovalStream(threadId, abortController);

    try {
      const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
      const stream = approveAgentToolCall(
        { runId: pending.runId, toolCallId: pending.toolCallId },
        headers,
        abortController.signal
      );

      for await (const chunk of stream) {
        chatStreamStore.handleChunk(threadId, chunk);
      }
    } catch {
      chatStreamStore.appendErrorToLastAssistant(threadId, '\n\n_Tool execution failed._');
    }
    chatStreamStore.finishApprovalStream(threadId);
  }, [threadId, token, organizationId]);

  const handleDeclineToolCall = useCallback(async () => {
    if (!threadId) return;
    const snap = chatStreamStore.getSnapshot(threadId);
    const pending = snap.pendingToolApproval;
    if (!pending) return;

    chatStreamStore.declineApproval(threadId);

    try {
      const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
      await declineAgentToolCall({ runId: pending.runId, toolCallId: pending.toolCallId }, headers);
      chatStreamStore.appendErrorToLastAssistant(threadId, '\n\n_Tool call was declined._');
    } catch {
      // ignore
    }
  }, [threadId, token, organizationId]);

  useEffect(() => {
    if (!pendingToolApproval) {
      lastAutoApprovedRef.current = null;
      return;
    }
    const isInternalTool =
      pendingToolApproval.toolName === 'search_tools' ||
      pendingToolApproval.toolName === 'load_tool';
    if (!autoRunTools && !isInternalTool) return;
    const key = `${pendingToolApproval.runId}-${pendingToolApproval.toolCallId}`;
    if (lastAutoApprovedRef.current === key) return;
    lastAutoApprovedRef.current = key;
    handleApproveToolCall();
  }, [pendingToolApproval, autoRunTools, handleApproveToolCall]);

  const handleSubmit = useCallback(
    (e?: React.FormEvent) => {
      e?.preventDefault();
      if (!inputValue.trim() || !threadId) return;
      if (chatStreamStore.getSnapshot(threadId).isStreaming) return;

      const content = inputValue.trim();
      const userMessage: ChatMessage = {
        id: uuid(),
        role: 'user',
        content,
        timestamp: new Date(),
        ...(contexts.length > 0 ? { contexts: [...contexts] } : {})
      };

      const snap = chatStreamStore.getSnapshot(threadId);
      if (snap.messages.length === 0 && onFirstMessage) {
        onFirstMessage(content);
      }

      chatStreamStore.addUserMessage(threadId, userMessage);
      setInputValue('');
      streamResponse(content);
    },
    [inputValue, threadId, streamResponse, onFirstMessage, contexts]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSubmit();
      }
    },
    [handleSubmit]
  );

  const handleSuggestionClick = useCallback(
    (text: string) => {
      if (!threadId) return;
      if (chatStreamStore.getSnapshot(threadId).isStreaming) return;

      const userMessage: ChatMessage = {
        id: uuid(),
        role: 'user',
        content: text,
        timestamp: new Date(),
        ...(contexts.length > 0 ? { contexts: [...contexts] } : {})
      };

      const snap = chatStreamStore.getSnapshot(threadId);
      if (snap.messages.length === 0 && onFirstMessage) {
        onFirstMessage(text);
      }

      chatStreamStore.addUserMessage(threadId, userMessage);
      setInputValue('');
      streamResponse(text);
    },
    [threadId, streamResponse, onFirstMessage, contexts]
  );

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(e.target.value);
  }, []);

  const submitQuestionResponse = useCallback(
    (answers: Record<string, string>) => {
      setActiveQuestion(null);
      if (!threadId) return;
      if (chatStreamStore.getSnapshot(threadId).isStreaming) return;

      const formatted = Object.entries(answers)
        .map(([key, value]) => `${key}: ${value}`)
        .join('\n');
      const content = `[user_response]\n${formatted}`;

      const userMessage: ChatMessage = {
        id: uuid(),
        role: 'user',
        content,
        timestamp: new Date()
      };
      chatStreamStore.addUserMessage(threadId, userMessage);
      streamResponse(content);
    },
    [threadId, streamResponse]
  );

  const dismissQuestion = useCallback(() => {
    setActiveQuestion(null);
  }, []);

  const stopStreaming = useCallback(() => {
    if (threadId) chatStreamStore.stopStream(threadId);
  }, [threadId]);

  return {
    messages,
    inputValue,
    setInputValue,
    isStreaming,
    isLoadingHistory,
    pendingToolApproval,
    activeQuestion,
    omStatus,
    scrollRef,
    textareaRef,
    handleSubmit,
    handleKeyDown,
    handleSuggestionClick,
    handleInputChange,
    handleApproveToolCall,
    handleDeclineToolCall,
    submitQuestionResponse,
    dismissQuestion,
    stopStreaming
  };
}
