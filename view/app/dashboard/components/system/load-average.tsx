'use client';

import React from 'react';
import { Activity } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { BarChartComponent } from '@/components/ui/bar-chart-component';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '../../hooks/use-system-metric';
import { createLoadAverageChartData, createLoadAverageChartConfig } from '../utils/utils';
import { DEFAULT_METRICS, CHART_COLORS } from '../utils/constants';
import { LoadAverageCardSkeletonContent } from './skeletons/load-average';

interface LoadAverageCardProps {
  systemStats: SystemStatsType | null;
}

const LoadAverageCard: React.FC<LoadAverageCardProps> = ({ systemStats }) => {
  const { data: load, isLoading, t } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.load,
    defaultData: DEFAULT_METRICS.load,
  });

  const chartData = createLoadAverageChartData(load);
  const chartConfig = createLoadAverageChartConfig();

  return (
    <SystemMetricCard
      title={t('dashboard.load.title')}
      icon={Activity}
      isLoading={isLoading}
      skeletonContent={<LoadAverageCardSkeletonContent />}
    >
      <br /><br /><br />
      <div className="space-y-4">
        <div>
          <BarChartComponent
            data={chartData}
            chartConfig={chartConfig}
            height="h-[180px]"
            yAxisLabel="Load"
            xAxisLabel="Time Period"
            showAxisLabels={true}
          />
        </div>

        {/* Summary Stats with Color Indicators */}
        <div className="grid grid-cols-3 gap-2 text-center">
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: CHART_COLORS.blue }} />
              <TypographyMuted className="text-xs">1 min</TypographyMuted>
            </div>
            <TypographySmall className="text-sm font-bold">{load.oneMin.toFixed(2)}</TypographySmall>
          </div>
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: CHART_COLORS.green }} />
              <TypographyMuted className="text-xs">5 min</TypographyMuted>
            </div>
            <TypographySmall className="text-sm font-bold">{load.fiveMin.toFixed(2)}</TypographySmall>
          </div>
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: CHART_COLORS.orange }} />
              <TypographyMuted className="text-xs">15 min</TypographyMuted>
            </div>
            <TypographySmall className="text-sm font-bold">{load.fifteenMin.toFixed(2)}</TypographySmall>
          </div>
        </div>
      </div>
    </SystemMetricCard>
  );
};

export default LoadAverageCard;
