import { CHART_COLORS } from './constants';
import { BarChartDataItem } from '@/components/ui/bar-chart-component';
import { DoughnutChartDataItem } from '@/components/ui/doughnut-chart-component';

export const formatGB = (value: number): string => `${value.toFixed(2)}`;
export const formatPercentage = (value: number): string => `${value.toFixed(1)}`;

export const createLoadAverageChartData = (load: {
  oneMin: number;
  fiveMin: number;
  fifteenMin: number;
}): BarChartDataItem[] => [
  {
    name: '1 min',
    value: load.oneMin,
    fill: CHART_COLORS.blue,
  },
  {
    name: '5 min',
    value: load.fiveMin,
    fill: CHART_COLORS.green,
  },
  {
    name: '15 min',
    value: load.fifteenMin,
    fill: CHART_COLORS.orange,
  },
];

export const createLoadAverageChartConfig = () => ({
  oneMin: {
    label: '1 min',
    color: CHART_COLORS.blue,
  },
  fiveMin: {
    label: '5 min',
    color: CHART_COLORS.green,
  },
  fifteenMin: {
    label: '15 min',
    color: CHART_COLORS.orange,
  },
});

export const createMemoryChartData = (
  used: number,
  free: number
): DoughnutChartDataItem[] => [
  {
    name: 'Used',
    value: used,
    fill: CHART_COLORS.blue,
  },
  {
    name: 'Free',
    value: free,
    fill: CHART_COLORS.green,
  },
];

export const createMemoryChartConfig = () => ({
  used: {
    label: 'Used Memory',
    color: CHART_COLORS.blue,
  },
  free: {
    label: 'Free Memory',
    color: CHART_COLORS.green,
  },
});

export const createCPUChartData = (
  perCore: Array<{ core_id: number; usage: number }>
): BarChartDataItem[] => {
  if (!perCore || perCore.length === 0) {
    return [];
  }

  const colors = [
    CHART_COLORS.blue,
    CHART_COLORS.green,
    CHART_COLORS.orange,
    CHART_COLORS.purple,
    CHART_COLORS.red,
    CHART_COLORS.yellow,
  ];

  return perCore.map((core) => ({
    name: `Core ${core.core_id}`,
    value: core.usage,
    fill: colors[core.core_id % colors.length],
  }));
};

export const createCPUChartConfig = (coreCount: number) => {
  const config: Record<string, { label: string; color: string }> = {};
  const colors = [
    CHART_COLORS.blue,
    CHART_COLORS.green,
    CHART_COLORS.orange,
    CHART_COLORS.purple,
    CHART_COLORS.red,
    CHART_COLORS.yellow,
  ];

  for (let i = 0; i < coreCount; i++) {
    config[`core${i}`] = {
      label: `Core ${i}`,
      color: colors[i % colors.length],
    };
  }

  return config;
};
