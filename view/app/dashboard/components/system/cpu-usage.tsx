'use client';

import React from 'react';
import { Cpu } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { BarChartComponent } from '@/components/ui/bar-chart-component';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '../../hooks/use-system-metric';
import { createCPUChartData, createCPUChartConfig, formatPercentage } from './utils';
import { DEFAULT_METRICS, CHART_COLORS } from './constants';

interface CPUUsageCardProps {
  systemStats: SystemStatsType | null;
}

const CPUUsageCard: React.FC<CPUUsageCardProps> = ({ systemStats }) => {
  const { data: cpu, isLoading, t } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.cpu || DEFAULT_METRICS.cpu,
    defaultData: DEFAULT_METRICS.cpu,
  });

  const perCoreData = cpu?.per_core || [];
  const chartData = createCPUChartData(perCoreData);
  const chartConfig = createCPUChartConfig(perCoreData.length);

  // Get top 3 cores by usage
  const topCores = perCoreData.length > 0
    ? [...perCoreData].sort((a, b) => b.usage - a.usage).slice(0, 3)
    : [];

  return (
    <SystemMetricCard
      title={t('dashboard.cpu.title')}
      icon={Cpu}
      isLoading={isLoading}
      skeletonContent={<CPUUsageCardSkeletonContent />}
    >
      <div className="space-y-4">
        {/* Overall CPU Usage */}
        <div className="text-center">
          <TypographyMuted className="text-xs">{t('dashboard.cpu.overall')}</TypographyMuted>
          <div className="text-3xl font-bold text-primary mt-1">
            {formatPercentage(cpu?.overall || 0)}%
          </div>
        </div>

        {/* Bar Chart for Per-Core Usage */}
        {perCoreData.length > 0 ? (
          <>
            <div>
              <BarChartComponent
                data={chartData}
                chartConfig={chartConfig}
                height="h-[180px]"
                yAxisLabel={t('dashboard.cpu.usage')}
                xAxisLabel={t('dashboard.cpu.cores')}
                showAxisLabels={true}
              />
            </div>

            {/* Top 3 Cores Summary */}
            {topCores.length > 0 && (
              <div className="grid grid-cols-3 gap-2 text-center">
                {topCores.map((core) => {
                  const colors = [
                    CHART_COLORS.blue,
                    CHART_COLORS.green,
                    CHART_COLORS.orange,
                    CHART_COLORS.purple,
                    CHART_COLORS.red,
                    CHART_COLORS.yellow,
                  ];
                  const color = colors[core.core_id % colors.length];

                  return (
                    <div key={core.core_id} className="flex flex-col items-center gap-1">
                      <div className="flex items-center gap-1">
                        <div className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
                        <TypographyMuted className="text-xs">Core {core.core_id}</TypographyMuted>
                      </div>
                      <TypographySmall className="text-sm font-bold">
                        {formatPercentage(core.usage)}%
                      </TypographySmall>
                    </div>
                  );
                })}
              </div>
            )}
          </>
        ) : (
          <div className="text-center py-8">
            <TypographyMuted className="text-sm">
              {t('dashboard.cpu.noData') || 'No CPU data available'}
            </TypographyMuted>
          </div>
        )}
      </div>
    </SystemMetricCard>
  );
};

export default CPUUsageCard;

function CPUUsageCardSkeletonContent() {
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
    defaultData: DEFAULT_METRICS.cpu,
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
