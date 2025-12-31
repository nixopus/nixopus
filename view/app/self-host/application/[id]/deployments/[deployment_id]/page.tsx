'use client';

import React, { useCallback } from 'react';
import { useParams } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { Skeleton } from '@/components/ui/skeleton';
import { DeploymentLogsTable } from '@/app/self-host/components/deployment-logs';
import PageLayout from '@/components/layout/page-layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { XCircle, Loader2, CheckCircle2, AlertCircle, Ban } from 'lucide-react';
import {
  useGetApplicationDeploymentByIdQuery,
  useCancelDeploymentMutation
} from '@/redux/services/deploy/applicationsApi';
import { toast } from 'sonner';
import { Status } from '@/redux/types/applications';

function page() {
  const { t } = useTranslation();
  const { deployment_id } = useParams();
  const deploymentId = deployment_id?.toString() || '';

  const { data: deployment, isLoading: isLoadingDeployment } = useGetApplicationDeploymentByIdQuery(
    { id: deploymentId },
    { skip: !deploymentId, pollingInterval: 5000 }
  );

  const [cancelDeployment, { isLoading: isCancelling }] = useCancelDeploymentMutation();

  const deploymentStatus = deployment?.status?.status;

  const isInProgress =
    deploymentStatus === 'cloning' ||
    deploymentStatus === 'building' ||
    deploymentStatus === 'deploying' ||
    deploymentStatus === 'started';

  const handleCancelDeployment = useCallback(async () => {
    if (!deploymentId) return;

    try {
      await cancelDeployment({ id: deploymentId }).unwrap();
      toast.success(t('selfHost.deployment.cancelSuccess') || 'Deployment cancelled successfully');
    } catch (error) {
      toast.error(t('selfHost.deployment.cancelError') || 'Failed to cancel deployment');
    }
  }, [deploymentId, cancelDeployment, t]);

  const getStatusIcon = (status?: Status) => {
    switch (status) {
      case 'deployed':
        return <CheckCircle2 className="h-4 w-4 text-green-600" />;
      case 'failed':
        return <AlertCircle className="h-4 w-4 text-red-600" />;
      case 'cancelled':
        return <Ban className="h-4 w-4 text-orange-600" />;
      case 'cloning':
      case 'building':
      case 'deploying':
      case 'started':
        return <Loader2 className="h-4 w-4 text-blue-600 animate-spin" />;
      default:
        return <CheckCircle2 className="h-4 w-4 text-muted-foreground" />;
    }
  };

  const getStatusVariant = (status?: Status) => {
    switch (status) {
      case 'deployed':
        return 'default';
      case 'failed':
        return 'destructive';
      case 'cancelled':
        return 'secondary';
      case 'cloning':
      case 'building':
      case 'deploying':
      case 'started':
        return 'outline';
      default:
        return 'secondary';
    }
  };

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-semibold">
              {t('selfHost.deployment.logsTitle') || 'Deployment Logs'}
            </h1>
            {!isLoadingDeployment && deploymentStatus && (
              <Badge
                variant={getStatusVariant(deploymentStatus)}
                className="flex items-center gap-1"
              >
                {getStatusIcon(deploymentStatus)}
                <span className="capitalize">{deploymentStatus}</span>
              </Badge>
            )}
          </div>
          {isInProgress && (
            <Button
              variant="destructive"
              size="sm"
              onClick={handleCancelDeployment}
              disabled={isCancelling}
            >
              {isCancelling ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  {t('selfHost.deployment.cancelling') || 'Cancelling...'}
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 mr-2" />
                  {t('selfHost.deployment.cancel') || 'Cancel Deployment'}
                </>
              )}
            </Button>
          )}
        </div>
        <DeploymentLogsTable id={deploymentId} isDeployment={true} />
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
