import { useState, useMemo, useCallback, useEffect } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

const LOGS_DENSE_MODE_KEY = 'nixopus_logs_dense_mode';

type LogLevel = 'error' | 'warn' | 'info' | 'debug';

export interface ParsedLogEntry {
  id: string;
  timestamp: string;
  formattedTime: string;
  message: string;
  level: LogLevel;
  raw: string;
}

export const useContainerLogs = (logs: string, containerName?: string) => {
  const { t } = useTranslation();
  const [expandedLogIds, setExpandedLogIds] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState('');
  const [levelFilter, setLevelFilter] = useState<LogLevel | 'all'>('all');
  const [isDense, setIsDense] = useState(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(LOGS_DENSE_MODE_KEY);
      return stored !== null ? stored === 'true' : true;
    }
    return true;
  });
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCopied, setIsCopied] = useState(false);
  const [allExpanded, setAllExpanded] = useState(false);

  useEffect(() => {
    localStorage.setItem(LOGS_DENSE_MODE_KEY, isDense.toString());
  }, [isDense]);

  const parsedLogs = useMemo(() => {
    return parseContainerLogs(logs, searchTerm, levelFilter);
  }, [logs, searchTerm, levelFilter]);

  const handleCopyLogs = useCallback(async () => {
    const logText = parsedLogs.map((log) => log.raw).join('\n');
    if (!logText) {
      toast.error(t('containers.logs.copyEmpty'));
      return;
    }
    try {
      await navigator.clipboard.writeText(logText);
      setIsCopied(true);
      toast.success(t('containers.logs.copySuccess'));
      setTimeout(() => setIsCopied(false), 2000);
    } catch {
      toast.error(t('containers.logs.copyError'));
    }
  }, [parsedLogs, t]);

  const handleDownloadLogs = useCallback(() => {
    const logText = parsedLogs.map((log) => log.raw).join('\n');
    if (!logText) {
      toast.error(t('containers.logs.downloadEmpty'));
      return;
    }
    const blob = new Blob([logText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${containerName || 'container'}-logs-${new Date().toISOString().split('T')[0]}.log`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success(t('containers.logs.downloadSuccess'));
  }, [parsedLogs, containerName, t]);

  const toggleLogExpansion = useCallback((logId: string) => {
    setExpandedLogIds((prev) => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
        setAllExpanded(false);
      } else {
        next.add(logId);
      }
      return next;
    });
  }, []);

  const isLogExpanded = useCallback((logId: string) => expandedLogIds.has(logId), [expandedLogIds]);

  const expandAll = useCallback(() => {
    setExpandedLogIds(new Set(parsedLogs.map((log) => log.id)));
    setAllExpanded(true);
  }, [parsedLogs]);

  const collapseAll = useCallback(() => {
    setExpandedLogIds(new Set());
    setAllExpanded(false);
  }, []);

  const handleExpandCollapseToggle = useCallback(() => {
    if (allExpanded) {
      collapseAll();
    } else {
      expandAll();
    }
  }, [allExpanded, expandAll, collapseAll]);

  const clearFilters = useCallback(() => {
    setSearchTerm('');
    setLevelFilter('all');
  }, []);

  const hasActiveFilters = searchTerm || levelFilter !== 'all';

  return {
    parsedLogs,
    searchTerm,
    setSearchTerm,
    levelFilter,
    setLevelFilter,
    isDense,
    setIsDense,
    isLoadingMore,
    setIsLoadingMore,
    isRefreshing,
    setIsRefreshing,
    isCopied,
    allExpanded,
    handleCopyLogs,
    handleDownloadLogs,
    toggleLogExpansion,
    isLogExpanded,
    handleExpandCollapseToggle,
    clearFilters,
    hasActiveFilters
  };
};

function parseContainerLogs(
  logsString: string,
  searchTerm: string,
  levelFilter: LogLevel | 'all'
): ParsedLogEntry[] {
  if (!logsString) return [];

  const lines = logsString.split('\n').filter((line) => line.trim());

  return lines
    .map((line, index) => {
      const timestamp = extractTimestamp(line);
      const level = detectLogLevel(line);
      const message = cleanLogMessage(line);

      return {
        id: `log-${index}`,
        timestamp: timestamp || new Date().toISOString(),
        formattedTime: formatTimestamp(timestamp),
        message,
        level,
        raw: line
      };
    })
    .filter((log) => {
      const matchesSearch =
        !searchTerm || log.message.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesLevel = levelFilter === 'all' || log.level === levelFilter;
      return matchesSearch && matchesLevel;
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
      hour12: false
    });
  } catch {
    return timestamp;
  }
}

export const levelOptions: { value: LogLevel | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'error', label: 'Error' },
  { value: 'warn', label: 'Warn' },
  { value: 'info', label: 'Info' },
  { value: 'debug', label: 'Debug' }
];

export const levelColors: Record<LogLevel, { bg: string; text: string }> = {
  error: { bg: 'bg-red-500/20', text: 'text-red-400' },
  warn: { bg: 'bg-amber-500/20', text: 'text-amber-400' },
  info: { bg: 'bg-blue-500/20', text: 'text-blue-400' },
  debug: { bg: 'bg-zinc-500/20', text: 'text-zinc-400' }
};
