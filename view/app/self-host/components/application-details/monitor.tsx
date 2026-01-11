'use client';

import { Application } from '@/redux/types/applications';
import React from 'react';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { useMonitoringData } from '../../hooks/use_monitoring_data';
import { DeploymentOverview, LatestDeployment } from './monitoring';
import { DeploymentHealthChart } from './monitoring/deployment-health-chart';
import { HealthCheckCard } from './monitoring/health-check-card';

interface MonitorProps {
  application?: Application;
}

function Monitor({ application }: MonitorProps) {
  const {
    totalDeployments,
    successfulDeployments,
    failedDeployments,
    currentStatus,
    latestDeployment,
    deploymentsByStatus,
    successRate
  } = useMonitoringData(application);

  if (!application) {
    return null;
  }

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <div className="space-y-8">
        <DeploymentOverview
          totalDeployments={totalDeployments}
          successfulDeployments={successfulDeployments}
          failedDeployments={failedDeployments}
          currentStatus={currentStatus}
        />

        <HealthCheckCard application={application} />

        <div className="grid grid-cols-1 lg:grid-cols-10 gap-6">
          <div className="lg:col-span-7">
            <DeploymentHealthChart
              deploymentsByStatus={deploymentsByStatus}
              totalDeployments={totalDeployments}
              successRate={successRate}
            />
          </div>
          <div className="lg:col-span-3">
            <LatestDeployment deployment={latestDeployment} />
          </div>
        </div>
      </div>
    </ResourceGuard>
  );
}

export default Monitor;
