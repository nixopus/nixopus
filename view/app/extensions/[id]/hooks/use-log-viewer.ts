import { useEffect, useMemo, useRef, useState } from 'react';
import type { ExtensionLog } from '@/redux/types/extension';
import { formatLog, type FormattedLog } from '../components/utils/log-formatter';

export function useLogViewer(params: {
  open: boolean;
  executionId: string | null;
  logs: ExtensionLog[];
}) {
  const { open, executionId, logs } = params;
  const [formattedLogs, setFormattedLogs] = useState<FormattedLog[]>([]);
  const [collapsedLogs, setCollapsedLogs] = useState<Set<string>>(new Set());
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const stepCompletedIds = new Set<string>(
      logs
        .filter((log) => log.message?.startsWith('step_completed'))
        .map((log) => log.step_id)
        .filter((id): id is string => Boolean(id))
    );

    const stepFailedIds = new Set<string>(
      logs
        .filter((log) => log.message?.startsWith('step_failed'))
        .map((log) => log.step_id)
        .filter((id): id is string => Boolean(id))
    );

    const formatted = logs.map((log) => {
      const isStepCompleted = Boolean(log.step_id && stepCompletedIds.has(log.step_id));
      const isStepFailed = Boolean(log.step_id && stepFailedIds.has(log.step_id));
      return formatLog(log, isStepCompleted, isStepFailed);
    });
    setFormattedLogs(formatted);

    setCollapsedLogs((prev) => {
      const next = new Set(prev);
      formatted.forEach((log) => {
        if (log.isVerbose && log.data != null && !next.has(log.id)) {
          next.add(log.id);
        }
      });
      return next;
    });
  }, [logs, executionId]);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [formattedLogs]);

  useEffect(() => {
    if (!open || !executionId) {
      return;
    }
    if (logs.length === 0 && formattedLogs.length === 0) {
      return;
    }
  }, [open, executionId, logs.length, formattedLogs.length]);

  const toggleCollapse = (logId: string) => {
    setCollapsedLogs((prev) => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
      } else {
        next.add(logId);
      }
      return next;
    });
  };

  const isEmpty = useMemo(() => formattedLogs.length === 0, [formattedLogs]);

  return {
    formattedLogs,
    collapsedLogs,
    toggleCollapse,
    logsEndRef,
    isEmpty
  };
}
