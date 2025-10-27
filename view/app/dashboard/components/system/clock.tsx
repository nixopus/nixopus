'use client';

import React from 'react';
import { Clock } from 'lucide-react';
import { SystemMetricCard } from './system-metric-card';
import { ClockCardSkeletonContent } from './skeletons/clock';
import useClock from '../../hooks/use-clock';

const ClockWidget: React.FC = () => {
  const { formattedTime, formattedDate } = useClock();

  return (
    <SystemMetricCard
      title="Clock"
      icon={Clock}
      isLoading={false}
      skeletonContent={<ClockCardSkeletonContent />}
    >
      <div className="flex flex-col items-center justify-center h-full space-y-3">
        <div className="text-5xl font-bold text-primary tabular-nums">
          {formattedTime}
        </div>
        <div className="text-sm text-muted-foreground">
          {formattedDate}
        </div>
      </div>
    </SystemMetricCard>
  );
};

export default ClockWidget;

