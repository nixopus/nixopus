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
  Package,
  StopCircle,
  RotateCw,
  Copy,
  Check,
  Globe,
  Lock,
  Settings2,
  Zap,
  Info
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog';
import { Form, FormControl, FormField, FormItem, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import SubPageHeader from '@/components/ui/sub-page-header';
import { cn } from '@/lib/utils';
import { useContainerActions } from '@/packages/hooks/containers/use-container-actions';
import { translationKey } from '@/packages/hooks/shared/use-translation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRouter } from 'next/navigation';
import { formatDistanceToNow } from 'date-fns';
import { Container } from '@/redux/services/container/containerApi';
import { ContainerData } from '@/redux/types/monitor';
import {
  getStatusIconClasses,
  getPortColors,
  getStatusColors
} from '@/packages/utils/container-styles';
import { useState } from 'react';
import { UseFormReturn, ControllerRenderProps } from 'react-hook-form';
import {
  useUpdateContainerResources,
  presetConfig,
  fieldConfigs,
  formatPresetValue,
  PresetType,
  FieldConfig,
  ResourceLimitsFormValues
} from '@/packages/hooks/containers/use-update-container-resources';
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
  ContainersWidgetProps,
  StatPillProps,
  ContainerDetailsHeaderProps,
  ResourceLimitsFormProps,
  PresetButtonProps,
  PresetGridProps,
  ResourceFieldProps,
  FormActionsProps,
  ResourceFieldsProps,
  StatusIndicatorProps,
  CopyButtonProps,
  PortDisplayProps,
  StatusBadgeProps,
  EmptyStateProps
} from '@/packages/types/containers';

// Re-export Action enum for backward compatibility
export { Action };

// ============================================================================
// Shared Components
// ============================================================================

export function StatusIndicator({
  isRunning,
  size = 'md',
  showPulse = true
}: StatusIndicatorProps) {
  const sizes = {
    sm: 'h-1.5 w-1.5',
    md: 'h-2 w-2',
    lg: 'h-3 w-3'
  };

  const sizeClass = sizes[size];
  const colors = getStatusColors(isRunning ? 'running' : 'stopped');

  return (
    <span className={cn('relative flex', sizeClass, 'flex-shrink-0')}>
      {showPulse && isRunning && (
        <span
          className={cn(
            'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
            colors.dotPulse
          )}
        />
      )}
      <span className={cn('relative inline-flex rounded-full', sizeClass, colors.dot)} />
    </span>
  );
}

export function CopyButton({
  copied,
  onCopy,
  size = 'sm',
  className,
  showText = false
}: CopyButtonProps) {
  const iconSizes = {
    sm: 'h-3 w-3',
    md: 'h-4 w-4'
  };

  return (
    <button
      onClick={onCopy}
      className={cn(
        'text-muted-foreground hover:text-foreground transition-colors flex-shrink-0',
        className
      )}
    >
      {copied ? (
        <>
          <Check className={cn(iconSizes[size], 'text-emerald-500')} />
          {showText && <span className="ml-1 text-xs">Copied</span>}
        </>
      ) : (
        <>
          <Copy className={iconSizes[size]} />
          {showText && <span className="ml-1 text-xs">Copy</span>}
        </>
      )}
    </button>
  );
}

export function PortDisplay({ port, variant = 'pill', showType = true }: PortDisplayProps) {
  const hasPublic = port.public_port > 0;
  const colors = getPortColors(hasPublic);

  if (variant === 'pill') {
    return (
      <span
        className={cn(
          'inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-mono',
          colors.pill
        )}
      >
        {hasPublic ? (
          <>
            <span>{port.public_port}</span>
            <ArrowRight className="h-2.5 w-2.5" />
            <span>{port.private_port}</span>
          </>
        ) : (
          <span>{port.private_port}</span>
        )}
      </span>
    );
  }

  if (variant === 'flow') {
    return (
      <div
        className={cn(
          'flex items-center gap-3 px-4 py-3 rounded-xl transition-colors',
          colors.flow
        )}
      >
        {hasPublic ? (
          <>
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-emerald-500" />
              <span className="font-mono text-lg font-semibold text-emerald-600 dark:text-emerald-400">
                {port.public_port}
              </span>
            </div>
            <ArrowRight className="h-4 w-4 text-muted-foreground" />
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-muted-foreground" />
              <span className="font-mono text-lg">{port.private_port}</span>
            </div>
          </>
        ) : (
          <div className="flex items-center gap-2">
            <Lock className="h-4 w-4 text-muted-foreground" />
            <span className="font-mono text-lg">{port.private_port}</span>
          </div>
        )}
        {showType && (
          <span className="text-xs text-muted-foreground uppercase ml-1">/{port.type}</span>
        )}
      </div>
    );
  }

  // inline variant
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 text-xs font-mono',
        hasPublic ? colors.text : 'text-muted-foreground'
      )}
    >
      {hasPublic ? (
        <>
          {port.public_port}
          <ArrowRight className="h-2.5 w-2.5" />
          {port.private_port}
        </>
      ) : (
        port.private_port
      )}
    </span>
  );
}

