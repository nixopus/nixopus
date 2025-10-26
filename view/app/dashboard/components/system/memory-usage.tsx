'use client';

import React from 'react';
import { BarChart } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { DoughnutChartComponent } from '@/components/ui/doughnut-chart-component';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '../../hooks/use-system-metric';
import { formatGB, createMemoryChartData, createMemoryChartConfig } from './utils';
import { DEFAULT_METRICS, CHART_COLORS } from './constants';

interface MemoryUsageCardProps {
  systemStats: SystemStatsType | null;
}

const MemoryUsageCard: React.FC<MemoryUsageCardProps> = ({ systemStats }) => {
  const { data: memory, isLoading, t } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.memory,
    defaultData: DEFAULT_METRICS.memory,
  });

  // Calculate free memory
  const freeMemory = memory.total - memory.used;

  const chartData = createMemoryChartData(memory.used, freeMemory);
  const chartConfig = createMemoryChartConfig();

  return (
    <SystemMetricCard
      title={t('dashboard.memory.title')}
      icon={BarChart}
      isLoading={isLoading}
      skeletonContent={<MemoryUsageCardSkeletonContent />}
    >
      <div className="space-y-4">
        {/* Doughnut Chart */}
        <div className="flex items-center justify-center h-[200px]">
          <DoughnutChartComponent
            data={chartData}
            chartConfig={chartConfig}
            centerLabel={{
              value: `${memory.percentage.toFixed(1)}%`,
              subLabel: 'Used'
            }}
            innerRadius={60}
            outerRadius={80}
            maxHeight="max-h-[200px]"
          />
        </div>

        {/* Summary Stats with Distinct Colors */}
        <div className="space-y-2">
          <div className="flex justify-between text-xs">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: CHART_COLORS.blue }} />
              <TypographyMuted>
                Used: {formatGB(memory.used)} GB
              </TypographyMuted>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: CHART_COLORS.green }} />
              <TypographyMuted>
                Free: {formatGB(freeMemory)} GB
              </TypographyMuted>
            </div>
          </div>

          {/* Additional Info */}
          <TypographyMuted className="text-xs text-center">
            Total: {formatGB(memory.total)} GB
          </TypographyMuted>
        </div>
      </div>
    </SystemMetricCard>
  );
};

export default MemoryUsageCard;

function MemoryUsageCardSkeletonContent() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-center h-[200px]">
        <Skeleton className="mx-auto aspect-square max-h-[200px] w-[200px] rounded-full" />
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-xs">
          <div className="flex items-center gap-2">
            <Skeleton className="h-3 w-3 rounded-sm" />
            <Skeleton className="h-4 w-20" />
          </div>
          <div className="flex items-center gap-2">
            <Skeleton className="h-3 w-3 rounded-sm" />
            <Skeleton className="h-4 w-20" />
          </div>
        </div>

        <Skeleton className="h-4 w-32 mx-auto" />
      </div>
    </div>
  );
}

export function MemoryUsageCardSkeleton() {
  const { t } = useSystemMetric({
    systemStats: null,
    extractData: (stats) => stats.memory,
    defaultData: DEFAULT_METRICS.memory,
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
