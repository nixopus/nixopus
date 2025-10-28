'use client';

import React from 'react';
import { Cpu } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographyMuted } from '@/components/ui/typography';
import { SystemMetricCard } from '../system-metric-card';
import { useSystemMetric } from '../../../hooks/use-system-metric';
import { DEFAULT_METRICS } from '../../utils/constants';

export function CPUUsageCardSkeletonContent() {
  return (
    <div className="space-y-4">
      <div className="text-center">
        <TypographyMuted className="text-xs">Overall</TypographyMuted>
        <Skeleton className="h-10 w-20 mx-auto mt-1" />
      </div>
      <div>
        <Skeleton className="h-[180px] w-full rounded-lg" />
      </div>
      <div className="grid grid-cols-3 gap-2 text-center">
        {[0, 1, 2].map((i) => (
          <div key={i} className="flex flex-col items-center gap-1">
            <TypographyMuted className="text-xs">Core {i}</TypographyMuted>
            <Skeleton className="h-4 w-12 mx-auto mt-1" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function CPUUsageCardSkeleton() {
  const { t } = useSystemMetric({
    systemStats: null,
    extractData: (stats) => stats.cpu,
    defaultData: DEFAULT_METRICS.cpu
  });

  return (
    <SystemMetricCard
      title={t('dashboard.cpu.title')}
      icon={Cpu}
      isLoading={true}
      skeletonContent={<CPUUsageCardSkeletonContent />}
    >
      <div />
    </SystemMetricCard>
  );
}
