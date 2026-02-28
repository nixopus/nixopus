'use client';

import { useCallback, useRef, useState, useEffect, useMemo } from 'react';

import { useRouter } from 'next/navigation';
import { useAppSelector } from '@/redux/hooks';
import {
  saveWorkflow,
  streamWorkflowRun,
  runWorkflow,
  getRunStatus
} from '@/packages/lib/workflow-storage';
import { WORKFLOW_STREAMING_MESSAGES } from '@/packages/lib/workflow-utils';
import { authClient } from '@/packages/lib/auth-client';
import { createAgentClient } from '@/packages/lib/agent-client';
import type {
  WorkflowNode,
  WorkflowEdge,
  WorkflowCanvasHandle,
  NodeExecutionStatus,
  ExecutionMessage,
  MastraStreamEvent,
  WorkflowRunStatusResponse
} from '@/packages/types/workflow';
import type { ImperativePanelHandle } from 'react-resizable-panels';

export function usePanelToggle(panelRef: React.RefObject<ImperativePanelHandle | null>) {
  const toggle = useCallback(() => {
    if (panelRef.current?.isCollapsed()) {
      panelRef.current.expand();
    } else {
      panelRef.current?.collapse();
    }
  }, [panelRef]);

  const expand = useCallback(() => {
    panelRef.current?.expand();
  }, [panelRef]);

  return { toggle, expand };
}

export function useScrollToBottom(
  scrollRef: React.RefObject<HTMLDivElement | null>,
  deps: unknown[]
): void {
  useEffect(() => {
    const el = scrollRef.current;
    if (el) {
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight;
      });
    }
  }, [scrollRef, ...deps]);
}

export function useStreamingPlaceholder(isActive: boolean, intervalMs = 3800) {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (!isActive) return;
    const id = setInterval(() => {
      setIndex((i) => (i + 1) % WORKFLOW_STREAMING_MESSAGES.length);
    }, intervalMs);
    return () => clearInterval(id);
  }, [isActive, intervalMs]);

  return WORKFLOW_STREAMING_MESSAGES[index];
}

const PLANNER_AGENT_ID = 'workflow-planner';

export interface PlannerMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  graph?: WorkflowGraph | null;
}

export interface WorkflowGraph {
  name: string;
  description?: string;
  nodes: any[];
  edges: any[];
}

function resourceIdFor(organizationId: string, applicationId: string) {
  return `nixopus:${organizationId}:${applicationId}`;
}

function resolveThreadId(
  workflowId: string,
  applicationId: string,
  chatThreadId?: string | null,
  draftSessionId?: string | null
) {
  if (chatThreadId) return chatThreadId;
  if (workflowId === 'new' && draftSessionId) return `draft:${applicationId}:${draftSessionId}`;
  if (workflowId === 'new') return `draft:${applicationId}`;
  return workflowId;
}

function extractText(content: unknown): string {
  if (typeof content === 'string') return content;
  if (content && typeof content === 'object') {
    const obj = content as Record<string, unknown>;
    if (typeof obj.content === 'string') return obj.content;
    if (Array.isArray(obj.parts)) {
      return (obj.parts as Array<{ type?: string; text?: string }>)
        .filter((c) => c && typeof c === 'object')
        .map((c) => (c.type === 'text' && typeof c.text === 'string' ? c.text : ''))
        .join('');
    }
  }
  if (Array.isArray(content)) {
    return (content as Array<{ type?: string; text?: string }>)
      .filter((c) => c && typeof c === 'object' && 'type' in c)
      .map((c) => (c.type === 'text' && typeof c.text === 'string' ? c.text : ''))
      .join('');
  }
  return '';
}

function extractGraphFromText(text: string): WorkflowGraph | null {
  const jsonBlockRegex = /```json\s*([\s\S]*?)```/g;
  let match: RegExpExecArray | null;
  while ((match = jsonBlockRegex.exec(text)) !== null) {
    try {
      const parsed = JSON.parse(match[1]!.trim());
      if (
        parsed.nodes &&
        Array.isArray(parsed.nodes) &&
        parsed.edges &&
        Array.isArray(parsed.edges)
      ) {
        return parsed as WorkflowGraph;
      }
    } catch {
      continue;
    }
  }

  try {
    const parsed = JSON.parse(text);
    if (parsed.nodes && Array.isArray(parsed.nodes)) return parsed as WorkflowGraph;
  } catch {
    // not pure JSON
  }

  return null;
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
  if (authToken) headers['Authorization'] = `Bearer ${authToken}`;
  if (organizationId) headers['X-Organization-Id'] = organizationId;
  return headers;
}

