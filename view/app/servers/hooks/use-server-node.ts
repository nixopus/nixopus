'use client';

import { useMemo } from 'react';
import { Cpu, MemoryStick, HardDriveIcon } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { MetricBarConfig, UseServerNodeProps } from '../types';

export function useServerNode({ server }: UseServerNodeProps) {
  const { t } = useTranslation();

  const metricBars: MetricBarConfig[] = useMemo(() => {
    if (!server.metrics) return [];
    return [
      { key: 'cpu', icon: Cpu, label: 'CPU', value: server.metrics.cpu_usage },
      { key: 'ram', icon: MemoryStick, label: 'RAM', value: server.metrics.memory_usage },
      { key: 'disk', icon: HardDriveIcon, label: 'Disk', value: server.metrics.disk_usage }
    ];
  }, [server.metrics]);

  const visibleLabels = server.labels.slice(0, 2);
  const remainingLabelsCount = server.labels.length > 2 ? server.labels.length - 2 : 0;

  const getMetricBarColor = (value: number): string => {
    if (value < 50) return 'bg-emerald-500';
    if (value < 75) return 'bg-amber-500';
    return 'bg-red-500';
  };

  const containerCountText = server.metrics
    ? `${server.metrics.container_count} ${t('servers.details.containers').toLowerCase()}`
    : '';

  const offlineText = t('servers.status.offline');
  const connectingText = `${t('servers.status.connecting')}...`;
  const roleLabel = server.role.charAt(0).toUpperCase() + server.role.slice(1);

  return {
    t,
    metricBars,
    visibleLabels,
    remainingLabelsCount,
    getMetricBarColor,
    containerCountText,
    offlineText,
    connectingText,
    roleLabel
  };
}
