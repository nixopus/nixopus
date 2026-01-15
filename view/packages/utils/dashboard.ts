import { BarChartDataItem } from '@/components/ui/bar-chart-component';
import { DoughnutChartDataItem } from '@/components/ui/doughnut-chart-component';

// Color constants for charts and visualizations
export const CHART_COLORS = {
  blue: '#3b82f6',
  green: '#10b981',
  orange: '#f59e0b',
  red: '#ef4444',
  purple: '#a855f7',
  yellow: '#eab308'
};

// Default values for system metrics
export const DEFAULT_METRICS = {
  load: {
    oneMin: 0 as number,
    fiveMin: 0 as number,
    fifteenMin: 0 as number
  },
  cpu: {
    overall: 0 as number,
    per_core: [] as Array<{
      core_id: number;
      usage: number;
    }>
  },
  memory: {
    total: 0 as number,
    used: 0 as number,
    percentage: 0 as number
  },
  disk: {
    percentage: 0 as number,
    used: 0 as number,
    total: 0 as number,
    allMounts: [] as any[]
  }
};

export const getStatusColor = (status: string) => {
  if (status?.toLowerCase().includes('running')) return 'bg-green-100 text-green-800 rounded-full';

  if (status?.toLowerCase().includes('exited')) return 'bg-red-100 text-red-800 rounded-full';

  if (status?.toLowerCase().includes('created')) return 'bg-blue-100 text-blue-800 rounded-full';

  return 'bg-gray-100 text-gray-800 rounded-full';
};

export const truncateId = (id: string) => {
  return id?.substring(0, 12) || '';
};

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
    fill: CHART_COLORS.blue
  },
  {
    name: '5 min',
    value: load.fiveMin,
    fill: CHART_COLORS.green
  },
  {
    name: '15 min',
    value: load.fifteenMin,
    fill: CHART_COLORS.orange
  }
];

export const createLoadAverageChartConfig = () => ({
  oneMin: {
    label: '1 min',
    color: CHART_COLORS.blue
  },
  fiveMin: {
    label: '5 min',
    color: CHART_COLORS.green
  },
  fifteenMin: {
    label: '15 min',
    color: CHART_COLORS.orange
  }
});

export const createMemoryChartData = (used: number, free: number): DoughnutChartDataItem[] => [
  {
    name: 'Used',
    value: used,
    fill: CHART_COLORS.blue
  },
  {
    name: 'Free',
    value: free,
    fill: CHART_COLORS.green
  }
];

export const createMemoryChartConfig = () => ({
  used: {
    label: 'Used Memory',
    color: CHART_COLORS.blue
  },
  free: {
    label: 'Free Memory',
    color: CHART_COLORS.green
  }
});

export const createCPUChartData = (
  perCore: Array<{
    core_id: number;
    usage: number;
  }>
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
    CHART_COLORS.yellow
  ];

  return perCore.map((core) => ({
    name: `Core ${core.core_id}`,
    value: core.usage,
    fill: colors[core.core_id % colors.length]
  }));
};

export const createCPUChartConfig = (coreCount: number) => {
  const config: Record<
    string,
    {
      label: string;
      color: string;
    }
  > = {};
  const colors = [
    CHART_COLORS.blue,
    CHART_COLORS.green,
    CHART_COLORS.orange,
    CHART_COLORS.purple,
    CHART_COLORS.red,
    CHART_COLORS.yellow
  ];

  for (let i = 0; i < coreCount; i++) {
    config[`core${i}`] = {
      label: `Core ${i}`,
      color: colors[i % colors.length]
    };
  }

  return config;
};
