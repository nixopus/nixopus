'use client';

import { useState, useCallback } from 'react';
import { useAppSelector } from '@/redux/hooks';
import { authClient } from '@/packages/lib/auth-client';
import { createAgentClient, AGENT_ID } from '@/packages/lib/agent-client';
import { useAgentConfigured } from '@/packages/hooks/shared/use-config';

export interface MemorySearchResult {
  id: string;
  role: string;
  content: string;
  createdAt: string;
  threadId?: string;
  threadTitle?: string;
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
      /* ignore */
    }
  }
  if (authToken) headers['Authorization'] = `Bearer ${authToken}`;
  if (organizationId) headers['X-Organization-Id'] = organizationId;
  return headers;
}

export function useMemorySearch(resourceId: string | undefined) {
  const [results, setResults] = useState<MemorySearchResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [query, setQuery] = useState('');
  const agentConfigured = useAgentConfigured() === true;

  const token = useAppSelector((state) => state.auth.token);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const organizationId = activeOrg?.id;

  const search = useCallback(
    async (searchQuery: string) => {
      const q = searchQuery.trim();
      if (!q || !resourceId || !agentConfigured) {
        setResults([]);
        setQuery('');
        return;
      }

      setIsSearching(true);
      setQuery(q);

      try {
        const headers = await getAuthHeaders(token ?? null, organizationId ?? null);
        const client = createAgentClient(headers);
        const response = await client.searchMemory({
          agentId: AGENT_ID,
          resourceId,
          searchQuery: q
        });

        const raw = (response as { results?: MemorySearchResult[] })?.results ?? [];
        setResults(raw);
      } catch {
        setResults([]);
      } finally {
        setIsSearching(false);
      }
    },
    [resourceId, token, organizationId, agentConfigured]
  );

  const clear = useCallback(() => {
    setResults([]);
    setQuery('');
  }, []);

  return { results, query, isSearching, search, clear };
}
