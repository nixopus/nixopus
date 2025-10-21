'use client';
import LogViewer from '@/app/self-host/components/application-details/log-viewer';
import useDeploymentDetails from '@/app/self-host/hooks/use_deployment_details';
import { useTranslation } from '@/hooks/use-translation';
import React, { useState } from 'react';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';

function page() {
  const { t } = useTranslation();
  const { deployment } = useDeploymentDetails();
  const [currentPage, setCurrentPage] = useState(1);

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <div className="container mx-auto py-6 space-y-8 max-w-4xl 2xl:max-w-7xl">
        <div className="mt-5 flex-col flex mb-4">
          <span className="text-2xl font-bold">{t('selfHost.deployment.title')}</span>
        </div>
        <LogViewer
          id={deployment?.id || ''}
          currentPage={currentPage}
          setCurrentPage={setCurrentPage}
          isDeployment={true}
        />
      </div>
    </ResourceGuard>
  );
}

export default page;
