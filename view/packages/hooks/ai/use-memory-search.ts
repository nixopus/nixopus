'use client';

import { useState, useCallback } from 'react';

export interface MemorySearchResult {
  id: string;
  role: string;
  content: string;
  createdAt: string;
  threadId?: string;
  threadTitle?: string;
}

/**
 * Memory search hook - currently a stub.
 * The Go agent does not expose a search endpoint yet.
 * This preserves the interface so existing UI components continue to compile.
 */
export function useMemorySearch(_resourceId: string | undefined) {
  const [results, setResults] = useState<MemorySearchResult[]>([]);
  const [isSearching] = useState(false);
  const [query, setQuery] = useState('');

  const search = useCallback(async (searchQuery: string) => {
    const q = searchQuery.trim();
    if (!q) {
      setResults([]);
      setQuery('');
      return;
    }
    setQuery(q);
    // TODO: implement when Go API exposes /api/v1/agent/threads/search
    setResults([]);
  }, []);

  const clear = useCallback(() => {
    setResults([]);
    setQuery('');
  }, []);

  return { results, query, isSearching, search, clear };
}
