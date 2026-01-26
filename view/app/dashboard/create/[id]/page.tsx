'use client';
import React from 'react';
import useFindRepository from '@/packages/hooks/applications/use_find_repository';
import { QuickDeployForm } from '@/packages/components/application-form';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/packages/layouts/page-layout';

function page() {
  const { repository } = useFindRepository();

  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <PageLayout
        maxWidth="full"
        padding="md"
        spacing="lg"
        className="justify-center items-center min-h-[80vh] flex-col flex w-full"
      >
        <QuickDeployForm
          repository={repository?.id.toString() || ''}
          application_name={repository?.name || ''}
          repository_full_name={repository?.full_name || ''}
        />
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
