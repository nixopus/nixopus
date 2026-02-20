'use client';
import React from 'react';
import TabsWrapper, { TabsWrapperList, Skeleton } from '@nixopus/ui';
import useApplicationDetails from '../../../../packages/hooks/applications/use_application_details';
import { ApplicationDetailsHeader } from '@/packages/components/application-details';
import { ResourceGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';

function ApplicationDetailsSkeleton() {
  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <div className="space-y-4 mb-6">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Skeleton className="h-7 w-48" />
            <Skeleton className="h-6 w-6 rounded-md" />
            <Skeleton className="h-6 w-6 rounded-md" />
          </div>

          <div className="flex items-center gap-2">
            <Skeleton className="h-9 w-24" />
            <Skeleton className="h-9 w-28" />
            <Skeleton className="h-9 w-9 rounded-md" />
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-5 w-16 rounded-full" />
          <Skeleton className="h-5 w-16 rounded-full" />
          <Skeleton className="h-5 w-16 rounded-md" />
        </div>
      </div>

      <div className="border-b mb-6">
        <div className="flex gap-2">
          <Skeleton className="h-10 w-32 rounded-none" />
          <Skeleton className="h-10 w-36 rounded-none" />
          <Skeleton className="h-10 w-32 rounded-none" />
          <Skeleton className="h-10 w-28 rounded-none" />
          <Skeleton className="h-10 w-24 rounded-none" />
        </div>
      </div>

      <div className="space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Skeleton className="h-64 lg:col-span-2 rounded-lg" />
          <Skeleton className="h-64 rounded-lg" />
          <Skeleton className="h-64 rounded-lg" />
        </div>
      </div>
    </PageLayout>
  );
}

function Page() {
  const { application, activeTab, setActiveTab, tabs, sharedTabTriggerClassName } =
    useApplicationDetails();

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<ApplicationDetailsSkeleton />}>
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
