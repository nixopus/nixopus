import { useMemo } from 'react';
import { ApplicationDeployment } from '@/redux/types/applications';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ChartConfig } from '@/components/ui/chart';

interface UseDeploymentStatusChartProps {
  deployments: ApplicationDeployment[];
}

export function useDeploymentStatusChart({ deployments }: UseDeploymentStatusChartProps) {
  const { t } = useTranslation();

  const statusCounts = useMemo(() => {
    const counts = {
      failed: 0,
      building: 0,
      deploying: 0,
      deployed: 0
    };

    deployments.forEach((deployment) => {
      if (deployment.status && deployment.status.status) {
        counts[deployment.status.status as keyof typeof counts]++;
      }
    });

    return [
      { status: t('selfHost.deployments.chart.status.building'), value: counts.building },
      { status: t('selfHost.deployments.chart.status.deployed'), value: counts.deployed },
      { status: t('selfHost.deployments.chart.status.deploying'), value: counts.deploying },
      { status: t('selfHost.deployments.chart.status.failed'), value: counts.failed }
    ];
  }, [deployments, t]);

  const statusChartConfig = {
    value: {
      label: 'Count',
      color: 'hsl(var(--color-desktop))'
    }
  } satisfies ChartConfig;

  return {
    statusCounts,
    statusChartConfig
  };
}
