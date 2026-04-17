'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';
import {
  useListBackupsQuery,
  useTriggerBackupMutation,
} from '@/redux/services/machine/machineBackupApi';
import type { BackupListParams } from '@/redux/types/machine-backup';

const PAGE_SIZE = 20;
const ACTIVE_POLL_INTERVAL = 10_000;

export function useMachineBackup(serverId?: string) {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState<string>('created_at');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [statusFilter, setStatusFilter] = useState<string>('');

  const params: BackupListParams = useMemo(
    () => ({
      page,
      page_size: PAGE_SIZE,
      search: search || undefined,
      sort_by: sortBy,
      sort_order: sortOrder,
      status: statusFilter || undefined,
      server_id: serverId,
    }),
    [page, search, sortBy, sortOrder, statusFilter, serverId],
  );

  const hasInProgressRef = useRef(false);

  const {
    data,
    isLoading: isListLoading,
    isFetching: isListFetching,
    error: listError,
    refetch,
  } = useListBackupsQuery(params, {
    refetchOnMountOrArgChange: 30,
    pollingInterval: hasInProgressRef.current ? ACTIVE_POLL_INTERVAL : 0,
  });

  const backups = data?.backups ?? [];
  const totalCount = data?.total_count ?? 0;
  const totalPages = Math.ceil(totalCount / PAGE_SIZE);

  const hasInProgress = useMemo(
    () => backups.some((b) => b.status === 'pending' || b.status === 'in_progress'),
    [backups],
  );

  hasInProgressRef.current = hasInProgress;

  const [triggerBackup, triggerState] = useTriggerBackupMutation();

  const handleTriggerBackup = useCallback(async () => {
    try {
      const result = await triggerBackup(serverId ? { server_id: serverId } : undefined).unwrap();
      toast.success(result.message || 'Backup initiated');
      refetch();
    } catch (err: unknown) {
      const e = err as { data?: { detail?: string; message?: string } };
      const detail = e?.data?.detail || e?.data?.message || 'Failed to trigger backup';
      toast.error(detail);
    }
  }, [triggerBackup, refetch, serverId]);

  const handleSearch = useCallback((term: string) => {
    setSearch(term);
    setPage(1);
  }, []);

  const handleSort = useCallback((column: string) => {
    setSortBy((prev) => {
      if (prev === column) {
        setSortOrder((o) => (o === 'asc' ? 'desc' : 'asc'));
        return prev;
      }
      setSortOrder('desc');
      return column;
    });
    setPage(1);
  }, []);

  const handleStatusFilter = useCallback((status: string) => {
    setStatusFilter(status);
    setPage(1);
  }, []);

  const handlePageChange = useCallback((newPage: number) => {
    setPage(newPage);
  }, []);

  return {
    backups,
    totalCount,
    totalPages,
    page,
    isListLoading,
    isListFetching,
    listError,
    isTriggerLoading: triggerState.isLoading,
    hasInProgress,
    search,
    sortBy,
    sortOrder,
    statusFilter,
    handleTriggerBackup,
    handleSearch,
    handleSort,
    handleStatusFilter,
    handlePageChange,
    refetch,
  };
}
