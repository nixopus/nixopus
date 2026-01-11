'use client';

import React from 'react';
import {
  Box,
  Clock,
  Network,
  ArrowRight,
  Play,
  Square,
  Trash2,
  RefreshCw,
  Loader2,
  Scissors,
  ChevronUp,
  ChevronDown,
  Package
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import { cn } from '@/lib/utils';
import { useContainerActions } from '@/packages/hooks/containers/use-container-actions';
import { translationKey } from '@/packages/hooks/shared/use-translation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRouter } from 'next/navigation';
import { formatDistanceToNow } from 'date-fns';
import { Container } from '@/redux/services/container/containerApi';
import { ContainerData } from '@/redux/types/monitor';
import {
  StatusIndicator,
  PortDisplay,
  EmptyState,
  StatusBadge
} from '@/packages/components/container-shared';
import { getStatusIconClasses } from '@/packages/utils/container-styles';
import {
  Action,
  SortField,
  ContainerActionsProps,
  ActionButtonProps,
  ActionHeaderProps,
  ContainerCardProps,
  ContainersTableProps,
  SortableHeaderProps,
  ContainerRowProps,
  ContainersWidgetProps
} from '@/packages/types/containers';

// Re-export Action enum for backward compatibility
export { Action };

// ============================================================================
// Container Actions
// ============================================================================

export const ContainerActions = ({ container, onAction }: ContainerActionsProps) => {
  const { containerId, isProtected, isRunning } = useContainerActions(container);

  function handleClick(e: React.MouseEvent, action: Action) {
    e.stopPropagation();
    onAction(containerId, action);
  }

  return (
    <div className="flex items-center gap-1">
      <ResourceGuard
        resource="container"
        action="update"
        loadingFallback={<Skeleton className="h-8 w-8 rounded-lg" />}
      >
        {isRunning ? (
          <ActionButton
            icon={Square}
            onClick={(e) => handleClick(e, Action.STOP)}
            disabled={isProtected}
            tooltip="Stop container"
            variant="warning"
          />
        ) : (
          <ActionButton
            icon={Play}
            onClick={(e) => handleClick(e, Action.START)}
            disabled={isProtected}
            tooltip="Start container"
            variant="success"
          />
        )}
      </ResourceGuard>
      <ResourceGuard
        resource="container"
        action="delete"
        loadingFallback={<Skeleton className="h-8 w-8 rounded-lg" />}
      >
        <ActionButton
          icon={Trash2}
          onClick={(e) => handleClick(e, Action.REMOVE)}
          disabled={isProtected}
          tooltip="Remove container"
          variant="danger"
        />
      </ResourceGuard>
    </div>
  );
};

function ActionButton({ icon: Icon, onClick, disabled, tooltip, variant }: ActionButtonProps) {
  const variantStyles = {
    success: 'hover:bg-emerald-500/10 hover:text-emerald-500',
    warning: 'hover:bg-amber-500/10 hover:text-amber-500',
    danger: 'hover:bg-red-500/10 hover:text-red-500'
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'h-8 w-8 text-muted-foreground transition-colors',
        variant && variantStyles[variant],
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      title={tooltip}
    >
      <Icon className="h-4 w-4" />
    </Button>
  );
}

export function ActionHeader({
  handleRefresh,
  isRefreshing,
  isFetching,
  t,
  setShowPruneImagesConfirm,
  setShowPruneBuildCacheConfirm
}: ActionHeaderProps) {
  return (
    <div className="flex items-center gap-2">
      <Button
        onClick={handleRefresh}
        variant="outline"
        size="sm"
        disabled={isRefreshing || isFetching}
      >
        {isRefreshing || isFetching ? (
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
        ) : (
          <RefreshCw className="mr-2 h-4 w-4" />
        )}
        {t('containers.refresh')}
      </Button>
      <AnyPermissionGuard
        permissions={['container:delete']}
        loadingFallback={<Skeleton className="h-9 w-20" />}
      >
        <Button variant="outline" size="sm" onClick={() => setShowPruneImagesConfirm(true)}>
          <Trash2 className="mr-2 h-4 w-4" />
          {t('containers.prune_images')}
        </Button>
        <Button variant="outline" size="sm" onClick={() => setShowPruneBuildCacheConfirm(true)}>
          <Scissors className="mr-2 h-4 w-4" />
          {t('containers.prune_build_cache')}
        </Button>
      </AnyPermissionGuard>
    </div>
  );
}

// ============================================================================
// Container Card
// ============================================================================

export const ContainerCard = ({ container, onClick, onAction }: ContainerCardProps) => {
  const isRunning = container.status === 'running';
  const hasPorts = container.ports && container.ports.length > 0;
  const iconClasses = getStatusIconClasses(isRunning);

  return (
    <div
      onClick={onClick}
      className={cn(
        'group relative rounded-xl p-5 cursor-pointer transition-all duration-200',
        'hover:bg-muted/50',
        'border border-transparent hover:border-border/50'
      )}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <div className={cn('p-2.5 rounded-xl flex-shrink-0', iconClasses.container)}>
            <Box className={cn('h-5 w-5', iconClasses.icon)} />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold truncate">{container.name}</h3>
              {isRunning && <StatusIndicator isRunning={true} size="md" />}
            </div>
            <p className="text-xs text-muted-foreground truncate mt-0.5 font-mono">
              {container.id.slice(0, 12)}
            </p>
          </div>
        </div>

        <div
          className="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={(e) => e.stopPropagation()}
        >
          <ContainerActions container={container} onAction={onAction} />
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
        <span className="truncate" title={container.image}>
          {container.image}
        </span>
      </div>

      <div className="mt-4 flex items-center justify-between gap-4">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {hasPorts ? (
            <div className="flex items-center gap-1.5 overflow-hidden">
              <Network className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
              <div className="flex items-center gap-1 overflow-hidden">
                {container.ports.slice(0, 2).map((port, idx) => (
                  <PortDisplay key={idx} port={port} variant="pill" showType={false} />
                ))}
                {container.ports.length > 2 && (
                  <span className="text-xs text-muted-foreground">
                    +{container.ports.length - 2}
                  </span>
                )}
              </div>
            </div>
          ) : (
            <span className="text-xs text-muted-foreground/50">No ports</span>
          )}
        </div>

        <div className="flex items-center gap-1.5 text-xs text-muted-foreground flex-shrink-0">
          <Clock className="h-3 w-3" />
          <span>{formatDistanceToNow(new Date(container.created), { addSuffix: true })}</span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Container Table
// ============================================================================

const ContainersTable = ({
  containersData,
  sortBy = 'name',
  sortOrder = 'asc',
  onSort,
  onAction
}: ContainersTableProps) => {
  const { t } = useTranslation();
  const router = useRouter();

  const handleRowClick = (container: Container) => {
    router.push(`/containers/${container.id}`);
  };

  if (containersData.length === 0) {
    return <EmptyState icon={Box} message={t('dashboard.containers.table.noContainers')} />;
  }

  return (
    <div className="rounded-xl border overflow-hidden">
      <div className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 bg-muted/30 text-xs font-medium text-muted-foreground uppercase tracking-wider">
        <SortableHeader
          label={t('dashboard.containers.table.headers.name')}
          field="name"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSort}
        />
        <div>Image</div>
        <SortableHeader
          label={t('dashboard.containers.table.headers.status')}
          field="status"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSort}
        />
        <div className="w-32">Ports</div>
        <div className="w-24"></div>
      </div>

      <div className="divide-y divide-border/50">
        {containersData.map((container) => (
          <ContainerRow
            key={container.id}
            container={container}
            onClick={() => handleRowClick(container)}
            onAction={onAction}
          />
        ))}
      </div>
    </div>
  );
};

function SortableHeader({ label, field, currentSort, currentOrder, onSort }: SortableHeaderProps) {
  const isActive = currentSort === field;

  return (
    <button
      onClick={() => onSort?.(field)}
      className="flex items-center gap-1 hover:text-foreground transition-colors"
    >
      {label}
      <span className="flex flex-col">
        <ChevronUp
          className={cn(
            'h-3 w-3 -mb-1',
            isActive && currentOrder === 'asc' ? 'text-foreground' : 'opacity-30'
          )}
        />
        <ChevronDown
          className={cn(
            'h-3 w-3',
            isActive && currentOrder === 'desc' ? 'text-foreground' : 'opacity-30'
          )}
        />
      </span>
    </button>
  );
}

function ContainerRow({ container, onClick, onAction }: ContainerRowProps) {
  const isRunning = container.status === 'running';
  const hasPorts = container.ports && container.ports.length > 0;
  const iconClasses = getStatusIconClasses(isRunning);

  return (
    <div
      onClick={onClick}
      className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 items-center cursor-pointer hover:bg-muted/30 transition-colors group"
    >
      <div className="flex items-center gap-3 min-w-0">
        <div className={cn('p-2 rounded-lg flex-shrink-0', iconClasses.container)}>
          <Box className={iconClasses.icon} />
        </div>
        <div className="min-w-0">
          <p className="font-medium truncate">{container.name}</p>
          <p className="text-xs text-muted-foreground font-mono">{container.id.slice(0, 12)}</p>
        </div>
      </div>

      <div className="min-w-0">
        <p className="text-sm truncate text-muted-foreground" title={container.image}>
          {container.image}
        </p>
        <p className="text-xs text-muted-foreground/60 flex items-center gap-1 mt-0.5">
          <Clock className="h-3 w-3" />
          {formatDistanceToNow(new Date(container.created), { addSuffix: true })}
        </p>
      </div>

      <div className="w-24">
        <StatusBadge status={container.state || container.status} showDot={isRunning} />
      </div>

      <div className="w-32">
        {hasPorts ? (
          <div className="flex flex-col gap-1">
            {container.ports.slice(0, 2).map((port, idx) => (
              <PortDisplay key={idx} port={port} variant="inline" showType={false} />
            ))}
            {container.ports.length > 2 && (
              <span className="text-xs text-muted-foreground">
                +{container.ports.length - 2} more
              </span>
            )}
          </div>
        ) : (
          <span className="text-xs text-muted-foreground/50">—</span>
        )}
      </div>

      <div
        className="w-24 flex justify-end opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={(e) => e.stopPropagation()}
      >
        {onAction && <ContainerActions container={container} onAction={onAction} />}
      </div>
    </div>
  );
}

export default ContainersTable;

// ============================================================================
// Container Widget
// ============================================================================

const ContainersWidget: React.FC<ContainersWidgetProps> = ({ containersData, columns }) => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <CardWrapper
      title={t('dashboard.containers.title')}
      icon={Package}
      compact
      actions={
        <Button variant="outline" size="sm" onClick={() => router.push('/containers')}>
          <ArrowRight className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          {t('dashboard.containers.viewAll')}
        </Button>
      }
    >
      <DataTable
        data={containersData}
        columns={columns}
        emptyMessage={t('dashboard.containers.table.noContainers')}
        showBorder={false}
        hoverable={false}
      />
    </CardWrapper>
  );
};

export { ContainersWidget };

export const ContainersWidgetSkeleton: React.FC = () => {
  const { t } = useTranslation();

  return (
    <CardWrapper
      title={t('dashboard.containers.title')}
      icon={Package}
      compact
      actions={<Skeleton className="h-8 w-24" />}
    >
      <div className="border-b pb-2 mb-2">
        <div className="grid grid-cols-6 gap-4">
          {['h-4 w-8', 'h-4 w-12', 'h-4 w-12', 'h-4 w-12', 'h-4 w-10', 'h-4 w-14'].map(
            (className, idx) => (
              <Skeleton key={idx} className={className} />
            )
          )}
        </div>
      </div>
      <div className="space-y-3">
        {[0, 1, 2].map((i) => (
          <div key={i} className="grid grid-cols-6 gap-4 items-center">
            {[
              'h-4 w-16 font-mono',
              'h-4 w-24',
              'h-4 w-32',
              'h-5 w-16 rounded-full',
              'h-4 w-12',
              'h-4 w-16'
            ].map((className, idx) => (
              <Skeleton key={idx} className={className} />
            ))}
          </div>
        ))}
      </div>
    </CardWrapper>
  );
};
