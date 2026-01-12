import { useMemo } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { CircleCheck, CircleAlert, CircleX } from 'lucide-react';
import { cn } from '@/lib/utils';

interface UseDeploymentHealthChartProps {
  deploymentsByStatus: Record<string, number>;
  totalDeployments: number;
  successRate: number;
}

const COLORS = {
  deployed: '#10b981',
  building: '#3b82f6',
  deploying: '#f59e0b',
  cloning: '#8b5cf6',
  failed: '#ef4444'
};

const STATUS_LABELS: Record<string, string> = {
  deployed: 'Deployed',
  building: 'Building',
  deploying: 'Deploying',
  cloning: 'Cloning',
  failed: 'Failed'
};

export function useDeploymentHealthChart({
  deploymentsByStatus,
  totalDeployments,
  successRate
}: UseDeploymentHealthChartProps) {
  const { t } = useTranslation();

  const chartData = useMemo(() => {
    return Object.entries(deploymentsByStatus)
      .filter(([_, count]) => count > 0)
      .map(([status, count]) => ({
        name: STATUS_LABELS[status] || status,
        value: count,
        color: COLORS[status as keyof typeof COLORS] || '#6b7280'
      }));
  }, [deploymentsByStatus]);

  const statusIcon = useMemo(() => {
    if (successRate >= 80) return <CircleCheck className="h-4 w-4 text-emerald-500" />;
    if (successRate >= 50) return <CircleAlert className="h-4 w-4 text-amber-500" />;
    return <CircleX className="h-4 w-4 text-red-500" />;
  }, [successRate]);

  const healthStatus = useMemo(() => {
    if (successRate >= 80)
      return {
        text: t('selfHost.monitoring.chart.healthStatus.healthy'),
        color: 'text-emerald-500'
      };
    if (successRate >= 50)
      return { text: t('selfHost.monitoring.chart.healthStatus.warning'), color: 'text-amber-500' };
    return { text: t('selfHost.monitoring.chart.healthStatus.critical'), color: 'text-red-500' };
  }, [successRate, t]);

  const statusLegendItems = useMemo(() => {
    return Object.entries(deploymentsByStatus).map(([status, count]) => ({
      status,
      count,
      label: STATUS_LABELS[status] || status,
      color: COLORS[status as keyof typeof COLORS] || '#6b7280'
    }));
  }, [deploymentsByStatus]);

  const headerActions = useMemo(
    () => (
      <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-muted">
        {statusIcon}
        <span className={cn('text-sm font-medium', healthStatus.color)}>{healthStatus.text}</span>
      </div>
    ),
    [statusIcon, healthStatus]
  );

  const customHeader = useMemo(
    () => (
      <div className="flex items-center justify-between w-full">
        <span>{t('selfHost.monitoring.chart.title')}</span>
        {headerActions}
      </div>
    ),
    [t, headerActions]
  );

  const emptyStateContent = useMemo(
    () => (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <div className="h-32 w-32 rounded-full border-4 border-dashed border-muted flex items-center justify-center mb-4">
          <span className="text-3xl font-bold text-muted-foreground/50">0</span>
        </div>
        <p className="font-medium">{t('selfHost.monitoring.chart.noData')}</p>
        <p className="text-sm text-muted-foreground/60 mt-1">
          {t('selfHost.monitoring.chart.noDataDescription')}
        </p>
      </div>
    ),
    [t]
  );

  const tooltipContent = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-popover border rounded-lg shadow-lg px-3 py-2">
          <p className="font-medium">{data.name}</p>
          <p className="text-sm text-muted-foreground">
            {data.value} {t('selfHost.monitoring.chart.deployments')}
          </p>
        </div>
      );
    }
    return null;
  };

  return {
    chartData,
    statusIcon,
    healthStatus,
    statusLegendItems,
    customHeader,
    emptyStateContent,
    tooltipContent,
    hasData: totalDeployments > 0,
    successRate
  };
}
