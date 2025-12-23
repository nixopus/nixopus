'use client';

import { useMemo } from 'react';
import {
  Cpu,
  MemoryStick,
  HardDriveIcon,
  Plug,
  Edit,
  Terminal,
  FolderOpen,
  Link2,
  Link2Off,
  ChevronUp,
  ChevronDown
} from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { MetricRowConfig, ActionButtonConfig, UseServerDetailsPanelProps } from '../types';

export function useServerDetailsPanel({
  server,
  onEdit,
  onTestConnection,
  onJoinCluster,
  onLeaveCluster,
  onPromote,
  onDemote,
  isTestingConnection,
  hasCluster
}: UseServerDetailsPanelProps) {
  const { t } = useTranslation();
  const isOnline = server.status === 'online';

  const metricRows: MetricRowConfig[] = useMemo(() => {
    if (!server.metrics) return [];
    return [
      { key: 'cpu', icon: Cpu, label: t('servers.details.cpu'), value: server.metrics.cpu_usage },
      {
        key: 'memory',
        icon: MemoryStick,
        label: t('servers.details.memory'),
        value: server.metrics.memory_usage
      },
      {
        key: 'disk',
        icon: HardDriveIcon,
        label: t('servers.details.disk'),
        value: server.metrics.disk_usage
      }
    ];
  }, [server.metrics, t]);

  const clusterActions: ActionButtonConfig[] = useMemo(() => {
    const actions: ActionButtonConfig[] = [];

    if (server.role === 'standalone' && hasCluster) {
      actions.push(
        {
          key: 'join-worker',
          icon: Link2,
          label: `${t('servers.cluster.joinCluster')} ${t('servers.actions.asWorker')}`,
          onClick: () => onJoinCluster(server.id, 'worker'),
          disabled: !isOnline
        },
        {
          key: 'join-manager',
          icon: Link2,
          label: `${t('servers.cluster.joinCluster')} ${t('servers.actions.asManager')}`,
          onClick: () => onJoinCluster(server.id, 'manager'),
          disabled: !isOnline
        }
      );
    }

    if (server.role === 'worker') {
      actions.push(
        {
          key: 'promote',
          icon: ChevronUp,
          label: t('servers.cluster.promote'),
          onClick: () => onPromote(server.id),
          disabled: !isOnline
        },
        {
          key: 'leave',
          icon: Link2Off,
          label: t('servers.cluster.leaveCluster'),
          onClick: () => onLeaveCluster(server.id),
          destructive: true
        }
      );
    }

    if (server.role === 'manager') {
      actions.push(
        {
          key: 'demote',
          icon: ChevronDown,
          label: t('servers.cluster.demote'),
          onClick: () => onDemote(server.id),
          disabled: !isOnline
        },
        {
          key: 'leave',
          icon: Link2Off,
          label: t('servers.cluster.leaveCluster'),
          onClick: () => onLeaveCluster(server.id),
          destructive: true
        }
      );
    }

    return actions;
  }, [
    server.role,
    server.id,
    hasCluster,
    isOnline,
    onJoinCluster,
    onPromote,
    onDemote,
    onLeaveCluster,
    t
  ]);

  const quickActions: ActionButtonConfig[] = useMemo(
    () => [
      {
        key: 'test',
        icon: Plug,
        label: t('servers.testConnection'),
        onClick: () => onTestConnection(server.id),
        disabled: isTestingConnection
      },
      {
        key: 'terminal',
        icon: Terminal,
        label: t('servers.actions.terminal'),
        disabled: !isOnline
      },
      {
        key: 'files',
        icon: FolderOpen,
        label: t('servers.actions.files'),
        disabled: !isOnline
      },
      {
        key: 'edit',
        icon: Edit,
        label: t('common.edit'),
        onClick: () => onEdit(server)
      }
    ],
    [server, isOnline, isTestingConnection, onTestConnection, onEdit, t]
  );

  const showNoClusterMessage = server.role === 'standalone' && !hasCluster;

  const getMetricBarColor = (value: number): string => {
    if (value < 50) return 'bg-emerald-500';
    if (value < 75) return 'bg-amber-500';
    return 'bg-red-500';
  };

  const getMetricTextColor = (value: number): string => {
    if (value < 50) return 'text-emerald-500';
    if (value < 75) return 'text-amber-500';
    return 'text-red-500';
  };

  return {
    t,
    metricRows,
    clusterActions,
    quickActions,
    showNoClusterMessage,
    getMetricBarColor,
    getMetricTextColor,
    isOnline
  };
}
