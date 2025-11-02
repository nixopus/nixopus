'use client';

import React from 'react';
import { Thermometer } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographyMuted } from '@/components/ui/typography';
import { SystemMetricCard } from '../system-metric-card';
import { useSystemMetric } from '../../../hooks/use-system-metric';
import { DEFAULT_METRICS } from '../../utils/constants';

export function CPUTemperatureCardSkeletonContent() {
  return (
    <div className="space-y-6">
      <div className="text-center">
        <TypographyMuted className="text-xs">Current Temperature</TypographyMuted>
        <Skeleton className="h-12 w-32 mx-auto mt-1" />
        <Skeleton className="h-4 w-16 mx-auto mt-1" />
      </div>
      <div className="space-y-2">
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>0°C</span>
          <span>50°C</span>
          <span>100°C</span>
        </div>
        <Skeleton className="h-4 w-full rounded-full" />
        <div className="flex justify-between">
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-3 w-16" />
        </div>
      </div>
    </div>
  );
}

export function CPUTemperatureCardSkeleton() {
  const { t } = useSystemMetric({
    systemStats: null,
    extractData: (stats) => stats.cpu,
    defaultData: DEFAULT_METRICS.cpu
  });

  return (
    <SystemMetricCard
      title={t('dashboard.cpu.temperature')}
      icon={Thermometer}
      isLoading={true}
      skeletonContent={<CPUTemperatureCardSkeletonContent />}
    >
      <div />
    </SystemMetricCard>
  );
}

