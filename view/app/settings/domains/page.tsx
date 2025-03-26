'use client';
import DashboardPageHeader from '@/components/dashboard-page-header';
import { Button } from '@/components/ui/button';
import React from 'react';
import DomainsTable from './components/domainsTable';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import UpdateDomainDialog from './components/update-domain';
import { useAppSelector } from '@/redux/hooks';
import { useResourcePermissions } from '@/lib/permission';

const Page = () => {
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const { data: domains } = useGetAllDomainsQuery(
    { organizationId: activeOrg?.id || '' },
    { skip: !activeOrg }
  );
  const [addDomainDialogOpen, setAddDomainDialogOpen] = React.useState(false);
  const user = useAppSelector((state) => state.auth.user);
  const { canCreate, canRead } = useResourcePermissions(user, 'organization', activeOrg?.id);

  if (!canRead) {
    return (
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <DashboardPageHeader label="Server and Domains" description="Configure your domains" />
        <div className="flex flex-col h-full justify-center items-center gap-4 mt-12">
          <h2 className="text-xl font-medium text-center text-foreground">Access Denied</h2>
          <p className="text-muted-foreground text-center">
            You don't have permission to view domains for this organization.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader label="Server and Domains" description="Configure your domains" />
      {domains ? (
        <>
          <div className="flex justify-between items-center mt-8">
            <h2 className="text-xl font-medium text-foreground">Domains</h2>
            {canCreate && (
              <Button variant="default" onClick={() => setAddDomainDialogOpen(true)}>
                Add Domain
              </Button>
            )}
          </div>
          <DomainsTable domains={domains || []} />
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
  return (
    <div className="flex flex-col h-full justify-center items-center gap-4">
      <h2 className="text-xl font-medium text-center text-foreground">No Domains Found</h2>
      {canCreate ? (
        <Button className="mx-auto" variant="default" onClick={onPressAddDomain}>
          Add Domain
        </Button>
      ) : (
        <p className="text-muted-foreground text-center">
          You don't have permission to add domains.
        </p>
      )}
    </div>
  );
};
