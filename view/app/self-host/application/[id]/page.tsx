'use client';
import React from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import ApplicationLogs from '../../components/application-details/logs';
import Monitor from '../../components/application-details/monitor';
import DeploymentsList from '../../components/application-details/deploymentsList';
import useApplicationDetails from '../../hooks/use_application_details';
import { DeployConfigureForm } from '../../components/application-details/configuration';
import { BuildPack, Environment } from '@/redux/types/deploy-form';
import ApplicationDetailsHeader from '../../components/application-details/header';
import { useAppSelector } from '@/redux/hooks';
import { hasPermission } from '@/lib/permission';
import { useTranslation } from '@/hooks/use-translation';

function Page() {
  const { t } = useTranslation();
  const user = useAppSelector((state) => state.auth.user);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const {
    application,
    currentPage,
    setCurrentPage,
    envVariables,
    buildVariables,
    defaultTab,
    deploymentsPage,
    setDeploymentsPage,
    deploymentsPerPage,
    totalDeployments
  } = useApplicationDetails();

  const canRead = hasPermission(user, 'deploy', 'read', activeOrg?.id);
  const totalPages = Math.ceil(totalDeployments / deploymentsPerPage);

  if (!canRead) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">{t('selfHost.application.accessDenied.title')}</h2>
          <p className="text-muted-foreground">
            {t('selfHost.application.accessDenied.description')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl 2xl:max-w-7xl">
      <ApplicationDetailsHeader application={application} />
      <Tabs defaultValue={defaultTab} className="w-full">
        <TabsList>
          <TabsTrigger value="monitoring">{t('selfHost.application.tabs.monitoring')}</TabsTrigger>
          <TabsTrigger value="configuration">
            {t('selfHost.application.tabs.configuration')}
          </TabsTrigger>
          <TabsTrigger value="deployments">
            {t('selfHost.application.tabs.deployments')}
          </TabsTrigger>
          <TabsTrigger value="logs">{t('selfHost.application.tabs.logs')}</TabsTrigger>
        </TabsList>
        <TabsContent value="deployments" className="mt-6">
          <DeploymentsList
            deployments={application?.deployments}
            currentPage={deploymentsPage}
            totalPages={totalPages}
            onPageChange={setDeploymentsPage}
          />
        </TabsContent>
        <TabsContent value="configuration" className="mt-6">
          <DeployConfigureForm
            application_name={application?.name}
            domain={application?.domain}
            environment={application?.environment as Environment | undefined}
            env_variables={envVariables}
            build_variables={buildVariables}
            build_pack={application?.build_pack as BuildPack}
            branch={application?.branch}
            port={application?.port?.toString()}
            repository={application?.repository}
            pre_run_commands={application?.pre_run_command}
            post_run_commands={application?.post_run_command}
            application_id={application?.id}
            dockerFilePath={application?.dockerfile_path}
            base_path={application?.base_path}
          />
        </TabsContent>
        <TabsContent value="logs" className="mt-6">
          <ApplicationLogs
            id={application?.id || ''}
            currentPage={currentPage}
            setCurrentPage={setCurrentPage}
          />
        </TabsContent>
        <TabsContent value="monitoring" className="mt-6">
          <Monitor application={application} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default Page;
