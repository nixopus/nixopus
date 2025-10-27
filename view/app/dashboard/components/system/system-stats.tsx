'use client';

import React from 'react';
import { SystemStatsType } from '@/redux/types/monitor';
import SystemInfoCard from './system-info';
import LoadAverageCard from './load-average';
import CPUUsageCard from './cpu-usage';
import MemoryUsageCard from './memory-usage';
import { SystemInfoCardSkeleton } from './skeletons/system-info';
import { LoadAverageCardSkeleton } from './skeletons/load-average';
import { CPUUsageCardSkeleton } from './skeletons/cpu-usage';
import { MemoryUsageCardSkeleton } from './skeletons/memory-usage';
import { DraggableGrid, DraggableItem } from '@/components/ui/draggable-grid';

export interface SystemStatsProps {
  systemStats: SystemStatsType | null;
}

const SystemStats: React.FC<SystemStatsProps> = ({ systemStats }) => {
  if (!systemStats) {
    return (
      <div className="space-y-4">
        <SystemInfoCardSkeleton />
        <LoadAverageCardSkeleton />
        <CPUUsageCardSkeleton />
        <MemoryUsageCardSkeleton />
      </div>
    );
  }

  const systemStatsItems: DraggableItem[] = [
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

  return <DraggableGrid items={systemStatsItems} storageKey="system-stats-card-order" />;
};

export default SystemStats;
