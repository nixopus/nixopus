'use client';

import React from 'react';
import { SystemStatsType } from '../hooks/use_monitor';
import SystemInfoCard, { SystemInfoCardSkeleton } from './system_info';
import LoadAverageCard, { LoadAverageCardSkeleton } from './load_average';
import MemoryUsageCard, { MemoryUsageCardSkeleton } from './memory_usage';

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
