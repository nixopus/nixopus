'use client';
import { Application } from '@/redux/types/applications';
import React from 'react';
import { DeploymentStatusChart } from './deploymentStatusChart';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';

function Monitor({ application }: { application?: Application }) {
  if (!application) {
    return null;
  }

  const deployments = application.deployments || [];

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <div className="space-y-6">
        <div className="">
          <DeploymentStatusChart deployments={deployments} />
        </div>
      </div>
    </ResourceGuard>
  );
}

export default Monitor;
