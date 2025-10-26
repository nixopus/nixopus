import { CHART_COLORS } from './constants';
import { BarChartDataItem } from '@/components/ui/bar-chart-component';
import { DoughnutChartDataItem } from '@/components/ui/doughnut-chart-component';

// Format utilities
export const formatGB = (value: number): string => `${value.toFixed(2)}`;
export const formatPercentage = (value: number): string => `${value.toFixed(1)}`;

// Chart data transformers
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
