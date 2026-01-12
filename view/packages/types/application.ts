import { ApplicationDeployment, Status, Application } from '@/redux/types/applications';
import { DashboardItem } from '@/packages/types/layout';

export interface InfoLineProps {
  icon: React.ElementType;
  label: string;
  value: string;
  displayValue?: string;
  sublabel?: string;
  mono?: boolean;
  copyable?: boolean;
}

export interface StatBlockProps {
  value: string | number;
  label: string;
  sublabel?: string;
  color?: 'emerald' | 'red' | 'amber' | 'blue' | 'purple';
  pulse?: boolean;
}

export interface StatusIndicatorProps {
  status?: Status;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

export interface SectionLabelProps {
  children: React.ReactNode;
}

export interface LatestDeploymentProps {
  deployment?: ApplicationDeployment;
}

export interface DeploymentHealthChartProps {
  deploymentsByStatus: Record<string, number>;
  totalDeployments: number;
  successRate: number;
}

export interface ProjectFamilySwitcherProps {
  application: Application;
}

export interface DeploymentOverviewProps {
  totalDeployments: number;
  successfulDeployments: number;
  failedDeployments: number;
  currentStatus?: string;
}

export interface MonitorProps {
  application?: Application;
}

export interface ApplicationLogsProps {
  id: string;
  currentPage?: number;
  setCurrentPage?: (page: number) => void;
}

export interface DuplicateProjectDialogProps {
  application: Application;
}

export interface UseMonitoringReturn {
  showDragHint: boolean;
  mounted: boolean;
  layoutResetKey: number;
  hasCustomLayout: boolean;
  dismissHint: () => void;
  handleResetLayout: () => void;
  handleLayoutChange: () => void;
  monitoringItems: Array<{
    id: string;
    component: React.ReactNode;
    className?: string;
    isDefault: boolean;
  }>;
}

export interface DeploymentsListProps {
  deployments?: ApplicationDeployment[];
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}
