'use client';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import { Button } from '@/components/ui/button';
import React from 'react';
import DomainsTable from './components/domainsTable';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import UpdateDomainDialog from './components/update-domain';
import { useAppSelector } from '@/redux/hooks';
import { useResourcePermissions } from '@/lib/permission';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

const Page = () => {
  const { t } = useTranslation();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const {
    data: domains,
    isLoading,
    error
  } = useGetAllDomainsQuery({ organizationId: activeOrg?.id || '' }, { skip: !activeOrg?.id });
  const [addDomainDialogOpen, setAddDomainDialogOpen] = React.useState(false);
  const user = useAppSelector((state) => state.auth.user);
  const { canCreate, canRead } = useResourcePermissions(user, 'organization', activeOrg?.id);

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

  if (!canRead) {
    return (
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <DashboardPageHeader
          label={t('settings.domains.page.title')}
          description={t('settings.domains.page.description')}
        />
        <div className="flex flex-col h-full justify-center items-center gap-4 mt-12">
          <h2 className="text-xl font-medium text-center text-foreground">
            {t('settings.domains.page.accessDenied.title')}
          </h2>
          <p className="text-muted-foreground text-center">
            {t('settings.domains.page.accessDenied.description')}
          </p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <DashboardPageHeader
          label={t('settings.domains.page.title')}
          description={t('settings.domains.page.description')}
        />
        <div className="flex flex-col h-full justify-center items-center gap-4 mt-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <p className="text-muted-foreground text-center">{t('settings.domains.page.loading')}</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <DashboardPageHeader
          label={t('settings.domains.page.title')}
          description={t('settings.domains.page.description')}
        />
        <Alert variant="destructive">
          <AlertDescription>{t('settings.domains.page.error')}</AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label={t('settings.domains.page.title')}
        description={t('settings.domains.page.description')}
      />
      {domains && domains.length > 0 ? (
        <>
          <div className="flex justify-between items-center mt-8">
            <h2 className="text-xl font-medium text-foreground">
              {t('settings.domains.page.domainsList.title')}
            </h2>
            {canCreate && (
              <Button variant="default" onClick={() => setAddDomainDialogOpen(true)}>
                {t('settings.domains.page.domainsList.addButton')}
              </Button>
            )}
          </div>
          <DomainsTable domains={domains} />
        </>
      ) : (
        <NoDomainsFound
          onPressAddDomain={() => setAddDomainDialogOpen(true)}
          canCreate={canCreate}
        />
      )}
      {addDomainDialogOpen && (
        <UpdateDomainDialog open={addDomainDialogOpen} setOpen={setAddDomainDialogOpen} />
      )}
    </div>
  );
};

export default Page;

interface NoDomainsFoundProps {
  onPressAddDomain: () => void;
  canCreate: boolean;
}

const NoDomainsFound = ({ onPressAddDomain, canCreate }: NoDomainsFoundProps) => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col h-full justify-center items-center gap-4">
      <h2 className="text-xl font-medium text-center text-foreground">
        {t('settings.domains.page.noDomains.title')}
      </h2>
      {canCreate ? (
        <Button className="mx-auto" variant="default" onClick={onPressAddDomain}>
          {t('settings.domains.page.domainsList.addButton')}
        </Button>
      ) : (
        <p className="text-muted-foreground text-center">
          {t('settings.domains.page.noDomains.noPermission')}
        </p>
      )}
    </div>
  );
};
