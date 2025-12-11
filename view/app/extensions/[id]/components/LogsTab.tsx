'use client';

import { useMemo } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { useParams } from 'next/navigation';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { LogViewer } from './LogViewer';
import { useExecutionLogs } from '../hooks/use-execution-logs';
import { Server } from 'lucide-react';

export default function ExecutionsTab() {
  const { t } = useTranslation();
  const params = useParams();
  const id = (params?.id as string) || '';

  const { executions, isLoading, allLogs, open, selectedExecId, setOpen, onOpenLogs } =
    useExecutionLogs(id);

  const StatusBadge = ({ status }: { status: string }) => {
    const s = (status || '').toLowerCase();
    const cls =
      s === 'completed'
        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
        : s === 'failed'
          ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
          : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400';
    return <Badge className={cls}>{status}</Badge>;
  };

  const ServerBadge = ({ hostname }: { hostname?: string }) => {
    if (!hostname) return null;
    const isLocal = hostname === 'localhost' || hostname === '127.0.0.1';
    return (
      <Badge
        variant="outline"
        className="text-xs flex items-center gap-1 bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border-blue-200 dark:border-blue-800"
      >
        <Server className="h-3 w-3" />
        {isLocal ? 'Local' : hostname}
      </Badge>
    );
  };

  const ExecutionRow = useMemo(
    () =>
      ({ e }: { e: any }) => (
        <div
          key={e.id}
          className="grid grid-cols-12 px-3 py-3 text-sm items-center cursor-pointer hover:bg-muted/30"
          onClick={() => onOpenLogs(e.id)}
        >
          <div className="col-span-3 truncate">{e.id}</div>
          <div className="col-span-2 capitalize">
            <StatusBadge status={e.status} />
          </div>
          <div className="col-span-2">
            <ServerBadge hostname={e.server_hostname} />
          </div>
          <div className="col-span-2 text-muted-foreground">
            {new Date(e.started_at).toLocaleString()}
          </div>
          <div className="col-span-3 text-muted-foreground">
            {e.completed_at ? new Date(e.completed_at).toLocaleString() : '-'}
          </div>
        </div>
      ),
    []
  );

  return (
    <div className="space-y-3">
      <div className="rounded-md border overflow-hidden">
        <div className="grid grid-cols-12 bg-muted/50 px-3 py-2 text-xs font-medium text-muted-foreground">
          <div className="col-span-3">{t('extensions.executionId') || 'Execution ID'}</div>
          <div className="col-span-2">{t('extensions.status') || 'Status'}</div>
          <div className="col-span-2">{t('extensions.server') || 'Server'}</div>
          <div className="col-span-2">{t('extensions.startedAt') || 'Started At'}</div>
          <div className="col-span-3">{t('extensions.completedAt') || 'Completed At'}</div>
        </div>
        <div className="divide-y">
          {isLoading &&
            Array.from({ length: 5 }).map((_, i) => (
              <div key={`s-${i}`} className="grid grid-cols-12 px-3 py-3 text-sm">
                <div className="col-span-3">
                  <Skeleton className="h-4 w-32" />
                </div>
                <div className="col-span-2">
                  <Skeleton className="h-4 w-20" />
                </div>
                <div className="col-span-2">
                  <Skeleton className="h-4 w-20" />
                </div>
                <div className="col-span-2">
                  <Skeleton className="h-4 w-24" />
                </div>
                <div className="col-span-3">
                  <Skeleton className="h-4 w-24" />
                </div>
              </div>
            ))}
          {!isLoading && (executions || []).map((e) => <ExecutionRow key={e.id} e={e} />)}
          {!isLoading && (!executions || executions.length === 0) && (
            <div className="px-3 py-6 text-sm text-muted-foreground">
              {t('extensions.noExecutions') || 'No executions yet.'}
            </div>
          )}
        </div>
      </div>

      <LogViewer open={open} onOpenChange={setOpen} executionId={selectedExecId} logs={allLogs} />
    </div>
  );
}
