'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { useParams, useSearchParams } from 'next/navigation';
import {
  useGetExecutionLogsQuery,
  useListExecutionsQuery
} from '@/redux/services/extensions/extensionsApi';
import { Skeleton } from '@/components/ui/skeleton';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import AceEditor from '@/components/ui/ace-editor';
import { Badge } from '@/components/ui/badge';

export default function ExecutionsTab() {
  const { t } = useTranslation();
  const params = useParams();
  const id = (params?.id as string) || '';
  const { data: executions, isLoading } = useListExecutionsQuery(
    { extensionId: id },
    { skip: !id }
  );
  const search = useSearchParams();

  const [selectedExecId, setSelectedExecId] = useState<string | null>(null);
  const [afterSeq, setAfterSeq] = useState<number>(0);
  const [lines, setLines] = useState<string[]>([]);
  const [open, setOpen] = useState(false);
  const editorRef = useRef<any>(null);

  const poll = !!selectedExecId && open;
  const { data: logsResp } = useGetExecutionLogsQuery(
    { executionId: selectedExecId || '', afterSeq, limit: 200 },
    { skip: !poll, pollingInterval: 2500 }
  );

  useEffect(() => {
    if (!logsResp) return;
    if (logsResp.logs && logsResp.logs.length > 0) {
      setAfterSeq(logsResp.next_after);
      setLines((prev) => {
        const appended = logsResp.logs.map((l) => {
          const ts = new Date(l.created_at).toLocaleTimeString();
          const dataStr = l.data
            ? ` ${typeof l.data === 'string' ? l.data : JSON.stringify(l.data)}`
            : '';
          return `[${ts}] [${l.level.toUpperCase()}] ${l.message}${dataStr}`;
        });
        return [...prev, ...appended];
      });
    }
  }, [logsResp]);

  // Auto open logs if URL contains exec and openLogs
  useEffect(() => {
    const exec = search?.get('exec');
    const openLogs = search?.get('openLogs') === '1';
    if (exec && openLogs) {
      setSelectedExecId(exec);
      setOpen(true);
    }
  }, [search]);

  useEffect(() => {
    if (!open) return;
    setAfterSeq(0);
    setLines([]);
  }, [selectedExecId, open]);

  useEffect(() => {
    if (!editorRef.current) return;
    const session = editorRef.current?.getSession?.();
    const total = session?.getLength?.();
    if (typeof total === 'number') {
      editorRef.current.scrollToLine(total, true, true, () => {});
    }
  }, [lines]);

  const onOpenLogs = (execId: string) => {
    setSelectedExecId(execId);
    setOpen(true);
  };

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

  const ExecutionRow = useMemo(
    () =>
      ({ e }: { e: any }) => (
        <div
          key={e.id}
          className="grid grid-cols-12 px-3 py-3 text-sm items-center cursor-pointer hover:bg-muted/30"
          onClick={() => onOpenLogs(e.id)}
        >
          <div className="col-span-4 truncate">{e.id}</div>
          <div className="col-span-2 capitalize">
            <StatusBadge status={e.status} />
          </div>
          <div className="col-span-3 text-muted-foreground">
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
          <div className="col-span-4">{t('extensions.executionId') || 'Execution ID'}</div>
          <div className="col-span-2">{t('extensions.status') || 'Status'}</div>
          <div className="col-span-3">{t('extensions.startedAt') || 'Started At'}</div>
          <div className="col-span-3">{t('extensions.completedAt') || 'Completed At'}</div>
        </div>
        <div className="divide-y">
          {isLoading &&
            Array.from({ length: 5 }).map((_, i) => (
              <div key={`s-${i}`} className="grid grid-cols-12 px-3 py-3 text-sm">
                <div className="col-span-4">
                  <Skeleton className="h-4 w-40" />
                </div>
                <div className="col-span-2">
                  <Skeleton className="h-4 w-20" />
                </div>
                <div className="col-span-3">
                  <Skeleton className="h-4 w-28" />
                </div>
                <div className="col-span-3">
                  <Skeleton className="h-4 w-28" />
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

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="right" className="sm:max-w-2xl">
          <SheetHeader>
            <SheetTitle>{t('extensions.logs') || 'Logs'}</SheetTitle>
          </SheetHeader>
          <div className="p-3">
            <div className="mb-2 text-xs text-muted-foreground">
              {t('extensions.executionId') || 'Execution ID'}: {selectedExecId}
            </div>
            <AceEditor
              onLoad={(editor: any) => {
                editorRef.current = editor;
              }}
              mode="sh"
              value={lines.join('\n')}
              onChange={() => {}}
              name="extension-logs"
              readOnly
              height="70vh"
            />
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
