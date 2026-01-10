import { useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import {
  useGetExecutionLogsQuery,
  useListExecutionsQuery
} from '@/redux/services/extensions/extensionsApi';
import type { ExtensionLog } from '@/redux/types/extension';

export function useExecutionLogs(extensionId: string) {
  const search = useSearchParams();
  const [selectedExecId, setSelectedExecId] = useState<string | null>(null);
  const [afterSeq, setAfterSeq] = useState<number>(0);
  const [allLogs, setAllLogs] = useState<ExtensionLog[]>([]);
  const [open, setOpen] = useState(false);
  const prevExecIdRef = useRef<string | null>(null);
  const prevExtensionIdRef = useRef<string>(extensionId);

  const { data: executions, isLoading } = useListExecutionsQuery(
    { extensionId },
    { skip: !extensionId }
  );

  const [executionStatus, setExecutionStatus] = useState<string | null>(null);
  const poll =
    !!selectedExecId &&
    open &&
    executionStatus !== 'completed' &&
    executionStatus !== 'failed' &&
    executionStatus !== 'cancelled';

  const shouldFetch = !!selectedExecId && open;
  const { data: logsResp, refetch } = useGetExecutionLogsQuery(
    { executionId: selectedExecId || '', afterSeq, limit: 200 },
    { skip: !shouldFetch, pollingInterval: poll ? 2500 : undefined }
  );

  useEffect(() => {
    if (prevExtensionIdRef.current !== extensionId) {
      setSelectedExecId(null);
      setAfterSeq(0);
      setAllLogs([]);
      setExecutionStatus(null);
      setOpen(false);
      prevExecIdRef.current = null;
      prevExtensionIdRef.current = extensionId;
      return;
    }
  }, [extensionId]);

  useEffect(() => {
    if (!selectedExecId) {
      setAfterSeq(0);
      setAllLogs([]);
      setExecutionStatus(null);
      prevExecIdRef.current = null;
      return;
    }

    const execIdChanged = prevExecIdRef.current !== selectedExecId;
    if (execIdChanged) {
      setAfterSeq(0);
      setAllLogs([]);
      setExecutionStatus(null);
      prevExecIdRef.current = selectedExecId;
      if (open) {
        refetch();
      }
    }
  }, [selectedExecId, open, refetch]);

  useEffect(() => {
    if (
      open &&
      selectedExecId &&
      prevExecIdRef.current === selectedExecId &&
      afterSeq === 0 &&
      allLogs.length === 0
    ) {
      refetch();
    }
  }, [open, selectedExecId, afterSeq, allLogs.length, refetch]);

  useEffect(() => {
    if (!logsResp || !selectedExecId) return;

    if (logsResp.execution_status) {
      setExecutionStatus(logsResp.execution_status);
    }

    if (logsResp.logs && logsResp.logs.length > 0) {
      setAfterSeq(logsResp.next_after);
      setAllLogs((prev) => {
        const existingIds = new Set(prev.map((l) => l.id));
        const newLogs = logsResp.logs.filter((l) => !existingIds.has(l.id));
        const merged = [...prev, ...newLogs];
        return merged.sort((a, b) => {
          const timeA = new Date(a.created_at).getTime();
          const timeB = new Date(b.created_at).getTime();
          if (timeA !== timeB) return timeA - timeB;
          return a.sequence - b.sequence;
        });
      });
    }
  }, [logsResp, selectedExecId]);

  useEffect(() => {
    const exec = search?.get('exec');
    const openLogs = search?.get('openLogs') === '1';
    if (exec && openLogs) {
      setSelectedExecId(exec);
      setOpen(true);
    }
  }, [search]);

  const onOpenLogs = (execId: string) => {
    setSelectedExecId(execId);
    setOpen(true);
  };

  return {
    executions,
    isLoading,
    allLogs,
    open,
    selectedExecId,
    setOpen,
    onOpenLogs
  };
}
