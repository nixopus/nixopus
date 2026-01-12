'use client';

import React from 'react';
import { HardDrive } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographySmall } from '@/components/ui/typography';
import { useSystemMetric } from '@/packages/hooks/dashboard/use-system-metric';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { DEFAULT_METRICS } from '@/packages/utils/dashboard';
import { Card } from '@/components/ui/card';
import { CardHeader } from '@/components/ui/card';
import { CardTitle } from '@/components/ui/card';
import { Server } from 'lucide-react';
import { TypographyMuted } from '@/components/ui/typography';
import { CardContent } from '@/components/ui/card';
import { Cpu } from 'lucide-react';
import { Activity } from 'lucide-react';
import { BarChart } from 'lucide-react';
import { SystemMetricCard } from './dashboard';

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

export const ClockCardSkeletonContent = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full space-y-3">
      <Skeleton className="h-12 w-40" /> {/* text-5xl time */}
      <Skeleton className="h-4 w-32" /> {/* text-sm date */}
    </div>
  );
};

export function CPUUsageCardSkeletonContent() {
  return (
    <div className="space-y-4">
      <div className="text-center">
        <TypographyMuted className="text-xs">Overall</TypographyMuted>
        <Skeleton className="h-9 w-20 mx-auto mt-1" /> {/* text-3xl percentage */}
      </div>
      <div>
        <Skeleton className="h-[180px] w-full rounded-lg" /> {/* bar chart */}
      </div>
      <div className="grid grid-cols-3 gap-2 text-center">
        {[0, 1, 2].map((i) => (
          <div key={i} className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <Skeleton className="h-2 w-2 rounded-full" /> {/* colored dot */}
              <TypographyMuted className="text-xs">Core {i}</TypographyMuted>
            </div>
            <Skeleton className="h-5 w-12" /> {/* text-sm font-bold percentage */}
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

export const NetworkCardSkeletonContent: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full space-y-4">
      {/* 2-column grid for download/upload */}
      <div className="grid grid-cols-2 gap-4 w-full">
        {/* Download column */}
        <div className="flex flex-col items-center text-center">
          <Skeleton className="h-8 w-8 rounded-full mb-2" /> {/* ArrowDownCircle icon */}
          <Skeleton className="h-3 w-14 mb-1" /> {/* "Download" label */}
          <Skeleton className="h-7 w-16" /> {/* text-2xl speed value */}
        </div>
        {/* Upload column */}
        <div className="flex flex-col items-center text-center">
          <Skeleton className="h-8 w-8 rounded-full mb-2" /> {/* ArrowUpCircle icon */}
          <Skeleton className="h-3 w-12 mb-1" /> {/* "Upload" label */}
          <Skeleton className="h-7 w-16" /> {/* text-2xl speed value */}
        </div>
      </div>
      {/* Total download/upload row */}
      <div className="flex gap-4 text-xs">
        <Skeleton className="h-3 w-16" /> {/* ↓ total download */}
        <Skeleton className="h-3 w-16" /> {/* ↑ total upload */}
      </div>
    </div>
  );
};

export function SystemInfoCardSkeleton() {
  const { t } = useTranslation();

  return (
    <Card className="overflow-hidden h-full flex flex-col w-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-bold flex items-center">
          <Server className="h-4 w-4 mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.system.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[...Array(8)].map((_, index) => (
            <div key={index} className="flex items-start gap-3 p-2 rounded-lg">
              {/* Icon placeholder */}
              <Skeleton className="h-4 w-4 mt-0.5 rounded" />
              <div className="flex-1 min-w-0 space-y-1">
                {/* Label */}
                <Skeleton className="h-3 w-16" />
                {/* Value */}
                <Skeleton className="h-3 w-20" />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
