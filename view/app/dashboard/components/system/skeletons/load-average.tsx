'use client';

import React from 'react';
import { Activity } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographyMuted } from '@/components/ui/typography';
import { SystemMetricCard } from '../system-metric-card';
import { useSystemMetric } from '../../../hooks/use-system-metric';
import { DEFAULT_METRICS } from '../../utils/constants';

export function LoadAverageCardSkeletonContent() {
  return (
    <div className="space-y-4">
      <div>
        <Skeleton className="h-[180px] w-full rounded-lg" />
      </div>
      <div className="grid grid-cols-3 gap-2 text-center">
        <div>
          <TypographyMuted className="text-xs">1 min</TypographyMuted>
          <Skeleton className="h-4 w-12 mx-auto mt-1" />
        </div>
        <div>
          <TypographyMuted className="text-xs">5 min</TypographyMuted>
          <Skeleton className="h-4 w-12 mx-auto mt-1" />
        </div>
        <div>
          <TypographyMuted className="text-xs">15 min</TypographyMuted>
          <Skeleton className="h-4 w-12 mx-auto mt-1" />
        </div>
      </div>
    </div>
  );
}

export function LoadAverageCardSkeleton() {
  const { t } = useSystemMetric({
    systemStats: null,
    extractData: (stats) => stats.load,
    defaultData: DEFAULT_METRICS.load
  });

  return (
    <SystemMetricCard
      title={t('dashboard.load.title')}
      icon={Activity}
      isLoading={true}
      skeletonContent={<LoadAverageCardSkeletonContent />}
    >
      <div />
    </SystemMetricCard>
  );
}
