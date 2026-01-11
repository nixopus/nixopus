'use client';

import React from 'react';
import { useParams } from 'next/navigation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/components/layout/page-layout';
import DeploymentLogsTable from '@/packages/components/deployment-logs';

function page() {
  const { deployment_id } = useParams();
  const deploymentId = deployment_id?.toString() || '';

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <DeploymentLogsTable id={deploymentId} isDeployment={true} title="Deployment Logs" />
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
