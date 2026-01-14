'use client';

import React, { useState, useRef, useEffect } from 'react';
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
  StopCircle,
  RotateCw,
  Copy,
  Check,
  Globe,
  Lock,
  Settings2,
  Zap,
  Info,
  ChevronRight
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
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
import type { Resource, Action as RBACAction } from '@/packages/utils/rbac';
import SubPageHeader from '@/components/ui/sub-page-header';
import { cn } from '@/lib/utils';
import { useContainerActions } from '@/packages/hooks/containers/use-container-actions';
import { useContainerNavigation } from '@/packages/hooks/containers/use-container-navigation';
import { useContainerActionHandlers } from '@/packages/hooks/containers/use-container-action-handlers';
import { useResourceLimitsDialog } from '@/packages/hooks/containers/use-resource-limits-dialog';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Container } from '@/redux/services/container/containerApi';
import {
  getStatusIconClasses,
  getPortColors,
  getStatusColors
} from '@/packages/utils/container-styles';
import {
  isRunning,
  formatDate,
  truncateId,
  getPortsDisplay
} from '@/packages/utils/container-helpers';
import {
  useUpdateContainerResources,
  presetConfig,
  fieldConfigs,
  formatPresetValue,
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
  EmptyStateProps,
  GroupedContainerViewProps
} from '@/packages/types/containers';

export { Action };

function GuardedButton({
  resource,
  action,
  children,
  loadingSize = 'h-8 w-8'
}: {
  resource: Resource;
  action: RBACAction;
  children: React.ReactNode;
  loadingSize?: string;
}) {
  return (
    <ResourceGuard
      resource={resource}
      action={action}
      loadingFallback={<Skeleton className={cn('rounded-lg', loadingSize)} />}
    >
      {children}
    </ResourceGuard>
  );
}

function IconButton({
  icon: Icon,
  onClick,
  disabled,
  tooltip,
  variant,
  className
}: ActionButtonProps & { className?: string }) {
  const variants = {
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
        variant && variants[variant],
        disabled && 'opacity-50 cursor-not-allowed',
        className
      )}
      title={tooltip}
    >
      <Icon className="h-4 w-4" />
    </Button>
  );
}

function ContainerMetadata({
  container,
  showIcon = true,
  iconSize = 'md',
  nameClassName,
  idClassName
}: {
  container: Container;
  showIcon?: boolean;
  iconSize?: 'sm' | 'md';
  nameClassName?: string;
  idClassName?: string;
}) {
  const running = isRunning(container.status);
  const iconClasses = getStatusIconClasses(running);
  const iconSizes = { sm: 'h-4 w-4', md: 'h-5 w-5' };
  const containerSizes = { sm: 'p-2 rounded-lg', md: 'p-2.5 rounded-xl' };

  return (
    <div className="flex items-center gap-3 min-w-0 flex-1">
      {showIcon && (
        <div className={cn(containerSizes[iconSize], 'flex-shrink-0', iconClasses.container)}>
          <Box className={cn(iconSizes[iconSize], iconClasses.icon)} />
        </div>
      )}
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <h3 className={cn('font-semibold truncate', nameClassName || 'font-medium')}>
            {container.name}
          </h3>
          {running && showIcon && <StatusIndicator isRunning={true} size={iconSize} />}
        </div>
        <p className={cn('text-xs text-muted-foreground truncate mt-0.5 font-mono', idClassName)}>
          {truncateId(container.id)}
        </p>
      </div>
    </div>
  );
}

function PortsList({
  ports,
  maxVisible = 2,
  variant = 'pill',
  showType = false,
  emptyText = 'No ports'
}: {
  ports: any[];
  maxVisible?: number;
  variant?: 'pill' | 'inline';
  showType?: boolean;
  emptyText?: string;
}) {
  const display = getPortsDisplay(ports, maxVisible, variant);
  if (!display) return <span className="text-xs text-muted-foreground/50">{emptyText}</span>;

  const isVertical = variant === 'inline';
  return (
    <div
      className={cn('flex', isVertical ? 'flex-col gap-1' : 'items-center gap-1 overflow-hidden')}
    >
      {display.visible.map((port, idx) => (
        <PortDisplay key={idx} port={port} variant={variant} showType={showType} />
      ))}
      {display.remaining > 0 && (
        <span className="text-xs text-muted-foreground">
          +{display.remaining}
          {isVertical ? ' more' : ''}
        </span>
      )}
    </div>
  );
}

