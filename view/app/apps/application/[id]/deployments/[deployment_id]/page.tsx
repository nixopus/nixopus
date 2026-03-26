'use client';

import React from 'react';
import { useParams } from 'next/navigation';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@nixopus/ui';
import PageLayout from '@/packages/layouts/page-layout';
import DeploymentLogsTable from '@/packages/components/deployment-logs';

function page() {
  const { deployment_id } = useParams();
  const deploymentId = deployment_id?.toString() || '';

  return (
    <ResourceGuard
      resource="deploy"
      action="read"
      loadingFallback={
        <PageLayout maxWidth="7xl" padding="md" spacing="lg">
          <Skeleton className="h-8 w-48" />
          <div className="flex items-center gap-4">
            <Skeleton className="h-10 flex-1 max-w-sm rounded-md" />
            <Skeleton className="h-10 w-36 rounded-md" />
          </div>
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full rounded-md" />
            ))}
          </div>
        </PageLayout>
      }
    >
      <PageLayout maxWidth="7xl" padding="md" spacing="lg">
        <DeploymentLogsTable id={deploymentId} isDeployment={true} title="Deployment Logs" />
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
