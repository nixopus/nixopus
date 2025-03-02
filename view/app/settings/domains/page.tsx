'use client';
import DashboardPageHeader from '@/components/dashboard-page-header';
import { Button } from '@/components/ui/button';
import React from 'react';
import DomainsTable from './components/domainsTable';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import UpdateDomainDialog from './components/update-domain';

const Page = () => {
  const { data: domains } = useGetAllDomainsQuery();

  const [addDomainDialogOpen, setAddDomainDialogOpen] = React.useState(false);

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader label="Server and Domains" description="Configure your domains" />
      {domains ? (
        <>
          <div className="flex justify-between items-center mt-8">
            <h2 className="text-xl font-medium text-foreground">Domains</h2>
            <Button variant="default" onClick={() => setAddDomainDialogOpen(true)}>
              {' '}
              Add Domain
            </Button>
          </div>
          <DomainsTable domains={domains || []} />
        </>
      ) : (
        <NoDomainsFound onPressAddDomain={() => setAddDomainDialogOpen(true)} />
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
}

const NoDomainsFound = ({ onPressAddDomain }: NoDomainsFoundProps) => {
  return (
    <div className="flex flex-col h-full justify-center items-center gap-4">
      <h2 className="text-xl font-medium text-center text-foreground">No Domains Found</h2>
      <Button className="mx-auto" variant="default" onClick={onPressAddDomain}>
        {' '}
        Add Domain
      </Button>
    </div>
  );
};
