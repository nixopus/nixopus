'use client';

import React from 'react';
import useMonitor from './hooks/use-monitor';
import ContainersTable from './components/containers/container-table';
import SystemInfoCard from './components/system/system-info';
import LoadAverageCard from './components/system/load-average';
import MemoryUsageCard from './components/system/memory-usage';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, ArrowRight, RefreshCw, Info } from 'lucide-react';
import DiskUsageCard from './components/system/disk-usage';
import { useTranslation } from '@/hooks/use-translation';
import { useAppSelector } from '@/redux/hooks';
import { useGetSMTPConfigurationsQuery } from '@/redux/services/settings/notificationApi';
import { SMTPBanner } from './components/smtp-banner';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/types/feature-flags';
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
  const { containersData, systemStats } = useMonitor();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const { data: smtpConfig } = useGetSMTPConfigurationsQuery(activeOrganization?.id, {
    skip: !activeOrganization
  });
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();
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

  const dismissHint = () => {
    setShowDragHint(false);
    localStorage.setItem('dashboard-drag-hint-seen', 'true');
  };

  const handleResetLayout = () => {
    localStorage.removeItem('dashboard-card-order');
    setLayoutResetKey(prev => prev + 1);
  };

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureMonitoring)) {
    return <DisabledFeature />;
  }

  return (
    <ResourceGuard
      resource="dashboard"
      action="read"
    // fallback={<div>You are not authorized to access this page</div>}
    >
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2 mb-4">
          <div>
            <TypographyH1>{t('dashboard.title')}</TypographyH1>
            <TypographyMuted>{t('dashboard.description')}</TypographyMuted>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleResetLayout}
            className="shrink-0"
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Reset Layout
          </Button>
        </div>
        {mounted && showDragHint && (
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
              onClick={dismissHint}
              className="shrink-0"
            >
              Got it
            </Button>
          </div>
        )}
        {!smtpConfig && <SMTPBanner />}
        <MonitoringSection
          systemStats={systemStats}
          containersData={containersData}
          t={t}
          layoutResetKey={layoutResetKey}
        />
      </PageLayout>
    </ResourceGuard>
  );
}

export default DashboardPage;

const MonitoringSection = ({
  systemStats,
  containersData,
  t,
  layoutResetKey
}: {
  systemStats: any;
  containersData: any;
  t: any;
  layoutResetKey: number;
}) => {
  const router = useRouter();

  if (!systemStats) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Skeleton className="h-48 w-full rounded-xl" />
          <Skeleton className="h-48 w-full rounded-xl" />
          <Skeleton className="h-48 w-full rounded-xl" />
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
    />
  );
};
