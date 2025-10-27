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
