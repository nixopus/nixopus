import React from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import ApplicationLogs from '../../components/application-details/logs';
import Monitor from '../../components/application-details/monitor';
import Configuration from '../../components/application-details/configuration';
import DeploymentsList from '../../components/application-details/deploymentsList';

function Page() {
  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl mt-10">
      <Tabs defaultValue="monitoring" className="w-full">
        <TabsList>
          <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="deployments">Deployments</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
        </TabsList>
        <TabsContent value="deployments" className="mt-6">
          <DeploymentsList />
        </TabsContent>

        <TabsContent value="configuration" className="mt-6">
          <Configuration />
        </TabsContent>

        <TabsContent value="logs" className="mt-6">
          <ApplicationLogs />
        </TabsContent>

        <TabsContent value="monitoring" className="mt-6">
          <Monitor />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default Page;