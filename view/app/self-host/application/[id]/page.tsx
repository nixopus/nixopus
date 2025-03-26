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

function Page() {
  const { application, currentPage, setCurrentPage, envVariables, buildVariables } =
    useApplicationDetails();

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl 2xl:max-w-7xl">
      <ApplicationDetailsHeader application={application} />
      <Tabs defaultValue="monitoring" className="w-full">
        <TabsList>
          <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="deployments">Deployments</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
        </TabsList>
        <TabsContent value="deployments" className="mt-6">
          <DeploymentsList deployments={application?.deployments} />
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
          />
        </TabsContent>
        <TabsContent value="logs" className="mt-6">
          <ApplicationLogs
            logs={application?.logs}
            onRefresh={() => {}}
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
