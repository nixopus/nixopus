'use client';

import React from 'react';
import { BarChart } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { TypographyMuted } from '@/components/ui/typography';
import { DoughnutChartComponent } from '@/components/ui/doughnut-chart-component';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '../../hooks/use-system-metric';
import { formatGB, createMemoryChartData, createMemoryChartConfig } from '../utils/utils';
import { DEFAULT_METRICS, CHART_COLORS } from '../utils/constants';
import { MemoryUsageCardSkeletonContent } from './skeletons/memory-usage';

interface MemoryUsageCardProps {
  systemStats: SystemStatsType | null;
}

const MemoryUsageCard: React.FC<MemoryUsageCardProps> = ({ systemStats }) => {
  const {
    data: memory,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.memory,
    defaultData: DEFAULT_METRICS.memory
  });

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

        <div className="space-y-2">
          <div className="flex justify-between text-xs">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: CHART_COLORS.blue }} />
              <TypographyMuted>Used: {formatGB(memory.used)} GB</TypographyMuted>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: CHART_COLORS.green }} />
              <TypographyMuted>Free: {formatGB(freeMemory)} GB</TypographyMuted>
            </div>
          </div>

          <TypographyMuted className="text-xs text-center">
            Total: {formatGB(memory.total)} GB
          </TypographyMuted>
        </div>
      </div>
    </SystemMetricCard>
  );
};

export default MemoryUsageCard;
