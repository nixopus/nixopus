'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  useGetContainersQuery,
  useLazyGetContainerLogsQuery
} from '@/redux/services/container/containerApi';
import { getAdvancedSettings } from '@/packages/utils/advanced-settings';
import type { FormattedLogEntry, LogLevel } from './use_deployment_logs_viewer';

export function useApplicationLogs(applicationId: string) {
  const { data } = useGetContainersQuery(
    { page: 1, page_size: 100, sort_by: 'name', sort_order: 'asc' },
    { refetchOnMountOrArgChange: true }
  );

  const containers = useMemo(() => {
    const group = data?.groups?.find((g) => g.application_id === applicationId);
    return group?.containers ?? [];
  }, [data?.groups, applicationId]);

  const containerSettings = getAdvancedSettings();
  const [fetchContainerLogs] = useLazyGetContainerLogsQuery();
  const [containerLogEntries, setContainerLogEntries] = useState<FormattedLogEntry[]>([]);

  const fetchAllContainerLogs = useCallback(async () => {
    if (containers.length === 0) {
      setContainerLogEntries([]);
      return;
    }

    const allEntries: FormattedLogEntry[] = [];

    for (const container of containers) {
      try {
        const result = await fetchContainerLogs({
          containerId: container.id,
          tail: containerSettings.containerLogTailLines
        }).unwrap();

        if (result) {
          const name = container.name?.replace(/^\//, '') || container.id.slice(0, 12);
          const entries = parseContainerLogsToEntries(result, name);
          allEntries.push(...entries);
        }
      } catch {
        // skip containers that fail to fetch logs
      }
    }

    setContainerLogEntries(allEntries);
  }, [containers, containerSettings.containerLogTailLines, fetchContainerLogs]);

  useEffect(() => {
    fetchAllContainerLogs();
  }, [fetchAllContainerLogs]);

  return { containerLogEntries, refreshContainerLogs: fetchAllContainerLogs };
}

function parseContainerLogsToEntries(
  logsString: string,
  containerName: string
): FormattedLogEntry[] {
  if (!logsString) return [];

  const lines = logsString.split('\n').filter((line) => line.trim());

  return lines.map((line, index) => {
    const timestamp = extractTimestamp(line);
    const level = detectLogLevel(line);
    const message = cleanLogMessage(line);

    const createdAt = timestamp || new Date().toISOString();

    return {
      id: `container-${containerName}-${index}`,
      timestamp: createdAt,
      formattedTime: formatTimestamp(timestamp),
      message: `[${containerName}] ${message}`,
      level,
      raw: {
        id: `container-${containerName}-${index}`,
        application_id: '',
        application_deployment_id: '',
        log: `[${containerName}] ${line}`,
        created_at: createdAt,
        updated_at: createdAt
      }
    };
  });
}

function extractTimestamp(line: string): string {
  const isoMatch = line.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
  if (isoMatch) return isoMatch[0];

  const commonMatch = line.match(/^\[?(\d{4}[-/]\d{2}[-/]\d{2}\s+\d{2}:\d{2}:\d{2})/);
  if (commonMatch) return commonMatch[1];

  return '';
}

function cleanLogMessage(line: string): string {
  return line
    .replace(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?\s*/, '')
    .replace(/^\[?\d{4}[-/]\d{2}[-/]\d{2}\s+\d{2}:\d{2}:\d{2}\]?\s*/, '')
    .trim();
}

function detectLogLevel(message: string): LogLevel {
  const lower = message.toLowerCase();
  if (
    lower.includes('error') ||
    lower.includes('failed') ||
    lower.includes('exception') ||
    lower.includes('fatal')
  ) {
    return 'error';
  }
  if (lower.includes('warn') || lower.includes('warning')) {
    return 'warn';
  }
  if (lower.includes('debug') || lower.includes('trace')) {
    return 'debug';
  }
  return 'info';
}

function formatTimestamp(timestamp: string): string {
  if (!timestamp) return '—';
  try {
    const date = new Date(timestamp);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: true
    });
  } catch {
    return timestamp;
  }
}
