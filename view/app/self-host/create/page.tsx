'use client';
import React from 'react';
import ListRepositories from '../components/github-repositories/list-repositories';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';

function page() {
  const { t } = useTranslation();

  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <div className="container mx-auto py-6 space-y-8 max-w-4xl">
        <ListRepositories />
      </div>
    </ResourceGuard>
  );
}

export default page;
