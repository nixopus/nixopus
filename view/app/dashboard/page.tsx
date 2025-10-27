'use client';

import React from 'react';
import { useDashboard } from './hooks/use-dashboard';
import ContainersTable from './components/containers/container-table';
import SystemInfoCard from './components/system/system-info';
import LoadAverageCard from './components/system/load-average';
import CPUUsageCard from './components/system/cpu-usage';
import MemoryUsageCard from './components/system/memory-usage';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, ArrowRight, RefreshCw, Info } from 'lucide-react';
import DiskUsageCard from './components/system/disk-usage';
import { useTranslation } from '@/hooks/use-translation';
import { SMTPBanner } from './components/smtp-banner';
import DisabledFeature from '@/components/features/disabled-feature';
import { Button } from '@/components/ui/button';
import { useRouter } from 'next/navigation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { TypographyH1, TypographyMuted, TypographySmall } from '@/components/ui/typography';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/components/layout/page-layout';
import { DraggableGrid, DraggableItem } from '@/components/ui/draggable-grid';

// for dashboard page, we need to check if the user has the dashboard:read permission
function DashboardPage() {
  const { t } = useTranslation();
  const {
    isFeatureFlagsLoading,
    isDashboardEnabled,
    containersData,
    systemStats,
    smtpConfig,
    showDragHint,
    mounted,
    layoutResetKey,
    dismissHint,
    handleResetLayout,
    hasCustomLayout,
    handleLayoutChange,
  } = useDashboard();

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isDashboardEnabled) {
    return <DisabledFeature />;
  }

  return (
    <ResourceGuard
      resource="dashboard"
      action="read"
    >
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <DashboardHeader
          hasCustomLayout={hasCustomLayout}
          onResetLayout={handleResetLayout}
          title={t('dashboard.title')}
          description={t('dashboard.description')}
        />
        <DragHintBanner
          mounted={mounted}
          showDragHint={showDragHint}
          onDismiss={dismissHint}
        />
        <SMTPBannerConditional hasSMTPConfig={!!smtpConfig} />
        <MonitoringSection
          systemStats={systemStats}
          containersData={containersData}
          t={t}
          layoutResetKey={layoutResetKey}
          onLayoutChange={handleLayoutChange}
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
  description
}: {
  hasCustomLayout: boolean;
  onResetLayout: () => void;
  title: string;
  description: string;
}) => (
  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2 mb-4">
    <div>
      <TypographyH1>{title}</TypographyH1>
      <TypographyMuted>{description}</TypographyMuted>
    </div>
    {hasCustomLayout && (
      <Button
        variant="outline"
        size="sm"
        onClick={onResetLayout}
        className="shrink-0"
      >
        <RefreshCw className="mr-2 h-4 w-4" />
        Reset Layout
      </Button>
    )}
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
            Hover over any card to see the drag handle on the left. Click and drag to rearrange cards in your preferred order. Your layout will be saved automatically.
          </p>
        </div>
      </div>
      <Button
        variant="ghost"
        size="default"
        onClick={onDismiss}
        className="shrink-0"
      >
        Got it
      </Button>
    </div>
  );
};

const SMTPBannerConditional = ({ hasSMTPConfig }: { hasSMTPConfig: boolean }) => {
  if (hasSMTPConfig) return null;
  return <SMTPBanner />;
};

const MonitoringSection = ({
  systemStats,
  containersData,
  t,
  layoutResetKey,
  onLayoutChange
}: {
  systemStats: any;
  containersData: any;
  t: any;
  layoutResetKey: number;
  onLayoutChange: () => void;
}) => {
  const router = useRouter();

  if (!systemStats) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Skeleton className="h-48 w-full rounded-xl md:col-span-2" />
          <Skeleton className="h-64 w-full rounded-xl" />
          <Skeleton className="h-64 w-full rounded-xl" />
          <Skeleton className="h-64 w-full rounded-xl" />
          <Skeleton className="h-64 w-full rounded-xl" />
        </div>
        <Skeleton className="h-96 w-full rounded-xl" />
      </div>
    );
  }


  const dashboardItems: DraggableItem[] = [
    {
      id: 'system-info',
      component: <SystemInfoCard systemStats={systemStats} />,
      className: 'md:col-span-2'
    },
    {
      id: 'load-average',
      component: <LoadAverageCard systemStats={systemStats} />
    },
    {
      id: 'cpu-usage',
      component: <CPUUsageCard systemStats={systemStats} />
    },
    {
      id: 'memory-usage',
      component: <MemoryUsageCard systemStats={systemStats} />
    },
    {
      id: 'disk-usage',
      component: <DiskUsageCard systemStats={systemStats} />
    },
    {
      id: 'containers',
      component: (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-xs sm:text-sm font-bold flex items-center">
              <Package className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
              <TypographySmall>{t('dashboard.containers.title')}</TypographySmall>
            </CardTitle>
            <Button variant="outline" size="sm" onClick={() => router.push('/containers')}>
              <ArrowRight className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
              {t('dashboard.containers.viewAll')}
            </Button>
          </CardHeader>
          <CardContent>
            <ContainersTable containersData={containersData} />
          </CardContent>
        </Card>
      ),
      className: 'md:col-span-2'
    }
  ];

  return (
    <DraggableGrid
      items={dashboardItems}
      storageKey="dashboard-card-order"
      gridCols="grid-cols-1 md:grid-cols-2"
      resetKey={layoutResetKey}
      onReorder={() => onLayoutChange()}
    />
  );
};
