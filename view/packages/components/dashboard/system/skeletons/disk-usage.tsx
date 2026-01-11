'use client';

import React from 'react';
import { HardDrive } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographySmall } from '@/components/ui/typography';
import { SystemMetricCard } from '../system-metric-card';
import { useSystemMetric } from '@/packages/hooks/dashboard/use-system-metric';
import { useTranslation } from '@/hooks/use-translation';
import { DEFAULT_METRICS } from '../../utils/constants';

export function DiskUsageCardSkeletonContent() {
  const { t } = useTranslation();

  return (
    <div className="space-y-2 sm:space-y-3">
      {/* Progress bar */}
      <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
        <Skeleton className="h-2 w-1/2 rounded-full" />
      </div>
      {/* Used, percentage, total row */}
      <div className="flex justify-between">
        <Skeleton className="h-3 w-16" /> {/* Used: X.XX GB */}
        <Skeleton className="h-3 w-8" /> {/* XX% */}
        <Skeleton className="h-3 w-16" /> {/* Total: X.XX GB */}
      </div>
      {/* Table */}
      <div className="text-xs font-mono mt-1 sm:mt-2 overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr>
              <th className="text-left pr-1 sm:pr-2">
                <TypographySmall className="text-xs">
                  {t('dashboard.disk.table.headers.mount')}
                </TypographySmall>
              </th>
              <th className="text-right pr-1 sm:pr-2">
                <TypographySmall className="text-xs">
                  {t('dashboard.disk.table.headers.size')}
                </TypographySmall>
              </th>
              <th className="text-right pr-1 sm:pr-2">
                <TypographySmall className="text-xs">
                  {t('dashboard.disk.table.headers.used')}
                </TypographySmall>
              </th>
              <th className="text-right">
                <TypographySmall className="text-xs">
                  {t('dashboard.disk.table.headers.percentage')}
                </TypographySmall>
              </th>
            </tr>
          </thead>
          <tbody className="text-xxs sm:text-xs">
            {[0, 1].map((i) => (
              <tr key={i}>
                <td className="text-left pr-1 sm:pr-2 py-1">
                  <Skeleton className="h-3 w-8" />
                </td>
                <td className="text-right pr-1 sm:pr-2 py-1">
                  <Skeleton className="h-3 w-10 ml-auto" />
                </td>
                <td className="text-right pr-1 sm:pr-2 py-1">
                  <Skeleton className="h-3 w-10 ml-auto" />
                </td>
                <td className="text-right py-1">
                  <Skeleton className="h-3 w-8 ml-auto" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function DiskUsageCardSkeleton() {
  const { t } = useSystemMetric({
    systemStats: null,
    extractData: (stats) => stats.disk,
    defaultData: DEFAULT_METRICS.disk
  });

  return (
    <SystemMetricCard
      title={t('dashboard.disk.title')}
      icon={HardDrive}
      isLoading={true}
      skeletonContent={<DiskUsageCardSkeletonContent />}
    >
      <div />
    </SystemMetricCard>
  );
}
