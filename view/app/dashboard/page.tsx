'use client';

import React from 'react';
import { useDashboard } from './hooks/use-dashboard';
import ContainersWidget from './components/containers/containers-widget';
import SystemInfoCard from './components/system/system-info';
import LoadAverageCard from './components/system/load-average';
import CPUUsageCard from './components/system/cpu-usage';
import MemoryUsageCard from './components/system/memory-usage';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, ArrowRight, RefreshCw, Info } from 'lucide-react';
import DiskUsageCard from './components/system/disk-usage';
import ClockWidget from './components/system/clock';
import NetworkWidget from './components/system/network';
// TODO: Add weather widget back in with configuration for api key
// import WeatherWidget from './components/system/weather';
import { useTranslation } from '@/hooks/use-translation';
// TODO: Re-enable SMTP banner when notifications feature is working
// import { SMTPBanner } from './components/smtp-banner';
import DisabledFeature from '@/components/features/disabled-feature';
import { Button } from '@/components/ui/button';
import { useRouter } from 'next/navigation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { TypographyH1, TypographyMuted, TypographySmall } from '@/components/ui/typography';
import { Skeleton } from '@/components/ui/skeleton';
import {
  SystemInfoCardSkeleton,
  CPUUsageCardSkeleton,
  LoadAverageCardSkeleton,
  MemoryUsageCardSkeleton,
  DiskUsageCardSkeleton
} from './components/system/skeletons';
import { ContainersWidgetSkeleton } from './components/containers/containers-widget-skeleton';
import PageLayout from '@/components/layout/page-layout';
import { DraggableGrid, DraggableItem } from '@/components/ui/draggable-grid';
import { WidgetSelector } from './components/widget-selector';

// for dashboard page, we need to check if the user has the dashboard:read permission
function DashboardPage() {
  const { t } = useTranslation();
  const {
    isFeatureFlagsLoading,
    isDashboardEnabled,
    containersData,
    systemStats,
    // TODO: Re-enable when SMTP banner is working
    // smtpConfig,
    showDragHint,
    mounted,
    layoutResetKey,
    dismissHint,
    handleResetLayout,
    hasCustomLayout,
    handleLayoutChange
  } = useDashboard();

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

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isDashboardEnabled) {
    return <DisabledFeature />;
  }

  const allWidgetLabels = [
    { id: 'system-info', label: 'System Information' },
    { id: 'clock', label: 'Clock' },
    { id: 'network', label: 'Network Traffic' },
    { id: 'load-average', label: 'Load Average' },
    { id: 'cpu-usage', label: 'CPU Usage' },
    { id: 'memory-usage', label: 'Memory Usage' },
    { id: 'disk-usage', label: 'Disk Usage' },
    { id: 'containers', label: 'Containers' }
  ];

  const availableWidgets = allWidgetLabels.filter((widget) => hiddenWidgets.includes(widget.id));

  return (
    <ResourceGuard resource="dashboard" action="read">
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <DashboardHeader
          hasCustomLayout={hasCustomLayout}
          onResetLayout={handleResetLayoutWithWidgets}
          title={t('dashboard.title')}
          description={t('dashboard.description')}
          onAddWidget={handleAddWidget}
          availableWidgets={availableWidgets}
        />
        <DragHintBanner mounted={mounted} showDragHint={showDragHint} onDismiss={dismissHint} />
        {/* TODO: Re-enable SMTP banner when notifications feature is working */}
        {/* <SMTPBannerConditional hasSMTPConfig={!!smtpConfig} /> */}
        <MonitoringSection
          systemStats={systemStats}
          containersData={containersData}
          t={t}
          layoutResetKey={layoutResetKey}
          onLayoutChange={handleLayoutChange}
          onDeleteWidget={handleDeleteWidget}
          hiddenWidgets={hiddenWidgets}
        />
      </PageLayout>
    </ResourceGuard>
  );
}

export default DashboardPage;

const DashboardHeader = ({
  hasCustomLayout,
  onResetLayout,
  title,
  description,
  onAddWidget,
  availableWidgets
}: {
  hasCustomLayout: boolean;
  onResetLayout: () => void;
  title: string;
  description: string;
  onAddWidget: (widgetId: string) => void;
  availableWidgets: Array<{ id: string; label: string }>;
}) => (
  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2 mb-4">
    <div>
      <TypographyH1>{title}</TypographyH1>
      <TypographyMuted>{description}</TypographyMuted>
    </div>
    <div className="flex items-center gap-2">
      <WidgetSelector availableWidgets={availableWidgets} onAddWidget={onAddWidget} />
      {hasCustomLayout && (
        <Button variant="outline" size="sm" onClick={onResetLayout} className="shrink-0">
          <RefreshCw className="mr-2 h-4 w-4" />
          Reset Layout
        </Button>
      )}
    </div>
  </div>
);

const DragHintBanner = ({
  mounted,
  showDragHint,
  onDismiss
}: {
  mounted: boolean;
  showDragHint: boolean;
  onDismiss: () => void;
}) => {
  if (!mounted || !showDragHint) return null;

  return (
    <div className="mb-4 p-4 bg-primary/5 border border-primary/20 rounded-lg flex items-start justify-between gap-4">
      <div className="flex items-start gap-3">
        <div className="mt-0.5 text-primary">
          <Info className="h-5 w-5" />
        </div>
        <div className="flex-1">
          <p className="text-sm font-medium text-foreground">Customize Your Dashboard</p>
          <p className="text-xs text-muted-foreground mt-1">
            Hover over any card to see the drag handle on the left. Click and drag to rearrange
            cards in your preferred order. Your layout will be saved automatically.
          </p>
        </div>
      </div>
      <Button variant="ghost" size="default" onClick={onDismiss} className="shrink-0">
        Got it
      </Button>
    </div>
  );
};

// TODO: Re-enable SMTP banner when notifications feature is working
// const SMTPBannerConditional = ({ hasSMTPConfig }: { hasSMTPConfig: boolean }) => {
//   if (hasSMTPConfig) return null;
//   return <SMTPBanner />;
// };

const MonitoringSection = ({
  systemStats,
  containersData,
  t,
  layoutResetKey,
  onLayoutChange,
  onDeleteWidget,
  hiddenWidgets
}: {
  systemStats: any;
  containersData: any;
  t: any;
  layoutResetKey: number;
  onLayoutChange: () => void;
  onDeleteWidget: (id: string) => void;
  hiddenWidgets: string[];
}) => {
  const router = useRouter();

  if (!systemStats) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="md:col-span-2">
          <SystemInfoCardSkeleton />
        </div>
        <LoadAverageCardSkeleton />
        <CPUUsageCardSkeleton />
        <MemoryUsageCardSkeleton />
        <DiskUsageCardSkeleton />
        <div className="md:col-span-2">
          <ContainersWidgetSkeleton />
        </div>
      </div>
    );
  }

  const allWidgetDefinitions = [
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
      component: <ContainersWidget containersData={containersData} />,
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
    className: w.className
  }));

  return (
    <DraggableGrid
      items={dashboardItems}
      storageKey="dashboard-card-order"
      gridCols="grid-cols-1 md:grid-cols-2"
      resetKey={layoutResetKey}
      onReorder={() => onLayoutChange()}
      onDelete={onDeleteWidget}
    />
  );
};
