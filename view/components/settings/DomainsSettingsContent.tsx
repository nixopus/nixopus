'use client';

import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import { useAppSelector } from '@/redux/hooks';
import DomainsTable from '@/app/settings/domains/components/domainsTable';
import UpdateDomainDialog from '@/app/settings/domains/components/update-domain';
import { Button } from '@/components/ui/button';
import { useState } from 'react';

export function DomainsSettingsContent() {
  const { t } = useTranslation();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const { data: domains, isLoading } = useGetAllDomainsQuery();
  const [addDomainDialogOpen, setAddDomainDialogOpen] = useState(false);

  if (!activeOrg?.id) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-semibold">{t('settings.domains.page.title')}</h2>
        <div className="text-center text-muted-foreground">
          {t('settings.domains.page.noOrganization.description')}
        </div>
      </div>
    );
  }

  return (
    <ResourceGuard resource="domain" action="read">
      <div className="space-y-6">
        <h2 className="text-2xl font-semibold">{t('settings.domains.page.title')}</h2>
        {isLoading ? (
          <div className="text-center text-muted-foreground">
            {t('settings.domains.page.loading')}
          </div>
        ) : domains && domains.length > 0 ? (
          <>
            <div className="flex justify-end">
              <ResourceGuard resource="domain" action="create">
                <Button onClick={() => setAddDomainDialogOpen(true)}>
                  {t('settings.domains.page.domainsList.addButton')}
                </Button>
              </ResourceGuard>
            </div>
            <DomainsTable domains={domains} />
          </>
        ) : (
          <div className="text-center text-muted-foreground">
            {t('settings.domains.page.noDomains.title')}
          </div>
        )}
        {addDomainDialogOpen && (
          <UpdateDomainDialog open={addDomainDialogOpen} setOpen={setAddDomainDialogOpen} />
        )}
      </div>
    </ResourceGuard>
  );
}
