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
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/components/layout/page-layout';

function Page() {
  const { t } = useTranslation();
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

  const totalPages = Math.ceil(totalDeployments / deploymentsPerPage);

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        <ApplicationDetailsHeader application={application} />
        <Tabs defaultValue={defaultTab} className="w-full">
          <TabsList>
            <TabsTrigger value="monitoring">
              {t('selfHost.application.tabs.monitoring')}
            </TabsTrigger>
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
      </PageLayout>
    </ResourceGuard>
  );
}

export default Page;