interface UseWorkflowPlannerOptions {
  workflowId: string;
  applicationId: string;
  organizationId?: string | null;
  chatThreadId?: string | null;
  onGraphUpdate?: (graph: WorkflowGraph) => void;
  getCurrentGraph?: () => { name?: string; nodes: any[]; edges: any[] } | null;
  initialMessages?: {
    role: 'user' | 'assistant';
    content: string;
    graph?: WorkflowGraph | null;
    timestamp?: number;
  }[];
  /** For fallback timestamp inference when planning messages have no timestamp (e.g. legacy saved workflows) */
  initialExecutionMessages?: ExecutionMessage[];
}

const MS_PER_MINUTE = 60_000;

function fallbackTimestampsForPlanning(
  count: number,
  executionMessages?: ExecutionMessage[]
): number[] {
  if (count === 0) return [];
  const base =
    executionMessages && executionMessages.length > 0
      ? Math.min(...executionMessages.map((m) => m.timestamp)) - MS_PER_MINUTE * count
      : Date.now() - MS_PER_MINUTE * count;
  return Array.from({ length: count }, (_, i) => base + i * MS_PER_MINUTE);
}

export function useWorkflowPlanner({
  workflowId,
  applicationId,
  organizationId,
  chatThreadId,
  onGraphUpdate,
  getCurrentGraph,
  initialMessages = [],
  initialExecutionMessages = []
}: UseWorkflowPlannerOptions) {
  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const orgId = organizationId ?? activeOrg?.id ?? null;
  const [draftSessionId] = useState(() => (workflowId === 'new' ? crypto.randomUUID() : null));
  const resourceId = orgId && applicationId ? resourceIdFor(orgId, applicationId) : undefined;
  const threadId = applicationId
    ? resolveThreadId(workflowId, applicationId, chatThreadId, draftSessionId)
    : undefined;

  const [messages, setMessages] = useState<PlannerMessage[]>([]);
  const [isLoadingHistory, setIsLoadingHistory] = useState(!!threadId);
  const [inputValue, setInputValue] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const onGraphUpdateRef = useRef(onGraphUpdate);
  const getCurrentGraphRef = useRef(getCurrentGraph);
  const initialMessagesRef = useRef(initialMessages);
  const initialExecutionMessagesRef = useRef(initialExecutionMessages);
  const skipFetchForThreadRef = useRef<string | null>(null);
  const fetchingThreadIdRef = useRef<string | null>(null);
  onGraphUpdateRef.current = onGraphUpdate;
  getCurrentGraphRef.current = getCurrentGraph;
  initialMessagesRef.current = initialMessages;
  initialExecutionMessagesRef.current = initialExecutionMessages;

  if (skipFetchForThreadRef.current && skipFetchForThreadRef.current !== threadId) {
    skipFetchForThreadRef.current = null;
  }

  const buildFallbackWithTimestamps = useCallback(() => {
    const fallback = initialMessagesRef.current;
    const execMsgs = initialExecutionMessagesRef.current;
    const inferredTs = fallbackTimestampsForPlanning(fallback.length, execMsgs);
    return fallback.map((m, i) => {
      const ts =
        m.timestamp != null && !Number.isNaN(Number(m.timestamp))
          ? Number(m.timestamp)
          : (inferredTs[i] ?? Date.now());
      return {
        id: `saved-${i}`,
        role: m.role,
        content: m.content,
        timestamp: new Date(ts),
        graph: m.graph ?? undefined
      };
    });
  }, []);

  useEffect(() => {
    if (!threadId || !resourceId) {
      setMessages(buildFallbackWithTimestamps());
      setIsLoadingHistory(false);
      return;
    }

    if (skipFetchForThreadRef.current === threadId) {
      setMessages(buildFallbackWithTimestamps());
      setIsLoadingHistory(false);
      return;
    }

    if (fetchingThreadIdRef.current === threadId) {
      setMessages(buildFallbackWithTimestamps());
      setIsLoadingHistory(false);
      return;
    }

    // Show planning from API immediately so timeline isn't execution-only while thread fetches
    const fallback = buildFallbackWithTimestamps();
    if (fallback.length > 0) {
      setMessages(fallback);
    }

    fetchingThreadIdRef.current = threadId;
    let cancelled = false;
    setIsLoadingHistory(true);

    (async () => {
      try {
        const headers = await getAuthHeaders(token ?? null, orgId);
        const client = createAgentClient(headers);
        const thread = client.getMemoryThread({
          threadId,
          agentId: PLANNER_AGENT_ID
        });
        const listResourceId = resourceId || threadId;
        const result = await thread.listMessages({ resourceId: listResourceId });

        if (cancelled) return;

        const rawMessages = result?.messages ?? [];
        if (rawMessages.length > 0) {
          const msgs: PlannerMessage[] = [];
          for (const msg of rawMessages) {
            const role =
              msg.role === 'user' ? 'user' : msg.role === 'assistant' ? 'assistant' : null;
            const text = extractText(msg.content);
            if (!role || !text) continue;
            const graph = extractGraphFromText(text);
            const createdAtRaw =
              (msg as { createdAt?: string | Date; created_at?: string }).createdAt ??
              (msg as { created_at?: string }).created_at;
            const createdAt =
              createdAtRaw instanceof Date
                ? createdAtRaw
                : createdAtRaw
                  ? new Date(String(createdAtRaw))
                  : new Date();
            msgs.push({
              id: msg.id ?? crypto.randomUUID(),
              role,
              content: text,
              timestamp: createdAt,
              graph: graph ?? undefined
            });
          }
          if (msgs.length > 0) {
            setMessages(msgs);
          } else {
            setMessages(buildFallbackWithTimestamps());
          }
        } else {
          setMessages(buildFallbackWithTimestamps());
        }
      } catch {
        skipFetchForThreadRef.current = threadId;
        setMessages(buildFallbackWithTimestamps());
      } finally {
        if (fetchingThreadIdRef.current === threadId) {
          fetchingThreadIdRef.current = null;
        }
        if (!cancelled) setIsLoadingHistory(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [threadId, resourceId, orgId, token, buildFallbackWithTimestamps]);

  const scrollToBottom = useCallback(() => {
    const el = scrollRef.current;
    if (el) {
      requestAnimationFrame(() => {
        if (scrollRef.current) {
          scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
      });
    }
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  const sendMessage = useCallback(
    async (content: string) => {
      if (!content.trim() || isStreaming) return;

      const userMsg: PlannerMessage = {
        id: crypto.randomUUID(),
        role: 'user',
        content: content.trim(),
        timestamp: new Date()
      };

      const assistantId = crypto.randomUUID();
      setMessages((prev) => [
        ...prev,
        userMsg,
        { id: assistantId, role: 'assistant', content: '', timestamp: new Date() }
      ]);
      setIsStreaming(true);

      const abortController = new AbortController();
      abortRef.current = abortController;

      try {
        const headers = await getAuthHeaders(token ?? null, orgId);
        const client = createAgentClient(headers);
        const agent = client.getAgent(PLANNER_AGENT_ID);

        const streamOptions =
          threadId && resourceId
            ? {
                memory: { thread: threadId, resource: resourceId }
              }
            : undefined;
        const current = getCurrentGraphRef.current?.();
        const prompt = current?.nodes?.length
          ? `Current workflow in the canvas:

\`\`\`json
${JSON.stringify({
  name: current.name,
  nodes: current.nodes.map((n: any) => ({
    id: n.id,
    type: n.type,
    position: n.position,
    data: n.data
  })),
  edges: current.edges.map((e: any) => ({
    id: e.id,
    source: e.source,
    target: e.target
  }))
})}
\`\`\`

User request: ${content.trim()}`
          : content.trim();
        const response = await agent.stream(prompt, streamOptions);

        let fullText = '';

        await response.processDataStream({
          onChunk: async (chunk: unknown) => {
            if (abortController.signal.aborted) return;
            const c = chunk as { type?: string; payload?: Record<string, unknown> };

            if (c.type === 'text-delta') {
              const text = c.payload?.text as string | undefined;
              if (text) {
                fullText += text;
                setMessages((prev) =>
                  prev.map((m) => (m.id === assistantId ? { ...m, content: fullText } : m))
                );
              }
            }
          }
        });

        const graph = extractGraphFromText(fullText);
        if (graph) {
          setMessages((prev) => prev.map((m) => (m.id === assistantId ? { ...m, graph } : m)));
          onGraphUpdateRef.current?.(graph);
        }
        if (threadId && skipFetchForThreadRef.current === threadId) {
          skipFetchForThreadRef.current = null;
        }
      } catch (err: unknown) {
        if (err instanceof Error && err.name === 'AbortError') return;
        const msg = err instanceof Error ? err.message : 'Failed to get response';
        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantId ? { ...m, content: m.content || `Error: ${msg}` } : m
          )
        );
      } finally {
        setIsStreaming(false);
        abortRef.current = null;
      }
    },
    [isStreaming, token, orgId, threadId, resourceId]
  );

  const handleSubmit = useCallback(
    (e?: React.FormEvent) => {
      e?.preventDefault();
      if (!inputValue.trim()) return;
      const content = inputValue;
      setInputValue('');
      sendMessage(content);
    },
    [inputValue, sendMessage]
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

  const stopStreaming = useCallback(() => {
    abortRef.current?.abort();
    setIsStreaming(false);
  }, []);

  const resetChat = useCallback(() => {
    setMessages([]);
  }, []);

  return {
    messages,
    isLoadingHistory,
    threadId,
    inputValue,
    isStreaming,
    scrollRef,
    setInputValue,
    handleSubmit,
    handleKeyDown,
    sendMessage,
    stopStreaming,
    resetChat
  };
}

export type TimelineItem =
  | { kind: 'planner'; msg: PlannerMessage; ts: number; idx: number }
  | { kind: 'execution'; msg: ExecutionMessage; ts: number };

export function useChatTimeline(
  messages: PlannerMessage[],
  executionMessages: ExecutionMessage[]
): TimelineItem[] {
  return useMemo(() => {
    const items: TimelineItem[] = [];
    const plannerTsList: number[] = [];

    messages.forEach((msg, idx) => {
      const ts = msg.timestamp instanceof Date ? msg.timestamp.getTime() : Number(msg.timestamp);
      const resolvedTs = Number.isFinite(ts) ? ts : Date.now() - (messages.length - idx) * 60_000;
      plannerTsList.push(resolvedTs);
      items.push({ kind: 'planner' as const, msg, ts: resolvedTs, idx });
    });

    const executionAnchorTs = (() => {
      const graphIndices = messages
        .map((m, i) => (m.role === 'assistant' && m.graph ? i : -1))
        .filter((i) => i >= 0);
      const graphIdx = graphIndices.length > 0 ? graphIndices[graphIndices.length - 1]! : -1;
      if (graphIdx >= 0 && plannerTsList[graphIdx] != null) {
        return plannerTsList[graphIdx]! + 1000;
      }
      const lastAsstIdx = messages
        .map((m, i) => (m.role === 'assistant' ? i : -1))
        .filter((i) => i >= 0)
        .pop();
      if (lastAsstIdx != null && plannerTsList[lastAsstIdx] != null) {
        return plannerTsList[lastAsstIdx]! + 1000;
      }
      return plannerTsList.length > 0 ? Math.max(...plannerTsList) + 1000 : Date.now();
    })();

    executionMessages.forEach((msg, idx) => {
      const raw = msg.timestamp;
      const ts =
        typeof raw === 'number' && Number.isFinite(raw)
          ? raw
          : typeof raw === 'string'
            ? new Date(raw).getTime()
            : executionAnchorTs + idx;
      items.push({ kind: 'execution' as const, msg, ts });
    });

    items.sort((a, b) => {
      const dt = a.ts - b.ts;
      if (dt !== 0) return dt;
      return a.kind === 'planner' && b.kind === 'execution'
        ? -1
        : a.kind === 'execution' && b.kind === 'planner'
          ? 1
          : 0;
    });

    return items;
  }, [messages, executionMessages]);
}

type WorkflowRunStatus = import('@/packages/types/workflow').WorkflowRunStatus;

const POLL_INTERVAL_MS = 2500;
const TERMINAL_STATUSES: WorkflowRunStatus[] = ['success', 'failed', 'cancelled'];

let msgCounter = 0;
function nextMsgId(): string {
  return `exec-${Date.now()}-${++msgCounter}`;
}

function resolveStepId(payload: Record<string, unknown>): string | undefined {
  return (payload.stepId ?? payload.stepName ?? payload.id ?? payload.step) as string | undefined;
}

function errorToString(val: unknown): string {
  if (typeof val === 'string') return val;
  if (val && typeof val === 'object') {
    const obj = val as Record<string, unknown>;
    if (typeof obj.message === 'string') return obj.message;
    try {
      return JSON.stringify(val);
    } catch {
      /* fall through */
    }
  }
  return String(val ?? '');
}

function findLastIndex<T>(arr: T[], pred: (item: T) => boolean): number {
  for (let i = arr.length - 1; i >= 0; i--) {
    if (pred(arr[i])) return i;
  }
  return -1;
}

function mapStatus(raw: string | undefined): WorkflowRunStatus {
  switch (raw) {
    case 'success':
      return 'success';
    case 'failed':
      return 'failed';
    case 'suspended':
      return 'suspended';
    case 'cancelled':
      return 'cancelled';
    case 'running':
      return 'running';
    default:
      return 'running';
  }
}

interface UseWorkflowRunnerOptions {
  applicationId: string;
  workflowId: string;
  initialExecutionMessages?: ExecutionMessage[];
  onStepUpdate?: (
    stepId: string,
    status: NodeExecutionStatus,
    output?: Record<string, unknown>
  ) => void;
  onRunComplete?: (status: WorkflowRunStatus, value?: Record<string, unknown>) => void;
}

export function useWorkflowRunner({
  applicationId,
  workflowId,
  initialExecutionMessages,
  onStepUpdate,
  onRunComplete
}: UseWorkflowRunnerOptions) {
  const [runId, setRunId] = useState<string | null>(null);
  const [status, setStatus] = useState<WorkflowRunStatus>('idle');
  const [error, setError] = useState<string | null>(null);
  const [runOutput, setRunOutput] = useState<Record<string, unknown> | null>(null);
  const [executionMessages, setExecutionMessages] = useState<ExecutionMessage[]>(
    initialExecutionMessages ?? []
  );

  const abortRef = useRef<AbortController | null>(null);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const onStepUpdateRef = useRef(onStepUpdate);
  const onRunCompleteRef = useRef(onRunComplete);
  onStepUpdateRef.current = onStepUpdate;
  onRunCompleteRef.current = onRunComplete;

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  const getHeaders = useCallback(() => {
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;
    return headers;
  }, [token, activeOrg?.id]);

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
  }, []);

  const stopStream = useCallback(() => {
    if (abortRef.current) {
      abortRef.current.abort();
      abortRef.current = null;
    }
  }, []);

  useEffect(() => {
    return () => {
      stopPolling();
      stopStream();
    };
  }, [stopPolling, stopStream]);

  const isNewWorkflow = workflowId === 'new';

  const appendMessage = useCallback((msg: Omit<ExecutionMessage, 'id' | 'timestamp'>) => {
    const full: ExecutionMessage = { ...msg, id: nextMsgId(), timestamp: Date.now() };
    setExecutionMessages((prev) => {
      if (msg.stepId && (msg.kind === 'step-result' || msg.kind === 'step-error')) {
        const startIdx = findLastIndex(
          prev,
          (m) => m.kind === 'step-start' && m.stepId === msg.stepId
        );
        if (startIdx >= 0) {
          const updated = [...prev];
          updated[startIdx] = full;
          return updated;
        }
      }
      return [...prev, full];
    });
  }, []);

  const appendReasoningChunk = useCallback((stepId: string, chunk: string) => {
    setExecutionMessages((prev) => {
      const lastIdx = findLastIndex(
        prev,
        (m) => m.kind === 'step-reasoning' && m.stepId === stepId
      );
      if (lastIdx >= 0) {
        const updated = [...prev];
        updated[lastIdx] = {
          ...updated[lastIdx],
          text: (updated[lastIdx].text || '') + chunk
        };
        return updated;
      }
      return [
        ...prev,
        {
          id: nextMsgId(),
          kind: 'step-reasoning' as const,
          stepId,
          stepLabel: stepId,
          text: chunk,
          timestamp: Date.now()
        }
      ];
    });
  }, []);

  const handleMastraEvent = useCallback(
    (event: MastraStreamEvent) => {
      const eventType = event.type;
      const payload = event.payload || {};

      if (event.runId) setRunId(event.runId);

      if (eventType === 'deployment-reasoning-chunk') {
        const step = (payload.step as string) || 'agent';
        const chunk = payload.chunk as string;
        if (chunk) appendReasoningChunk(step, chunk);
        return;
      }

      if (eventType === 'deployment-progress') {
        const step = (payload.step as string) || 'agent';
        const message = (payload.message as string) || '';
        if (message) {
          appendMessage({ kind: 'step-progress', stepId: step, stepLabel: step, text: message });
        }
        return;
      }

      if (eventType === 'build-log') {
        return;
      }

      switch (eventType) {
        case 'run-start': {
          appendMessage({ kind: 'workflow-start' });
          break;
        }

        case 'workflow-start': {
          break;
        }

        case 'workflow-step-start': {
          const stepId = resolveStepId(payload);
          if (stepId) {
            onStepUpdateRef.current?.(stepId, 'running');
            appendMessage({
              kind: 'step-start',
              stepId,
              stepLabel: (payload.stepName as string) || stepId
            });
          }
          break;
        }

        case 'workflow-step-output':
        case 'workflow-step-progress': {
          const output = payload.output as Record<string, unknown> | undefined;
          if (output) {
            const nestedType = output.type as string;
            if (nestedType === 'deployment-reasoning-chunk') {
              const step = (output.step as string) || resolveStepId(payload) || 'agent';
              const chunk = output.chunk as string;
              if (chunk) appendReasoningChunk(step, chunk);
              return;
            }
            if (nestedType === 'deployment-progress') {
              const step = (output.step as string) || resolveStepId(payload) || 'agent';
              const message = (output.message as string) || '';
              if (message) {
                appendMessage({
                  kind: 'step-progress',
                  stepId: step,
                  stepLabel: step,
                  text: message
                });
              }
              return;
            }
          }

          const stepId = resolveStepId(payload);
          if (stepId && payload.message) {
            appendMessage({
              kind: 'step-progress',
              stepId,
              stepLabel: (payload.stepName as string) || stepId,
              text: payload.message as string
            });
          }
          break;
        }

        case 'workflow-step-result': {
          const stepId = resolveStepId(payload);
          const output = (payload.output ?? payload.result) as Record<string, unknown> | undefined;
          const isFailed = payload.status === 'failed';

          if (stepId) {
            const stepStatus: NodeExecutionStatus = isFailed ? 'failed' : 'success';
            onStepUpdateRef.current?.(stepId, stepStatus, output);

            if (isFailed) {
              appendMessage({
                kind: 'step-error',
                stepId,
                stepLabel: (payload.stepName as string) || stepId,
                status: 'failed',
                error: errorToString(payload.error) || 'Step failed'
              });
            } else {
              appendMessage({
                kind: 'step-result',
                stepId,
                stepLabel: (payload.stepName as string) || stepId,
                status: 'success',
                output
              });
            }
          }
          break;
        }

        case 'workflow-step-finish':
          break;

        case 'workflow-step-suspended': {
          const stepId = (payload.stepName as string) || (payload.id as string);
          if (stepId) {
            onStepUpdateRef.current?.(stepId, 'suspended');
            setStatus('suspended');
            appendMessage({
              kind: 'step-progress',
              stepId,
              stepLabel: stepId,
              status: 'suspended',
              text: 'Waiting for approval'
            });
          }
          break;
        }

        case 'workflow-finish':
        case 'finish': {
          const workflowStatus = (payload.workflowStatus as string) || 'success';
          const finalStatus: WorkflowRunStatus =
            workflowStatus === 'failed'
              ? 'failed'
              : workflowStatus === 'suspended'
                ? 'suspended'
                : 'success';
          setStatus(finalStatus);

          if (payload.output) {
            setRunOutput(payload.output as Record<string, unknown>);
          }

          if (finalStatus === 'suspended') {
            break;
          }

          appendMessage({
            kind: finalStatus === 'failed' ? 'workflow-error' : 'workflow-complete',
            status: finalStatus,
            output: (payload.output as Record<string, unknown>) || undefined,
            error: finalStatus === 'failed' ? 'Workflow failed' : undefined
          });
          onRunCompleteRef.current?.(
            finalStatus,
            (payload.output as Record<string, unknown>) || undefined
          );
          break;
        }

        case 'workflow-suspend': {
          setStatus('suspended');
          break;
        }

        case 'run-complete': {
          const finalStatus: WorkflowRunStatus =
            (payload.status as string) === 'failed' ? 'failed' : 'success';
          setStatus(finalStatus);
          if (payload.result || payload.steps) {
            setRunOutput((payload.result ?? payload.steps ?? payload) as Record<string, unknown>);
          }
          appendMessage({
            kind: finalStatus === 'failed' ? 'workflow-error' : 'workflow-complete',
            status: finalStatus,
            output: (payload.result ?? payload) as Record<string, unknown>
          });
          onRunCompleteRef.current?.(
            finalStatus,
            (payload.result ?? payload) as Record<string, unknown>
          );
          break;
        }

        case 'run-error': {
          const errMsg =
            errorToString(payload.error) || errorToString(payload.message) || 'Workflow failed';
          setStatus('failed');
          setError(errMsg);
          appendMessage({ kind: 'workflow-error', error: errMsg });
          onRunCompleteRef.current?.('failed');
          break;
        }

        default:
          break;
      }
    },
    [appendMessage, appendReasoningChunk]
  );

  const parseStream = useCallback(
    async (response: Response) => {
      const reader = response.body?.getReader();
      if (!reader) throw new Error('No readable stream');

      const decoder = new TextDecoder();
      let buffer = '';

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });

          const chunks = buffer.split(/\x1e|\n/);
          buffer = chunks.pop() || '';

          for (const chunk of chunks) {
            let raw = chunk.trim();
            if (!raw || raw === '[DONE]') continue;

            if (raw.startsWith('data:')) {
              raw = raw.slice(5).trim();
            }
            if (raw.startsWith('event:')) continue;
            if (!raw) continue;

            try {
              const parsed = JSON.parse(raw) as MastraStreamEvent;
              if (parsed.type) {
                handleMastraEvent(parsed);
              }
            } catch {
              // Non-JSON chunk, skip
            }
          }
        }

        if (buffer.trim()) {
          let raw = buffer.trim();
          if (raw.startsWith('data:')) raw = raw.slice(5).trim();
          if (raw && raw !== '[DONE]') {
            try {
              const parsed = JSON.parse(raw) as MastraStreamEvent;
              if (parsed.type) handleMastraEvent(parsed);
            } catch {
              // ignore
            }
          }
        }
      } finally {
        reader.releaseLock();
      }
    },
    [handleMastraEvent]
  );

  const pollStatus = useCallback(
    (currentRunId: string) => {
      stopPolling();

      pollingRef.current = setInterval(async () => {
        try {
          const result: WorkflowRunStatusResponse = await getRunStatus(workflowId, currentRunId, {
            applicationId,
            headers: getHeaders()
          });

          const mappedStatus = mapStatus(result.status);
          setStatus(mappedStatus);

          if (result.value) {
            setRunOutput(result.value);
            for (const [stepId, stepData] of Object.entries(result.value)) {
              if (!stepData || typeof stepData !== 'object') continue;
              const s = stepData as Record<string, unknown>;
              const stepStatus: NodeExecutionStatus =
                s.status === 'success'
                  ? 'success'
                  : s.status === 'failed'
                    ? 'failed'
                    : s.status === 'suspended'
                      ? 'suspended'
                      : s.status === 'skipped'
                        ? 'skipped'
                        : 'running';
              onStepUpdateRef.current?.(
                stepId,
                stepStatus,
                s.output as Record<string, unknown> | undefined
              );
            }
          }

          if (TERMINAL_STATUSES.includes(mappedStatus)) {
            stopPolling();
            onRunCompleteRef.current?.(mappedStatus, result.value ?? undefined);
          }
        } catch {
          // keep polling on transient errors
        }
      }, POLL_INTERVAL_MS);
    },
    [workflowId, applicationId, getHeaders, stopPolling]
  );

  const startRun = useCallback(
    async (inputData: Record<string, unknown> = {}) => {
      if (isNewWorkflow) return;

      try {
        setError(null);
        setStatus('running');
        setRunOutput(null);

        const headers = getHeaders();
        const controller = new AbortController();
        abortRef.current = controller;

        const response = await streamWorkflowRun(workflowId, {
          applicationId,
          inputData,
          headers,
          signal: controller.signal
        });

        await parseStream(response);

        setStatus((prev) => (TERMINAL_STATUSES.includes(prev) ? prev : 'success'));
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          setStatus('cancelled');
          appendMessage({ kind: 'workflow-error', error: 'Run cancelled' });
          return;
        }

        try {
          const headers = getHeaders();
          const result = await runWorkflow(workflowId, {
            applicationId,
            inputData,
            headers
          });
          setRunId(result.runId);
          appendMessage({ kind: 'workflow-start' });
          pollStatus(result.runId);
        } catch (fallbackErr) {
          setStatus('failed');
          const msg = fallbackErr instanceof Error ? fallbackErr.message : 'Failed to run workflow';
          setError(msg);
          appendMessage({ kind: 'workflow-error', error: msg });
        }
      }
    },
    [workflowId, applicationId, isNewWorkflow, getHeaders, parseStream, appendMessage, pollStatus]
  );

  const cancelRun = useCallback(() => {
    stopStream();
    stopPolling();
    setStatus('cancelled');
    appendMessage({ kind: 'workflow-error', error: 'Run cancelled' });
  }, [stopStream, stopPolling, appendMessage]);

  const reset = useCallback(() => {
    stopStream();
    stopPolling();
    setRunId(null);
    setStatus('idle');
    setError(null);
    setRunOutput(null);
    setExecutionMessages([]);
  }, [stopStream, stopPolling]);

  return {
    runId,
    status,
    error,
    runOutput,
    executionMessages,
    startRun,
    cancelRun,
    reset,
    isRunning: status === 'running'
  };
}

