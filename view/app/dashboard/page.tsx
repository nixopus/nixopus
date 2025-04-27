'use client';

import React from 'react';
import useMonitor from './hooks/use-monitor';
import ContainersTable from './components/containers/container-table';
import SystemStats from './components/system/system-stats';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package } from 'lucide-react';
import DiskUsageCard from './components/system/disk-usage';
import { useTranslation } from '@/hooks/use-translation';
import { useAppSelector } from '@/redux/hooks';
import { useGetSMTPConfigurationsQuery } from '@/redux/services/settings/notificationApi';
import { SMTPBanner } from './components/smtp-banner';
import { useFeatureFlags } from '@/hooks/features_provider';
import Skeleton from '../file-manager/components/skeleton/Skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/types/feature-flags';

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
    <div className="container mx-auto py-6 space-y-8 max-w-6xl">
      <div className="p-4 md:p-6 space-y-4 md:space-y-6">
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2">
          <div>
            <h1 className="text-xl sm:text-2xl font-bold">{t('dashboard.title')}</h1>
            <p className="text-sm sm:text-base text-muted-foreground">
              {t('dashboard.description')}
            </p>
          </div>
        </div>
        {!smtpConfig && <SMTPBanner />}
        <MonitoringSection systemStats={systemStats} containersData={containersData} t={t} />
      </div>
    </div>
  );
}

export default DashboardPage;

const MonitoringSection = ({ systemStats, containersData, t }: { systemStats: any, containersData: any, t: any }) => {
  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <SystemStats systemStats={systemStats} />
        </div>
        <DiskUsageCard systemStats={systemStats} />
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
            <Package className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
            {t('dashboard.containers.title')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ContainersTable containersData={containersData} />
        </CardContent>
      </Card>
    </>
  );
};
