'use client';

import { useState, useRef, useEffect, useCallback, useSyncExternalStore } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import {
  streamAgent,
  cancelStream,
  getThreadMessages,
  type StreamChunk,
  type RawThreadMessage
} from '@/packages/lib/agent-client';
import { type ChatContext, formatContextsForAgent } from './chat-context';
import { v4 as uuid } from 'uuid';
import { chatStreamStore, isRateLimitError } from './chat-stream-store';

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
  errorKind?: 'rate-limited' | 'generic-error';
}

interface UseAgentChatOptions {
  threadId: string | null;
  resourceId?: string;
  agentId?: string;
  readOnly?: boolean;
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

/**
 * Merge raw DB rows back into ChatMessages with `parts`, matching what
 * the streaming path produces.  A single agent turn in the DB may look like:
 *   assistant (tool_calls) → tool → tool → assistant (tool_calls) → tool → assistant (final text)
 * We collapse that whole sequence into one ChatMessage with interleaved parts.
 */
function rebuildChatMessages(rawMessages: RawThreadMessage[]): ChatMessage[] {
  const result: ChatMessage[] = [];
  let i = 0;

  while (i < rawMessages.length) {
    const msg = rawMessages[i];

    if (msg.role === 'user') {
      result.push({
        id: msg.id,
        role: 'user',
        content: msg.content || '',
        timestamp: msg.created_at ? new Date(msg.created_at) : new Date()
      });
      i++;
      continue;
    }

    if (msg.role === 'assistant') {
      const parts: MessagePart[] = [];
      let fullContent = '';
      const firstId = msg.id;
      const firstTs = msg.created_at;

      while (i < rawMessages.length && rawMessages[i].role !== 'user') {
        const cur = rawMessages[i];

        if (cur.role === 'assistant') {
          if (cur.content) {
            parts.push({ type: 'text', content: cur.content });
            fullContent += (fullContent ? '\n' : '') + cur.content;
          }
          for (const tc of cur.tool_calls ?? []) {
            let args: unknown;
            try {
              args = tc.function.arguments ? JSON.parse(tc.function.arguments) : undefined;
            } catch {
              args = tc.function.arguments;
            }
            parts.push({
              type: 'tool-call',
              toolName: tc.function.name || 'tool',
              toolCallId: tc.id,
              args,
              status: 'done' as const
            });
          }
        }

        i++;
      }

      if (!fullContent && parts.length === 0) continue;

      const hasToolParts = parts.some((p) => p.type === 'tool-call');

      result.push({
        id: firstId,
        role: 'assistant',
        content: fullContent,
        ...(hasToolParts ? { parts } : {}),
        timestamp: firstTs ? new Date(firstTs) : new Date()
      });
      continue;
    }

    i++;
  }

  return result;
}

export function useAgentChat({
  threadId,
  resourceId,
  readOnly = false,
  contexts = [],
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

  const { messages, isStreaming } = snapshot;

  const [inputValue, setInputValue] = useState('');
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [activeQuestion, setActiveQuestion] = useState<AgentQuestion | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

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
        const rawMessages = await getThreadMessages(threadId, headers);

        if (cancelled) return;
        if (chatStreamStore.hasActiveStream(threadId)) return;

        const msgs = rebuildChatMessages(rawMessages);
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

      if (waitForThread) await waitForThread(threadId);

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
          model
        );

        for await (const chunk of stream) {
          chatStreamStore.handleChunk(threadId, chunk);
        }
      } catch (err: unknown) {
        if (err instanceof Error && err.name === 'AbortError') return;
        const errorMessage =
          err instanceof Error ? err.message : 'Failed to get response from AI agent';
        chatStreamStore.finishStream(
          threadId,
          errorMessage,
          isRateLimitError(err) ? 'rate-limited' : 'generic-error'
        );
        return;
      }

      const snap = chatStreamStore.getSnapshot(threadId);
      if (snap.isStreaming) {
        chatStreamStore.finishStream(threadId);
      }
    },
    [threadId, resourceId, token, organizationId, contexts, model, waitForThread]
  );

  const handleSubmit = useCallback(
    (e?: React.FormEvent) => {
      e?.preventDefault();
      if (readOnly) return;
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

  const stopStreaming = useCallback(async () => {
    if (!threadId) return;
    chatStreamStore.stopStream(threadId);
    try {
      const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
      await cancelStream(threadId, headers);
    } catch {
      // best-effort cancel on server
    }
  }, [threadId, token, organizationId]);

  return {
    messages,
    inputValue,
    setInputValue,
    isStreaming,
    isLoadingHistory,
    readOnly,
    pendingToolApproval: null as PendingToolApproval | null,
    activeQuestion,
    omStatus: null as OmStatus | null,
    scrollRef,
    textareaRef,
    handleSubmit,
    handleKeyDown,
    handleSuggestionClick,
    handleInputChange,
    handleApproveToolCall: () => {},
    handleDeclineToolCall: () => {},
    submitQuestionResponse,
    dismissQuestion,
    stopStreaming
  };
}
