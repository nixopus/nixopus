'use client';

import React from 'react';
import useMonitor from './hooks/use-monitor';
import ContainersTable from './components/containers/container-table';
import SystemStats from './components/system/system-stats';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, ArrowRight } from 'lucide-react';
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

// for dashboard page, we need to check if the user has the dashboard:read permission
function DashboardPage() {
  const { t } = useTranslation();
  const { containersData, systemStats } = useMonitor();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const { data: smtpConfig } = useGetSMTPConfigurationsQuery(activeOrganization?.id, {
    skip: !activeOrganization
  });
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

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
        </div>
        {!smtpConfig && <SMTPBanner />}
        <MonitoringSection systemStats={systemStats} containersData={containersData} t={t} />
      </PageLayout>
    </ResourceGuard>
  );
}

export default DashboardPage;

const MonitoringSection = ({
  systemStats,
  containersData,
  t
}: {
  systemStats: any;
  containersData: any;
  t: any;
}) => {
  const router = useRouter();
  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <SystemStats systemStats={systemStats} />
        </div>
        <DiskUsageCard systemStats={systemStats} />
      </div>
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
    </>
  );
};
