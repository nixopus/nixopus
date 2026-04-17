'use client';

import React, { useMemo } from 'react';
import PageLayout from '@/packages/layouts/page-layout';
import {
  TypographyMuted,
  DataTable,
  Button,
  Skeleton,
  SelectWrapper,
  Label,
  DialogWrapper,
  PaginationWrapper,
  SearchBar,
} from '@nixopus/ui';
import { HardDrive, Loader2, CalendarClock, RotateCcw, Plus } from 'lucide-react';
import {
  useBackupsView,
  useBackupSchedule,
  statusConfig,
  formatBytes,
  formatDate,
  formatDuration,
  buildRestoreMailto,
  HOUR_OPTIONS,
  DAY_OPTIONS,
  FREQUENCY_OPTIONS,
  RETENTION_OPTIONS,
  STATUS_FILTER_OPTIONS,
} from '@/packages/hooks/backups/use-backups-view';
import type { MachineBackup } from '@/packages/hooks/backups/use-backups-view';
import {
  DATA_TABLE_CLASS,
  TablePanel,
  TableSkeleton,
  EmptyState,
  ListToolbar,
  RefreshButton,
} from '@/components/ui/list-page-chrome';

const BACKUP_TABLE_SKELETON = {
  headerWidths: ['w-20', 'w-28', 'w-14', 'w-14', 'w-24', 'w-14'],
  rows: 4,
  rowConfig: {
    widths: [
      'h-4 w-20',
      'h-3.5 w-32 flex-1',
      'h-3.5 w-14',
      'h-3.5 w-14',
      'h-3.5 w-28',
      'h-3.5 w-14',
    ],
  },
};

function BackupStatusBadge({ status }: { status: MachineBackup['status'] }) {
  const config = statusConfig[status] || statusConfig.pending;
  const Icon = config.icon;
  return (
    <div className={`flex items-center gap-1.5 text-xs font-medium ${config.className}`}>
      <Icon className={`h-3.5 w-3.5 ${status === 'in_progress' ? 'animate-spin' : ''}`} />
      <span>{config.label}</span>
    </div>
  );
}

function BackupScheduleDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { form, isLoading, actions, update } = useBackupSchedule(() => onOpenChange(false));

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title="Automatic Backup Schedule"
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[500px]"
    >
      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-full" />
        </div>
      ) : (
        <div className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Frequency</Label>
              <SelectWrapper
                value={form.frequency}
                onValueChange={(v) => update({ frequency: v as 'daily' | 'weekly' })}
                options={FREQUENCY_OPTIONS}
                placeholder="Frequency"
                className="w-full"
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Time (UTC)</Label>
              <SelectWrapper
                value={String(form.hour_utc)}
                onValueChange={(v) => update({ hour_utc: Number(v) })}
                options={HOUR_OPTIONS}
                placeholder="Hour"
                className="w-full"
              />
            </div>
            {form.frequency === 'weekly' && (
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">Day of Week</Label>
                <SelectWrapper
                  value={String(form.day_of_week)}
                  onValueChange={(v) => update({ day_of_week: Number(v) })}
                  options={DAY_OPTIONS}
                  placeholder="Day"
                  className="w-full"
                />
              </div>
            )}
            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Retention</Label>
              <SelectWrapper
                value={String(form.retention_count)}
                onValueChange={(v) => update({ retention_count: Number(v) })}
                options={RETENTION_OPTIONS}
                placeholder="Keep"
                className="w-full"
              />
            </div>
          </div>
        </div>
      )}
    </DialogWrapper>
  );
}

