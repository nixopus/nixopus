import type { StreamChunk } from '@/packages/lib/agent-client';
import type { ChatMessage, PendingToolApproval, OmStatus, TokenUsage } from './use-agent-chat';

export function isRateLimitError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false;
  const e = error as Record<string, unknown>;
  if (e.statusCode === 429 || e.status === 429) return true;
  const msg = typeof e.message === 'string' ? e.message.toLowerCase() : '';
  return (
    msg.includes('quota') ||
    msg.includes('rate limit') ||
    msg.includes('rate_limit') ||
    msg.includes('resource_exhausted') ||
    msg.includes('too_many_requests')
  );
}

export interface SessionSnapshot {
  messages: ChatMessage[];
  isStreaming: boolean;
  pendingToolApproval: PendingToolApproval | null;
  omStatus: OmStatus | null;
}

interface SessionInternal {
  snapshot: SessionSnapshot;
  abortController: AbortController | null;
  firstTextDeltaTime: number | null;
  assistantMessageId: string | null;
  threadId: string;
}

type Listener = () => void;

const EMPTY_SNAPSHOT: SessionSnapshot = Object.freeze({
  messages: [],
  isStreaming: false,
  pendingToolApproval: null,
  omStatus: null
});

class ChatStreamStore {
  private sessions = new Map<string, SessionInternal>();
  private listeners = new Map<string, Set<Listener>>();

  private notify(threadId: string) {
    this.listeners.get(threadId)?.forEach((fn) => fn());
  }

  private getOrCreate(threadId: string): SessionInternal {
    let s = this.sessions.get(threadId);
    if (!s) {
      s = {
        snapshot: { messages: [], isStreaming: false, pendingToolApproval: null, omStatus: null },
        abortController: null,
        firstTextDeltaTime: null,
        assistantMessageId: null,
        threadId
      };
      this.sessions.set(threadId, s);
    }
    return s;
  }

  subscribe = (threadId: string | null, listener: Listener): (() => void) => {
    if (!threadId) return () => {};
    let set = this.listeners.get(threadId);
    if (!set) {
      set = new Set();
      this.listeners.set(threadId, set);
    }
    set.add(listener);
    return () => {
      set!.delete(listener);
      if (set!.size === 0) this.listeners.delete(threadId);
    };
  };

  getSnapshot = (threadId: string | null): SessionSnapshot => {
    if (!threadId) return EMPTY_SNAPSHOT;
    return this.sessions.get(threadId)?.snapshot ?? EMPTY_SNAPSHOT;
  };

  getEmptySnapshot = (): SessionSnapshot => EMPTY_SNAPSHOT;

  hasActiveStream(threadId: string | null): boolean {
    if (!threadId) return false;
    return this.sessions.get(threadId)?.snapshot.isStreaming ?? false;
  }

  setMessages(threadId: string, messages: ChatMessage[]) {
    const s = this.getOrCreate(threadId);
    s.snapshot = { ...s.snapshot, messages };
    this.notify(threadId);
  }

  addUserMessage(threadId: string, message: ChatMessage) {
    const s = this.getOrCreate(threadId);
    s.snapshot = { ...s.snapshot, messages: [...s.snapshot.messages, message] };
    this.notify(threadId);
  }

  beginStream(threadId: string, assistantMessageId: string, abortController: AbortController) {
    const s = this.getOrCreate(threadId);
    s.abortController = abortController;
    s.assistantMessageId = assistantMessageId;
    s.firstTextDeltaTime = null;
    s.snapshot = {
      ...s.snapshot,
      messages: [
        ...s.snapshot.messages,
        {
          id: assistantMessageId,
          role: 'assistant',
          content: '',
          timestamp: new Date(),
          parts: []
        }
      ],
      isStreaming: true,
      pendingToolApproval: null
    };
    this.notify(threadId);
  }