export function EmptyState({ icon: Icon, message, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-16 text-muted-foreground',
        className
      )}
    >
      <Icon className="h-12 w-12 mb-4 opacity-30" />
      <p className="text-sm">{message}</p>
    </div>
  );
}

export function StatusBadge({ status, showDot = false, className }: StatusBadgeProps) {
  const colors = getStatusColors(status);
  const isRunning = (status || '').toLowerCase() === 'running';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
        colors.badge,
        className
      )}
    >
      {showDot && isRunning && <StatusIndicator isRunning={true} size="sm" />}
      {status}
    </span>
  );
}

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

// ============================================================================
// Stat Pill
// ============================================================================

export function StatPill({ value, label, color }: StatPillProps) {
  return (
    <div className="flex items-center gap-2">
      {color && (
        <span
          className={cn(
            'w-2 h-2 rounded-full',
            color === 'emerald' ? 'bg-emerald-500' : 'bg-zinc-500'
          )}
        />
      )}
      <span className="text-xl font-bold">{value}</span>
      <span className="text-sm text-muted-foreground">{label}</span>
    </div>
  );
}

// ============================================================================
// Container Details Header
// ============================================================================

export function ContainerDetailsHeader({
  container,
  isLoading,
  isProtected,
  handleContainerAction,
  t
}: ContainerDetailsHeaderProps) {
  const statusColors = getStatusColors(container.status);

  const icon = (
    <div className={cn('w-12 h-12 rounded-xl flex items-center justify-center', statusColors.bg)}>
      <div className={cn('w-3 h-3 rounded-full', statusColors.dot)} />
    </div>
  );

  const metadata = (
    <div className="flex items-center gap-2">
      <code className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded">
        {container.id.slice(0, 12)}
      </code>
      <Badge variant="outline" className={cn('text-xs', statusColors.border)}>
        {container.status}
      </Badge>
    </div>
  );

  const actions = (
    <>
      <ResourceGuard
        resource="container"
        action="update"
        loadingFallback={<Skeleton className="h-9 w-24" />}
      >
        {container.status !== 'running' ? (
          <Button
            variant="default"
            size="sm"
            onClick={() => handleContainerAction('start')}
            disabled={isLoading || isProtected}
            className="bg-emerald-600 hover:bg-emerald-700"
          >
            <Play className="mr-2 h-4 w-4" />
            {t('containers.start')}
          </Button>
        ) : (
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleContainerAction('stop')}
              disabled={isLoading || isProtected}
            >
              <StopCircle className="mr-2 h-4 w-4" />
              {t('containers.stop')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleContainerAction('restart')}
              disabled={isLoading || isProtected}
            >
              <RotateCw className="mr-2 h-4 w-4" />
              {t('containers.restart')}
            </Button>
          </>
        )}
      </ResourceGuard>
      <ResourceGuard
        resource="container"
        action="delete"
        loadingFallback={<Skeleton className="h-9 w-20" />}
      >
        <Button
          variant="outline"
          size="sm"
          onClick={() => handleContainerAction('remove')}
          disabled={isLoading || isProtected}
          className="text-red-500 hover:text-red-600 hover:bg-red-500/10 border-red-500/20"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          {t('containers.remove')}
        </Button>
      </ResourceGuard>
    </>
  );

  return <SubPageHeader icon={icon} title={container.name} metadata={metadata} actions={actions} />;
}

// ============================================================================
// Resource Limits Form
// ============================================================================

function PresetButton({ presetKey, memory, isActive, onSelect }: PresetButtonProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(presetKey)}
      className={cn(
        'flex flex-col items-center gap-1 p-3 rounded-lg border-2 transition-all text-xs',
        isActive
          ? 'border-primary bg-primary/5 text-primary'
          : 'border-muted hover:border-muted-foreground/20 hover:bg-muted/50'
      )}
    >
      <span className="font-medium">{formatPresetValue(presetKey, memory)}</span>
      <span className={cn('capitalize', isActive ? 'text-primary/70' : 'text-muted-foreground')}>
        {presetKey}
      </span>
    </button>
  );
}

