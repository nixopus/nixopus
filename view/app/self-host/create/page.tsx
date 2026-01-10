'use client';
import React from 'react';
import ListRepositories from '../components/github-repositories/list-repositories';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/components/layout/page-layout';
import MainPageHeader from '@/components/ui/main-page-header';

function page() {
  const { t } = useTranslation();

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
