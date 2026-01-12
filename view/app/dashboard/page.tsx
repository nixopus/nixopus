'use client';

import React from 'react';
import { useDashboard } from '@/packages/hooks/dashboard/use-dashboard';
import { RefreshCw } from 'lucide-react';
import DisabledFeature from '@/packages/components/rbac';
import { Button } from '@/components/ui/button';
import { Banner } from '@/components/ui/banner';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';
import {
  SystemInfoCardSkeleton,
  CPUUsageCardSkeleton,
  LoadAverageCardSkeleton,
  MemoryUsageCardSkeleton,
  DiskUsageCardSkeleton
} from '@/packages/components/dashboard-skeletons';
import { ContainersWidgetSkeleton } from '@/packages/components/container-skeletons';
import PageLayout from '@/packages/layouts/page-layout';
import { DraggableGrid } from '@/components/ui/draggable-grid';
import { WidgetSelector } from '@/packages/components/dashboard';
import { DashboardItem } from '@/packages/types/layout';
import { ContainerData, SystemStatsType } from '@/redux/types/monitor';
import MainPageHeader from '@/components/ui/main-page-header';

// for dashboard page, we need to check if the user has the dashboard:read permission
function DashboardPage() {
  const {
    availableWidgets,
    handleDeleteWidget,
    handleAddWidget,
    handleResetLayoutWithWidgets,
    isFeatureFlagsLoading,
    isDashboardEnabled,
    containersData,
    systemStats,
    showDragHint,
    mounted,
    layoutResetKey,
    hasCustomLayout,
    dismissHint,
    handleLayoutChange,
    t,
    dashboardItems
  } = useDashboard();

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isDashboardEnabled) {
    return <DisabledFeature />;
  }

  return (
    <ResourceGuard resource="dashboard" action="read">
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <MainPageHeader
          label={t('dashboard.title')}
          description={t('dashboard.description')}
          actions={
            <div className="flex items-center gap-2">
              <WidgetSelector availableWidgets={availableWidgets} onAddWidget={handleAddWidget} />
              {hasCustomLayout && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleResetLayoutWithWidgets}
                  className="shrink-0"
                >
                  <RefreshCw className="mr-2 h-4 w-4" />
                  Reset Layout
                </Button>
              )}
            </div>
          }
        />
        {showDragHint && mounted && (
          <Banner
            variant="info"
            title="Customize Your Dashboard"
            description="Hover over any card to see the drag handle on the left. Click and drag to rearrange cards in your preferred order. Your layout will be saved automatically."
            dismissible
            onDismiss={dismissHint}
            className="mb-4"
          />
        )}
        {/* TODO: Re-enable SMTP banner when notifications feature is working */}
        {/* <SMTPBannerConditional hasSMTPConfig={!!smtpConfig} /> */}
        <MonitoringSection
          systemStats={systemStats}
          containersData={containersData}
          layoutResetKey={layoutResetKey}
          onLayoutChange={handleLayoutChange}
          onDeleteWidget={handleDeleteWidget}
          dashboardItems={dashboardItems}
        />
      </PageLayout>
    </ResourceGuard>
  );
}

export default DashboardPage;

// TODO: Re-enable SMTP banner when notifications feature is working
// const SMTPBannerConditional = ({ hasSMTPConfig }: { hasSMTPConfig: boolean }) => {
//   if (hasSMTPConfig) return null;
//   return <SMTPBanner />;
// };

export interface MonitoringSectionProps extends React.HTMLAttributes<HTMLDivElement> {
  systemStats: SystemStatsType | null;
  containersData: ContainerData[] | null;
  layoutResetKey: number;
  onLayoutChange: () => void;
  onDeleteWidget: (id: string) => void;
  dashboardItems: DashboardItem[];
}

const MonitoringSection = ({
  systemStats,
  layoutResetKey,
  onLayoutChange,
  onDeleteWidget,
  dashboardItems
}: MonitoringSectionProps) => {
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
