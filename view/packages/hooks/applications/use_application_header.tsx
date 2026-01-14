import React, { useState, useRef, useEffect, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { Application } from '@/redux/types/applications';
import {
  useRedeployApplicationMutation,
  useRestartApplicationMutation,
  useDeleteApplicationMutation,
  useUpdateApplicationLabelsMutation,
  useDeployProjectMutation
} from '@/redux/services/deploy/applicationsApi';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { ExternalLink, RotateCcw, Trash2, Rocket, RefreshCw, X, Plus } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { MoreVertical } from 'lucide-react';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import {
  ProjectFamilySwitcher,
  DuplicateProjectDialog
} from '@/packages/components/application-details';
import { DomainDropdown } from '@/packages/components/multi-domains';
import { AnyPermissionGuard } from '@/packages/components/rbac';
import { useTranslation } from '../shared/use-translation';

interface UseApplicationHeaderProps {
  application?: Application;
}

const getStatusConfig = (status?: string) => {
  switch (status) {
    case 'deployed':
      return { bg: 'bg-emerald-500/10', dot: 'bg-emerald-500', pulse: true };
    case 'running':
      return { bg: 'bg-emerald-500/10', dot: 'bg-emerald-500', pulse: true };
    case 'failed':
      return { bg: 'bg-red-500/10', dot: 'bg-red-500', pulse: false };
    case 'building':
    case 'deploying':
    case 'cloning':
    case 'started':
      return { bg: 'bg-amber-500/10', dot: 'bg-amber-500', pulse: true };
    case 'draft':
      return { bg: 'bg-blue-500/10', dot: 'bg-blue-500', pulse: false };
    case 'stopped':
      return { bg: 'bg-zinc-500/10', dot: 'bg-zinc-500', pulse: false };
    default:
      return { bg: 'bg-zinc-500/10', dot: 'bg-zinc-500', pulse: false };
  }
};

interface HeaderLabelBadgeProps {
  label: string;
  onRemove: () => void;
}

function HeaderLabelBadge({ label, onRemove }: HeaderLabelBadgeProps) {
  const [isHovered, setIsHovered] = useState(false);

  return (
    <Badge
      variant="outline"
      className={cn(
        'text-xs px-2 py-0.5 gap-1 relative',
        'transition-all duration-200',
        'border-violet-500/30 text-violet-500 bg-violet-500/10',
        'pr-1.5'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <span className={cn('transition-opacity duration-200', isHovered && 'opacity-70')}>
        {label}
      </span>
      <button
        onClick={(e) => {
          e.stopPropagation();
          onRemove();
        }}
        className={cn(
          'transition-all duration-200',
          'hover:text-destructive',
          'flex items-center justify-center',
          isHovered ? 'opacity-100 scale-100' : 'opacity-0 scale-75 w-0'
        )}
      >
        <X size={12} />
      </button>
    </Badge>
  );
}

export function useApplicationHeader({ application }: UseApplicationHeaderProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const [redeployApplication, { isLoading: isRedeploying }] = useRedeployApplicationMutation();
  const [deleteApplication, { isLoading: isDeleting }] = useDeleteApplicationMutation();
  const [updateLabels, { isLoading: isUpdatingLabels }] = useUpdateApplicationLabelsMutation();
  const [deployProject, { isLoading: isDeployingProject }] = useDeployProjectMutation();
  const [restartApplication, { isLoading: isRestarting }] = useRestartApplicationMutation();

  const [isAddingLabel, setIsAddingLabel] = useState(false);
  const [newLabel, setNewLabel] = useState('');
  const labelInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isAddingLabel && labelInputRef.current) {
      labelInputRef.current.focus();
    }
  }, [isAddingLabel]);

  const latestDeployment = application?.deployments?.[0];
  const currentStatus = latestDeployment?.status?.status || application?.status?.status;
  const isDraft = currentStatus === 'draft';
  const statusConfig = getStatusConfig(currentStatus);

  const handleDeployProject = async () => {
    if (!application?.id) return;
    try {
      await deployProject({ id: application.id }).unwrap();
      toast.success(t('selfHost.applicationDetails.header.actions.redeploy.success'));
      router.push('/self-host/application/' + application.id + '?logs=true');
    } catch {
      toast.error(t('selfHost.applicationDetails.header.actions.redeploy.error'));
    }
  };

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

  const handleRemoveLabel = async (labelToRemove: string) => {
    if (!application?.id || !application?.labels) return;
    const updated = application.labels.filter((l) => l !== labelToRemove);
    await updateLabels({
      id: application.id,
      labels: updated
    }).unwrap();
  };

  const handleAddLabel = async () => {
    const value = newLabel.trim();
    if (!value || !application?.id) {
      setIsAddingLabel(false);
      setNewLabel('');
      return;
    }
    const currentLabels = application.labels || [];
    const updated = [...new Set([...currentLabels, value])];
    await updateLabels({
      id: application.id,
      labels: updated
    }).unwrap();
    setIsAddingLabel(false);
    setNewLabel('');
  };

  const handleLabelKeyDown = async (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      await handleAddLabel();
    } else if (e.key === 'Escape') {
      setIsAddingLabel(false);
      setNewLabel('');
    }
  };

  const icon = useMemo(
    () => (
      <div className={cn('w-12 h-12 rounded-xl flex items-center justify-center', statusConfig.bg)}>
        <div
          className={cn(
            'w-3 h-3 rounded-full',
            statusConfig.dot,
            statusConfig.pulse && 'animate-pulse'
          )}
        />
      </div>
    ),
    [statusConfig]
  );

  const title = useMemo(
    () => (
      <div className="flex items-center gap-2">
        <span className="capitalize">{application?.name}</span>
        {application && <ProjectFamilySwitcher application={application} />}
        {application && <DomainDropdown domains={application.domains} variant="icon" />}
      </div>
    ),
    [application]
  );

  const metadata = useMemo(
    () => (
      <div className="flex flex-wrap items-center gap-2">
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
        {application?.labels && application.labels.length > 0 && (
          <>
            {application.labels.map((label, index) => (
              <HeaderLabelBadge
                key={index}
                label={label}
                onRemove={() => handleRemoveLabel(label)}
              />
            ))}
          </>
        )}
        <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
          {isAddingLabel ? (
            <Input
              ref={labelInputRef}
              value={newLabel}
              onChange={(e) => setNewLabel(e.target.value)}
              onKeyDown={handleLabelKeyDown}
              onBlur={handleAddLabel}
              className="h-5 w-24 text-xs px-2 py-0"
              placeholder="New label"
              disabled={isUpdatingLabels}
            />
          ) : (
            <button
              type="button"
              onClick={() => setIsAddingLabel(true)}
              className="inline-flex items-center gap-1 h-5 px-2 rounded-md border border-dashed border-muted-foreground/40 text-xs text-muted-foreground hover:bg-muted hover:border-muted-foreground/60 transition-colors"
            >
              <Plus size={10} />
              Add
            </button>
          )}
        </AnyPermissionGuard>
      </div>
    ),
    [application, isAddingLabel, newLabel, isUpdatingLabels, handleLabelKeyDown, handleAddLabel]
  );

  const primaryActions = useMemo(() => {
    if (isDraft) {
      return (
        <AnyPermissionGuard permissions={['deploy:update']} loadingFallback={null}>
          <Button
            variant="default"
            size="sm"
            disabled={isDeployingProject}
            onClick={handleDeployProject}
            className="gap-2"
          >
            <Rocket className="h-4 w-4" />
            {isDeployingProject ? 'Deploying...' : 'Deploy Now'}
          </Button>
        </AnyPermissionGuard>
      );
    }

    return (
      <>
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
      </>
    );
  }, [
    isDraft,
    isDeployingProject,
    isRestarting,
    isRedeploying,
    t,
    handleDeployProject,
    handleRestart,
    handleRedeploy
  ]);

  const dropdownMenuItems = useMemo(
    () => (
      <>
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
        <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
          {application && <DuplicateProjectDialog application={application} />}
        </AnyPermissionGuard>
        <AnyPermissionGuard permissions={['deploy:delete']} loadingFallback={null}>
          <DropdownMenuSeparator />
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
      </>
    ),
    [application, isRedeploying, isDeleting, t, handleRedeploy, handleDelete]
  );

  const actions = useMemo(
    () => (
      <div className="flex items-center gap-2">
        {primaryActions}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="icon" className="h-9 w-9">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            {dropdownMenuItems}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
    [primaryActions, dropdownMenuItems]
  );

  return {
    icon,
    title,
    metadata,
    actions,
    primaryActions,
    dropdownMenuItems
  };
}
