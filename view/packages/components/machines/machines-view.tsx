'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import PageLayout from '@/packages/layouts/page-layout';
import {
  TypographyMuted,
  CardWrapper,
  CardTitle,
  SearchBar,
  PaginationWrapper,
  Skeleton,
  DataTable,
  Badge,
  Button,
  Input,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@nixopus/ui';
import {
  Loader2,
  AlertCircle,
  Plus,
  RotateCw,
  LayoutGrid,
  List,
  MoreVertical,
  RefreshCw,
  Pause,
  Play,
  HardDrive,
  Trash2,
  ChartColumnDecreasing,
  DatabaseBackup,
  Layers,
  Pencil
} from 'lucide-react';
import {
  useDeleteMachineMutation,
  useRenameMachineMutation
} from '@/redux/services/servers/serversApi';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { toast } from 'sonner';
import { AddMachineWizard } from '@/packages/components/machines/add-machine-wizard';
import { useMachineCard } from '@/packages/hooks/machines/use-machine-card';
import { useMachinesView } from '@/packages/hooks/machines/use-machines-view';
import { useMachineLifecycle } from '@/packages/hooks/machines/use-machine-lifecycle';
import { useMachineBackup } from '@/packages/hooks/backups/use-machine-backup';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import type { Machine } from '@/packages/hooks/machines/use-machines';
import type { LiveState } from '@/packages/hooks/machines/use-machine-lifecycle';

export function formatRam(mb: number): string {
  return mb >= 1024 ? `${(mb / 1024).toFixed(0)} GB` : `${mb} MB`;
}

const stateConfig: Record<
  LiveState,
  { color: string; pulse: boolean; label: string; useBadge?: boolean }
> = {
  Running: { color: 'bg-green-500', pulse: true, label: 'Running', useBadge: true },
  Paused: { color: 'bg-yellow-500', pulse: false, label: 'Paused' },
  Stopped: { color: 'bg-red-500', pulse: false, label: 'Stopped' },
  unknown: { color: 'bg-gray-400', pulse: false, label: 'Unknown' }
};

const MachineStateIndicator = React.memo(function MachineStateIndicator({
  state
}: {
  state: LiveState;
}) {
  const config = stateConfig[state];

  if (config.useBadge) {
    return (
      <Badge
        variant="secondary"
        className="bg-green-500/10 text-green-200 uppercase border-green-500/20 text-xs rounded-full"
      >
        {config.label}
      </Badge>
    );
  }

  return (
    <div className="flex items-center gap-1.5 shrink-0" title={config.label}>
      <span className="text-xs text-muted-foreground">{config.label}</span>
    </div>
  );
});

interface MachineActionsDropdownProps {
  liveState: LiveState;
  isActionInFlight: boolean;
  onAction: (action: 'restart' | 'pause' | 'resume') => void;
  onBackup?: () => void;
  isBackupInProgress?: boolean;
  onDelete?: () => void;
  isDeleting?: boolean;
  machineLabel: string;
  /** Extra menu items rendered after the backup action (used by premium for "Upgrade" link) */
  extraMenuItems?: React.ReactNode;
}

const MachineActionsDropdown = React.memo(function MachineActionsDropdown({
  liveState,
  isActionInFlight,
  onAction,
  onBackup,
  isBackupInProgress,
  onDelete,
  isDeleting,
  machineLabel,
  extraMenuItems
}: MachineActionsDropdownProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 shrink-0"
          disabled={isActionInFlight}
          onPointerDown={(e: React.PointerEvent) => e.stopPropagation()}
          onClick={(e: React.MouseEvent) => e.stopPropagation()}
          aria-label={`Actions for ${machineLabel}`}
        >
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        onPointerDown={(e: React.PointerEvent) => e.stopPropagation()}
        onClick={(e: React.MouseEvent) => e.stopPropagation()}
      >
        <DropdownMenuItem onClick={() => onAction('restart')}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Restart
        </DropdownMenuItem>
        {onBackup && (
          <DropdownMenuItem onClick={onBackup} disabled={isActionInFlight || isBackupInProgress}>
            <HardDrive className="mr-2 h-4 w-4" />
            {isBackupInProgress ? 'Backup in progress...' : 'Backup'}
          </DropdownMenuItem>
        )}
        {extraMenuItems}
        {liveState === 'Running' && (
          <DropdownMenuItem onClick={() => onAction('pause')} disabled={isActionInFlight}>
            <Pause className="mr-2 h-4 w-4" />
            Pause
          </DropdownMenuItem>
        )}
        {liveState === 'Paused' && (
          <DropdownMenuItem onClick={() => onAction('resume')} disabled={isActionInFlight}>
            <Play className="mr-2 h-4 w-4" />
            Resume
          </DropdownMenuItem>
        )}
        {onDelete && (
          <DropdownMenuItem
            onClick={onDelete}
            disabled={isDeleting}
            className="text-destructive focus:text-destructive"
          >
            <Trash2 className="mr-2 h-4 w-4" />
            {isDeleting ? 'Removing...' : 'Remove'}
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
});

const actionLabels: Record<string, string> = {
  restart: 'Restarting...',
  pause: 'Pausing...',
  resume: 'Resuming...'
};

const MachineActionOverlay = React.memo(function MachineActionOverlay({
  activeAction
}: {
  activeAction: string | null;
}) {
  if (!activeAction) return null;

  return (
    <div className="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-background/60 backdrop-blur-[1px]">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span>{actionLabels[activeAction] || 'Processing...'}</span>
      </div>
    </div>
  );
});

const InlineEditName = React.memo(function InlineEditName({
  machineId,
  currentName,
  className
}: {
  machineId: string;
  currentName: string;
  className?: string;
}) {
  const { t } = useTranslation();
  const [isEditing, setIsEditing] = React.useState(false);
  const [isSaving, setIsSaving] = React.useState(false);
  const [value, setValue] = React.useState(currentName);
  const [renameMachine, { isLoading }] = useRenameMachineMutation();
  const inputRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    setValue(currentName);
  }, [currentName]);

  React.useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [isEditing]);

  const handleSave = async () => {
    if (isSaving) return;

    const trimmed = value.trim();
    if (!trimmed || trimmed === currentName) {
      setValue(currentName);
      setIsEditing(false);
      return;
    }

    setIsSaving(true);
    try {
      await renameMachine({ id: machineId, name: trimmed }).unwrap();
      setIsEditing(false);
    } catch {
      toast.error(t('machines.renameError'));
      setValue(currentName);
      setIsEditing(false);
    } finally {
      setIsSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (isSaving) return;

    if (e.key === 'Enter') {
      e.preventDefault();
      handleSave();
    }
    if (e.key === 'Escape') {
      setValue(currentName);
      setIsEditing(false);
    }
  };

  if (isEditing) {
    return (
      <div
        data-machine-action-zone="true"
        onPointerDown={(e) => e.stopPropagation()}
        onClick={(e) => e.stopPropagation()}
      >
        <Input
          ref={inputRef}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onBlur={handleSave}
          onKeyDown={(e) => {
            e.stopPropagation();
            handleKeyDown(e);
          }}
          disabled={isLoading || isSaving}
          className={`h-7 text-base font-semibold px-1.5 py-0 ${className || ''}`}
          maxLength={255}
        />
      </div>
    );
  }

  return (
    <Button
      variant="ghost"
      data-machine-action-zone="true"
      onPointerDown={(e) => e.stopPropagation()}
      onClick={(e) => {
        e.stopPropagation();
        setIsEditing(true);
      }}
      className={`group/edit inline-flex items-center gap-1.5 text-left cursor-pointer hover:opacity-80 transition-opacity h-auto p-0 ${className || ''}`}
    >
      <span
        className="text-base font-semibold"
        style={{ wordBreak: 'break-word', overflowWrap: 'break-word' }}
      >
        {currentName}
      </span>
      <Pencil className="h-3 w-3 shrink-0 opacity-0 group-hover/edit:opacity-60 transition-opacity" />
    </Button>
  );
});

const MACHINE_NAV_OPTIONS = [
  { label: 'Charts', icon: ChartColumnDecreasing, path: 'charts' },
  { label: 'Apps', icon: Layers, path: 'apps' },
  { label: 'Backups', icon: DatabaseBackup, path: 'backups' }
];

export const MachineCard = React.memo(function MachineCard({ machine }: { machine: Machine }) {
  const router = useRouter();
  const { isInteractive, isFailed } = useMachineCard(machine);
  const machineLabel = machine.name;
  const isProvisioning = machine.status === 'provisioning';
  const isActive = machine.status === 'active' || (!isProvisioning && !isFailed);

  const { liveState, isActionInFlight, activeAction, handleAction } = useMachineLifecycle(
    !isActive,
    machine.ssh_key_id
  );
  const { handleTriggerBackup, hasInProgress: isBackupInProgress } = useMachineBackup(
    machine.ssh_key_id
  );
  const [deleteMachine, { isLoading: isDeleting }] = useDeleteMachineMutation();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isNavOpen, setIsNavOpen] = useState(false);

  const handleDeleteConfirm = React.useCallback(async () => {
    try {
      await deleteMachine(machine.ssh_key_id).unwrap();
      toast.success('Machine removed');
      setIsDeleteDialogOpen(false);
    } catch (e: unknown) {
      const err = e as { data?: { error?: string } };
      toast.error(err?.data?.error || 'Failed to remove machine');
    }
  }, [deleteMachine, machine.ssh_key_id]);

  const handleCardClick = React.useCallback((event: React.MouseEvent<HTMLElement>) => {
    const target = event.target as HTMLElement;
    if (target.closest('[data-machine-action-zone="true"]')) {
      return;
    }
    setIsNavOpen(true);
  }, []);

  return (
    <>
      <DropdownMenu open={isNavOpen} onOpenChange={setIsNavOpen} modal={false}>
        <DropdownMenuTrigger asChild>
          <CardWrapper
            onClick={handleCardClick}
            className={`group relative rounded-md p-5 w-full overflow-hidden flex min-h-32 md:min-h-44 h-44 flex-col border-black-900/70 transition-colors cursor-pointer ${!isInteractive ? 'opacity-75' : ''} ${isFailed ? 'border-destructive/30' : ''}`}
            header={
              <div className="flex-1 min-w-0 w-full space-y-1.5">
                <div className="flex w-full min-w-0 items-start justify-between gap-2">
                  <div className="min-w-0 flex-1 max-w-full">
                    <div className="flex items-start gap-2 flex-wrap">
                      <div className="min-w-0 max-w-full">
                        <InlineEditName machineId={machine.ssh_key_id} currentName={machineLabel} />
                      </div>
                    </div>
                  </div>
                  <div
                    data-machine-action-zone="true"
                    className="flex items-center gap-1 ml-auto shrink-0"
                    onPointerDown={(e: React.PointerEvent) => e.stopPropagation()}
                    onClick={(e) => e.stopPropagation()}
                  >
                    {isActive && (
                      <MachineActionsDropdown
                        liveState={liveState}
                        isActionInFlight={isActionInFlight}
                        onAction={handleAction}
                        onBackup={handleTriggerBackup}
                        isBackupInProgress={isBackupInProgress}
                        onDelete={() => setIsDeleteDialogOpen(true)}
                        isDeleting={isDeleting}
                        machineLabel={machineLabel}
                      />
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {isActive ? (
                    <MachineStateIndicator state={liveState} />
                  ) : (
                    <MachineStatusBadge status={machine.status} />
                  )}
                </div>
                {machine.is_active === false && machine.status === 'active' && (
                  <Badge variant="outline" className="text-amber-500 border-amber-500/30 text-xs">
                    Offline
                  </Badge>
                )}
              </div>
            }
            headerClassName="w-full min-w-0 pb-0 px-0"
            contentClassName="space-y-1.5"
          >
            <MachineActionOverlay activeAction={activeAction} />
            <MachineCardContent status={machine.status} />
            {(machine.total_vcpu > 0 || machine.total_ram_mb > 0 || machine.total_disk_gb > 0) && (
              <div className="flex items-center gap-3 pt-1.5 border-t border-border/50 flex-wrap">
                {machine.total_vcpu > 0 && (
                  <span className="text-xs text-muted-foreground">{machine.total_vcpu} CPU</span>
                )}
                {machine.total_ram_mb > 0 && (
                  <span className="text-xs text-muted-foreground">
                    {formatRam(machine.total_ram_mb)} RAM
                  </span>
                )}
                {machine.total_disk_gb > 0 && (
                  <span className="text-xs text-muted-foreground">
                    {machine.total_disk_gb} GB Disk
                  </span>
                )}
              </div>
            )}
          </CardWrapper>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-44">
          {MACHINE_NAV_OPTIONS.map((opt) => (
            <DropdownMenuItem
              key={opt.path}
              onClick={() => router.push(`/machines/${machine.ssh_key_id}/${opt.path}`)}
            >
              <opt.icon className="h-4 w-4 mr-2" />
              {opt.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      <DeleteDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        onConfirm={handleDeleteConfirm}
        isDeleting={isDeleting}
        title="Remove Machine"
        description={`Are you sure you want to remove "${machineLabel}"? This action cannot be undone.`}
        confirmText="Remove"
        variant="destructive"
        icon={Trash2}
      />
    </>
  );
});

const MachineStatusBadge = React.memo(function MachineStatusBadge({
  status
}: {
  status?: Machine['status'];
}) {
  if (status === 'provisioning') {
    return (
      <div className="flex items-center gap-2 text-xs text-muted-foreground shrink-0">
        <Loader2 className="h-3 w-3 animate-spin text-primary" />
        <span>Provisioning</span>
      </div>
    );
  }

  if (status === 'failed') {
    return (
      <div className="flex items-center gap-2 text-xs text-destructive shrink-0">
        <AlertCircle className="h-3 w-3" />
        <span>Failed</span>
      </div>
    );
  }

  return null;
});

const MachineCardContent = React.memo(function MachineCardContent({
  status
}: {
  status?: Machine['status'];
}) {
  if (status === 'provisioning') {
    return (
      <TypographyMuted className="text-xs text-muted-foreground">
        Your machine is being set up. This may take a few moments...
      </TypographyMuted>
    );
  }

  if (status === 'failed') {
    return (
      <TypographyMuted className="text-xs text-destructive/80">
        Machine setup failed. It will be retried automatically.
      </TypographyMuted>
    );
  }

  return null;
});

function MachinesSkeleton({
  showHeader = true,
  showSearch = true
}: {
  showHeader?: boolean;
  showSearch?: boolean;
}) {
  return (
    <div>
      {showHeader && (
        <>
          <Skeleton className="h-8 w-36 mb-4" />
          {showSearch && (
            <div className="flex items-center justify-between flex-wrap gap-3 pb-4 pt-6">
              <Skeleton className="h-9 w-64" />
              <Skeleton className="h-8 w-8 rounded" />
            </div>
          )}
        </>
      )}
      {!showHeader && showSearch && (
        <div className="mb-6">
          <Skeleton className="h-9 w-64" />
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <CardWrapper
            key={i}
            className="group relative rounded-md p-5 w-full overflow-hidden flex min-h-32 md:min-h-44 h-44 flex-col border-black-900/70"
            header={
              <div className="flex-1 min-w-0 w-full space-y-1.5">
                <div className="flex w-full min-w-0 items-start justify-between gap-2">
                  <Skeleton className="h-5 w-40" />
                </div>
                <Skeleton className="h-3.5 w-16" />
              </div>
            }
            headerClassName="w-full min-w-0 pb-0 px-0"
            contentClassName="space-y-1.5"
          >
            <div className="flex items-center gap-3 pt-1.5 border-t border-border/50">
              <Skeleton className="h-3.5 w-12" />
              <Skeleton className="h-3.5 w-16" />
              <Skeleton className="h-3.5 w-16" />
            </div>
          </CardWrapper>
        ))}
      </div>
    </div>
  );
}

const ProvisioningMachineCard = React.memo(function ProvisioningMachineCard({
  progress = 0
}: {
  progress?: number;
}) {
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <CardWrapper
      className="group relative rounded-md p-5 w-full overflow-hidden flex min-h-32 md:min-h-44 h-44 flex-col border-black-900/70 cursor-default opacity-75"
      header={
        <div className="flex-1 min-w-0 w-full space-y-1.5">
          <div className="flex w-full min-w-0 items-start justify-between gap-2">
            <div className="min-w-0 flex-1 max-w-full">
              <CardTitle
                className="text-base font-semibold wrap-break-word max-w-full"
                style={{ wordBreak: 'break-word', overflowWrap: 'break-word', maxWidth: '100%' }}
              >
                Domain pending
              </CardTitle>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2 text-xs text-muted-foreground shrink-0">
              <Loader2 className="h-3 w-3 animate-spin text-primary" />
              <span>Provisioning</span>
              {clampedProgress > 0 && (
                <span className="text-primary font-medium">{clampedProgress}%</span>
              )}
            </div>
          </div>
        </div>
      }
      headerClassName="w-full min-w-0 pb-0 px-0"
      contentClassName="relative flex-1"
    >
      <div
        className="absolute bottom-0 left-0 right-0 bg-primary/10 transition-[height] duration-700 ease-in-out rounded-b-md"
        style={{ height: `${clampedProgress}%` }}
      />
    </CardWrapper>
  );
});

interface MachinesEmptyStateProps {
  isProvisioning?: boolean;
  provisioningHasError?: boolean;
  onRetry?: () => void;
  canRetry?: boolean;
  isRetrying?: boolean;
  retryCount?: number;
  maxRetries?: number;
}

function MachinesEmptyState({
  isProvisioning,
  provisioningHasError,
  onRetry,
  canRetry = true,
  isRetrying,
  retryCount = 0,
  maxRetries = 3
}: MachinesEmptyStateProps) {
  if (isProvisioning || isRetrying) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] gap-3">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <TypographyMuted className="text-sm text-muted-foreground">
          Setting up your machine...
        </TypographyMuted>
      </div>
    );
  }

  if (provisioningHasError) {
    const exhaustedRetries = retryCount >= maxRetries;
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] gap-3">
        <AlertCircle className="h-8 w-8 text-destructive" />
        <TypographyMuted className="text-sm text-muted-foreground text-center max-w-md">
          {exhaustedRetries
            ? 'Machine provisioning failed after multiple attempts. Please '
            : 'Machine provisioning failed. You can retry or '}
          <a
            href="mailto:raghav@nixopus.com?subject=Machine%20Provisioning%20Issue"
            className="text-primary hover:underline"
          >
            contact support
          </a>
          {exhaustedRetries ? '.' : ' if the issue persists.'}
        </TypographyMuted>
        {onRetry && canRetry && (
          <button
            onClick={onRetry}
            className="mt-2 px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Retry Provisioning ({maxRetries - retryCount} left)
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] gap-3">
      <TypographyMuted className="text-lg">No machines yet</TypographyMuted>
      <TypographyMuted className="text-sm text-muted-foreground text-center max-w-md">
        Machines are your servers where apps are deployed and run. Get started by adding your first
        machine.
      </TypographyMuted>
    </div>
  );
}

function AddMachineCard({ onClick }: { onClick?: () => void }) {
  return (
    <button
      onClick={onClick}
      className="group relative w-full h-full min-h-[120px] rounded-lg border-2 border-dashed border-border/50 hover:border-primary/40 hover:bg-muted/50 transition-colors cursor-pointer flex flex-col items-center justify-center gap-2 p-6"
    >
      <div className="h-10 w-10 rounded-full bg-muted group-hover:bg-primary/10 flex items-center justify-center transition-colors">
        <Plus className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
      </div>
      <span className="text-sm font-medium text-muted-foreground group-hover:text-foreground transition-colors">
        Add Machine
      </span>
    </button>
  );
}

const MachineRowActions = React.memo(function MachineRowActions({ machine }: { machine: Machine }) {
  const { isFailed } = useMachineCard(machine);
  const isProvisioning = machine.status === 'provisioning';
  const isActive = machine.status === 'active' || (!isProvisioning && !isFailed);
  const { liveState, isActionInFlight, handleAction } = useMachineLifecycle(
    !isActive,
    machine.ssh_key_id
  );
  const { handleTriggerBackup, hasInProgress: isBackupInProgress } = useMachineBackup(
    machine.ssh_key_id
  );

  if (!isActive) return null;

  return (
    <div className="flex items-center justify-end" onClick={(e) => e.stopPropagation()}>
      <MachineActionsDropdown
        liveState={liveState}
        isActionInFlight={isActionInFlight}
        onAction={handleAction}
        onBackup={handleTriggerBackup}
        isBackupInProgress={isBackupInProgress}
        machineLabel={machine.name}
      />
    </div>
  );
});

const MachineRowStatus = React.memo(function MachineRowStatus({ machine }: { machine: Machine }) {
  const { isFailed } = useMachineCard(machine);
  const isProvisioning = machine.status === 'provisioning';
  const isActive = machine.status === 'active' || (!isProvisioning && !isFailed);
  const { liveState } = useMachineLifecycle(!isActive, machine.ssh_key_id);

  if (isActive) return <MachineStateIndicator state={liveState} />;
  if (isProvisioning) {
    return (
      <Badge variant="secondary" className="rounded-sm opacity-80">
        <Loader2 className="h-3 w-3 animate-spin mr-1" />
        Provisioning
      </Badge>
    );
  }
  if (isFailed) {
    return (
      <Badge variant="destructive" className="rounded-sm opacity-80">
        Failed
      </Badge>
    );
  }
  return null;
});

const MachineTableName = React.memo(function MachineTableName({ machine }: { machine: Machine }) {
  return (
    <InlineEditName
      machineId={machine.ssh_key_id}
      currentName={machine.name}
      className="max-w-[260px]"
    />
  );
});

const MACHINE_TABLE_COLUMNS = [
  {
    key: 'name',
    title: 'Name',
    render: (_: unknown, machine: Machine) => <MachineTableName machine={machine} />
  },
  {
    key: 'status',
    title: 'Status',
    render: (_: unknown, machine: Machine) => <MachineRowStatus machine={machine} />
  },
  {
    key: 'cpu',
    title: 'CPU',
    render: (_: unknown, machine: Machine) => (
      <span className="text-sm text-muted-foreground tabular-nums">
        {machine.total_vcpu > 0 ? `${machine.total_vcpu} vCPU` : '—'}
      </span>
    )
  },
  {
    key: 'ram',
    title: 'RAM',
    render: (_: unknown, machine: Machine) => (
      <span className="text-sm text-muted-foreground tabular-nums">
        {machine.total_ram_mb > 0 ? formatRam(machine.total_ram_mb) : '—'}
      </span>
    )
  },
  {
    key: 'disk',
    title: 'Disk',
    render: (_: unknown, machine: Machine) => (
      <span className="text-sm text-muted-foreground tabular-nums">
        {machine.total_disk_gb > 0 ? `${machine.total_disk_gb} GB` : '—'}
      </span>
    )
  },
  {
    key: 'actions',
    title: '',
    render: (_: unknown, machine: Machine) => <MachineRowActions machine={machine} />
  }
];

function MachinesTable({
  machines,
  tableClassName
}: {
  machines: Machine[];
  tableClassName?: string;
}) {
  const router = useRouter();
  return (
    <div className={`overflow-hidden rounded-md border ${tableClassName || ''}`}>
      <DataTable
        data={machines}
        columns={MACHINE_TABLE_COLUMNS}
        showBorder={false}
        onRowClick={(machine: Machine) => router.push(`/machines/${machine.ssh_key_id}/charts`)}
        hoverable
      />
    </div>
  );
}

export interface MachinesViewProps {
  showHeader?: boolean;
  showSearch?: boolean;
  headerTitle?: string;
  searchLabel?: string;
  className?: string;
  isLoading?: boolean;
  isProvisioning?: boolean;
  provisioningHasError?: boolean;
  onRetryProvisioning?: () => void;
  canRetry?: boolean;
  isRetrying?: boolean;
  retryCount?: number;
  maxRetries?: number;
  provisioningProgress?: number;
  tableClassName?: string;
  /** Override the default data hook. Premium injects its billing-gated version here. */
  useDataHook?: (isLoadingProp?: boolean) => ReturnType<typeof useMachinesView>;
}

export function MachinesView({
  showHeader = true,
  showSearch = true,
  headerTitle = 'Machines',
  searchLabel = 'Search machines...',
  className = '',
  isLoading: isLoadingProp,
  isProvisioning,
  provisioningHasError,
  onRetryProvisioning,
  canRetry,
  isRetrying,
  retryCount,
  maxRetries,
  provisioningProgress = 0,
  tableClassName,
  useDataHook = useMachinesView
}: MachinesViewProps) {
  const {
    searchTerm,
    currentPage,
    machines,
    paginatedMachines,
    totalPages,
    isLoading,
    error,
    handleSearchChange,
    handlePageChange,
    isFetching,
    refetch
  } = useDataHook(isLoadingProp);

  const [isWizardOpen, setIsWizardOpen] = React.useState(false);
  const [viewMode, setViewMode] = React.useState<'grid' | 'table'>('grid');

  if (isLoading) {
    return <MachinesSkeleton showHeader={showHeader} showSearch={showSearch} />;
  }

  if (error) {
    return (
      <div className={className}>
        {showHeader && <h1 className="text-2xl tracking-wide pb-4 sm:pb-6">{headerTitle}</h1>}
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center space-y-3">
            <AlertCircle className="h-10 w-10 text-destructive mx-auto" />
            <TypographyMuted className="text-lg">Failed to load machines</TypographyMuted>
            <TypographyMuted className="text-sm text-muted-foreground">
              Something went wrong while loading your machines. Please try again.
            </TypographyMuted>
            <Button variant="outline" size="sm" onClick={() => refetch()} className="gap-2">
              <RefreshCw className="h-4 w-4" />
              Retry
            </Button>
          </div>
        </div>
      </div>
    );
  }

  const headerContent = showHeader && (
    <>
      <h1 className="text-2xl tracking-wide pb-4 sm:pb-6">{headerTitle}</h1>
      <div className="flex items-center justify-between flex-wrap gap-3 pb-4 pt-6">
        <div className="w-full max-w-xs">
          {showSearch && machines.length > 0 && (
            <SearchBar
              searchTerm={searchTerm}
              handleSearchChange={handleSearchChange}
              label={searchLabel}
              isLoading={false}
              className="[&_input]:w-full!"
            />
          )}
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center border rounded-md">
            <Button
              variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
              size="icon"
              className="h-8 w-8 rounded-r-none border-0"
              onClick={() => setViewMode('grid')}
              aria-label="Grid view"
            >
              <LayoutGrid className="h-4 w-4" />
            </Button>
            <Button
              variant={viewMode === 'table' ? 'secondary' : 'ghost'}
              size="icon"
              className="h-8 w-8 rounded-l-none border-0"
              onClick={() => setViewMode('table')}
              aria-label="Table view"
            >
              <List className="h-4 w-4" />
            </Button>
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => refetch()}
            className="h-8 w-8"
            aria-label="Refresh machines"
          >
            <RotateCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>
    </>
  );

  if (machines.length === 0) {
    if (isProvisioning || isRetrying) {
      return (
        <div className={className}>
          {headerContent}
          <div className="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            <ProvisioningMachineCard progress={provisioningProgress} />
          </div>
        </div>
      );
    }

    return (
      <div className={className}>
        {headerContent}
        <MachinesEmptyState
          isProvisioning={false}
          provisioningHasError={provisioningHasError}
          onRetry={onRetryProvisioning}
          canRetry={canRetry}
          isRetrying={false}
          retryCount={retryCount}
          maxRetries={maxRetries}
        />
        {!provisioningHasError && (
          <div className="flex justify-center mt-4">
            <Button onClick={() => setIsWizardOpen(true)} className="gap-2">
              <Plus className="h-4 w-4" />
              Add Machine
            </Button>
          </div>
        )}
        <AddMachineWizard open={isWizardOpen} onOpenChange={setIsWizardOpen} />
      </div>
    );
  }

  return (
    <div className={className}>
      {headerContent}
      {!showHeader && showSearch && (
        <div className="mb-6">
          <SearchBar
            searchTerm={searchTerm}
            handleSearchChange={handleSearchChange}
            label={searchLabel}
            isLoading={false}
          />
        </div>
      )}

      {viewMode === 'grid' ? (
        <div className="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
          {paginatedMachines.map((machine) => (
            <MachineCard key={machine.id} machine={machine} />
          ))}
          <AddMachineCard onClick={() => setIsWizardOpen(true)} />
        </div>
      ) : (
        <MachinesTable machines={paginatedMachines} tableClassName={tableClassName} />
      )}

      {totalPages > 1 && (
        <div className="mt-8 flex justify-center">
          <PaginationWrapper
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      )}
      <AddMachineWizard open={isWizardOpen} onOpenChange={setIsWizardOpen} />
    </div>
  );
}

export function MachinesScreen() {
  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <MachinesView />
    </PageLayout>
  );
}
