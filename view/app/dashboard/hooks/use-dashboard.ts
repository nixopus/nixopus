'use client';

import React from 'react';
import { useFeatureFlags } from '@/hooks/features_provider';
import { useAppSelector } from '@/redux/hooks';
import { useGetSMTPConfigurationsQuery } from '@/redux/services/settings/notificationApi';
import { useCheckForUpdatesQuery } from '@/redux/services/users/userApi';
import { FeatureNames } from '@/types/feature-flags';
import useMonitor from './use-monitor';
import { useGetContainersQuery, Container } from '@/redux/services/container/containerApi';
import { ContainerData } from '@/redux/types/monitor';

function transformContainerToContainerData(container: Container): ContainerData {
  return {
    Id: container.id,
    Names: [container.name],
    Image: container.image,
    ImageID: '',
    Command: container.command || '',
    Created: container.created ? Math.floor(new Date(container.created).getTime() / 1000) : 0,
    Ports: container.ports.map((p) => ({
      IP: '',
      PrivatePort: p.private_port,
      PublicPort: p.public_port,
      Type: p.type
    })),
    Labels: {},
    State: container.state,
    Status: container.status,
    HostConfig: {
      NetworkMode: container.host_config?.memory ? 'default' : undefined,
      Annotations: {}
    }
  };
}

export function useDashboard() {
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
  const { containersData: wsContainersData, systemStats } = useMonitor();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);

  const { data: containersApiData } = useGetContainersQuery(
    { page: 1, page_size: 4, search: '', sort_by: 'name', sort_order: 'asc' },
    { refetchOnMountOrArgChange: true }
  );

  const containersData = React.useMemo(() => {
    if (wsContainersData && wsContainersData.length > 0) {
      return wsContainersData;
    }
    if (containersApiData?.containers) {
      return containersApiData.containers.slice(0, 4).map(transformContainerToContainerData);
    }
    return [];
  }, [wsContainersData, containersApiData]);

  const { data: smtpConfig } = useGetSMTPConfigurationsQuery(activeOrganization?.id, {
    skip: !activeOrganization
  });

  // Check for updates on dashboard load and auto update if user has auto_update enabled
  useCheckForUpdatesQuery(undefined, { skip: !isAuthenticated });

  const [showDragHint, setShowDragHint] = React.useState(false);
  const [mounted, setMounted] = React.useState(false);
  const [layoutResetKey, setLayoutResetKey] = React.useState(0);
  const [hasCustomLayout, setHasCustomLayout] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
    // Check if user has seen the hint before
    const hasSeenHint = localStorage.getItem('dashboard-drag-hint-seen');
    if (!hasSeenHint) {
      setShowDragHint(true);
    }

    // Check if layout has been modified
    const savedOrder = localStorage.getItem('dashboard-card-order');
    setHasCustomLayout(!!savedOrder);
  }, []);

  const dismissHint = React.useCallback(() => {
    setShowDragHint(false);
    localStorage.setItem('dashboard-drag-hint-seen', 'true');
  }, []);

  const handleResetLayout = React.useCallback(() => {
    localStorage.removeItem('dashboard-card-order');
    setLayoutResetKey((prev) => prev + 1);
    setHasCustomLayout(false);
  }, []);

  const handleLayoutChange = React.useCallback(() => {
    const savedOrder = localStorage.getItem('dashboard-card-order');
    setHasCustomLayout(!!savedOrder);
  }, []);

  const isDashboardEnabled = React.useMemo(() => {
    return isFeatureEnabled(FeatureNames.FeatureMonitoring);
  }, [isFeatureEnabled]);

  return {
    // Feature flags
    isFeatureFlagsLoading,
    isDashboardEnabled,

    // Data
    containersData,
    systemStats,
    smtpConfig,

    // UI state
    showDragHint,
    mounted,
    layoutResetKey,
    hasCustomLayout,

    // Actions
    dismissHint,
    handleResetLayout,
    handleLayoutChange
  };
}
