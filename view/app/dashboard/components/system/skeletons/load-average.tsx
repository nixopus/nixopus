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
    <>
      <br />
      <br />
      <br />
      <div className="space-y-4">
        <div>
          <Skeleton className="h-[180px] w-full rounded-lg" /> {/* bar chart */}
        </div>
        <div className="grid grid-cols-3 gap-2 text-center">
          {['1 min', '5 min', '15 min'].map((label, i) => (
            <div key={i} className="flex flex-col items-center gap-1">
              <div className="flex items-center gap-1">
                <Skeleton className="h-2 w-2 rounded-full" /> {/* colored dot */}
                <TypographyMuted className="text-xs">{label}</TypographyMuted>
              </div>
              <Skeleton className="h-5 w-10" /> {/* text-sm font-bold value */}
            </div>
          ))}
        </div>
      </div>
    </>
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
