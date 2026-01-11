'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { Container } from '@/redux/services/container/containerApi';
import { useRouter } from 'next/navigation';
import { ContainerActions } from './container-actions';
import { Action } from './container-card';
import { cn } from '@/lib/utils';
import { Box, ChevronUp, ChevronDown, ArrowRight, Clock } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

type SortField = 'name' | 'status';

const ContainersTable = ({
  containersData,
  sortBy = 'name',
  sortOrder = 'asc',
  onSort,
  onAction
}: {
  containersData: Container[];
  sortBy?: SortField;
  sortOrder?: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
  onAction?: (id: string, action: Action) => void;
}) => {
  const { t } = useTranslation();
  const router = useRouter();

  const handleRowClick = (container: Container) => {
    router.push(`/containers/${container.id}`);
  };

  if (containersData.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <Box className="h-12 w-12 mb-4 opacity-30" />
        <p className="text-sm">{t('dashboard.containers.table.noContainers')}</p>
      </div>
    );
  }

  return (
    <div className="rounded-xl border overflow-hidden">
      {/* Header */}
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

      {/* Rows */}
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

function SortableHeader({
  label,
  field,
  currentSort,
  currentOrder,
  onSort
}: {
  label: string;
  field: SortField;
  currentSort: SortField;
  currentOrder: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
}) {
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

function ContainerRow({
  container,
  onClick,
  onAction
}: {
  container: Container;
  onClick: () => void;
  onAction?: (id: string, action: Action) => void;
}) {
  const isRunning = container.status === 'running';
  const hasPorts = container.ports && container.ports.length > 0;

  return (
    <div
      onClick={onClick}
      className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 items-center cursor-pointer hover:bg-muted/30 transition-colors group"
    >
      {/* Name & ID */}
      <div className="flex items-center gap-3 min-w-0">
        <div
          className={cn(
            'p-2 rounded-lg flex-shrink-0',
            isRunning ? 'bg-emerald-500/10' : 'bg-zinc-500/10'
          )}
        >
          <Box className={cn('h-4 w-4', isRunning ? 'text-emerald-500' : 'text-zinc-500')} />
        </div>
        <div className="min-w-0">
          <p className="font-medium truncate">{container.name}</p>
          <p className="text-xs text-muted-foreground font-mono">{container.id.slice(0, 12)}</p>
        </div>
      </div>

      {/* Image */}
      <div className="min-w-0">
        <p className="text-sm truncate text-muted-foreground" title={container.image}>
          {container.image}
        </p>
        <p className="text-xs text-muted-foreground/60 flex items-center gap-1 mt-0.5">
          <Clock className="h-3 w-3" />
          {formatDistanceToNow(new Date(container.created), { addSuffix: true })}
        </p>
      </div>

      {/* Status */}
      <div className="w-24">
        <span
          className={cn(
            'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
            isRunning
              ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
              : 'bg-zinc-500/10 text-zinc-600 dark:text-zinc-400'
          )}
        >
          {isRunning && (
            <span className="relative flex h-1.5 w-1.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
            </span>
          )}
          {container.state || container.status}
        </span>
      </div>

      {/* Ports */}
      <div className="w-32">
        {hasPorts ? (
          <div className="flex flex-col gap-1">
            {container.ports.slice(0, 2).map((port, idx) => (
              <span
                key={idx}
                className={cn(
                  'inline-flex items-center gap-1 text-xs font-mono',
                  port.public_port > 0
                    ? 'text-emerald-600 dark:text-emerald-400'
                    : 'text-muted-foreground'
                )}
              >
                {port.public_port > 0 ? (
                  <>
                    {port.public_port}
                    <ArrowRight className="h-2.5 w-2.5" />
                    {port.private_port}
                  </>
                ) : (
                  port.private_port
                )}
              </span>
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

      {/* Actions */}
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
