import React from 'react';
import { Application } from '@/redux/types/applications';
import { useMonitoringData } from '@/packages/hooks/applications/use_monitoring_data';
import {
  DeploymentOverview,
  DeploymentHealthChart,
  LatestDeployment
} from '@/packages/components/application-details';
import { HealthCheckCard } from '@/packages/components/application-healthcheck';
import { DashboardItem } from '@/packages/types/layout';

interface UseMonitoringReturn {
  showDragHint: boolean;
  mounted: boolean;
  layoutResetKey: number;
  hasCustomLayout: boolean;
  dismissHint: () => void;
  handleResetLayout: () => void;
  handleLayoutChange: () => void;
  monitoringItems: DashboardItem[];
}

export function useMonitoring(application?: Application): UseMonitoringReturn {
  const {
    totalDeployments,
    successfulDeployments,
    failedDeployments,
    currentStatus,
    latestDeployment,
    deploymentsByStatus,
    successRate
  } = useMonitoringData(application);

  const [showDragHint, setShowDragHint] = React.useState(false);
  const [mounted, setMounted] = React.useState(false);
  const [layoutResetKey, setLayoutResetKey] = React.useState(0);
  const [hasCustomLayout, setHasCustomLayout] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
    // Check if user has seen the hint before
    const hasSeenHint = localStorage.getItem('monitoring-drag-hint-seen');
    if (!hasSeenHint) {
      setShowDragHint(true);
    }

    // Check if layout has been modified
    const savedOrder = localStorage.getItem('monitoring-card-order');
    setHasCustomLayout(!!savedOrder);
  }, []);

  const dismissHint = React.useCallback(() => {
    setShowDragHint(false);
    localStorage.setItem('monitoring-drag-hint-seen', 'true');
  }, []);

  const handleResetLayout = React.useCallback(() => {
    localStorage.removeItem('monitoring-card-order');
    setLayoutResetKey((prev) => prev + 1);
    setHasCustomLayout(false);
  }, []);

  const handleLayoutChange = React.useCallback(() => {
    const savedOrder = localStorage.getItem('monitoring-card-order');
    setHasCustomLayout(!!savedOrder);
  }, []);

  const allWidgetDefinitions: DashboardItem[] = React.useMemo(() => {
    if (!application) {
      return [];
    }

    return [
      {
        id: 'deployment-overview',
        component: (
          <DeploymentOverview
            totalDeployments={totalDeployments}
            successfulDeployments={successfulDeployments}
            failedDeployments={failedDeployments}
            currentStatus={currentStatus}
          />
        ),
        className: 'lg:col-span-2',
        isDefault: true
      },
      {
        id: 'health-check',
        component: <HealthCheckCard application={application} />,
        isDefault: true
      },
      {
        id: 'deployment-health-chart',
        component: (
          <DeploymentHealthChart
            deploymentsByStatus={deploymentsByStatus}
            totalDeployments={totalDeployments}
            successRate={successRate}
          />
        ),
        isDefault: true
      },
      {
        id: 'latest-deployment',
        component: <LatestDeployment deployment={latestDeployment} />,
        isDefault: true
      }
    ];
  }, [
    application,
    totalDeployments,
    successfulDeployments,
    failedDeployments,
    currentStatus,
    latestDeployment,
    deploymentsByStatus,
    successRate
  ]);

  const monitoringItems: DashboardItem[] = React.useMemo(
    () =>
      allWidgetDefinitions.map((w) => ({
        id: w.id,
        component: w.component,
        className: w.className,
        isDefault: w.isDefault
      })),
    [allWidgetDefinitions]
  );

  return {
    showDragHint,
    mounted,
    layoutResetKey,
    hasCustomLayout,
    dismissHint,
    handleResetLayout,
    handleLayoutChange,
    monitoringItems
  };
}
