import React from 'react';
import { DraggableItem } from '@/components/ui/draggable-grid';
import { SystemStatsType } from '@/redux/types/monitor';

export interface SystemStatsComponentCreators {
  SystemInfoCard: React.FC<{ systemStats: SystemStatsType }>;
  LoadAverageCard: React.FC<{ systemStats: SystemStatsType }>;
  CPUUsageCard: React.FC<{ systemStats: SystemStatsType }>;
  MemoryUsageCard: React.FC<{ systemStats: SystemStatsType }>;
}

export function useSystemStatsItems(
  systemStats: SystemStatsType,
  components: SystemStatsComponentCreators
): DraggableItem[] {
  const { SystemInfoCard, LoadAverageCard, CPUUsageCard, MemoryUsageCard } = components;

  return [
    {
      id: 'system-info',
      component: <SystemInfoCard systemStats={systemStats} />
    },
    {
      id: 'load-average',
      component: <LoadAverageCard systemStats={systemStats} />
    },
    {
      id: 'cpu-usage',
      component: <CPUUsageCard systemStats={systemStats} />
    },
    {
      id: 'memory-usage',
      component: <MemoryUsageCard systemStats={systemStats} />
    }
  ];
}
