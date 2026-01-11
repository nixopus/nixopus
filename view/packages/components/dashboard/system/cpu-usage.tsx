'use client';

import React from 'react';
import { Cpu } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { BarChartComponent } from '@/components/ui/bar-chart-component';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '@/packages/hooks/dashboard/use-system-metric';
import { createCPUChartData, createCPUChartConfig, formatPercentage } from '../utils/utils';
import { DEFAULT_METRICS, CHART_COLORS } from '../utils/constants';
import { CPUUsageCardSkeletonContent } from './skeletons/cpu-usage';

interface CPUUsageCardProps {
  systemStats: SystemStatsType | null;
}

interface CPUUsageHeaderProps {
  overallUsage: number;
  label: string;
}

interface CPUUsageChartProps {
  chartData: ReturnType<typeof createCPUChartData>;
  chartConfig: ReturnType<typeof createCPUChartConfig>;
  yAxisLabel: string;
  xAxisLabel: string;
}

interface TopCoresListProps {
  cores: Array<{ core_id: number; usage: number }>;
}

interface CoreItemProps {
  coreId: number;
  usage: number;
  color: string;
}

const CPU_COLORS = [
  CHART_COLORS.blue,
  CHART_COLORS.green,
  CHART_COLORS.orange,
  CHART_COLORS.purple,
  CHART_COLORS.red,
  CHART_COLORS.yellow
];

const CPUUsageHeader: React.FC<CPUUsageHeaderProps> = ({ overallUsage, label }) => {
  return (
    <div className="text-center">
      <TypographyMuted className="text-xs">{label}</TypographyMuted>
      <div className="text-3xl font-bold text-primary mt-1">{formatPercentage(overallUsage)}%</div>
    </div>
  );
};

const CPUUsageChart: React.FC<CPUUsageChartProps> = ({
  chartData,
  chartConfig,
  yAxisLabel,
  xAxisLabel
}) => {
  return (
    <div>
      <BarChartComponent
        data={chartData}
        chartConfig={chartConfig}
        height="h-[180px]"
        yAxisLabel={yAxisLabel}
        xAxisLabel={xAxisLabel}
        showAxisLabels={true}
      />
    </div>
  );
};

const CoreItem: React.FC<CoreItemProps> = ({ coreId, usage, color }) => {
  return (
    <div className="flex flex-col items-center gap-1">
      <div className="flex items-center gap-1">
        <div className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
        <TypographyMuted className="text-xs">Core {coreId}</TypographyMuted>
      </div>
      <TypographySmall className="text-sm font-bold">{formatPercentage(usage)}%</TypographySmall>
    </div>
  );
};

const TopCoresList: React.FC<TopCoresListProps> = ({ cores }) => {
  return (
    <div className="grid grid-cols-3 gap-2 text-center">
      {cores.map((core) => {
        const color = CPU_COLORS[core.core_id % CPU_COLORS.length];
        return (
          <CoreItem key={core.core_id} coreId={core.core_id} usage={core.usage} color={color} />
        );
      })}
    </div>
  );
};

const CPUUsageCard: React.FC<CPUUsageCardProps> = ({ systemStats }) => {
  const {
    data: cpu,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.cpu,
    defaultData: DEFAULT_METRICS.cpu
  });

  const perCoreData = cpu.per_core;
  const chartData = createCPUChartData(perCoreData);
  const chartConfig = createCPUChartConfig(perCoreData.length);
  const topCores = [...perCoreData].sort((a, b) => b.usage - a.usage).slice(0, 3);

  return (
    <SystemMetricCard
      title={t('dashboard.cpu.title')}
      icon={Cpu}
      isLoading={isLoading}
      skeletonContent={<CPUUsageCardSkeletonContent />}
    >
      <div className="space-y-4">
        <CPUUsageHeader overallUsage={cpu.overall} label={t('dashboard.cpu.overall')} />

        <CPUUsageChart
          chartData={chartData}
          chartConfig={chartConfig}
          yAxisLabel={t('dashboard.cpu.usage')}
          xAxisLabel={t('dashboard.cpu.cores')}
        />

        <TopCoresList cores={topCores} />
      </div>
    </SystemMetricCard>
  );
};

export default CPUUsageCard;
