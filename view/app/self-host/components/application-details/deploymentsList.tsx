import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useRollbackApplicationMutation } from '@/redux/services/deploy/applicationsApi';
import { ApplicationDeployment } from '@/redux/types/applications';
import { Redo, Undo } from 'lucide-react';
import { useRouter } from 'next/navigation';
import React from 'react';

function DeploymentsList({ deployments }: { deployments?: ApplicationDeployment[] }) {
  const router = useRouter();
  const [rollBackApplication, { isLoading }] = useRollbackApplicationMutation();

  const formatDate = (created_at: string) =>
    deployments
      ? new Date(created_at).toLocaleString('en-US', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
      : 'N/A';

  const calculateRunTime = (updated_at: string, created_at: string) => {
    const createdTime = new Date(created_at).getTime();
    const updatedTime = new Date(updated_at).getTime();
    const duration = updatedTime - createdTime;

    const minutes = Math.floor(duration / 60000);
    const seconds = Math.floor((duration % 60000) / 1000);

    return `${minutes}m ${seconds}s`;
  };

  return (
    <div className="space-y-6">
      {deployments && deployments.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {deployments.map((deployment) => (
            <Card
              key={deployment.id}
              className="w-full hover:shadow-md transition-shadow">
              <CardHeader className="pb-2">
                <div className="flex justify-between items-center">
                  <div className='flex flex-col items-center'>
                    <CardTitle className="text-lg font-medium">
                      #{deployment.id.slice(0, 6)}
                    </CardTitle>
                    <CardDescription>
                      <span className="text-xs px-2 py-1 rounded-full">
                        {deployment.status?.status || 'Completed'}
                      </span>
                    </CardDescription>
                  </div>
                  <Button size={"icon"} onClick={() => rollBackApplication({ id: deployment.id })}>
                    <Undo className="h-4 w-4" />
                  </Button>
                </div>
                <CardDescription>{formatDate(deployment.created_at)}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex flex-col space-y-1">
                  <div className="flex justify-between text-sm">
                    <span className="">Run Time:</span>
                    <span className="font-medium">
                      {calculateRunTime(deployment.updated_at, deployment.created_at)}
                    </span>
                  </div>
                </div>
                <Button className='w-full mt-3 cursor-pointer' variant={"secondary"} onClick={() => {
                  router.push(
                    `/self-host/application/${deployment.application_id}/deployments/${deployment.id}`
                  );
                }}>View</Button>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <div className="text-center py-8  rounded-lg">
          <p className="">No deployments found</p>
        </div>
      )}
    </div>
  );
}

export default DeploymentsList;
