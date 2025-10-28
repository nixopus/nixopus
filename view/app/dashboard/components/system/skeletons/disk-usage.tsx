'use client';

import React from 'react';
import { HardDrive } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographySmall } from '@/components/ui/typography';
import { SystemMetricCard } from '../system-metric-card';
import { useSystemMetric } from '../../../hooks/use-system-metric';
import { useTranslation } from '@/hooks/use-translation';
import { DEFAULT_METRICS } from '../../utils/constants';

export function DiskUsageCardSkeletonContent() {
  const { t } = useTranslation();

  return (
    <div className="space-y-2 sm:space-y-3">
      <div className="w-full h-2 bg-gray-200 rounded-full">
        <div className="h-2 rounded-full bg-gray-400" />
      </div>
      <div className="flex justify-between">
        <Skeleton className="h-3 w-20" />
        <Skeleton className="h-3 w-10" />
        <Skeleton className="h-3 w-20" />
      </div>
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
            <tr>
              <td className="text-left pr-1 sm:pr-2">
                <Skeleton className="h-3 w-10" />
              </td>
              <td className="text-right pr-1 sm:pr-2">
                <Skeleton className="h-3 w-10" />
              </td>
              <td className="text-right pr-1 sm:pr-2">
                <Skeleton className="h-3 w-10" />
              </td>
              <td className="text-right">
                <Skeleton className="h-3 w-10" />
              </td>
            </tr>
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
