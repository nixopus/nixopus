'use client';

import { useState, useEffect, useCallback } from 'react';
import { getThreads, type Thread } from '@/redux/services/agents/agentsApi';

export function useThreads() {
  const [threads, setThreads] = useState<Thread[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchThreads = useCallback(async (signal?: AbortSignal) => {
    setIsLoading(true);
    setError(null);

    try {
      const fetchedThreads = await getThreads(signal);
      setThreads(fetchedThreads);
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        console.error('Error fetching threads:', err);
        setError('Failed to load threads');
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    const abortController = new AbortController();

    fetchThreads(abortController.signal);

    return () => {
      abortController.abort();
    };
  }, [fetchThreads]);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return {
    threads,
    isLoading,
    error,
    refetch: () => fetchThreads(),
    formatDate
  };
}