function ActionButtonGroup({
  isRunning: running,
  isProtected,
  onStart,
  onStop,
  onRemove
}: {
  isRunning: boolean;
  isProtected: boolean;
  onStart: (e: React.MouseEvent) => void;
  onStop: (e: React.MouseEvent) => void;
  onRemove: (e: React.MouseEvent) => void;
}) {
  return (
    <div className="flex items-center gap-1">
      <GuardedButton resource="container" action="update">
        {running ? (
          <IconButton
            icon={Square}
            onClick={onStop}
            disabled={isProtected}
            tooltip="Stop container"
            variant="warning"
          />
        ) : (
          <IconButton
            icon={Play}
            onClick={onStart}
            disabled={isProtected}
            tooltip="Start container"
            variant="success"
          />
        )}
      </GuardedButton>
      <GuardedButton resource="container" action="delete">
        <IconButton
          icon={Trash2}
          onClick={onRemove}
          disabled={isProtected}
          tooltip="Remove container"
          variant="danger"
        />
      </GuardedButton>
    </div>
  );
}

export function StatusIndicator({
  isRunning,
  size = 'md',
  showPulse = true
}: StatusIndicatorProps) {
  const sizes = { sm: 'h-1.5 w-1.5', md: 'h-2 w-2', lg: 'h-3 w-3' };
  const colors = getStatusColors(isRunning ? 'running' : 'stopped');
  const sizeClass = sizes[size];

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
  const iconSizes = { sm: 'h-3 w-3', md: 'h-4 w-4' };
  const Icon = copied ? Check : Copy;
  return (
    <button
      onClick={onCopy}
      className={cn(
        'text-muted-foreground hover:text-foreground transition-colors flex-shrink-0',
        className
      )}
    >
      <Icon className={cn(iconSizes[size], copied && 'text-emerald-500')} />
      {showText && <span className="ml-1 text-xs">{copied ? 'Copied' : 'Copy'}</span>}
    </button>
  );
}

export function PortDisplay({ port, variant = 'pill', showType = true }: PortDisplayProps) {
  const hasPublic = port.public_port > 0;
  const colors = getPortColors(hasPublic);
  const portContent = hasPublic ? (
    <>
      <span>{port.public_port}</span>
      <ArrowRight className={variant === 'flow' ? 'h-4 w-4' : 'h-2.5 w-2.5'} />
      <span>{port.private_port}</span>
    </>
  ) : (
    <span>{port.private_port}</span>
  );

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

  return (
    <span
      className={cn(
        variant === 'pill'
          ? 'inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-mono'
          : 'inline-flex items-center gap-1 text-xs font-mono',
        variant === 'pill' ? colors.pill : hasPublic ? colors.text : 'text-muted-foreground'
      )}
    >
      {portContent}
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
  const running = isRunning(status);
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
        colors.badge,
        className
      )}
    >
      {showDot && running && <StatusIndicator isRunning={true} size="sm" />}
      {status}
    </span>
  );
}

export const ContainerActions = ({ container, onAction }: ContainerActionsProps) => {
  const { containerId, isProtected, isRunning: running } = useContainerActions(container);
  const { handleStart, handleStop, handleRemove } = useContainerActionHandlers(
    containerId,
    onAction
  );

  return (
    <ActionButtonGroup
      isRunning={running}
      isProtected={isProtected}
      onStart={handleStart}
      onStop={handleStop}
      onRemove={handleRemove}
    />
  );
};

