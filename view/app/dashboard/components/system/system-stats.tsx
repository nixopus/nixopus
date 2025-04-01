'use client';

import React from 'react';
import { SystemStatsType } from '@/redux/types/monitor';
import SystemInfoCard, { SystemInfoCardSkeleton } from './system-info';
import LoadAverageCard, { LoadAverageCardSkeleton } from './load-average';
import MemoryUsageCard, { MemoryUsageCardSkeleton } from './memory-usage';

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

  return (
    <div className="space-y-4">
      <SystemInfoCard systemStats={systemStats} />
      <LoadAverageCard systemStats={systemStats} />
      <MemoryUsageCard systemStats={systemStats} />
    </div>
  );
};

export default SystemStats;
