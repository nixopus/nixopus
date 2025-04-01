'use client';
import React from 'react';
import useFindRepository from '../../hooks/use_find_repository';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import { DeployForm } from '../../components/create-form/deploy-form';
import { useTranslation } from '@/hooks/use-translation';

function page() {
  const { repository } = useFindRepository();
  const { t } = useTranslation();

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader
        label={repository?.name || t('selfHost.create.title')}
        description={t('selfHost.create.description')}
      />
      <DeployForm
        repository={repository?.id.toString() || ''}
        application_name={repository?.name || ''}
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
  );
}

export default page;
