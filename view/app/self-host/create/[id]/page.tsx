'use client';
import React from 'react';
import useFindRepository from '../../hooks/use_find_repository';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import { DeployForm } from '../../components/create-form/deploy-form';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';

function page() {
  const { repository } = useFindRepository();
  const { t } = useTranslation();

  return (
    <ResourceGuard
      resource="deploy"
      action="create"
      loadingFallback={<Skeleton className="h-96" />}
    >
      <div className="container mx-auto py-6 space-y-8 max-w-4xl justify-center items-center h-screen flex-col flex">
        <DashboardPageHeader
          label={repository?.name || t('selfHost.create.title')}
          description={t('selfHost.create.description')}
          className="justify-center text-center mb-16"
        />
        <DeployForm
          repository={repository?.id.toString() || ''}
          application_name={repository?.name || ''}
          repository_full_name={repository?.full_name || ''}
          environment={Environment.Production}
          branch="main"
          port={'3000'}
          domain=""
          build_pack={BuildPack.Dockerfile}
          env_variables={{}}
          build_variables={{}}
          pre_run_commands=""
          post_run_commands=""
        />
      </div>
    </ResourceGuard>
  );
}

export default page;
