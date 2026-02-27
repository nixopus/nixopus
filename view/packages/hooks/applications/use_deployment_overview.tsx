import { useMemo } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

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

  return {
    statBlocks
  };
}
