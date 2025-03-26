'use client';

import React from 'react';
import { SystemStatsType } from '../hooks/use_monitor';
import SystemInfoCard from './system_info';
import LoadAverageCard from './load_average';
import MemoryUsageCard from './memory_usage';

export interface SystemStatsProps {
  systemStats: SystemStatsType;
}

const SystemStats: React.FC<SystemStatsProps> = ({ systemStats }) => {
  if (!systemStats) return null;

  return (
    <div className="space-y-4">
      <SystemInfoCard systemStats={systemStats} />
      <LoadAverageCard systemStats={systemStats} />
      <MemoryUsageCard systemStats={systemStats} />
    </div>
  );
};

export default SystemStats;
