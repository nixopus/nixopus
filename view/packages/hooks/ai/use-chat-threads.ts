'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import { createAgentClient, AGENT_ID } from '@/packages/lib/agent-client';
import { useAgentConfigured } from '@/packages/hooks/shared/use-config';

export interface ChatThread {
  id: string;
  title: string;
  createdAt: Date;
  updatedAt: Date;
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

export function useChatThreads() {
  const [threads, setThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadIdState] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const organizationId = activeOrg?.id;
  const authUser = useAppSelector((state) => state.auth.user);
  const resourceId = authUser?.id || 'default';

  const agentConfigured = useAgentConfigured() === true;
  const headersRef = useRef<Record<string, string>>({});

  useEffect(() => {
    (async () => {
      headersRef.current = await getAuthHeaders(token ?? null, organizationId ?? null);
    })();
  }, [token, organizationId]);

  useEffect(() => {
    if (!agentConfigured) {
      setIsInitialized(true);
      return;
    }

    let cancelled = false;

    (async () => {
      try {
        const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
        const client = createAgentClient(headers);
        const result = await client.listMemoryThreads({
          resourceId,
          agentId: AGENT_ID
        });

        if (cancelled) return;

        const threadList = Array.isArray(result) ? result : (result?.threads ?? []);
        const mapped: ChatThread[] = threadList.map(
          (t: {
            id: string;
            title?: string;
            createdAt: string | Date;
            updatedAt: string | Date;
          }) => ({
            id: t.id,
            title: t.title || 'New Chat',
            createdAt: new Date(t.createdAt),
            updatedAt: new Date(t.updatedAt)
          })
        );

        mapped.sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime());
        setThreads(mapped);

        const savedActiveId = loadActiveThreadId();
        if (savedActiveId && mapped.some((t) => t.id === savedActiveId)) {
          setActiveThreadIdState(savedActiveId);
        } else if (mapped.length > 0) {
          setActiveThreadIdState(mapped[0].id);
          saveActiveThreadId(mapped[0].id);
        }
      } catch {
        // agent may be unreachable
      } finally {
        if (!cancelled) setIsInitialized(true);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [token, organizationId, resourceId, agentConfigured]);

  const setActiveThreadId = useCallback((id: string | null) => {
    setActiveThreadIdState(id);
    saveActiveThreadId(id);
  }, []);

  const createThread = useCallback(
    (title?: string): ChatThread => {
      const threadId = crypto.randomUUID();
      const now = new Date();
      const thread: ChatThread = {
        id: threadId,
        title: title || 'New Chat',
        createdAt: now,
        updatedAt: now
      };

      setThreads((prev) => [thread, ...prev]);
      setActiveThreadId(thread.id);

      if (agentConfigured) {
        (async () => {
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
          }
        })();
      }

      return thread;
    },
    [resourceId, setActiveThreadId, agentConfigured]
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

      if (agentConfigured) {
        (async () => {
          try {
            const client = createAgentClient(headersRef.current);
            const thread = client.getMemoryThread({ threadId: id, agentId: AGENT_ID });
            await thread.delete();
          } catch {
            // ignore
          }
        })();
      }
    },
    [activeThreadId, setActiveThreadId, agentConfigured]
  );

  const updateThreadTitle = useCallback(
    (id: string, title: string) => {
      setThreads((prev) =>
        prev.map((t) => (t.id === id ? { ...t, title, updatedAt: new Date() } : t))
      );

      if (agentConfigured) {
        (async () => {
          try {
            const client = createAgentClient(headersRef.current);
            const thread = client.getMemoryThread({ threadId: id, agentId: AGENT_ID });
            await thread.update({ title, metadata: {}, resourceId });
          } catch {
            // ignore
          }
        })();
      }
    },
    [agentConfigured, resourceId]
  );

  const touchThread = useCallback((id: string) => {
    setThreads((prev) => prev.map((t) => (t.id === id ? { ...t, updatedAt: new Date() } : t)));
  }, []);

  const activeThread = threads.find((t) => t.id === activeThreadId) ?? null;

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
    touchThread
  };
}
