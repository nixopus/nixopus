'use client';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import { Button } from '@/components/ui/button';
import React from 'react';
import DomainsTable from './components/domainsTable';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import UpdateDomainDialog from './components/update-domain';
import { useAppSelector } from '@/redux/hooks';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import Skeleton from '@/app/file-manager/components/skeleton/Skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/types/feature-flags';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import PageLayout from '@/components/layout/page-layout';

const Page = () => {
  const { t } = useTranslation();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const { data: domains, isLoading, error } = useGetAllDomainsQuery();
  const [addDomainDialogOpen, setAddDomainDialogOpen] = React.useState(false);
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  if (!activeOrg?.id) {
    return (
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <DashboardPageHeader
          label={t('settings.domains.page.title')}
          description={t('settings.domains.page.description')}
        />
        <div className="flex flex-col h-full justify-center items-center gap-4 mt-12">
          <h2 className="text-xl font-medium text-center text-foreground">
            {t('settings.domains.page.noOrganization.title')}
          </h2>
          <p className="text-muted-foreground text-center">
            {t('settings.domains.page.noOrganization.description')}
          </p>
        </div>
      </div>
    );
  }

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureDomain)) {
    return <DisabledFeature />;
  }

  return (
    <ResourceGuard resource="domain" action="read">
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <DashboardPageHeader
          label={t('settings.domains.page.title')}
          description={t('settings.domains.page.description')}
        />
        {isLoading ? (
          <div className="flex flex-col h-full justify-center items-center gap-4 mt-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <p className="text-muted-foreground text-center">
              {t('settings.domains.page.loading')}
            </p>
          </div>
        ) : error ? (
          <Alert variant="destructive">
            <AlertDescription>{t('settings.domains.page.error')}</AlertDescription>
          </Alert>
        ) : domains && domains.length > 0 ? (
          <>
            <div className="flex justify-end items-center mt-8">
              <ResourceGuard resource="domain" action="create">
                <Button variant="default" onClick={() => setAddDomainDialogOpen(true)}>
                  {t('settings.domains.page.domainsList.addButton')}
                </Button>
              </ResourceGuard>
            </div>
            <DomainsTable domains={domains} />
          </>
        ) : (
          <NoDomainsFound onPressAddDomain={() => setAddDomainDialogOpen(true)} />
        )}
        {addDomainDialogOpen && (
          <UpdateDomainDialog open={addDomainDialogOpen} setOpen={setAddDomainDialogOpen} />
        )}
      </PageLayout>
    </ResourceGuard>
  );
};

export default Page;

interface NoDomainsFoundProps {
  onPressAddDomain: () => void;
}

const NoDomainsFound = ({ onPressAddDomain }: NoDomainsFoundProps) => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col h-full justify-center items-center gap-4">
      <h2 className="text-xl font-medium text-center text-foreground">
        {t('settings.domains.page.noDomains.title')}
      </h2>
      <ResourceGuard resource="domain" action="create">
        <Button className="mx-auto" variant="default" onClick={onPressAddDomain}>
          {t('settings.domains.page.domainsList.addButton')}
        </Button>
      </ResourceGuard>
      <ResourceGuard
        resource="domain"
        action="create"
        fallback={
          <p className="text-muted-foreground text-center">
            {t('settings.domains.page.noDomains.noPermission')}
          </p>
        }
      >
        <></>
      </ResourceGuard>
    </div>
  );
};
