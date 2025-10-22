'use client';
import React from 'react';
import { Button } from '@/components/ui/button';
import { ExternalLink, MoreVertical, RotateCcw, TrashIcon } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipProvider, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip';
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

const ApplicationDetailsHeader = ({ application }: { application?: Application }) => {
  const { t } = useTranslation();
  const [redeployApplication, { isLoading: isRedeploying }] = useRedeployApplicationMutation();
  const [deleteApplication, { isLoading: isDeleting }] = useDeleteApplicationMutation();
  const router = useRouter();
  const [restartApplication, { isLoading: isRestarting }] = useRestartApplicationMutation();

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
      await restartApplication({ id: application?.deployments?.[0]?.id || '' }).unwrap();
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
      <div className="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
        <div className="flex items-start">
          <div className="mr-2">
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold capitalize sm:text-3xl">{application?.name}</h1>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => window.open('https://' + application?.domain, '_blank')}
                aria-label={t('selfHost.applicationDetails.header.actions.open')}
              >
                <ExternalLink className="h-5 w-5" />
              </Button>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="secondary"
                    size="icon"
                    disabled={isRestarting}
                    onClick={handleRestart}
                  >
                    <RotateCcw className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{t('selfHost.applicationDetails.header.actions.restart.button')}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </AnyPermissionGuard>
          <AnyPermissionGuard permissions={['deploy:delete']} loadingFallback={null}>
            <DeleteDialog
              title={t('selfHost.applicationDetails.header.actions.delete.dialog.title').replace(
                '{name}',
                application?.name || ''
              )}
              description={t(
                'selfHost.applicationDetails.header.actions.delete.dialog.description'
              ).replace('{name}', application?.name || '')}
              onConfirm={handleDelete}
              trigger={
                <Button variant="outline" size="icon">
                  <TrashIcon className="h-4 w-4" />
                </Button>
              }
              isDeleting={isDeleting}
              variant="destructive"
              icon={TrashIcon}
            />
          </AnyPermissionGuard>
          <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="icon">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => handleRedeploy(true)} disabled={isRedeploying}>
                  {t('selfHost.applicationDetails.header.actions.redeploy.button')}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => handleRedeploy(false)} disabled={isRedeploying}>
                  {t('selfHost.applicationDetails.header.actions.redeploy.forceButton')}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </AnyPermissionGuard>
        </div>
      </div>
    </ResourceGuard>
  );
};

export default ApplicationDetailsHeader;
