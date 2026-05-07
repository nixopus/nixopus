'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import {
  createThread as apiCreateThread,
  listThreads as apiListThreads,
  updateThread as apiUpdateThread,
  deleteThread as apiDeleteThread,
  type AgentThread
} from '@/packages/lib/agent-client';
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

const THREAD_CHANGE_EVENT = 'nixopus_active_thread_change';

function saveActiveThreadId(id: string | null) {
  if (typeof window === 'undefined') return;
  if (id) {
    localStorage.setItem(ACTIVE_THREAD_KEY, id);
  } else {
    localStorage.removeItem(ACTIVE_THREAD_KEY);
  }
  window.dispatchEvent(new CustomEvent(THREAD_CHANGE_EVENT, { detail: id }));
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

function mapApiThreads(raw: AgentThread[]): ChatThread[] {
  return raw.map((t) => ({
    id: t.id,
    title: t.title || 'New Chat',
    createdAt: new Date(t.created_at),
    updatedAt: new Date(t.updated_at)
  }));
}

export function useChatThreads() {
  const [threads, setThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadIdState] = useState<string | null>(loadActiveThreadId);
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
        const rawThreads = await apiListThreads(headers);

        if (cancelled) return;

        const all = mapApiThreads(rawThreads);
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

  useEffect(() => {
    const handler = (e: Event) => {
      const newId = (e as CustomEvent<string | null>).detail;
      setActiveThreadIdState((prev) => (prev === newId ? prev : newId));
    };
    window.addEventListener(THREAD_CHANGE_EVENT, handler);
    return () => window.removeEventListener(THREAD_CHANGE_EVENT, handler);
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
          await apiCreateThread(headersRef.current, { id: threadId, title: title || 'New Chat' });
        } catch {
          // will be created automatically on first message
        } finally {
          threadCreationPromises.current.delete(threadId);
        }
      })();
      threadCreationPromises.current.set(threadId, creationPromise);

      return thread;
    },
    [setActiveThreadId]
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
          await apiDeleteThread(id, headersRef.current);
        } catch {
          // ignore
        }
      })();
    },
    [activeThreadId, setActiveThreadId]
  );

  const updateThreadTitle = useCallback((id: string, title: string) => {
    setThreads((prev) =>
      prev.map((t) => (t.id === id ? { ...t, title, updatedAt: new Date() } : t))
    );

    (async () => {
      try {
        const pending = threadCreationPromises.current.get(id);
        if (pending) await pending;
        await apiUpdateThread(id, title, headersRef.current);
      } catch {
        // non-critical — title persists locally even if backend call fails
      }
    })();
  }, []);

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
