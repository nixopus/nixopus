'use client';
import React from 'react';
import TabsWrapper, { TabsWrapperList } from '@/components/ui/tabs-wrapper';
import useApplicationDetails from '../../../../packages/hooks/applications/use_application_details';
import { ApplicationDetailsHeader } from '@/packages/components/application-details';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/packages/layouts/page-layout';

function Page() {
  const { application, activeTab, setActiveTab, tabs, sharedTabTriggerClassName } =
    useApplicationDetails();

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <ApplicationDetailsHeader application={application} />
        <TabsWrapper
          value={activeTab}
          onValueChange={setActiveTab}
          tabs={tabs}
          tabsListClassName="w-full justify-start rounded-none h-auto p-0 bg-transparent gap-2"
          tabsTriggerClassName={sharedTabTriggerClassName}
        >
          <TabsWrapperList />
        </TabsWrapper>
      </PageLayout>
    </ResourceGuard>
  );
}

export default Page;
