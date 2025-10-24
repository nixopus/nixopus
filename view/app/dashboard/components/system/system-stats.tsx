'use client';

import React from 'react';
import { SystemStatsType } from '@/redux/types/monitor';
import SystemInfoCard, { SystemInfoCardSkeleton } from './system-info';
import LoadAverageCard, { LoadAverageCardSkeleton } from './load-average';
import MemoryUsageCard, { MemoryUsageCardSkeleton } from './memory-usage';
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
      id: 'memory-usage',
      component: <MemoryUsageCard systemStats={systemStats} />
    }
  ];

  return <DraggableGrid items={systemStatsItems} storageKey="system-stats-card-order" />;
};

export default SystemStats;
