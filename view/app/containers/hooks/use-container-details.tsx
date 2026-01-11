'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import {
  useGetContainerQuery,
  useStartContainerMutation,
  useStopContainerMutation,
  useRemoveContainerMutation,
  useGetContainerLogsQuery
} from '@/redux/services/container/containerApi';
import { useRouter, useParams } from 'next/navigation';
import { useEffect, useState, useMemo, useCallback } from 'react';
import { getAdvancedSettings } from '@/lib/advanced-settings';
import { TabItem } from '@/components/ui/tabs-wrapper';
import { Info, ScrollText, Terminal, Layers } from 'lucide-react';
import { OverviewTab } from '../[id]/components/OverviewTab';
import { LogsTab } from '../[id]/components/LogsTab';
import { Terminal as TerminalComponent } from '../[id]/components/Terminal';
import { Images } from '../[id]/components/images';

function useContainerDetails() {
  const { t } = useTranslation();
  const router = useRouter();
  const params = useParams();
  const containerId = params.id as string;
  const [tab, setTab] = useState('overview');
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

  const tabs: TabItem[] = useMemo(
    () => [
      {
        value: 'overview',
        label: t('containers.overview') || 'Overview',
        icon: Info,
        content: container ? <OverviewTab container={container} /> : null
      },
      {
        value: 'logs',
        label: t('containers.logs.title') || 'Logs',
        icon: ScrollText,
        content: container ? (
          <LogsTab
            container={container}
            logs={allLogs}
            onLoadMore={handleLoadMoreLogs}
            onRefresh={handleRefreshLogs}
          />
        ) : null
      },
      {
        value: 'terminal',
        label: t('terminal.title') || 'Terminal',
        icon: Terminal,
        disabled: container?.status !== 'running',
        content:
          container && container.status === 'running' ? (
            <div className="rounded-xl overflow-hidden">
              <TerminalComponent containerId={containerId} />
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <Terminal className="h-12 w-12 mb-4 opacity-30" />
              <p className="font-medium">Container not running</p>
              <p className="text-sm text-muted-foreground/60 mt-1">
                Start the container to access the terminal
              </p>
            </div>
          )
      },
      {
        value: 'images',
        label: t('containers.images.title') || 'Images',
        icon: Layers,
        content: container ? (
          <Images containerId={containerId} imagePrefix={(container.image || '') + '*'} />
        ) : null
      }
    ],
    [container, allLogs, containerId, t, handleLoadMoreLogs, handleRefreshLogs]
  );

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
    setIsDeleteDialogOpen,
    tab,
    setTab,
    tabs
  };
}

export default useContainerDetails;
