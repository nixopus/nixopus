'use client';

import { useState, useMemo } from 'react';
import { useGetServersQuery } from '@/redux/services/servers/serversApi';
import type { Server } from '@/redux/types/servers';

export type Machine = {
  id: string;
  ssh_key_id: string;
  name: string;
  domain: string;
  status?: 'provisioning' | 'active' | 'failed';
  is_active?: boolean;
  total_vcpu: number;
  total_ram_mb: number;
  total_disk_gb: number;
};

const PAGE_SIZE = 6;

function mapServerToMachine(server: Server): Machine {
  const { provision } = server;

  let status: 'provisioning' | 'active' | 'failed' | undefined;
  if (provision) {
    if (provision.status === 'ACTIVE') {
      status = 'active';
    } else if (provision.status === 'FAILED') {
      status = 'failed';
    } else if (provision.status === 'PROVISIONING' || provision.status === 'NOT_STARTED') {
      status = 'provisioning';
    }
  }

  const displayName =
    server.name?.trim() || provision?.domain?.trim() || server.host?.trim() || server.id;

  return {
    id: provision?.id || server.id,
    ssh_key_id: server.id,
    name: displayName,
    domain: provision?.domain?.trim() || server.host?.trim() || server.id,
    status,
    is_active: server.is_active,
    total_vcpu: server.total_vcpu,
    total_ram_mb: server.total_ram_mb,
    total_disk_gb: server.total_disk_gb
  };
}

export function useMachines() {
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);

  const {
    data: serversData,
    isLoading,
    isFetching,
    error,
    refetch
  } = useGetServersQuery({
    sort_by: 'created_at',
    sort_order: 'desc'
  });

  const machines = useMemo(() => {
    if (!serversData?.servers) return [];
    return serversData.servers.map(mapServerToMachine);
  }, [serversData]);

  const filteredMachines = useMemo(
    () =>
      machines.filter(
        (machine) =>
          machine.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          machine.domain.toLowerCase().includes(searchTerm.toLowerCase())
      ),
    [machines, searchTerm]
  );

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(filteredMachines.length / PAGE_SIZE)),
    [filteredMachines.length]
  );

  const paginatedMachines = useMemo(() => {
    const startIndex = (currentPage - 1) * PAGE_SIZE;
    const endIndex = startIndex + PAGE_SIZE;
    return filteredMachines.slice(startIndex, endIndex);
  }, [filteredMachines, currentPage]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  return {
    searchTerm,
    machines,
    currentPage,
    isLoading,
    isFetching,
    error,
    refetch,
    filteredMachines,
    paginatedMachines,
    totalPages,
    handleSearchChange,
    handlePageChange
  };
}
