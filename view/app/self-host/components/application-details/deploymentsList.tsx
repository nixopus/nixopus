import { Button } from '@/components/ui/button';
import { useRollbackApplicationMutation } from '@/redux/services/deploy/applicationsApi';
import { ApplicationDeployment } from '@/redux/types/applications';
import { Undo, Eye, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import PaginationWrapper from '@/components/ui/pagination';
import { Badge } from '@/components/ui/badge';
import { DataTable, TableColumn } from '@/components/ui/data-table';

interface DeploymentsListProps {
  deployments?: ApplicationDeployment[];
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

function DeploymentsList({
  deployments,
  currentPage,
  totalPages,
  onPageChange
}: DeploymentsListProps) {
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
    router.push(
      `/self-host/application/${deployment.application_id}/deployments/${deployment.id}`
    );
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
          onClick={(e) => {
            e.stopPropagation();
            rollBackApplication({ id: deployment.id });
          }}
          disabled={isLoading}
          className="text-destructive hover:text-destructive hover:bg-destructive/10"
        >
          <Undo className="h-4 w-4 mr-2" />
          {t('selfHost.deployment.list.card.rollback.title')}
        </Button>
      )
    }
  ];

  return (
    <div className="space-y-6">
      {deployments && deployments.length > 0 ? (
        <>
          <DataTable
            data={deployments}
            columns={columns}
            onRowClick={handleRowClick}
            showBorder={true}
            hoverable={true}
          />
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
        <div className="text-center py-12 rounded-lg border">
          <p className="text-muted-foreground">{t('selfHost.deployment.list.noDeployments')}</p>
        </div>
      )}
    </div>
  );
}

export default DeploymentsList;
