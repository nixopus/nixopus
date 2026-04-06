import type { StreamChunk } from '@/packages/lib/agent-client';
import type { ChatMessage, PendingToolApproval, OmStatus, TokenUsage } from './use-agent-chat';

export interface SessionSnapshot {
  messages: ChatMessage[];
  isStreaming: boolean;
  pendingToolApproval: PendingToolApproval | null;
  omStatus: OmStatus | null;
}

interface SessionInternal {
  snapshot: SessionSnapshot;
  abortController: AbortController | null;
  needsStepSeparator: boolean;
  needsNewTextPart: boolean;
  firstTextDeltaTime: number | null;
  pendingApproval: boolean;
  omStatusCache: OmStatus | null;
  assistantMessageId: string | null;
  runId: string;
}

type Listener = () => void;

const EMPTY_SNAPSHOT: SessionSnapshot = Object.freeze({
  messages: [],
  isStreaming: false,
  pendingToolApproval: null,
  omStatus: null
});

function extractUsageFromPayload(payload: unknown): TokenUsage | null {
  if (!payload || typeof payload !== 'object') return null;
  const p = payload as Record<string, unknown>;

  const output = p.output as Record<string, unknown> | undefined;
  const outputUsage = output?.usage as Record<string, unknown> | undefined;

  const metadata = p.metadata as Record<string, unknown> | undefined;
  const providerMeta = metadata?.providerMetadata as Record<string, unknown> | undefined;
  const orMeta = providerMeta?.openrouter as Record<string, unknown> | undefined;
  const orUsage = orMeta?.usage as Record<string, unknown> | undefined;

  const usage = outputUsage ?? orUsage;
  if (!usage) return null;

  const prompt = (usage.promptTokens as number) ?? (usage.inputTokens as number) ?? undefined;
  const completion =
    (usage.completionTokens as number) ?? (usage.outputTokens as number) ?? undefined;

  if (typeof prompt !== 'number' && typeof completion !== 'number') return null;

  const promptTokens = prompt ?? 0;
  const completionTokens = completion ?? 0;
  const costUsd = typeof orUsage?.cost === 'number' ? (orUsage.cost as number) : undefined;

  return {
    promptTokens,
    completionTokens,
    totalTokens:
      typeof (usage.totalTokens as number) === 'number'
        ? (usage.totalTokens as number)
        : promptTokens + completionTokens,
    costUsd
  };
}

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
        needsStepSeparator: false,
        needsNewTextPart: false,
        firstTextDeltaTime: null,
        pendingApproval: false,
        omStatusCache: null,
        assistantMessageId: null,
        runId: ''
      };
      this.sessions.set(threadId, s);
    }
    return s;
  }

  private accumulateUsage(s: SessionInternal, messageId: string, usage: TokenUsage) {
    s.snapshot = {
      ...s.snapshot,
      messages: s.snapshot.messages.map((m) => {
        if (m.id !== messageId) return m;
        const existing = m.usage;
        if (existing) {
          const costUsd =
            existing.costUsd != null || usage.costUsd != null
              ? (existing.costUsd ?? 0) + (usage.costUsd ?? 0)
              : undefined;
          return {
            ...m,
            usage: {
              promptTokens: existing.promptTokens + usage.promptTokens,
              completionTokens: existing.completionTokens + usage.completionTokens,
              totalTokens: existing.totalTokens + usage.totalTokens,
              costUsd
            }
          };
        }
        return { ...m, usage };
      })
    };
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
    s.runId = '';
    s.needsStepSeparator = false;
    s.needsNewTextPart = false;
    s.firstTextDeltaTime = null;
    s.pendingApproval = false;
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

    if (chunk.type === 'start' && chunk.runId) {
      s.runId = chunk.runId;
      return;
    }

    if (chunk.type === 'text-delta') {
      if (s.firstTextDeltaTime === null) {
        s.firstTextDeltaTime = Date.now();
      }
      const text = (chunk.payload as Record<string, unknown> | undefined)?.text as
        | string
        | undefined;
      if (!text) return;

      const insertSep = s.needsStepSeparator;
      const startNewPart = s.needsNewTextPart;
      s.needsStepSeparator = false;
      s.needsNewTextPart = false;

      s.snapshot = {
        ...s.snapshot,
        messages: s.snapshot.messages.map((m) => {
          if (m.id !== amId) return m;
          const sep = insertSep && m.content.length > 0 ? '\n\n' : '';
          const newContent = m.content + sep + text;

          let parts = [...(m.parts || [])];
          parts = parts.map((p) =>
            p.type === 'tool-call' && p.status === 'running' ? { ...p, status: 'done' as const } : p
          );
          const lastPart = parts[parts.length - 1];
          if (!startNewPart && lastPart?.type === 'text') {
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

    if (chunk.type === 'step-finish') {
      s.needsStepSeparator = true;
      s.needsNewTextPart = true;
      const stepUsage = extractUsageFromPayload(chunk.payload);
      if (stepUsage) {
        this.accumulateUsage(s, amId, stepUsage);
        this.notify(threadId);
      }
      return;
    }

    if (chunk.type === 'text-end') {
      s.needsStepSeparator = true;
      s.needsNewTextPart = true;
      return;
    }

    if (
      chunk.type === 'tool-call' ||
      chunk.type === 'tool-call-start' ||
      chunk.type === 'tool-call-approval'
    ) {
      const p = chunk.payload as Record<string, unknown> | undefined;
      const toolCallId = (p?.toolCallId ?? p?.id) as string | undefined;
      const toolName = (p?.toolName as string) ?? 'tool';

      if (toolName === 'ask_user' || toolName === 'askUser') return;

      if (toolCallId) {
        s.needsNewTextPart = true;
        const tcId = String(toolCallId);
        s.snapshot = {
          ...s.snapshot,
          messages: s.snapshot.messages.map((m) => {
            if (m.id !== amId) return m;
            const parts = [...(m.parts || [])];
            if (!parts.some((pt) => pt.type === 'tool-call' && pt.toolCallId === tcId)) {
              parts.push({
                type: 'tool-call' as const,
                toolName,
                toolCallId: tcId,
                args: p?.args,
                status: 'running' as const
              });
            }
            return { ...m, parts };
          })
        };

        if (chunk.type === 'tool-call-approval') {
          s.pendingApproval = true;
          s.snapshot = {
            ...s.snapshot,
            pendingToolApproval: {
              runId: s.runId,
              toolCallId: tcId,
              toolName,
              args: p?.args ?? {}
            }
          };
        }

        this.notify(threadId);
      }
      return;
    }

    if (chunk.type === 'tool-result') {
      const p = chunk.payload as Record<string, unknown> | undefined;
      const tcId = p?.toolCallId as string | undefined;
      if (tcId) {
        s.needsNewTextPart = true;
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
      }
      return;
    }

    if (chunk.type === 'finish' && chunk.payload) {
      const finishPayload = chunk.payload as Record<string, unknown>;

      const msg = s.snapshot.messages.find((m) => m.id === amId);
      if (!msg?.usage) {
        const finishUsage = extractUsageFromPayload(finishPayload);
        if (finishUsage) {
          s.snapshot = {
            ...s.snapshot,
            messages: s.snapshot.messages.map((m) =>
              m.id === amId ? { ...m, usage: finishUsage } : m
            )
          };
        }
      }

      const finishReason = finishPayload.finishReason as string | undefined;
      const sp = finishPayload.suspendPayload as Record<string, unknown> | undefined;
      if (finishReason === 'suspended' && sp) {
        s.pendingApproval = true;
        s.snapshot = {
          ...s.snapshot,
          pendingToolApproval: {
            runId: (sp.runId as string) ?? '',
            toolCallId: (sp.toolCallId as string) ?? '',
            toolName: (sp.toolName as string) ?? 'tool',
            args: sp.args ?? {}
          }
        };
      }
      this.notify(threadId);
      return;
    }

    if (chunk.type === 'data-om-status') {
      const d = chunk.payload as Record<string, unknown> | undefined;
      const windows = d?.windows as Record<string, unknown> | undefined;
      const active = windows?.active as Record<string, unknown> | undefined;
      if (active) {
        const msgs = active.messages as Record<string, unknown> | undefined;
        const obs = active.observations as Record<string, unknown> | undefined;
        const next: OmStatus = {
          messages: {
            tokens: (msgs?.tokens as number) ?? 0,
            threshold: (msgs?.threshold as number) ?? 30000
          },
          observations: {
            tokens: (obs?.tokens as number) ?? 0,
            threshold: (obs?.threshold as number) ?? 40000
          },
          isObserving: s.omStatusCache?.isObserving ?? false,
          observationsText: s.omStatusCache?.observationsText ?? null
        };
        s.omStatusCache = next;
        s.snapshot = { ...s.snapshot, omStatus: next };
        this.notify(threadId);
      }
      return;
    }

    if (chunk.type === 'data-om-observation-start') {
      if (s.omStatusCache) {
        const next = { ...s.omStatusCache, isObserving: true };
        s.omStatusCache = next;
        s.snapshot = { ...s.snapshot, omStatus: next };
        this.notify(threadId);
      }
      return;
    }

    if (chunk.type === 'data-om-observation-end' || chunk.type === 'data-om-activation') {
      const d = chunk.payload as Record<string, unknown> | undefined;
      if (s.omStatusCache) {
        const next: OmStatus = {
          ...s.omStatusCache,
          isObserving: false,
          observationsText: (d?.observations as string) ?? s.omStatusCache.observationsText
        };
        s.omStatusCache = next;
        s.snapshot = { ...s.snapshot, omStatus: next };
        this.notify(threadId);
      }
      return;
    }
  }

  finishStream(threadId: string, errorMessage?: string) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    const amId = s.assistantMessageId;

    if (!s.pendingApproval) {
      s.snapshot = { ...s.snapshot, isStreaming: false };
    }
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
          return { ...m, content, parts, usage };
        })
      };
    }

    s.needsStepSeparator = false;
    s.needsNewTextPart = false;
    this.notify(threadId);
  }

  stopStream(threadId: string) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    s.abortController?.abort();
    s.snapshot = { ...s.snapshot, isStreaming: false };
    this.notify(threadId);
  }

  prepareApprovalStream(threadId: string): string | null {
    const s = this.sessions.get(threadId);
    if (!s) return null;
    s.pendingApproval = false;
    s.snapshot = { ...s.snapshot, pendingToolApproval: null };
    const amId = s.snapshot.messages.filter((m) => m.role === 'assistant').pop()?.id ?? null;
    s.assistantMessageId = amId;
    this.notify(threadId);
    return amId;
  }

  startApprovalStream(threadId: string, abortController: AbortController) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    s.abortController = abortController;
    s.needsNewTextPart = true;
  }

  finishApprovalStream(threadId: string) {
    const s = this.sessions.get(threadId);
    if (!s) return;

    if (!s.pendingApproval) {
      s.snapshot = { ...s.snapshot, isStreaming: false };
    }
    s.abortController = null;
    s.needsStepSeparator = false;
    s.needsNewTextPart = false;

    const amId = s.snapshot.messages.filter((m) => m.role === 'assistant').pop()?.id;
    if (amId) {
      s.snapshot = {
        ...s.snapshot,
        messages: s.snapshot.messages.map((m) => {
          if (m.id !== amId || !m.parts) return m;
          if (!m.parts.some((p) => p.type === 'tool-call' && p.status === 'running')) return m;
          return {
            ...m,
            parts: m.parts.map((p) =>
              p.type === 'tool-call' && p.status === 'running'
                ? { ...p, status: 'done' as const }
                : p
            )
          };
        })
      };
    }
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

  declineApproval(threadId: string) {
    const s = this.sessions.get(threadId);
    if (!s) return;
    s.pendingApproval = false;
    s.snapshot = { ...s.snapshot, pendingToolApproval: null, isStreaming: false };
    this.notify(threadId);
  }

  clearSession(threadId: string) {
    this.sessions.delete(threadId);
  }
}

export const chatStreamStore = new ChatStreamStore();
