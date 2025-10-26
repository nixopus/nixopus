'use client';

import React from 'react';
import { BarChart } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { DoughnutChartComponent, DoughnutChartDataItem } from '@/components/ui/doughnut-chart-component';
import { SystemMetricCard } from './system-metric-card';

interface MemoryUsageCardProps {
  systemStats: SystemStatsType | null;
}

const formatGB = (value: number) => `${value.toFixed(2)}`;

const MemoryUsageCard: React.FC<MemoryUsageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  const isLoading = !systemStats;

  const { memory } = systemStats || {
    memory: {
      total: 0,
      used: 0,
      percentage: 0
    }
  };

  // Calculate free memory
  const freeMemory = memory.total - memory.used;

  // Prepare data for doughnut chart with distinct colors
  const chartData: DoughnutChartDataItem[] = [
    {
      name: 'Used',
      value: memory.used,
      fill: '#3b82f6' // Bright blue for used memory
    },
    {
      name: 'Free',
      value: freeMemory,
      fill: '#10b981' // Bright green for free memory
    }
  ];

  const chartConfig = {
    used: {
      label: 'Used Memory',
      color: '#3b82f6'
    },
    free: {
      label: 'Free Memory',
      color: '#10b981'
    }
  };

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
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: '#3b82f6' }} />
              <TypographyMuted>
                Used: {formatGB(memory.used)} GB
              </TypographyMuted>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: '#10b981' }} />
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
  const { t } = useTranslation();

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
