'use client';
import React from 'react';
import useFindRepository from '../../hooks/use_find_repository';
import { QuickDeployForm } from '../../components/create-form/quick-deploy-form';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/components/layout/page-layout';

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
        className="justify-center items-center min-h-[80vh] flex-col flex"
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
