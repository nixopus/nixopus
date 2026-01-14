'use client';

import React from 'react';
import { ApplicationDeployment } from '@/redux/types/applications';
import { CHART_COLORS } from '@/packages/utils/dashboard';
import type { ChartConfig } from '@/components/ui/chart';
import { getDeploymentStatusColor } from '@/packages/utils/colors';
import { GitBranch } from 'lucide-react';

export interface DeploymentStats {
  total: number;
  deployed: number;
  failed: number;
  inProgress: number;
  successRate: number;
}

export interface PieDataItem {
  status: string;
  count: number;
  fill: string;
}

export function useDeploymentStats(deploymentsData: ApplicationDeployment[]) {
  const calculateStats = React.useCallback((): DeploymentStats => {
    if (!deploymentsData || deploymentsData.length === 0) {
      return {
        total: 0,
        deployed: 0,
        failed: 0,
        inProgress: 0,
        successRate: 0
      };
    }

    const deployed = deploymentsData.filter(
      (d) => d.status?.status?.toLowerCase() === 'deployed'
    ).length;
    const failed = deploymentsData.filter(
      (d) => d.status?.status?.toLowerCase() === 'failed'
    ).length;
    const inProgress = deploymentsData.filter((d) => {
      const status = d.status?.status?.toLowerCase() || '';
      return status === 'in_progress' || status === 'building' || status === 'deploying';
    }).length;

    const total = deploymentsData.length;
    const completed = deployed + failed;
    const successRate = completed > 0 ? (deployed / completed) * 100 : 0;

    return {
      total,
      deployed,
      failed,
      inProgress,
      successRate: Math.round(successRate)
    };
  }, [deploymentsData]);

  const stats = React.useMemo(() => calculateStats(), [calculateStats]);

  const pieData = React.useMemo<PieDataItem[]>(() => {
    return [
      {
        status: 'deployed',
        count: stats.deployed,
        fill: 'var(--color-deployed)'
      },
      {
        status: 'failed',
        count: stats.failed,
        fill: 'var(--color-failed)'
      },
      {
        status: 'inProgress',
        count: stats.inProgress,
        fill: 'var(--color-inProgress)'
      }
    ].filter((item) => item.count > 0);
  }, [stats]);

  const chartConfig = React.useMemo<ChartConfig>(() => {
    return {
      deployed: {
        label: 'Deployed',
        color: CHART_COLORS.green
      },
      failed: {
        label: 'Failed',
        color: CHART_COLORS.red || '#ef4444'
      },
      inProgress: {
        label: 'In Progress',
        color: CHART_COLORS.blue
      }
    };
  }, []);

  const [activeStatus, setActiveStatus] = React.useState<string | null>(
    pieData.length > 0 ? pieData[0].status : null
  );

  React.useEffect(() => {
    if (
      pieData.length > 0 &&
      (!activeStatus || !pieData.find((item) => item.status === activeStatus))
    ) {
      setActiveStatus(pieData[0].status);
    }
  }, [pieData, activeStatus]);

  const activeIndex = React.useMemo(
    () => pieData.findIndex((item) => item.status === activeStatus),
    [activeStatus, pieData]
  );

  const activeData = React.useMemo(() => {
    if (activeStatus && activeIndex >= 0) {
      return pieData[activeIndex];
    }
    return null;
  }, [activeStatus, activeIndex, pieData]);

  return {
    stats,
    pieData,
    chartConfig,
    activeStatus,
    activeIndex,
    activeData,
    setActiveStatus
  };
}

export type MetadataItem =
  | { icon: typeof GitBranch; content: string; className: string }
  | { content: string; className: string };

export interface DeploymentItem {
  deployment: ApplicationDeployment;
  status: string;
  statusColor: string;
  isLast: boolean;
  applicationName: string;
  metadataItems: MetadataItem[];
}

export function useDeploymentsWidget(deploymentsData: ApplicationDeployment[]) {
  const deploymentItems = React.useMemo<DeploymentItem[]>(() => {
    return deploymentsData.map((deployment, index) => {
      const status = String(deployment.status?.status || 'unknown');
      const statusColor = getDeploymentStatusColor(status);
      const isLast = index === deploymentsData.length - 1;

      const metadataItems: MetadataItem[] = [];
      if (deployment.commit_hash) {
        metadataItems.push({
          icon: GitBranch,
          content: deployment.commit_hash.substring(0, 7),
          className: 'font-mono truncate max-w-[80px]'
        });
      }
      if (deployment.container_name) {
        metadataItems.push({
          content: deployment.container_name,
          className: 'truncate max-w-[120px]'
        });
      }

      return {
        deployment,
        status,
        statusColor,
        isLast,
        applicationName: deployment.application?.name || 'Unknown Application',
        metadataItems
      };
    });
  }, [deploymentsData]);

  const isEmpty = React.useMemo(() => {
    return !deploymentsData || deploymentsData.length === 0;
  }, [deploymentsData]);

  return {
    deploymentItems,
    isEmpty
  };
}
