'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import { createAgentClient, AGENT_ID } from '@/packages/lib/agent-client';
import { useAgentConfigured } from '@/packages/hooks/shared/use-config';
import { type ChatContext, formatContextsForAgent } from './chat-context';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  contexts?: ChatContext[];
}

interface UseAgentChatOptions {
  threadId: string | null;
  resourceId?: string;
  contexts?: ChatContext[];
  autoRunTools?: boolean;
  onFirstMessage?: (content: string) => void;
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

export interface PendingToolApproval {
  runId: string;
  toolCallId: string;
  toolName: string;
  args: unknown;
}

export function useAgentChat({
  threadId,
  resourceId,
  contexts = [],
  autoRunTools = false,
  onFirstMessage
}: UseAgentChatOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [pendingToolApproval, setPendingToolApproval] = useState<PendingToolApproval | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const pendingApprovalRef = useRef(false);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const organizationId = activeOrg?.id;

  const isAgentEnabled = useAgentConfigured() === true;

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
    if (!threadId || !isAgentEnabled) {
      setMessages([]);
      return;
    }

    let cancelled = false;
    setIsLoadingHistory(true);
    setMessages([]);

    (async () => {
      try {
        const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
        const client = createAgentClient(headers);
        const thread = client.getMemoryThread({ threadId, agentId: AGENT_ID });
        const result = await thread.listMessages({
          resourceId: resourceId || threadId
        });

        if (cancelled) return;

        const msgs: ChatMessage[] = [];
        const rawMessages = result?.messages ?? [];
        for (const msg of rawMessages) {
          const role = msg.role === 'user' ? 'user' : msg.role === 'assistant' ? 'assistant' : null;
          if (!role) continue;
          const text = extractText(msg.content);
          if (!text) continue;
          msgs.push({
            id: msg.id,
            role,
            content: text,
            timestamp: msg.createdAt ? new Date(msg.createdAt) : new Date()
          });
        }
        setMessages(msgs);
      } catch {
        // thread may not exist on server yet
      } finally {
        if (!cancelled) setIsLoadingHistory(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [threadId, resourceId, token, organizationId, isAgentEnabled]);

  const streamResponse = useCallback(
    async (userContent: string) => {
      if (!isAgentEnabled || !threadId) return;

      const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
      const client = createAgentClient(headers);
      const agent = client.getAgent(AGENT_ID);

      const abortController = new AbortController();
      abortRef.current = abortController;
      const runIdRef = { current: '' };

      const assistantMessageId = crypto.randomUUID();
      setMessages((prev) => [
        ...prev,
        { id: assistantMessageId, role: 'assistant', content: '', timestamp: new Date() }
      ]);

      const processNext = async (response: {
        processDataStream: (opts: { onChunk: (chunk: unknown) => Promise<void> }) => Promise<void>;
      }): Promise<void> => {
        let runId = runIdRef.current;
        const chunkHandler = async (chunk: unknown) => {
          if (abortController.signal.aborted) return;
          const c = chunk as { type?: string; runId?: string; payload?: Record<string, unknown> };
          if (c.type === 'start' && c.runId) {
            runIdRef.current = c.runId;
          }
          if (c.type === 'text-delta') {
            const text = c.payload?.text as string | undefined;
            if (text) {
              setMessages((prev) =>
                prev.map((m) =>
                  m.id === assistantMessageId ? { ...m, content: m.content + text } : m
                )
              );
            }
          }
          if (c.type === 'tool-call' || c.type === 'tool-call-approval') {
            const payload = c.payload as
              | {
                  toolCallId?: string;
                  toolName?: string;
                  args?: unknown;
                  runId?: string;
                  id?: string;
                }
              | undefined;
            const rid = payload?.runId ?? runIdRef.current;
            const toolCallId = payload?.toolCallId ?? payload?.id;
            if (rid && toolCallId) {
              pendingApprovalRef.current = true;
              setPendingToolApproval({
                runId: rid,
                toolCallId: String(toolCallId),
                toolName: (payload?.toolName as string) ?? 'tool',
                args: payload?.args ?? {}
              });
            }
          }
          if (c.type === 'finish' && c.payload) {
            const finishReason = (c.payload as { finishReason?: string }).finishReason;
            const suspendPayload = (
              c.payload as {
                suspendPayload?: {
                  toolCallId?: string;
                  runId?: string;
                  toolName?: string;
                  args?: unknown;
                };
              }
            ).suspendPayload;
            if (finishReason === 'suspended' && suspendPayload) {
              pendingApprovalRef.current = true;
              setPendingToolApproval({
                runId: suspendPayload.runId ?? '',
                toolCallId: suspendPayload.toolCallId ?? '',
                toolName: (suspendPayload.toolName as string) ?? 'tool',
                args: suspendPayload.args ?? {}
              });
            }
          }
        };
        await response.processDataStream({ onChunk: chunkHandler });
      };

      try {
        const contextPrefix = formatContextsForAgent(contexts);
        const response = await agent.stream(contextPrefix + userContent, {
          memory: {
            thread: threadId,
            resource: resourceId || threadId
          }
        });
        await processNext(response);
      } catch (err: unknown) {
        if (err instanceof Error && err.name === 'AbortError') return;

        const errorMessage =
          err instanceof Error ? err.message : 'Failed to get response from AI agent';

        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantMessageId
              ? { ...m, content: m.content || `Error: ${errorMessage}` }
              : m
          )
        );
      } finally {
        if (!pendingApprovalRef.current) {
          setIsStreaming(false);
        }
        abortRef.current = null;
      }
    },
    [threadId, resourceId, token, organizationId, isAgentEnabled, contexts]
  );

  const handleApproveToolCall = useCallback(async () => {
    const pending = pendingToolApproval;
    if (!pending || !isAgentEnabled) return;

    setPendingToolApproval(null);
    pendingApprovalRef.current = false;

    const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
    const client = createAgentClient(headers);
    const agent = client.getAgent(AGENT_ID);
    const assistantMessageId = messages.filter((m) => m.role === 'assistant').pop()?.id;
    if (!assistantMessageId) {
      setIsStreaming(false);
      return;
    }

    const abortController = new AbortController();
    abortRef.current = abortController;
    const runIdRef = { current: '' };

    const processNext = async (response: {
      processDataStream: (opts: { onChunk: (chunk: unknown) => Promise<void> }) => Promise<void>;
    }): Promise<void> => {
      const chunkHandler = async (chunk: unknown) => {
        if (abortController.signal.aborted) return;
        const c = chunk as { type?: string; runId?: string; payload?: Record<string, unknown> };
        if (c.type === 'start' && c.runId) runIdRef.current = c.runId;
        if (c.type === 'text-delta') {
          const text = c.payload?.text as string | undefined;
          if (text) {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantMessageId ? { ...m, content: m.content + text } : m
              )
            );
          }
        }
        if (c.type === 'tool-call' || c.type === 'tool-call-approval') {
          const payload = c.payload as
            | {
                toolCallId?: string;
                toolName?: string;
                args?: unknown;
                runId?: string;
                id?: string;
              }
            | undefined;
          const rid = payload?.runId ?? runIdRef.current;
          const toolCallId = payload?.toolCallId ?? payload?.id;
          if (rid && toolCallId) {
            pendingApprovalRef.current = true;
            setPendingToolApproval({
              runId: rid,
              toolCallId: String(toolCallId),
              toolName: (payload?.toolName as string) ?? 'tool',
              args: payload?.args ?? {}
            });
          }
        }
        if (c.type === 'finish' && c.payload) {
          const finishReason = (c.payload as { finishReason?: string }).finishReason;
          const suspendPayload = (
            c.payload as {
              suspendPayload?: {
                toolCallId?: string;
                runId?: string;
                toolName?: string;
                args?: unknown;
              };
            }
          ).suspendPayload;
          if (finishReason === 'suspended' && suspendPayload) {
            pendingApprovalRef.current = true;
            setPendingToolApproval({
              runId: suspendPayload.runId ?? '',
              toolCallId: suspendPayload.toolCallId ?? '',
              toolName: (suspendPayload.toolName as string) ?? 'tool',
              args: suspendPayload.args ?? {}
            });
          }
        }
      };
      await response.processDataStream({ onChunk: chunkHandler });
    };

    try {
      const response = await agent.approveToolCall({
        runId: pending.runId,
        toolCallId: pending.toolCallId
      });
      await processNext(response);
    } catch {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === assistantMessageId
            ? { ...m, content: m.content + '\n\n_Tool execution failed._' }
            : m
        )
      );
    } finally {
      if (!pendingApprovalRef.current) setIsStreaming(false);
      abortRef.current = null;
    }
  }, [pendingToolApproval, isAgentEnabled, token, organizationId, messages]);

  const handleDeclineToolCall = useCallback(async () => {
    const pending = pendingToolApproval;
    if (!pending || !isAgentEnabled) return;

    setPendingToolApproval(null);
    pendingApprovalRef.current = false;

    try {
      const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
      const client = createAgentClient(headers);
      const agent = client.getAgent(AGENT_ID);
      await agent.declineToolCall({
        runId: pending.runId,
        toolCallId: pending.toolCallId
      });
      const assistantMessageId = messages.filter((m) => m.role === 'assistant').pop()?.id;
      if (assistantMessageId) {
        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantMessageId
              ? { ...m, content: m.content + '\n\n_Tool call was declined._' }
              : m
          )
        );
      }
    } catch {
      // ignore
    } finally {
      setIsStreaming(false);
    }
  }, [pendingToolApproval, isAgentEnabled, token, organizationId, messages]);

  const lastAutoApprovedRef = useRef<string | null>(null);
  useEffect(() => {
    if (!pendingToolApproval) {
      lastAutoApprovedRef.current = null;
      return;
    }
    if (!autoRunTools) return;
    const key = `${pendingToolApproval.runId}-${pendingToolApproval.toolCallId}`;
    if (lastAutoApprovedRef.current === key) return;
    lastAutoApprovedRef.current = key;
    handleApproveToolCall();
  }, [pendingToolApproval, autoRunTools, handleApproveToolCall]);

  const handleSubmit = useCallback(
    (e?: React.FormEvent) => {
      e?.preventDefault();
      if (!inputValue.trim() || isStreaming || !threadId) return;

      const content = inputValue.trim();
      const userMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'user',
        content,
        timestamp: new Date(),
        ...(contexts.length > 0 ? { contexts: [...contexts] } : {})
      };

      setMessages((prev) => {
        const isFirst = prev.length === 0;
        if (isFirst && onFirstMessage) {
          onFirstMessage(content);
        }
        return [...prev, userMessage];
      });
      setInputValue('');
      setIsStreaming(true);
      streamResponse(content);
    },
    [inputValue, isStreaming, threadId, streamResponse, onFirstMessage, contexts]
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
      if (!isStreaming && threadId) {
        setInputValue('');
        const userMessage: ChatMessage = {
          id: crypto.randomUUID(),
          role: 'user',
          content: text,
          timestamp: new Date(),
          ...(contexts.length > 0 ? { contexts: [...contexts] } : {})
        };
        setMessages((prev) => {
          const isFirst = prev.length === 0;
          if (isFirst && onFirstMessage) {
            onFirstMessage(text);
          }
          return [...prev, userMessage];
        });
        setIsStreaming(true);
        streamResponse(text);
      }
    },
    [isStreaming, threadId, streamResponse, onFirstMessage, contexts]
  );

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(e.target.value);
  }, []);

  const stopStreaming = useCallback(() => {
    abortRef.current?.abort();
    setIsStreaming(false);
  }, []);

  return {
    messages,
    inputValue,
    isStreaming,
    isLoadingHistory,
    isAgentConfigured: isAgentEnabled,
    pendingToolApproval,
    scrollRef,
    textareaRef,
    handleSubmit,
    handleKeyDown,
    handleSuggestionClick,
    handleInputChange,
    handleApproveToolCall,
    handleDeclineToolCall,
    stopStreaming
  };
}
