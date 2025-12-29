'use client';

import React, { useState, useMemo, useCallback } from 'react';
import {
  Search,
  ChevronsUpDown,
  RefreshCw,
  X,
  ChevronRight,
  Rows3,
  Rows4,
  Loader2,
  Copy,
  Download,
  Check
} from 'lucide-react';
import { toast } from 'sonner';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Container } from '@/redux/services/container/containerApi';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';

interface LogsTabProps {
  container: Container;
  logs: string;
  onLoadMore: () => void;
}

type LogLevel = 'error' | 'warn' | 'info' | 'debug';

interface ParsedLogEntry {
  id: string;
  timestamp: string;
  formattedTime: string;
  message: string;
  level: LogLevel;
  raw: string;
}

export function LogsTab({ container, logs, onLoadMore }: LogsTabProps) {
  const { t } = useTranslation();
  const [expandedLogIds, setExpandedLogIds] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState('');
  const [levelFilter, setLevelFilter] = useState<LogLevel | 'all'>('all');
  const [isDense, setIsDense] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [isCopied, setIsCopied] = useState(false);
  const [allExpanded, setAllExpanded] = useState(false);

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
    a.download = `${container.name || 'container'}-logs-${new Date().toISOString().split('T')[0]}.log`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success(t('containers.logs.downloadSuccess'));
  }, [parsedLogs, container.name, t]);

  const toggleLogExpansion = useCallback((logId: string) => {
    setExpandedLogIds((prev) => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
        // If we're collapsing a log, reset allExpanded state
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

  const handleLoadMore = async () => {
    setIsLoadingMore(true);
    await onLoadMore();
    setIsLoadingMore(false);
  };

  const hasActiveFilters = searchTerm || levelFilter !== 'all';

  return (
    <div className="space-y-4 overflow-x-hidden">
      {/* Toolbar */}
      <div className="flex flex-col gap-3 min-w-0">
        <div className="flex items-center gap-3 min-w-0">
          <div className="relative flex-1 max-w-md min-w-0">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 bg-transparent"
            />
          </div>
          <div className="flex items-center gap-2 ml-auto flex-shrink-0">
            {hasActiveFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearFilters}
                className="text-muted-foreground h-9"
              >
                <X className="h-4 w-4 mr-1" />
                Clear
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={handleLoadMore}
              disabled={isLoadingMore}
              className="h-9"
            >
              {isLoadingMore ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4 mr-2" />
              )}
              Load more
            </Button>
          </div>
        </div>

        <div className="flex items-center justify-between min-w-0 gap-2">
          <div className="flex items-center gap-1.5 flex-wrap min-w-0">
            {levelOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => setLevelFilter(option.value)}
                className={cn(
                  'px-3 py-1 text-xs font-medium rounded-full transition-colors flex-shrink-0',
                  levelFilter === option.value
                    ? 'bg-foreground text-background'
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                )}
              >
                {option.label}
              </button>
            ))}
          </div>
          <div className="flex items-center gap-1 flex-shrink-0">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCopyLogs}
              className="h-8 px-2 text-muted-foreground"
              title={t('containers.logs.copy')}
              disabled={parsedLogs.length === 0}
            >
              {isCopied ? (
                <Check className="h-4 w-4 text-green-500" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleDownloadLogs}
              className="h-8 px-2 text-muted-foreground"
              title={t('containers.logs.download')}
              disabled={parsedLogs.length === 0}
            >
              <Download className="h-4 w-4" />
            </Button>
            <div className="w-px h-4 bg-border mx-1" />
            <Button
              variant="ghost"
              size="sm"
              onClick={handleExpandCollapseToggle}
              className="h-8 px-2 text-muted-foreground"
              title={allExpanded ? 'Collapse all' : 'Expand all'}
            >
              <ChevronsUpDown className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsDense(!isDense)}
              className="h-8 px-2 text-muted-foreground"
              title={isDense ? 'Normal view' : 'Dense view'}
            >
              {isDense ? <Rows3 className="h-4 w-4" /> : <Rows4 className="h-4 w-4" />}
            </Button>
          </div>
        </div>
      </div>

      {/* Logs List */}
      <div className="rounded-lg border overflow-hidden bg-zinc-950 min-w-0">
        <div className="flex items-center gap-3 px-4 py-2 text-xs font-medium text-zinc-500 uppercase tracking-wider border-b border-zinc-800 min-w-0">
          <div className="w-4 flex-shrink-0" />
          <div className="w-14 flex-shrink-0">Level</div>
          <div className="w-40 flex-shrink-0">Time</div>
          <div className="flex-1 min-w-0">Message</div>
        </div>

        {parsedLogs.length === 0 ? (
          <div className="py-16 text-center text-zinc-500">
            <p className="text-sm">{t('containers.no_logs')}</p>
          </div>
        ) : (
          <div className="max-h-[600px] overflow-y-auto overflow-x-hidden">
            {parsedLogs.map((log) => (
              <LogEntry
                key={log.id}
                log={log}
                isExpanded={isLogExpanded(log.id)}
                onToggle={() => toggleLogExpansion(log.id)}
                isDense={isDense}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

const levelOptions: { value: LogLevel | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'error', label: 'Error' },
  { value: 'warn', label: 'Warn' },
  { value: 'info', label: 'Info' },
  { value: 'debug', label: 'Debug' }
];

const levelColors: Record<LogLevel, { bg: string; text: string }> = {
  error: { bg: 'bg-red-500/20', text: 'text-red-400' },
  warn: { bg: 'bg-amber-500/20', text: 'text-amber-400' },
  info: { bg: 'bg-blue-500/20', text: 'text-blue-400' },
  debug: { bg: 'bg-zinc-500/20', text: 'text-zinc-400' }
};

function LogEntry({
  log,
  isExpanded,
  onToggle,
  isDense
}: {
  log: ParsedLogEntry;
  isExpanded: boolean;
  onToggle: () => void;
  isDense: boolean;
}) {
  const colors = levelColors[log.level];

  return (
    <div className="border-b border-zinc-800/50 last:border-0 min-w-0">
      <div
        className={cn(
          'flex items-start gap-3 px-4 cursor-pointer transition-colors min-w-0',
          'hover:bg-zinc-900/50',
          isExpanded && 'bg-zinc-900/30',
          isDense ? 'py-1' : 'py-2.5'
        )}
        onClick={onToggle}
      >
        <ChevronRight
          className={cn(
            'h-3.5 w-3.5 text-zinc-500 transition-transform duration-150 flex-shrink-0 mt-0.5',
            isExpanded && 'rotate-90'
          )}
        />
        <span
          className={cn(
            'px-2 py-0.5 rounded text-[10px] font-semibold uppercase w-14 text-center flex-shrink-0',
            colors.bg,
            colors.text
          )}
        >
          {log.level}
        </span>
        <span
          className={cn(
            'w-40 font-mono text-zinc-500 flex-shrink-0',
            isDense ? 'text-[11px]' : 'text-xs'
          )}
        >
          {log.formattedTime}
        </span>
        <span
          className={cn(
            'flex-1 text-zinc-300 break-words line-clamp-1 overflow-hidden min-w-0',
            isDense ? 'text-xs' : 'text-sm'
          )}
        >
          {log.message}
        </span>
      </div>

      {isExpanded && (
        <div className="px-4 pb-3 pt-1 min-w-0 overflow-x-hidden">
          <pre
            className={cn(
              'ml-8 p-3 rounded bg-zinc-900 text-zinc-300 font-mono whitespace-pre-wrap break-words overflow-wrap-anywhere',
              isDense ? 'text-[11px]' : 'text-xs'
            )}
          >
            {log.raw}
          </pre>
        </div>
      )}
    </div>
  );
}

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