  handleChunk(threadId: string, chunk: StreamChunk) {
    const s = this.sessions.get(threadId);
    if (!s || !s.assistantMessageId) return;
    const amId = s.assistantMessageId;

    if (s.abortController?.signal.aborted) return;

    switch (chunk.type) {
      case 'content': {
        if (s.firstTextDeltaTime === null) {
          s.firstTextDeltaTime = Date.now();
        }
        const text = typeof chunk.data === 'string' ? chunk.data : '';
        if (!text) return;

        s.snapshot = {
          ...s.snapshot,
          messages: s.snapshot.messages.map((m) => {
            if (m.id !== amId) return m;
            const newContent = m.content + text;
            const parts = [...(m.parts || [])];
            const lastPart = parts[parts.length - 1];
            if (lastPart?.type === 'text') {
              parts[parts.length - 1] = { ...lastPart, content: lastPart.content + text };
            } else {
              parts.push({ type: 'text' as const, content: text });
            }
            return { ...m, content: newContent, parts };
          })
        };
        this.notify(threadId);
        return;
      }

      case 'tool_calls': {
        const toolCalls = Array.isArray(chunk.data) ? chunk.data : [];
        s.snapshot = {
          ...s.snapshot,
          messages: s.snapshot.messages.map((m) => {
            if (m.id !== amId) return m;
            const parts = [...(m.parts || [])];
            for (const tc of toolCalls as Array<{
              id?: string;
              function?: { name?: string; arguments?: string };
            }>) {
              const toolCallId = tc.id || '';
              const toolName = tc.function?.name || 'tool';
              let args: unknown = undefined;
              try {
                args = tc.function?.arguments ? JSON.parse(tc.function.arguments) : undefined;
              } catch {
                args = tc.function?.arguments;
              }
              parts.push({
                type: 'tool-call' as const,
                toolName,
                toolCallId,
                args,
                status: 'running' as const
              });
            }
            return { ...m, parts };
          })
        };
        this.notify(threadId);
        return;
      }

      case 'tool_result': {
        const result = chunk.data as { tool_call_id?: string; content?: string } | undefined;
        const tcId = result?.tool_call_id;
        if (!tcId) return;

        s.snapshot = {
          ...s.snapshot,
          messages: s.snapshot.messages.map((m) => {
            if (m.id !== amId) return m;
            const parts = (m.parts || []).map((pt) =>
              pt.type === 'tool-call' && pt.toolCallId === tcId
                ? { ...pt, status: 'done' as const }
                : pt
            );
            return { ...m, parts };
          })
        };
        this.notify(threadId);
        return;
      }

      case 'done': {
        const payload = chunk.data as {
          content?: string;
          usage?: { prompt_tokens?: number; completion_tokens?: number; total_tokens?: number };
          thread_id?: string;
          steps?: number;
        } | null;

        const usage = payload?.usage;
        let tokenUsage: TokenUsage | undefined;
        if (usage) {
          tokenUsage = {
            promptTokens: usage.prompt_tokens ?? 0,
            completionTokens: usage.completion_tokens ?? 0,
            totalTokens: usage.total_tokens ?? 0
          };
        }

        const durationMs = s.firstTextDeltaTime ? Date.now() - s.firstTextDeltaTime : undefined;
        s.firstTextDeltaTime = null;

        s.snapshot = {
          ...s.snapshot,
          isStreaming: false,
          messages: s.snapshot.messages.map((m) => {
            if (m.id !== amId) return m;
            const parts = m.parts?.map((p) =>
              p.type === 'tool-call' && p.status === 'running'
                ? { ...p, status: 'done' as const }
                : p
            );
            return {
              ...m,
              parts,
              usage: tokenUsage ? { ...tokenUsage, durationMs } : m.usage
            };
          })
        };
        s.abortController = null;
        this.notify(threadId);
        return;
      }

      case 'cancelled': {
        const durationMs = s.firstTextDeltaTime ? Date.now() - s.firstTextDeltaTime : undefined;
        s.firstTextDeltaTime = null;

        s.snapshot = {
          ...s.snapshot,
          isStreaming: false,
          messages: s.snapshot.messages.map((m) => {
            if (m.id !== amId) return m;
            const parts = m.parts?.map((p) =>
              p.type === 'tool-call' && p.status === 'running'
                ? { ...p, status: 'done' as const }
                : p
            );
            const usage = m.usage
              ? { ...m.usage, durationMs }
              : durationMs != null
                ? { promptTokens: 0, completionTokens: 0, totalTokens: 0, durationMs }
                : undefined;
            return { ...m, parts, usage };
          })
        };
        s.abortController = null;
        this.notify(threadId);
        return;
      }

      case 'error': {
        const payload = chunk.data as { message?: string } | string | null;
        const errMsg =
          typeof payload === 'string'
            ? payload
            : ((payload as { message?: string })?.message ?? 'An unexpected error occurred');
        this.finishStream(threadId, errMsg, 'generic-error');
        return;
      }
    }
  }

  finishStream(
    threadId: string,
    errorMessage?: string,
    errorKind?: 'rate-limited' | 'generic-error'
  ) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    const amId = s.assistantMessageId;

    s.snapshot = { ...s.snapshot, isStreaming: false };
    s.abortController = null;

    if (amId) {
      const durationMs = s.firstTextDeltaTime ? Date.now() - s.firstTextDeltaTime : undefined;
      s.firstTextDeltaTime = null;

      s.snapshot = {
        ...s.snapshot,
        messages: s.snapshot.messages.map((m) => {
          if (m.id !== amId) return m;
          let { content } = m;
          if (errorMessage && !content) {
            content = `Error: ${errorMessage}`;
          }
          const parts = m.parts?.some((p) => p.type === 'tool-call' && p.status === 'running')
            ? m.parts.map((p) =>
                p.type === 'tool-call' && p.status === 'running'
                  ? { ...p, status: 'done' as const }
                  : p
              )
            : m.parts;
          const usage = m.usage
            ? { ...m.usage, durationMs }
            : durationMs != null
              ? { promptTokens: 0, completionTokens: 0, totalTokens: 0, durationMs }
              : m.usage;
          return { ...m, content, parts, usage, ...(errorKind ? { errorKind } : {}) };
        })
      };
    }

    this.notify(threadId);
  }

  stopStream(threadId: string) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    s.abortController?.abort();
    s.snapshot = { ...s.snapshot, isStreaming: false };
    this.notify(threadId);
  }

  appendErrorToLastAssistant(threadId: string, text: string) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    const amId = s.snapshot.messages.filter((m) => m.role === 'assistant').pop()?.id;
    if (!amId) return;
    s.snapshot = {
      ...s.snapshot,
      messages: s.snapshot.messages.map((m) =>
        m.id === amId ? { ...m, content: m.content + text } : m
      )
    };
    this.notify(threadId);
  }

  clearSession(threadId: string) {
    this.sessions.delete(threadId);
  }
}

export const chatStreamStore = new ChatStreamStore();
