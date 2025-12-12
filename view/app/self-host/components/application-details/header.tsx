'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { ExternalLink, MoreVertical, RotateCcw, Trash2, Rocket, RefreshCw } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Application } from '@/redux/types/applications';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import {
  useRedeployApplicationMutation,
  useRestartApplicationMutation
} from '@/redux/services/deploy/applicationsApi';
import { useDeleteApplicationMutation } from '@/redux/services/deploy/applicationsApi';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

const ApplicationDetailsHeader = ({ application }: { application?: Application }) => {
  const { t } = useTranslation();
  const [redeployApplication, { isLoading: isRedeploying }] = useRedeployApplicationMutation();
  const [deleteApplication, { isLoading: isDeleting }] = useDeleteApplicationMutation();
  const router = useRouter();
  const [restartApplication, { isLoading: isRestarting }] = useRestartApplicationMutation();

  const latestDeployment = application?.deployments?.[0];
  const currentStatus = latestDeployment?.status?.status;

  const getStatusConfig = (status?: string) => {
    switch (status) {
      case 'deployed':
        return { bg: 'bg-emerald-500/10', dot: 'bg-emerald-500', pulse: true };
      case 'failed':
        return { bg: 'bg-red-500/10', dot: 'bg-red-500', pulse: false };
      case 'building':
      case 'deploying':
      case 'cloning':
        return { bg: 'bg-amber-500/10', dot: 'bg-amber-500', pulse: true };
      default:
        return { bg: 'bg-zinc-500/10', dot: 'bg-zinc-500', pulse: false };
    }
  };

  const statusConfig = getStatusConfig(currentStatus);

  const handleDelete = async () => {
    try {
      await deleteApplication({
        id: application?.id || ''
      }).unwrap();
      toast.success(t('selfHost.applicationDetails.header.actions.delete.success'));
      router.push('/self-host');
    } catch (error) {
      toast.error(t('selfHost.applicationDetails.header.actions.delete.error'));
    }
  };

  const handleRestart = async () => {
    try {
      await restartApplication({ id: latestDeployment?.id || '' }).unwrap();
      toast.success(t('selfHost.applicationDetails.header.actions.restart.success'));
    } catch (error) {
      toast.error(t('selfHost.applicationDetails.header.actions.restart.error'));
    }
  };

  const handleRedeploy = async (forceWithoutCache: boolean) => {
    try {
      await redeployApplication({
        id: application?.id || '',
        force: true,
        force_without_cache: forceWithoutCache
      }).unwrap();
      router.push('/self-host/application/' + application?.id + '?logs=true');
      toast.success(t('selfHost.applicationDetails.header.actions.redeploy.success'));
    } catch (error) {
      toast.error(t('selfHost.applicationDetails.header.actions.redeploy.error'));
    }
  };

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={null}>
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-4">
          <div
            className={cn('w-12 h-12 rounded-xl flex items-center justify-center', statusConfig.bg)}
          >
            <div
              className={cn(
                'w-3 h-3 rounded-full',
                statusConfig.dot,
                statusConfig.pulse && 'animate-pulse'
              )}
            />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold tracking-tight capitalize">{application?.name}</h1>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-muted-foreground hover:text-foreground"
                onClick={() => window.open('https://' + application?.domain, '_blank')}
                aria-label={t('selfHost.applicationDetails.header.actions.open')}
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            </div>
            <div className="flex items-center gap-2 mt-1">
              <a
                href={'https://' + application?.domain}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-muted-foreground hover:text-foreground font-mono bg-muted px-2 py-0.5 rounded transition-colors"
              >
                {application?.domain}
              </a>
              <Badge
                variant="outline"
                className={cn(
                  'text-xs capitalize',
                  application?.environment === 'production'
                    ? 'border-emerald-500/30 text-emerald-500 bg-emerald-500/10'
                    : application?.environment === 'staging'
                      ? 'border-amber-500/30 text-amber-500 bg-amber-500/10'
                      : 'border-blue-500/30 text-blue-500 bg-blue-500/10'
                )}
              >
                {application?.environment}
              </Badge>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
            <Button
              variant="outline"
              size="sm"
              disabled={isRestarting}
              onClick={handleRestart}
              className="gap-2"
            >
              <RotateCcw className={cn('h-4 w-4', isRestarting && 'animate-spin')} />
              {t('selfHost.applicationDetails.header.actions.restart.button')}
            </Button>
          </AnyPermissionGuard>

          <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
            <Button
              variant="default"
              size="sm"
              disabled={isRedeploying}
              onClick={() => handleRedeploy(true)}
              className="gap-2"
            >
              <Rocket className="h-4 w-4" />
              {t('selfHost.applicationDetails.header.actions.redeploy.button')}
            </Button>
          </AnyPermissionGuard>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon" className="h-9 w-9">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
                <DropdownMenuItem
                  onClick={() => handleRedeploy(false)}
                  disabled={isRedeploying}
                  className="gap-2"
                >
                  <RefreshCw className="h-4 w-4" />
                  {t('selfHost.applicationDetails.header.actions.redeploy.forceButton')}
                </DropdownMenuItem>
              </AnyPermissionGuard>
              <AnyPermissionGuard permissions={['deploy:delete']} loadingFallback={null}>
                <DropdownMenuSeparator />
                <DeleteDialog
                  title={t(
                    'selfHost.applicationDetails.header.actions.delete.dialog.title'
                  ).replace('{name}', application?.name || '')}
                  description={t(
                    'selfHost.applicationDetails.header.actions.delete.dialog.description'
                  ).replace('{name}', application?.name || '')}
                  onConfirm={handleDelete}
                  trigger={
                    <DropdownMenuItem
                      onSelect={(e) => e.preventDefault()}
                      className="gap-2 text-red-500 focus:text-red-500"
                    >
                      <Trash2 className="h-4 w-4" />
                      {t('selfHost.applicationDetails.header.actions.delete.button')}
                    </DropdownMenuItem>
                  }
                  isDeleting={isDeleting}
                  variant="destructive"
                  icon={Trash2}
                />
              </AnyPermissionGuard>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </ResourceGuard>
  );
};

export default ApplicationDetailsHeader;
