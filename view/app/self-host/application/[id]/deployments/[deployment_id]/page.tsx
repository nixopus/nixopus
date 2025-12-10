'use client';

import React from 'react';
import { useParams } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { DeploymentLogsTable } from '@/app/self-host/components/deployment-logs';

function page() {
  const { t } = useTranslation();
  const { deployment_id } = useParams();
  const deploymentId = deployment_id?.toString() || '';

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <div className="py-6 space-y-8 w-full">
        <DeploymentLogsTable id={deploymentId} isDeployment={true} title="Deployment Logs" />
      </div>
    </ResourceGuard>
  );
}

export default page;