export function ActionHeader({
  handleRefresh,
  isRefreshing,
  isFetching,
  t,
  setShowPruneImagesConfirm,
  setShowPruneBuildCacheConfirm
}: ActionHeaderProps) {
  const loading = isRefreshing || isFetching;
  return (
    <div className="flex items-center gap-2">
      <Button onClick={handleRefresh} variant="outline" size="sm" disabled={loading}>
        {loading ? (
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

export const ContainerCard = ({ container, onClick, onAction }: ContainerCardProps) => {
  return (
    <div
      onClick={onClick}
      className={cn(
        'group relative rounded-xl p-5 cursor-pointer transition-all duration-200',
        'hover:bg-muted/50 border border-transparent hover:border-border/50'
      )}
    >
      <div className="flex items-start justify-between gap-4">
        <ContainerMetadata container={container} nameClassName="font-semibold" />
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
        <div className="flex items-center gap-1.5 min-w-0 flex-1 overflow-hidden">
          <Network className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
          <PortsList
            ports={container.ports || []}
            maxVisible={2}
            variant="pill"
            emptyText="No ports"
          />
        </div>
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground flex-shrink-0">
          <Clock className="h-3 w-3" />
          <span>{formatDate(container.created)}</span>
        </div>
      </div>
    </div>
  );
};

const ContainersTable = ({
  containersData,
  sortBy = 'name',
  sortOrder = 'asc',
  onSort,
  onAction
}: ContainersTableProps) => {
  const { t } = useTranslation();
  const { handleRowClick } = useContainerNavigation();

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
        <div className="w-24 text-left">
          <SortableHeader
            label={t('dashboard.containers.table.headers.status')}
            field="status"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSort}
          />
        </div>
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
      className="flex items-center justify-start gap-1 hover:text-foreground transition-colors"
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
  const running = isRunning(container.status);
  return (
    <div
      onClick={onClick}
      className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 items-center cursor-pointer hover:bg-muted/30 transition-colors group"
    >
      <ContainerMetadata container={container} iconSize="sm" />
      <div className="min-w-0">
        <p className="text-sm truncate text-muted-foreground" title={container.image}>
          {container.image}
        </p>
        <p className="text-xs text-muted-foreground/60 flex items-center gap-1 mt-0.5">
          <Clock className="h-3 w-3" />
          {formatDate(container.created)}
        </p>
      </div>
      <div className="w-24">
        <StatusBadge status={container.state || container.status} showDot={running} />
      </div>
      <div className="w-32">
        <PortsList ports={container.ports || []} maxVisible={2} variant="inline" emptyText="—" />
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

export function ContainerDetailsHeader({
  container,
  isLoading,
  isProtected,
  handleContainerAction,
  t
}: ContainerDetailsHeaderProps) {
  const statusColors = getStatusColors(container.status);
  const running = isRunning(container.status);
  const disabled = isLoading || isProtected;

  const icon = (
    <div className={cn('w-12 h-12 rounded-xl flex items-center justify-center', statusColors.bg)}>
      <div className={cn('w-3 h-3 rounded-full', statusColors.dot)} />
    </div>
  );

  const metadata = (
    <div className="flex items-center gap-2">
      <code className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded">
        {truncateId(container.id)}
      </code>
      <Badge variant="outline" className={cn('text-xs', statusColors.border)}>
        {container.status}
      </Badge>
    </div>
  );

  const actions = (
    <>
      <GuardedButton resource="container" action="update" loadingSize="h-9 w-24">
        {running ? (
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleContainerAction('stop')}
              disabled={disabled}
            >
              <StopCircle className="mr-2 h-4 w-4" />
              {t('containers.stop')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleContainerAction('restart')}
              disabled={disabled}
            >
              <RotateCw className="mr-2 h-4 w-4" />
              {t('containers.restart')}
            </Button>
          </>
        ) : (
          <Button
            variant="default"
            size="sm"
            onClick={() => handleContainerAction('start')}
            disabled={disabled}
            className="bg-emerald-600 hover:bg-emerald-700"
          >
            <Play className="mr-2 h-4 w-4" />
            {t('containers.start')}
          </Button>
        )}
      </GuardedButton>
      <GuardedButton resource="container" action="delete" loadingSize="h-9 w-20">
        <Button
          variant="outline"
          size="sm"
          onClick={() => handleContainerAction('remove')}
          disabled={disabled}
          className="text-red-500 hover:text-red-600 hover:bg-red-500/10 border-red-500/20"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          {t('containers.remove')}
        </Button>
      </GuardedButton>
    </>
  );

  return <SubPageHeader icon={icon} title={container.name} metadata={metadata} actions={actions} />;
}

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
  const closeDialogRef = useRef<() => void>(() => {});

  const { form, onSubmit, isLoading, resetToCurrentValues, applyPreset } =
    useUpdateContainerResources({
      containerId: container.id,
      currentMemory: container.host_config.memory,
      currentMemorySwap: container.host_config.memory_swap,
      currentCpuShares: container.host_config.cpu_shares,
      onSuccess: () => closeDialogRef.current()
    });

  const { open, handleOpenChange, handleCancel, closeDialog } =
    useResourceLimitsDialog(resetToCurrentValues);

  useEffect(() => {
    closeDialogRef.current = closeDialog;
  }, [closeDialog]);

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
                onCancel={handleCancel}
              />
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </ResourceGuard>
  );
}

export function GroupedContainerView({
  groups,
  ungrouped = [],
  viewMode,
  onContainerClick,
  onContainerAction,
  sortBy,
  sortOrder,
  onSort
}: GroupedContainerViewProps) {
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(
    new Set(groups.map((g) => g.application_id))
  );

  const toggleGroup = (applicationId: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(applicationId)) {
        next.delete(applicationId);
      } else {
        next.add(applicationId);
      }
      return next;
    });
  };

  if (groups.length === 0 && ungrouped.length === 0) {
    return null;
  }

  return (
    <div className="space-y-6">
      {groups.map((group) => {
        const isExpanded = expandedGroups.has(group.application_id);
        const runningCount = group.containers.filter((c) => c.status === 'running').length;
        const totalCount = group.containers.length;

        return (
          <div key={group.application_id} className="border rounded-lg overflow-hidden">
            <button
              onClick={() => toggleGroup(group.application_id)}
              className={cn(
                'w-full flex items-center justify-between p-4 text-left',
                'hover:bg-muted/50 transition-colors',
                'border-b border-border'
              )}
            >
              <div className="flex items-center gap-3 flex-1 min-w-0">
                {isExpanded ? (
                  <ChevronDown className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                )}
                <Box className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                <div className="min-w-0 flex-1">
                  <h3 className="font-semibold truncate">{group.application_name}</h3>
                  <p className="text-sm text-muted-foreground">
                    {totalCount} container{totalCount !== 1 ? 's' : ''} • {runningCount} running
                  </p>
                </div>
              </div>
            </button>

            {isExpanded && (
              <div className="">
                {viewMode === 'card' ? (
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2 p-2">
                    {group.containers.map((container) => (
                      <ContainerCard
                        key={container.id}
                        container={container}
                        onClick={() => onContainerClick(container)}
                        onAction={onContainerAction}
                      />
                    ))}
                  </div>
                ) : (
                  <ContainersTable
                    containersData={group.containers}
                    sortBy={sortBy}
                    sortOrder={sortOrder}
                    onSort={onSort}
                    onAction={onContainerAction}
                  />
                )}
              </div>
            )}
          </div>
        );
      })}

      {ungrouped.length > 0 && (
        <>
          {viewMode === 'card' ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
              {ungrouped.map((container) => (
                <ContainerCard
                  key={container.id}
                  container={container}
                  onClick={() => onContainerClick(container)}
                  onAction={onContainerAction}
                />
              ))}
            </div>
          ) : (
            <div className="border rounded-lg">
              <ContainersTable
                containersData={ungrouped}
                sortBy={sortBy}
                sortOrder={sortOrder}
                onSort={onSort}
                onAction={onContainerAction}
              />
            </div>
          )}
        </>
      )}
    </div>
  );
}