export const BackupsView = React.memo(function BackupsView() {
  const {
    backups,
    totalCount,
    totalPages,
    page,
    isListLoading,
    isListFetching,
    listError,
    isTriggerLoading,
    hasInProgress,
    search,
    statusFilter,
    sortConfig,
    scheduleOpen,
    setScheduleOpen,
    handleTriggerBackup,
    handleSearch,
    handleSort,
    handleStatusFilter,
    handlePageChange,
    refetch,
  } = useBackupsView();

  const columns = useMemo(
    () => [
      {
        key: 'status',
        title: 'Status',
        sortable: true,
        render: (_: unknown, backup: MachineBackup) => (
          <BackupStatusBadge status={backup.status} />
        ),
      },
      {
        key: 'machine_name',
        title: 'Machine',
        className: 'w-[30%]',
        render: (_: unknown, backup: MachineBackup) => (
          <div className="flex flex-col gap-0.5">
            <span className="text-sm font-medium text-foreground truncate">
              {backup.machine_name}
            </span>
            {backup.error && (
              <span className="text-[11px] text-destructive/80 truncate leading-none">
                {backup.error}
              </span>
            )}
          </div>
        ),
      },
      {
        key: 'size_bytes',
        title: 'Size',
        sortable: true,
        align: 'right' as const,
        render: (_: unknown, backup: MachineBackup) => (
          <span className="text-sm text-muted-foreground">
            {backup.size_bytes > 0 ? formatBytes(backup.size_bytes) : '-'}
          </span>
        ),
      },
      {
        key: 'duration',
        title: 'Duration',
        align: 'right' as const,
        render: (_: unknown, backup: MachineBackup) => (
          <span className="text-sm text-muted-foreground">
            {formatDuration(backup.started_at, backup.completed_at)}
          </span>
        ),
      },
      {
        key: 'created_at',
        title: 'Created',
        sortable: true,
        render: (_: unknown, backup: MachineBackup) => (
          <span className="text-sm text-muted-foreground">{formatDate(backup.created_at)}</span>
        ),
      },
      {
        key: 'actions',
        title: '',
        align: 'right' as const,
        render: (_: unknown, backup: MachineBackup) => {
          if (backup.status !== 'completed') return null;
          return (
            <a
              href={buildRestoreMailto(backup)}
              onClick={(e) => e.stopPropagation()}
              className="inline-flex items-center gap-1.5 text-xs font-medium text-primary hover:text-primary/80 transition-colors"
            >
              <RotateCcw className="h-3.5 w-3.5" />
              Restore
            </a>
          );
        },
      },
    ],
    [],
  );

  return (
    <div>
      <h1 className="text-2xl tracking-wide pb-4 sm:pb-6">Backups</h1>

      <div className="pb-4 pt-6 min-w-0">
        <ListToolbar
          left={
            <>
              <SearchBar
                label="Search backups"
                searchTerm={search}
                handleSearchChange={(e) => handleSearch(e.target.value)}
                className="w-52 sm:w-64 [&_input]:w-full!"
              />
              <SelectWrapper
                value={statusFilter || 'all'}
                onValueChange={(v) => handleStatusFilter(v === 'all' ? '' : v)}
                options={STATUS_FILTER_OPTIONS}
                placeholder="Status"
                className="w-40"
              />
              {totalCount > 1 && (
                <TypographyMuted className="text-xs hidden sm:block">
                  {totalCount} backups
                </TypographyMuted>
              )}
            </>
          }
        >
          <RefreshButton onClick={refetch} isFetching={isListFetching} ariaLabel="Refresh backups" />
          <Button
            variant="outline"
            size="sm"
            className="gap-2"
            onClick={() => setScheduleOpen(true)}
          >
            <CalendarClock className="h-4 w-4" />
            Schedule
          </Button>
          <Button
            onClick={handleTriggerBackup}
            disabled={isTriggerLoading || hasInProgress}
            size="sm"
            className="gap-2"
          >
            {isTriggerLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Plus className="h-4 w-4" />
            )}
            {hasInProgress ? 'Backup in progress...' : 'Backup'}
          </Button>
        </ListToolbar>
      </div>

      <BackupScheduleDialog open={scheduleOpen} onOpenChange={setScheduleOpen} />

      <TablePanel>
        {isListLoading ? (
          <TableSkeleton {...BACKUP_TABLE_SKELETON} />
        ) : listError ? (
          <EmptyState
            icon={HardDrive}
            message="Failed to load backups"
            description="Something went wrong. Please try again."
            onRetry={refetch}
            bordered={false}
          />
        ) : backups.length === 0 ? (
          <EmptyState
            icon={HardDrive}
            message={search || statusFilter ? 'No backups match your filters' : 'No backups yet'}
            description={
              search || statusFilter
                ? undefined
                : 'Create a backup to snapshot your machine and upload to S3.'
            }
            onRetry={
              search || statusFilter
                ? () => {
                    handleSearch('');
                    handleStatusFilter('');
                  }
                : undefined
            }
            retryLabel="Clear filters"
            bordered={false}
          />
        ) : (
          <DataTable
            data={backups}
            columns={columns}
            showBorder={false}
            hoverable
            onSort={handleSort}
            sortConfig={sortConfig}
            tableClassName={DATA_TABLE_CLASS}
          />
        )}
      </TablePanel>

      {totalPages > 1 && (
        <div className="mt-6 flex justify-center">
          <PaginationWrapper
            currentPage={page}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      )}
    </div>
  );
});

export const BackupsWithLayout = React.memo(function BackupsWithLayout() {
  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <BackupsView />
    </PageLayout>
  );
});
