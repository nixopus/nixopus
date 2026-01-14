'use client';

import React from 'react';
import { Clock, Network, ArrowDown, ArrowUp } from 'lucide-react';
import useClock from '@/packages/hooks/dashboard/use-clock';
import { useSystemStats } from '@/packages/hooks/shared/system-stats-provider';
import { useNetworkSpeeds } from '@/packages/hooks/dashboard/use-network-speeds';

interface NetworkWidgetProps {
  networkSpeeds: {
    downloadSpeed: string;
    uploadSpeed: string;
  };
}

const NetworkWidget: React.FC<NetworkWidgetProps> = ({ networkSpeeds }) => {
  return (
    <div className="hidden lg:flex items-center gap-3 px-2 text-sm">
      <Network className="h-4 w-4 text-muted-foreground" />
      <div className="flex items-center gap-1.5">
        <ArrowDown className="h-3.5 w-3.5 text-blue-500" />
        <span className="font-medium">{networkSpeeds.downloadSpeed}</span>
      </div>
      <div className="flex items-center gap-1.5">
        <ArrowUp className="h-3.5 w-3.5 text-green-500" />
        <span className="font-medium">{networkSpeeds.uploadSpeed}</span>
      </div>
    </div>
  );
};

interface ClockWidgetProps {
  formattedTime: string;
  formattedDate: string;
}

const ClockWidget: React.FC<ClockWidgetProps> = ({ formattedTime, formattedDate }) => {
  return (
    <div className="hidden xl:flex items-center gap-3 px-2 text-sm border-l border-border pl-3">
      <Clock className="h-4 w-4 text-muted-foreground" />
      <div className="flex flex-col">
        <span className="font-medium tabular-nums leading-tight">{formattedTime}</span>
        <span className="text-muted-foreground text-[11px] hidden 2xl:block leading-tight">
          {formattedDate.split(',')[0]}
        </span>
      </div>
    </div>
  );
};

export function TopbarWidgets() {
  const { formattedTime, formattedDate } = useClock();
  const { systemStats } = useSystemStats();
  const { networkSpeeds } = useNetworkSpeeds(systemStats);

  return (
    <div className="flex items-center gap-3 pr-4">
      <NetworkWidget networkSpeeds={networkSpeeds} />
      <ClockWidget formattedTime={formattedTime} formattedDate={formattedDate} />
    </div>
  );
}
