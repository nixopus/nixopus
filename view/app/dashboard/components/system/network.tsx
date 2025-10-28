'use client';

import React from 'react';
import { Network } from 'lucide-react';
import { SystemMetricCard } from './system-metric-card';
import { NetworkCardSkeletonContent } from './skeletons/network';
import { useNetwork } from '../../hooks/use-network';
import { SystemStatsType } from '@/redux/types/monitor';
import { ArrowDownCircle, ArrowUpCircle } from 'lucide-react';

interface NetworkWidgetProps {
  systemStats: SystemStatsType | null;
}

const NetworkWidget: React.FC<NetworkWidgetProps> = ({ systemStats }) => {
  const { networkData } = useNetwork({ systemStats });

  const isLoading = !systemStats || !systemStats.network;

  return (
    <SystemMetricCard
      title="Network Traffic"
      icon={Network}
      isLoading={isLoading}
      skeletonContent={<NetworkCardSkeletonContent />}
    >
      <div className="flex flex-col items-center justify-center h-full space-y-4">
        <div className="grid grid-cols-2 gap-4 w-full">
          <div className="flex flex-col items-center text-center">
            <ArrowDownCircle className="h-8 w-8 text-blue-500 mb-2" />
            <div className="text-xs text-muted-foreground mb-1">Download</div>
            <div className="text-2xl font-bold text-primary tabular-nums">
              {networkData.downloadSpeed}
            </div>
          </div>
          <div className="flex flex-col items-center text-center">
            <ArrowUpCircle className="h-8 w-8 text-green-500 mb-2" />
            <div className="text-xs text-muted-foreground mb-1">Upload</div>
            <div className="text-2xl font-bold text-primary tabular-nums">
              {networkData.uploadSpeed}
            </div>
          </div>
        </div>
        <div className="flex gap-4 text-xs text-muted-foreground">
          <span>↓ {networkData.totalDownload}</span>
          <span>↑ {networkData.totalUpload}</span>
        </div>
      </div>
    </SystemMetricCard>
  );
};

export default NetworkWidget;