function PresetGrid({ currentMemory, onPresetSelect }: PresetGridProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-3">
      <label className="text-sm font-medium">{t('containers.resourceLimits.presets.label')}</label>
      <div className="grid grid-cols-5 gap-2">
        {presetConfig.map(({ key, memory }) => (
          <PresetButton
            key={key}
            presetKey={key}
            memory={memory}
            isActive={currentMemory === memory}
            onSelect={onPresetSelect}
          />
        ))}
      </div>
    </div>
  );
}

function ResourceField({ config, field }: ResourceFieldProps) {
  const { t } = useTranslation();
  const {
    icon: Icon,
    labelKey,
    placeholderKey,
    unitKey,
    descriptionKey,
    unlimitedDescKey,
    min,
    isUnlimited
  } = config;
  const description =
    isUnlimited(field.value) && unlimitedDescKey ? t(unlimitedDescKey) : t(descriptionKey);

  return (
    <FormItem>
      <FormLabel className="flex items-center gap-2">
        <Icon className="h-4 w-4 text-muted-foreground" />
        {t(labelKey)}
        <Tooltip>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 text-muted-foreground cursor-help" />
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-xs bg-popover text-popover-foreground border">
            {description}
          </TooltipContent>
        </Tooltip>
      </FormLabel>
      <div className={cn(unitKey && 'flex gap-2')}>
        <FormControl>
          <Input
            type="number"
            min={min}
            placeholder={t(placeholderKey)}
            {...field}
            onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
            className={cn(unitKey && 'flex-1')}
          />
        </FormControl>
        {unitKey && (
          <span className="flex items-center px-3 bg-muted rounded-md text-sm text-muted-foreground">
            {t(unitKey)}
          </span>
        )}
      </div>
    </FormItem>
  );
}

function FormActions({ isLoading, isDirty, onReset, onCancel }: FormActionsProps) {
  const { t } = useTranslation();

  return (
    <div className="flex justify-between pt-4">
      {isDirty ? (
        <Button type="button" variant="ghost" onClick={onReset} disabled={isLoading}>
          {t('containers.resourceLimits.buttons.reset')}
        </Button>
      ) : (
        <div />
      )}
      <div className="flex gap-2">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
          {t('containers.resourceLimits.buttons.cancel')}
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading
            ? t('containers.resourceLimits.buttons.saving')
            : t('containers.resourceLimits.buttons.save')}
        </Button>
      </div>
    </div>
  );
}

function ResourceFields({ form }: ResourceFieldsProps) {
  return (
    <>
      {fieldConfigs.map((config) => (
        <FormField
          key={config.name}
          control={form.control}
          name={config.name}
          render={({ field }) => <ResourceField config={config} field={field} />}
        />
      ))}
    </>
  );
}

export function ResourceLimitsForm({ container }: ResourceLimitsFormProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  const { form, onSubmit, isLoading, resetToCurrentValues, applyPreset } =
    useUpdateContainerResources({
      containerId: container.id,
      currentMemory: container.host_config.memory,
      currentMemorySwap: container.host_config.memory_swap,
      currentCpuShares: container.host_config.cpu_shares,
      onSuccess: () => setOpen(false)
    });

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (newOpen) resetToCurrentValues();
  };

  return (
    <ResourceGuard
      resource="container"
      action="update"
      loadingFallback={<Skeleton className="h-9 w-28" />}
    >
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogTrigger asChild>
          <Button variant="outline" size="sm" className="gap-2">
            <Settings2 className="h-4 w-4" />
            {t('containers.resourceLimits.editButton')}
          </Button>
        </DialogTrigger>

        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5 text-primary" />
              {t('containers.resourceLimits.title')}
            </DialogTitle>
            <DialogDescription>{t('containers.resourceLimits.description')}</DialogDescription>
          </DialogHeader>

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <PresetGrid currentMemory={form.watch('memoryMB')} onPresetSelect={applyPreset} />
              <ResourceFields form={form} />
              <FormActions
                isLoading={isLoading}
                isDirty={form.formState.isDirty}
                onReset={resetToCurrentValues}
                onCancel={() => setOpen(false)}
              />
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </ResourceGuard>
  );
}
