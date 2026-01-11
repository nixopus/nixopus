'use client';
import React from 'react';
import { ListRepositories } from '@/packages/components/github-repositories';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/packages/layouts/page-layout';
import MainPageHeader from '@/components/ui/main-page-header';

function page() {
  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <MainPageHeader label="Repository" description="Browse and manage projects" />
        <ListRepositories />
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
