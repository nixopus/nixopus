import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useRollbackApplicationMutation } from '@/redux/services/deploy/applicationsApi';
import { ApplicationDeployment } from '@/redux/types/applications';
import { Undo, Eye, Clock, GitBranch, Terminal, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import PaginationWrapper from '@/components/ui/pagination';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

interface DeploymentsListProps {
  deployments?: ApplicationDeployment[];
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

function DeploymentsList({ deployments, currentPage, totalPages, onPageChange }: DeploymentsListProps) {
  const { t } = useTranslation();
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

  const getStatusIcon = (status?: string) => {
    switch (status?.toLowerCase()) {
      case 'completed':
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      case 'in_progress':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
      default:
        return <CheckCircle2 className="h-4 w-4 text-gray-500" />;
    }
  };

  return (
    <div className="space-y-6">
      {deployments && deployments.length > 0 ? (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {deployments.map((deployment) => (
              <Card 
                key={deployment.id} 
                className="w-full hover:shadow-lg transition-all duration-200 border border-gray-200 dark:border-gray-800"
              >
                <CardHeader className="pb-2">
                  <div className="flex justify-between items-start gap-4">
                    <div className="space-y-1">
                      <CardTitle className="text-lg font-semibold flex items-center gap-2">
                        {getStatusIcon(deployment.status?.status)}
                        {deployment.container_name?.startsWith('/')
                          ? deployment.container_name.slice(1)
                          : deployment.container_name}
                      </CardTitle>
                      <CardDescription className="text-sm">
                        {formatDate(deployment.created_at)}
                      </CardDescription>
                    </div>
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={() => rollBackApplication({ id: deployment.id })}
                      disabled={isLoading}
                      title={t('selfHost.deployments.list.card.rollback.title')}
                      className="hover:bg-red-50 dark:hover:bg-red-900/20"
                    >
                      <Undo className="h-4 w-4" />
                    </Button>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex flex-col space-y-3">
                    <div className="flex justify-between items-center text-sm">
                      <span className="flex items-center gap-2 text-muted-foreground">
                        <Clock className="h-4 w-4" />
                        {t('selfHost.deployments.list.card.runTime')}
                      </span>
                      <Badge variant="secondary" className="font-medium">
                        {calculateRunTime(
                          deployment.status?.updated_at as string,
                          deployment.status?.created_at as string
                        )}
                      </Badge>
                    </div>
                    <div 
                      className="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400 hover:underline cursor-pointer"
                      onClick={() => {
                        router.push(
                          `/self-host/application/${deployment.application_id}/deployments/${deployment.id}`
                        );
                      }}
                    >
                      <Eye className="h-4 w-4" />
                      {t('selfHost.deployments.list.card.viewButton')}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
          {totalPages > 1 && (
            <div className="mt-8 flex justify-center">
              <PaginationWrapper
                currentPage={currentPage}
                totalPages={totalPages}
                onPageChange={onPageChange}
              />
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-12 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
          <p className="text-muted-foreground">{t('selfHost.deployments.list.noDeployments')}</p>
        </div>
      )}
    </div>
  );
}

export default DeploymentsList;
