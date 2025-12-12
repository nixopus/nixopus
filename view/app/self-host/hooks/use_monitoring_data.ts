import { useMemo } from 'react';
import { Application, ApplicationDeployment, Status } from '@/redux/types/applications';

interface MonitoringData {
  totalDeployments: number;
  successfulDeployments: number;
  failedDeployments: number;
  currentStatus?: Status;
  latestDeployment?: ApplicationDeployment;
  deploymentsByStatus: Record<Status, number>;
  successRate: number;
}

export function useMonitoringData(application?: Application): MonitoringData {
  return useMemo(() => {
    const deployments = application?.deployments || [];

    const deploymentsByStatus = deployments.reduce(
      (acc, deployment) => {
        const status = deployment.status?.status;
        if (status) {
          acc[status] = (acc[status] || 0) + 1;
        }
        return acc;
      },
      {} as Record<Status, number>
    );

    const totalDeployments = deployments.length;
    const successfulDeployments = deploymentsByStatus['deployed'] || 0;
    const failedDeployments = deploymentsByStatus['failed'] || 0;

    const latestDeployment = deployments.length > 0 ? deployments[0] : undefined;
    const currentStatus = latestDeployment?.status?.status;

    const successRate =
      totalDeployments > 0 ? Math.round((successfulDeployments / totalDeployments) * 100) : 0;

    return {
      totalDeployments,
      successfulDeployments,
      failedDeployments,
      currentStatus,
      latestDeployment,
      deploymentsByStatus,
      successRate
    };
  }, [application?.deployments]);
}
