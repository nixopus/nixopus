'use client';

import React from 'react';
import { useFeatureFlags } from '@/hooks/features_provider';
import { useAppSelector } from '@/redux/hooks';
import { useGetSMTPConfigurationsQuery } from '@/redux/services/settings/notificationApi';
import { FeatureNames } from '@/types/feature-flags';
import useMonitor from './use-monitor';

export function useDashboard() {
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
  const { containersData, systemStats } = useMonitor();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);

  const { data: smtpConfig } = useGetSMTPConfigurationsQuery(activeOrganization?.id, {
    skip: !activeOrganization
  });

  const [showDragHint, setShowDragHint] = React.useState(false);
  const [mounted, setMounted] = React.useState(false);
  const [layoutResetKey, setLayoutResetKey] = React.useState(0);

  React.useEffect(() => {
    setMounted(true);
    // Check if user has seen the hint before
    const hasSeenHint = localStorage.getItem('dashboard-drag-hint-seen');
    if (!hasSeenHint) {
      setShowDragHint(true);
    }
  }, []);

  const dismissHint = React.useCallback(() => {
    setShowDragHint(false);
    localStorage.setItem('dashboard-drag-hint-seen', 'true');
  }, []);

  const handleResetLayout = React.useCallback(() => {
    localStorage.removeItem('dashboard-card-order');
    setLayoutResetKey((prev) => prev + 1);
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

    // Actions
    dismissHint,
    handleResetLayout
  };
}
