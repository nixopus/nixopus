import { LucideIcon } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { CHART_COLORS } from '@/packages/utils/dashboard';
import { createCPUChartData, createCPUChartConfig } from '@/packages/utils/dashboard';

export interface AvailableWidget {
  id: string;
  label: string;
}

export interface WidgetSelectorProps {
  availableWidgets: AvailableWidget[];
  onAddWidget: (widgetId: string) => void;
}

export interface CPUUsageCardProps {
  systemStats: SystemStatsType | null;
}

export interface CPUUsageHeaderProps {
  overallUsage: number;
  label: string;
}

export interface CPUUsageChartProps {
  chartData: ReturnType<typeof createCPUChartData>;
  chartConfig: ReturnType<typeof createCPUChartConfig>;
  yAxisLabel: string;
  xAxisLabel: string;
}

export interface TopCoresListProps {
  cores: Array<{ core_id: number; usage: number }>;
}

export interface CoreItemProps {
  coreId: number;
  usage: number;
  color: string;
}

export const CPU_COLORS = [
  CHART_COLORS.blue,
  CHART_COLORS.green,
  CHART_COLORS.orange,
  CHART_COLORS.purple,
  CHART_COLORS.red,
  CHART_COLORS.yellow
];

export interface DiskUsageCardProps {
  systemStats: SystemStatsType | null;
}

export interface MountData {
  mountPoint: string;
  size: string;
  used: string;
  capacity: string;
}

export interface LoadAverageCardProps {
  systemStats: SystemStatsType | null;
}

export interface MemoryUsageCardProps {
  systemStats: SystemStatsType | null;
}

export interface NetworkWidgetProps {
  systemStats: SystemStatsType | null;
}

export interface SystemInfoCardProps {
  systemStats: SystemStatsType | null;
}

export interface SystemInfoItemProps {
  icon: React.ReactNode;
  label: string;
  value: string;
}

export interface SystemMetricCardProps {
  title: string;
  icon: LucideIcon;
  isLoading?: boolean;
  children: React.ReactNode;
  skeletonContent?: React.ReactNode;
}

export interface SystemStatsProps {
  systemStats: SystemStatsType | null;
}
