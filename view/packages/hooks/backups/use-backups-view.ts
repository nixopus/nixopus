'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { toast } from 'sonner';
import { Clock, Loader2, CheckCircle2, XCircle } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { SelectOption } from '@nixopus/ui';
import { useMachineBackup } from '@/packages/hooks/backups/use-machine-backup';
import { useMachineId } from '@/packages/contexts/machine-context';
import {
  useGetBackupScheduleQuery,
  useUpdateBackupScheduleMutation,
} from '@/redux/services/machine/machineBackupApi';
import type { MachineBackup, BackupScheduleData } from '@/redux/types/machine-backup';

export type { MachineBackup, BackupScheduleData };

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function formatDate(dateStr: string | null): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    year: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export function formatDuration(start: string | null, end: string | null): string {
  if (!start || !end) return '-';
  const ms = new Date(end).getTime() - new Date(start).getTime();
  const sec = Math.round(ms / 1000);
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  const rem = sec % 60;
  return `${min}m ${rem}s`;
}

export function buildRestoreMailto(backup: MachineBackup): string {
  const subject = `Restore Request: ${backup.machine_name}`;
  const body = [
    `Machine: ${backup.machine_name}`,
    `Backup ID: ${backup.id}`,
    `Backup Date: ${formatDate(backup.created_at)}`,
    `Size: ${backup.size_bytes > 0 ? formatBytes(backup.size_bytes) : 'N/A'}`,
    '',
    'Please restore this backup.',
  ].join('\n');
  return `mailto:raghav@nixopus.com?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
}

export const statusConfig: Record<string, { label: string; icon: LucideIcon; className: string }> =
  {
    pending: { label: 'Pending', icon: Clock, className: 'text-muted-foreground' },
    in_progress: { label: 'In Progress', icon: Loader2, className: 'text-primary animate-pulse' },
    completed: {
      label: 'Completed',
      icon: CheckCircle2,
      className: 'text-green-600 dark:text-green-400',
    },
    failed: { label: 'Failed', icon: XCircle, className: 'text-destructive' },
  };

export const HOUR_OPTIONS: SelectOption[] = Array.from({ length: 24 }, (_, i) => ({
  value: String(i),
  label: `${String(i).padStart(2, '0')}:00 UTC`,
}));

export const DAY_OPTIONS: SelectOption[] = [
  { value: '0', label: 'Sunday' },
  { value: '1', label: 'Monday' },
  { value: '2', label: 'Tuesday' },
  { value: '3', label: 'Wednesday' },
  { value: '4', label: 'Thursday' },
  { value: '5', label: 'Friday' },
  { value: '6', label: 'Saturday' },
];

export const FREQUENCY_OPTIONS: SelectOption[] = [
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
];

export const RETENTION_OPTIONS: SelectOption[] = [
  { value: '3', label: '3 backups' },
  { value: '7', label: '7 backups' },
  { value: '14', label: '14 backups' },
  { value: '30', label: '30 backups' },
];

export const STATUS_FILTER_OPTIONS: SelectOption[] = [
  { value: 'all', label: 'Status' },
  { value: 'pending', label: 'Pending' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
];

export function useBackupSchedule(onClose: () => void) {
  const machineId = useMachineId();
  const scheduleArg = machineId ? { server_id: machineId } : undefined;
  const { data: schedule, isLoading } = useGetBackupScheduleQuery(scheduleArg);
  const [updateSchedule, { isLoading: isSaving }] = useUpdateBackupScheduleMutation();

  const [form, setForm] = useState<BackupScheduleData>({
    enabled: false,
    frequency: 'daily',
    hour_utc: 2,
    day_of_week: 0,
    retention_count: 7,
  });
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    if (schedule) {
      setForm(schedule);
      setDirty(false);
    }
  }, [schedule]);

  const update = useCallback((patch: Partial<BackupScheduleData>) => {
    setForm((prev) => ({ ...prev, ...patch }));
    setDirty(true);
  }, []);

  const handleSave = useCallback(async () => {
    try {
      await updateSchedule({ ...form, enabled: true, server_id: machineId ?? undefined }).unwrap();
      toast.success('Backup schedule updated');
      setDirty(false);
      onClose();
    } catch (err: unknown) {
      const e = err as { data?: { detail?: string } };
      toast.error(e?.data?.detail || 'Failed to update schedule');
    }
  }, [form, updateSchedule, onClose, machineId]);

  const handleDisable = useCallback(async () => {
    try {
      await updateSchedule({ ...form, enabled: false, server_id: machineId ?? undefined }).unwrap();
      toast.success('Backup schedule disabled');
      setDirty(false);
      onClose();
    } catch (err: unknown) {
      const e = err as { data?: { detail?: string } };
      toast.error(e?.data?.detail || 'Failed to disable schedule');
    }
  }, [form, updateSchedule, onClose, machineId]);

  const actions = useMemo(
    () => [
      ...(schedule?.enabled
        ? [
            {
              label: isSaving ? 'Disabling...' : 'Disable Schedule',
              onClick: handleDisable,
              variant: 'outline' as const,
              disabled: isSaving,
              loading: isSaving,
            },
          ]
        : []),
      { label: 'Cancel', onClick: onClose, variant: 'outline' as const },
      {
        label: isSaving ? 'Saving...' : 'Save Schedule',
        onClick: handleSave,
        variant: 'default' as const,
        disabled: isSaving || !dirty,
        loading: isSaving,
      },
    ],
    [onClose, isSaving, handleSave, handleDisable, schedule?.enabled, dirty],
  );

  return { form, dirty, isLoading, isSaving, schedule, actions, update, handleSave, handleDisable };
}

export function useBackupsView() {
  const machineId = useMachineId();
  const backup = useMachineBackup(machineId ?? undefined);
  const [scheduleOpen, setScheduleOpen] = useState(false);

  const sortConfig = useMemo(
    () => ({ field: backup.sortBy, order: backup.sortOrder }),
    [backup.sortBy, backup.sortOrder],
  );

  return { ...backup, scheduleOpen, setScheduleOpen, sortConfig };
}
