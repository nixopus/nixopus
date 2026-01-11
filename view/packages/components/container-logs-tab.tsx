'use client';

import React from 'react';
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
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Container } from '@/redux/services/container/containerApi';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { cn } from '@/lib/utils';
import {
  useContainerLogs,
  levelOptions,
  levelColors,
  ParsedLogEntry
} from '@/packages/hooks/containers/use-container-logs';
import { LogsTabProps, LogEntryProps } from '@/packages/types/containers';

export function LogsTab({ container, logs, onLoadMore, onRefresh }: LogsTabProps) {
  const { t } = useTranslation();
  const {
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
  } = useContainerLogs(logs, container.name);

  const handleLoadMore = async () => {
    setIsLoadingMore(true);
    await onLoadMore();
    setIsLoadingMore(false);
  };

  const handleRefresh = async () => {
    if (!onRefresh) return;
    setIsRefreshing(true);
    try {
      await onRefresh();
    } finally {
      setIsRefreshing(false);
    }
  };

  return (
    <div className="space-y-4 overflow-x-hidden">
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
            {onRefresh && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefresh}
                disabled={isRefreshing}
                className="h-9"
                title="Refresh logs"
              >
                {isRefreshing ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4 mr-2" />
                )}
                Refresh
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

function LogEntry({ log, isExpanded, onToggle, isDense }: LogEntryProps) {
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
