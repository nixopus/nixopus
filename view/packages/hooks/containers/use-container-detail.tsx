'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import {
  useGetContainerQuery,
  useStartContainerMutation,
  useStopContainerMutation,
  useRemoveContainerMutation,
  useGetContainerLogsQuery
} from '@/redux/services/container/containerApi';
import { useRouter, useParams } from 'next/navigation';
import { useEffect, useState, useCallback } from 'react';
import { getAdvancedSettings } from '@/packages/utils/advanced-settings';

export function useContainerDetail() {
  const { t } = useTranslation();
  const router = useRouter();
  const params = useParams();
  const containerId = params.id as string;
  const { data: container, isLoading, error } = useGetContainerQuery(containerId);
  const [startContainer] = useStartContainerMutation();
  const [stopContainer] = useStopContainerMutation();
  const [removeContainer] = useRemoveContainerMutation();
  const containerSettings = getAdvancedSettings();
  const [logsTail, setLogsTail] = useState(containerSettings.containerLogTailLines);
  const [allLogs, setAllLogs] = useState<string>('');
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const { data: logs, refetch: refetchLogs } = useGetContainerLogsQuery(
    { containerId, tail: logsTail },
    {
      skip: !containerId,
      refetchOnMountOrArgChange: true
    }
  );

  useEffect(() => {
    if (logs) {
      setAllLogs(logs);
    }
  }, [logs]);

  const handleLoadMoreLogs = useCallback(async () => {
    const newTail = logsTail + containerSettings.containerLogTailLines;
    setLogsTail(newTail);
    await refetchLogs();
  }, [logsTail, containerSettings.containerLogTailLines, refetchLogs]);

  const handleRefreshLogs = useCallback(async () => {
    await refetchLogs();
  }, [refetchLogs]);

  const handleContainerAction = async (action: 'start' | 'stop' | 'remove' | 'restart') => {
    try {
      switch (action) {
        case 'start':
          await startContainer(containerId).unwrap();
          toast.success(t(`containers.${action}_success`));
          break;
        case 'stop':
          await stopContainer(containerId).unwrap();
          toast.success(t(`containers.${action}_success`));
          break;
        case 'restart':
          await stopContainer(containerId).unwrap();
          await startContainer(containerId).unwrap();
          toast.success(t('containers.restart_success'));
          break;
        case 'remove':
          setIsDeleteDialogOpen(true);
          break;
      }
    } catch (error) {
      if (action === 'restart') {
        toast.error(t('containers.restart_error'));
      } else {
        toast.error(t(`containers.${action}_error`));
      }
    }
  };

  const handleDeleteConfirm = async () => {
    try {
      await removeContainer(containerId).unwrap();
      toast.success(t('containers.remove_success'));
      router.push('/containers');
    } catch (error) {
      toast.error(t('containers.remove_error'));
    }
  };

  return {
    handleDeleteConfirm,
    handleContainerAction,
    handleLoadMoreLogs,
    handleRefreshLogs,
    isDeleteDialogOpen,
    container,
    isLoading,
    allLogs,
    containerId,
    t,
    setIsDeleteDialogOpen
  };
}
