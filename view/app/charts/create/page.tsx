'use client';
import React from 'react';
import {
  ListRepositories,
  GithubRepositoriesSkeletonLoader
} from '@/packages/components/github-repositories';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@nixopus/ui';
import PageLayout from '@/packages/layouts/page-layout';
import { MainPageHeader } from '@nixopus/ui';

function page() {
  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={
        <PageLayout maxWidth="7xl" padding="md" spacing="lg">
          <Skeleton className="h-8 w-48" />
          <div className="flex items-center justify-between gap-4 flex-wrap mb-4">
            <Skeleton className="h-10 w-[280px] rounded-md" />
            <div className="flex items-center gap-2">
              <Skeleton className="h-9 w-36 rounded-md" />
              <Skeleton className="h-9 w-28 rounded-md" />
            </div>
          </div>
          <GithubRepositoriesSkeletonLoader />
        </PageLayout>
      }
    >
      <PageLayout maxWidth="7xl" padding="md" spacing="lg">
        <MainPageHeader label="Repository" highlightLabel={false} />
        <ListRepositories />
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
