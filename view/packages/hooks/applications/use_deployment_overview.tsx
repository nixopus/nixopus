import { useMemo } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Activity } from 'lucide-react';

interface UseDeploymentOverviewProps {
  totalDeployments: number;
  successfulDeployments: number;
  failedDeployments: number;
  currentStatus?: string;
}

const getStatusColor = (status?: string): 'emerald' | 'red' | 'amber' | 'blue' | 'purple' => {
  switch (status) {
    case 'deployed':
      return 'emerald';
    case 'failed':
      return 'red';
    case 'building':
      return 'amber';
    case 'deploying':
      return 'blue';
    case 'cloning':
      return 'purple';
    default:
      return 'amber';
  }
};

export function useDeploymentOverview({
  totalDeployments,
  successfulDeployments,
  failedDeployments,
  currentStatus
}: UseDeploymentOverviewProps) {
  const { t } = useTranslation();

  const isActive = currentStatus === 'deployed';

  const statBlocks = useMemo(
    () => [
      {
        key: 'status',
        value: currentStatus || t('selfHost.monitoring.overview.noDeployment'),
        label: t('selfHost.monitoring.overview.status'),
        color: getStatusColor(currentStatus),
        pulse: isActive
      },
      {
        key: 'total',
        value: totalDeployments,
        label: t('selfHost.monitoring.overview.totalDeployments'),
        sublabel: t('selfHost.monitoring.overview.allTime')
      },
      {
        key: 'successful',
        value: successfulDeployments,
        label: t('selfHost.monitoring.overview.successful'),
        color: 'emerald' as const
      },
      {
        key: 'failed',
        value: failedDeployments,
        label: t('selfHost.monitoring.overview.failed'),
        color: (failedDeployments > 0 ? 'red' : undefined) as 'red' | undefined
      }
    ],
    [totalDeployments, successfulDeployments, failedDeployments, currentStatus, isActive, t]
  );

  const title = useMemo(
    () => (
      <div className="flex items-center gap-2">
        <Activity className="h-5 w-5 text-muted-foreground" />
        <span>{t('selfHost.monitoring.overview.title')}</span>
      </div>
    ),
    [t]
  );

  return {
    title,
    statBlocks
  };
}
