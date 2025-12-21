'use client';

import React from 'react';
import { BarChart } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { SystemMetricCard } from '../system-metric-card';
import { useSystemMetric } from '../../../hooks/use-system-metric';
import { DEFAULT_METRICS } from '../../utils/constants';

export function MemoryUsageCardSkeletonContent() {
  return (
    <div className="space-y-4">
      {/* Doughnut chart placeholder */}
      <div className="flex items-center justify-center h-[200px]">
        <Skeleton className="mx-auto aspect-square max-h-[200px] w-[160px] rounded-full" />
      </div>

      {/* Legend section */}
      <div className="space-y-2">
        <div className="flex justify-between text-xs">
          {/* Used legend item */}
          <div className="flex items-center gap-2">
            <Skeleton className="h-3 w-3 rounded-sm" /> {/* colored square */}
            <Skeleton className="h-4 w-24" /> {/* "Used: X.XX GB" */}
          </div>
          {/* Free legend item */}
          <div className="flex items-center gap-2">
            <Skeleton className="h-3 w-3 rounded-sm" /> {/* colored square */}
            <Skeleton className="h-4 w-24" /> {/* "Free: X.XX GB" */}
          </div>
        </div>
        {/* Total centered */}
        <Skeleton className="h-4 w-28 mx-auto" /> {/* "Total: X.XX GB" */}
      </div>
    </div>
  );
}

export function MemoryUsageCardSkeleton() {
  const { t } = useSystemMetric({
    systemStats: null,
    extractData: (stats) => stats.memory,
    defaultData: DEFAULT_METRICS.memory
  });

  return (
    <SystemMetricCard
      title={t('dashboard.memory.title')}
      icon={BarChart}
      isLoading={true}
      skeletonContent={<MemoryUsageCardSkeletonContent />}
    >
      <div />
    </SystemMetricCard>
  );
}
