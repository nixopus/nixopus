'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import { createAgentClient, AGENT_ID, INCIDENT_AGENT_ID } from '@/packages/lib/agent-client';
import { v4 as uuid } from 'uuid';

export interface ChatThread {
  id: string;
  title: string;
  createdAt: Date;
  updatedAt: Date;
  isIncident?: boolean;
  agentId?: string;
  threadResourceId?: string;
}

const ACTIVE_THREAD_KEY = 'nixopus_active_thread';

function loadActiveThreadId(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(ACTIVE_THREAD_KEY);
}

function saveActiveThreadId(id: string | null) {
  if (typeof window === 'undefined') return;
  if (id) {
    localStorage.setItem(ACTIVE_THREAD_KEY, id);
  } else {
    localStorage.removeItem(ACTIVE_THREAD_KEY);
  }
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

function mapThreads(
  raw: unknown,
  incident: boolean,
  agent: string,
  resource: string
): ChatThread[] {
  const list = Array.isArray(raw)
    ? raw
    : (((raw as Record<string, unknown>)?.threads ?? []) as Array<{
        id: string;
        title?: string;
        createdAt: string | Date;
        updatedAt: string | Date;
      }>);
  return list.map((t) => {
    let title = t.title || 'New Chat';
    if (incident && !t.title) {
      const parts = t.id.replace(/^incident-/, '').split('-');
      const source = parts[0] || 'unknown';
      const eventId = parts.slice(1).join('-') || t.id;
      title = `${source}: ${eventId}`;
    }
    return {
      id: t.id,
      title,
      createdAt: new Date(t.createdAt),
      updatedAt: new Date(t.updatedAt),
      agentId: agent,
      threadResourceId: resource,
      ...(incident ? { isIncident: true } : {})
    };
  });
}

export function useChatThreads() {
  const [threads, setThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadIdState] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const organizationId = activeOrg?.id;
  const authUser = useAppSelector((state) => state.auth.user);
  const resourceId = authUser?.id || 'default';

  const headersRef = useRef<Record<string, string>>({});
  const threadCreationPromises = useRef<Map<string, Promise<void>>>(new Map());

  useEffect(() => {
    (async () => {
      headersRef.current = await getAuthHeaders(token ?? null, organizationId ?? null);
    })();
  }, [token, organizationId]);

  const refreshThreads = useCallback(() => {
    setIsRefreshing(true);
    setRefreshKey((k) => k + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
        const client = createAgentClient(headers);

        const [chatResult, incidentResult] = await Promise.all([
          client.listMemoryThreads({ resourceId, agentId: AGENT_ID }),
          organizationId
            ? client
                .listMemoryThreads({ resourceId: organizationId, agentId: INCIDENT_AGENT_ID })
                .catch(() => [])
            : Promise.resolve([])
        ]);

        if (cancelled) return;

        const all = [
          ...mapThreads(chatResult, false, AGENT_ID, resourceId),
          ...mapThreads(incidentResult, true, INCIDENT_AGENT_ID, organizationId!)
        ];
        all.sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime());
        setThreads(all);
      } catch {
        // agent may be unreachable
      } finally {
        if (!cancelled) {
          setIsInitialized(true);
          setIsRefreshing(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [token, organizationId, resourceId, refreshKey]);

  const setActiveThreadId = useCallback((id: string | null) => {
    setActiveThreadIdState(id);
    saveActiveThreadId(id);
  }, []);

  const createThread = useCallback(
    (title?: string): ChatThread => {
      const threadId = uuid();
      const now = new Date();
      const thread: ChatThread = {
        id: threadId,
        title: title || 'New Chat',
        createdAt: now,
        updatedAt: now
      };

      setThreads((prev) => [thread, ...prev]);
      setActiveThreadId(thread.id);

      const creationPromise = (async () => {
        try {
          const client = createAgentClient(headersRef.current);
          await client.createMemoryThread({
            threadId,
            title: thread.title,
            resourceId,
            agentId: AGENT_ID
          });
        } catch {
          // will be created automatically on first message
        } finally {
          threadCreationPromises.current.delete(threadId);
        }
      })();
      threadCreationPromises.current.set(threadId, creationPromise);

      return thread;
    },
    [resourceId, setActiveThreadId]
  );

  const deleteThread = useCallback(
    (id: string) => {
      setThreads((prev) => prev.filter((t) => t.id !== id));

      if (activeThreadId === id) {
        setThreads((prev) => {
          const nextId = prev.length > 0 ? prev[0].id : null;
          setActiveThreadId(nextId);
          return prev;
        });
      }

      (async () => {
        try {
          const client = createAgentClient(headersRef.current);
          const thread = client.getMemoryThread({ threadId: id, agentId: AGENT_ID });
          await thread.delete();
        } catch {
          // ignore
        }
      })();
    },
    [activeThreadId, setActiveThreadId]
  );

  const updateThreadTitle = useCallback(
    (id: string, title: string) => {
      setThreads((prev) =>
        prev.map((t) => (t.id === id ? { ...t, title, updatedAt: new Date() } : t))
      );

      (async () => {
        try {
          const pending = threadCreationPromises.current.get(id);
          if (pending) await pending;
          const client = createAgentClient(headersRef.current);
          const thread = client.getMemoryThread({ threadId: id, agentId: AGENT_ID });
          await thread.update({ title, metadata: {}, resourceId });
        } catch {
          // ignore
        }
      })();
    },
    [resourceId]
  );

  const touchThread = useCallback((id: string) => {
    setThreads((prev) => prev.map((t) => (t.id === id ? { ...t, updatedAt: new Date() } : t)));
  }, []);

  const activeThread = threads.find((t) => t.id === activeThreadId) ?? null;

  const waitForThread = useCallback(async (id: string) => {
    const pending = threadCreationPromises.current.get(id);
    if (pending) await pending;
  }, []);

  return {
    threads,
    activeThread,
    activeThreadId,
    resourceId,
    isInitialized,
    setActiveThreadId,
    createThread,
    deleteThread,
    updateThreadTitle,
    touchThread,
    waitForThread,
    refreshThreads,
    isRefreshing
  };
}
