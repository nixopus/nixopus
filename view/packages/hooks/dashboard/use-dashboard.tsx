import { useAppSelector } from '@/redux/hooks';
import React from 'react';
import { useFeatureFlags } from '@/packages/hooks/shared/features_provider';
import useMonitor from '@/packages/hooks/dashboard/use-monitor';
import { useGetSMTPConfigurationsQuery } from '@/redux/services/settings/notificationApi';
import { useCheckForUpdatesQuery } from '@/redux/services/users/userApi';
import { FeatureNames } from '@/packages/types/feature-flags';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { DashboardItem } from '@/packages/types/layout';
import { useContainer } from '@/packages/hooks/dashboard/use-container';
import {
  ClockWidget,
  ContainersWidget,
  NetworkWidget,
  SystemInfoCard
} from '@/packages/components/dashboard';
import { LoadAverageCard } from '@/packages/components/dashboard';
import { CPUUsageCard } from '@/packages/components/dashboard';
import { MemoryUsageCard } from '@/packages/components/dashboard';
import { DiskUsageCard } from '@/packages/components/dashboard';

export const useDashboard = () => {
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
  const { containersData, systemStats } = useMonitor();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const { t } = useTranslation();
  const columns = useContainer();
  const { data: smtpConfig } = useGetSMTPConfigurationsQuery(activeOrganization?.id, {
    skip: !activeOrganization
  });

  // Check for updates on dashboard load and auto update if user has auto_update enabled
  useCheckForUpdatesQuery(undefined, {
    skip: !isAuthenticated
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

  const defaultHiddenWidgets = ['clock', 'network'];

  const [hiddenWidgets, setHiddenWidgets] = React.useState<string[]>(defaultHiddenWidgets);

  React.useEffect(() => {
    const saved = localStorage.getItem('dashboard-hidden-widgets');
    if (saved) {
      setHiddenWidgets(JSON.parse(saved));
    } else {
      setHiddenWidgets(defaultHiddenWidgets);
    }
  }, []);

  const handleDeleteWidget = (widgetId: string) => {
    const newHidden = [...hiddenWidgets, widgetId];
    setHiddenWidgets(newHidden);
    localStorage.setItem('dashboard-hidden-widgets', JSON.stringify(newHidden));
  };

  const handleAddWidget = (widgetId: string) => {
    const newHidden = hiddenWidgets.filter((id) => id !== widgetId);
    setHiddenWidgets(newHidden);
    localStorage.setItem('dashboard-hidden-widgets', JSON.stringify(newHidden));
  };

  const handleResetLayoutWithWidgets = () => {
    setHiddenWidgets([]);
    localStorage.removeItem('dashboard-hidden-widgets');
    handleResetLayout();
  };

  const allWidgetLabels = [
    {
      id: 'system-info',
      label: 'System Information'
    },
    {
      id: 'clock',
      label: 'Clock'
    },
    {
      id: 'network',
      label: 'Network Traffic'
    },
    {
      id: 'load-average',
      label: 'Load Average'
    },
    {
      id: 'cpu-usage',
      label: 'CPU Usage'
    },
    {
      id: 'memory-usage',
      label: 'Memory Usage'
    },
    {
      id: 'disk-usage',
      label: 'Disk Usage'
    },
    {
      id: 'containers',
      label: 'Containers'
    }
  ];

  const availableWidgets = allWidgetLabels.filter((widget) => hiddenWidgets.includes(widget.id));

  const allWidgetDefinitions: DashboardItem[] = [
    {
      id: 'system-info',
      component: <SystemInfoCard systemStats={systemStats} />,
      className: 'md:col-span-2',
      isDefault: true
    },
    {
      id: 'clock',
      component: <ClockWidget />,
      isDefault: false
    },
    {
      id: 'network',
      component: <NetworkWidget systemStats={systemStats} />,
      isDefault: false
    },
    {
      id: 'load-average',
      component: <LoadAverageCard systemStats={systemStats} />,
      isDefault: true
    },
    {
      id: 'cpu-usage',
      component: <CPUUsageCard systemStats={systemStats} />,
      isDefault: true
    },
    {
      id: 'memory-usage',
      component: <MemoryUsageCard systemStats={systemStats} />,
      isDefault: true
    },
    {
      id: 'disk-usage',
      component: <DiskUsageCard systemStats={systemStats} />,
      isDefault: true
    },
    {
      id: 'containers',
      component: <ContainersWidget containersData={containersData} columns={columns} />,
      className: 'md:col-span-2',
      isDefault: true
    }
  ];

  const visibleItems = allWidgetDefinitions.filter((widget) => {
    return !hiddenWidgets.includes(widget.id);
  });

  const dashboardItems = visibleItems.map((w) => ({
    id: w.id,
    component: w.component,
    className: w.className,
    isDefault: w.isDefault
  }));

  return {
    allWidgetLabels,
    availableWidgets,
    hiddenWidgets,
    setHiddenWidgets,
    handleDeleteWidget,
    handleAddWidget,
    handleResetLayoutWithWidgets,
    isFeatureFlagsLoading,
    isDashboardEnabled,
    containersData,
    systemStats,
    smtpConfig,
    showDragHint,
    mounted,
    layoutResetKey,
    hasCustomLayout,
    dismissHint,
    handleResetLayout,
    handleLayoutChange,
    t,
    dashboardItems
  };
};

export default useDashboard;