interface UseWorkflowEditorOptions {
  canvasRef: React.RefObject<WorkflowCanvasHandle | null>;
  applicationId: string;
  workflowId: string;
  workflowName: string;
  initialNodes?: WorkflowNode[];
  initialEdges?: WorkflowEdge[];
  initialPlanningMessages?: { role: 'user' | 'assistant'; content: string; graph?: any }[];
  initialExecutionMessages?: ExecutionMessage[];
  chatThreadId?: string | null;
  isDraft?: boolean;
}

export function useWorkflowEditor({
  canvasRef,
  applicationId,
  workflowId,
  workflowName: initialName,
  initialNodes = [],
  initialEdges = [],
  initialPlanningMessages = [],
  initialExecutionMessages,
  chatThreadId,
  isDraft = false
}: UseWorkflowEditorOptions) {
  const router = useRouter();
  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  const [savedId, setSavedId] = useState<string | null>(isDraft ? null : workflowId);
  const [currentName, setCurrentName] = useState(initialName);
  const [isSaving, setIsSaving] = useState(false);
  const [selectedNode, setSelectedNode] = useState<{
    id: string;
    type: string;
    data: Record<string, unknown>;
  } | null>(null);

  const handleGraphUpdate = useCallback(
    (graph: WorkflowGraph) => {
      if (!canvasRef.current) return;

      const nodes = (graph.nodes || []).map((n: any) => ({
        id: n.id,
        type: n.type || 'tool',
        position: n.position || { x: 300, y: 0 },
        data: {
          executionStatus: 'idle',
          ...n.data
        }
      }));

      const edges = (graph.edges || []).map((e: any) => ({
        id: e.id,
        source: e.source,
        target: e.target
      }));

      canvasRef.current.replaceGraph(nodes, edges);
      if (graph.name) setCurrentName(graph.name);
    },
    [canvasRef]
  );

  const getCurrentGraph = useCallback(() => {
    if (!canvasRef.current) return null;
    const nodes = canvasRef.current.getNodes();
    const edges = canvasRef.current.getEdges();
    if (!nodes?.length) return null;
    return {
      name: currentName,
      nodes: nodes.map((n: any) => ({
        id: n.id,
        type: n.type,
        position: n.position ?? { x: 0, y: 0 },
        data: n.data
      })),
      edges: edges.map((e: any) => ({
        id: e.id,
        source: e.source,
        target: e.target
      }))
    };
  }, [canvasRef, currentName]);

  const planner = useWorkflowPlanner({
    workflowId,
    applicationId,
    organizationId: activeOrg?.id,
    chatThreadId,
    onGraphUpdate: handleGraphUpdate,
    getCurrentGraph,
    initialMessages: initialPlanningMessages,
    initialExecutionMessages
  });

  const handleStepUpdate = useCallback(
    (stepId: string, status: NodeExecutionStatus, output?: Record<string, unknown>) => {
      canvasRef.current?.setNodeData(stepId, { executionStatus: status, executionOutput: output });
    },
    [canvasRef]
  );

  const saveRef = useRef<() => void>(() => {});

  const runner = useWorkflowRunner({
    applicationId,
    workflowId: savedId || workflowId,
    initialExecutionMessages,
    onStepUpdate: handleStepUpdate,
    onRunComplete: () => {
      setTimeout(() => saveRef.current(), 0);
    }
  });

  // Auto-save planning messages when they change (debounced), so "hello" etc. persist on refresh
  useEffect(() => {
    if (planner.isLoadingHistory || planner.isStreaming) return;
    if (planner.messages.length === 0) return;
    const t = setTimeout(() => saveRef.current(), 2000);
    return () => clearTimeout(t);
  }, [planner.messages, planner.isLoadingHistory, planner.isStreaming]);

  const handleSave = useCallback(async () => {
    const nodes = canvasRef.current?.getNodes() ?? [];
    const edges = canvasRef.current?.getEdges() ?? [];

    if (nodes.length === 0) return;

    setIsSaving(true);
    try {
      const headers: Record<string, string> = {};
      if (token) headers['Authorization'] = `Bearer ${token}`;
      if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;

      const planningMessages = planner.messages.map((m) => ({
        role: m.role,
        content: m.content,
        graph: m.graph ?? undefined,
        timestamp:
          m.timestamp instanceof Date ? m.timestamp.getTime() : Number(m.timestamp) || undefined
      }));

      const result = await saveWorkflow({
        id: savedId || workflowId,
        name: currentName,
        applicationId,
        chatThreadId: isDraft ? planner.threadId : (chatThreadId ?? undefined),
        planningMessages: planningMessages.length > 0 ? planningMessages : undefined,
        executionMessages:
          runner.executionMessages.length > 0 ? runner.executionMessages : undefined,
        nodes: nodes.map((n: any) => ({
          id: n.id,
          type: n.type,
          position: n.position ?? { x: 0, y: 0 },
          data: n.data
        })),
        edges: edges.map((e: any) => ({
          id: e.id,
          source: e.source,
          target: e.target
        })),
        headers
      });

      setSavedId(result.id);

      if (isDraft && result.id) {
        router.replace(`/apps/application/${applicationId}/workflows/${result.id}`);
      }
    } catch (err) {
      console.error('Failed to save workflow:', err);
    } finally {
      setIsSaving(false);
    }
  }, [
    applicationId,
    savedId,
    workflowId,
    currentName,
    isDraft,
    token,
    activeOrg?.id,
    router,
    planner.messages,
    planner.threadId,
    chatThreadId,
    canvasRef,
    runner.executionMessages
  ]);

  saveRef.current = handleSave;

  const handleRun = useCallback(() => runner.startRun({}), [runner]);

  const handleReset = useCallback(() => {
    runner.reset();
    canvasRef.current?.resetAllNodes();
  }, [runner, canvasRef]);

  const handleNodeClick = useCallback(
    (node: { id: string; type: string; data: Record<string, unknown> } | null) => {
      setSelectedNode(node);
    },
    []
  );

  const authHeaders = useCallback(() => {
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    if (activeOrg?.id) headers['X-Organization-Id'] = activeOrg.id;
    return headers;
  }, [token, activeOrg?.id]);

  return {
    planner,
    runner,
    savedId,
    currentName,
    isSaving,
    isStillDraft: isDraft && !savedId,
    selectedNode,
    setSelectedNode,
    handleSave,
    handleRun,
    handleReset,
    handleNodeClick,
    authHeaders
  };
}
