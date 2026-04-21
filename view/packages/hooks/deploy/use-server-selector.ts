'use client';

import { useState, useCallback, useEffect } from 'react';
import { useGetServersQuery } from '@/redux/services/servers/serversApi';
import { Server } from '@/redux/types/servers';
import { RoutingStrategy } from '@/redux/types/applications';

export function useServerSelector() {
  const { data, isLoading } = useGetServersQuery({ page: 1, page_size: 100 });
  const servers = (data?.servers ?? []).filter((s: Server) => s.is_active);

  const [selectedServerIds, setSelectedServerIds] = useState<string[]>([]);
  const [routingStrategy, setRoutingStrategy] = useState<RoutingStrategy | null>(null);
  const [primaryServerId, setPrimaryServerId] = useState<string | null>(null);

  const isMultiServer = servers.length >= 2;

  useEffect(() => {
    if (selectedServerIds.length < 2) {
      setRoutingStrategy(null);
      setPrimaryServerId(null);
    }
  }, [selectedServerIds]);

  useEffect(() => {
    if (routingStrategy !== 'primary_failover') {
      setPrimaryServerId(null);
    }
  }, [routingStrategy]);

  const toggleServer = useCallback((id: string) => {
    setSelectedServerIds((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]
    );
  }, []);

  const selectAll = useCallback(() => {
    setSelectedServerIds(servers.map((s) => s.id));
  }, [servers]);

  const clearAll = useCallback(() => {
    setSelectedServerIds([]);
  }, []);

  return {
    servers,
    isLoading,
    isMultiServer,
    selectedServerIds,
    setSelectedServerIds,
    toggleServer,
    selectAll,
    clearAll,
    routingStrategy,
    setRoutingStrategy,
    primaryServerId,
    setPrimaryServerId
  };
}
