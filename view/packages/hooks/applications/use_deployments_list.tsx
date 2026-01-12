import React from 'react';
import { useRouter } from 'next/navigation';
import { useRollbackApplicationMutation } from '@/redux/services/deploy/applicationsApi';
import { ApplicationDeployment } from '@/redux/types/applications';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { TableColumn } from '@/components/ui/data-table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { CheckCircle2, AlertCircle, Loader2, Undo } from 'lucide-react';

interface UseDeploymentsListProps {
  deployments?: ApplicationDeployment[];
}

export function useDeploymentsList({ deployments }: UseDeploymentsListProps) {
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
      case 'deployed':
        return <CheckCircle2 className="h-4 w-4 text-green-600" />;
      case 'failed':
        return <AlertCircle className="h-4 w-4 text-red-600" />;
      case 'in_progress':
        return <Loader2 className="h-4 w-4 text-blue-600 animate-spin" />;
      default:
        return <CheckCircle2 className="h-4 w-4 text-muted-foreground" />;
    }
  };

  const handleRowClick = (deployment: ApplicationDeployment) => {
    router.push(`/self-host/application/${deployment.application_id}/deployments/${deployment.id}`);
  };

  const handleRollback = (deploymentId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    rollBackApplication({ id: deploymentId });
  };

  const columns: TableColumn<ApplicationDeployment>[] = [
    {
      key: 'status',
      title: t('selfHost.deployment.list.table.status'),
      render: (_, deployment) => (
        <div className="flex items-center gap-2">
          {getStatusIcon(deployment.status?.status)}
          <Badge
            variant={
              deployment.status?.status?.toLowerCase() === 'deployed'
                ? 'default'
                : deployment.status?.status?.toLowerCase() === 'failed'
                  ? 'destructive'
                  : 'secondary'
            }
          >
            {deployment.status?.status || t('selfHost.deployment.list.table.unknown')}
          </Badge>
        </div>
      )
    },
    {
      key: 'container',
      title: t('selfHost.deployment.list.table.container'),
      dataIndex: 'container_name',
      className: 'font-medium',
      render: (containerName) =>
        containerName?.startsWith('/') ? containerName.slice(1) : containerName
    },
    {
      key: 'created',
      title: t('selfHost.deployment.list.table.created'),
      dataIndex: 'created_at',
      render: (createdAt) => formatDate(createdAt)
    },
    {
      key: 'runTime',
      title: t('selfHost.deployment.list.table.runTime'),
      render: (_, deployment) => (
        <Badge variant="outline">
          {calculateRunTime(
            deployment.status?.updated_at as string,
            deployment.status?.created_at as string
          )}
        </Badge>
      )
    },
    {
      key: 'actions',
      title: t('selfHost.deployment.list.table.actions'),
      render: (_, deployment) => (
        <Button
          size="sm"
          variant="outline"
          onClick={(e) => handleRollback(deployment.id, e)}
          disabled={isLoading}
          className="text-destructive hover:text-destructive hover:bg-destructive/10"
        >
          <Undo className="h-4 w-4 mr-2" />
          {t('selfHost.deployment.list.card.rollback.title')}
        </Button>
      )
    }
  ];

  return {
    columns,
    handleRowClick,
    isLoading,
    formatDate,
    calculateRunTime,
    getStatusIcon
  };
}
