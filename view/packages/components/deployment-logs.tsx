'use client';

import React from 'react';
import { RefreshCw, X, ChevronRight } from 'lucide-react';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import {
  useDeploymentLogsViewer,
  FormattedLogEntry,
  LogLevel
} from '@/packages/hooks/applications/use_deployment_logs_viewer';
import { useDeploymentLogsTable } from '@/packages/hooks/applications/use_deployment_logs_table';

interface DeploymentLogsTableProps {
  id: string;
  isDeployment?: boolean;
  title?: string;
}

export function DeploymentLogsTable({ id, isDeployment = false, title }: DeploymentLogsTableProps) {
  const {
    logs,
    isLoading,
    toggleLogExpansion,
    isLogExpanded,
    expandAll,
    collapseAll,
    searchTerm,
    setSearchTerm,
    currentPage,
    setCurrentPage,
    totalPages,
    filters,
    setFilters,
    clearFilters,
    isDense,
    setIsDense,
    refreshLogs
  } = useDeploymentLogsViewer({ id, isDeployment });

  const {
    hasLogs,
    hasActiveFilters,
    showLoadMore,
    levelOptions,
    dateFilterOptions,
    tableHeaderColumns,
    loadingSkeletons,
    actionButtons,
    handleLoadMore,
    handleLevelChange,
    handleDateFilterChange
  } = useDeploymentLogsTable({
    id,
    isDeployment,
    logs,
    refreshLogs,
    searchTerm,
    filters,
    currentPage,
    totalPages,
    isDense,
    expandAll,
    collapseAll,
    setCurrentPage,
    setFilters,
    clearFilters,
    setIsDense
  });

  return (
    <CardWrapper className="border-0 shadow-none overflow-x-hidden bg-transparent">
      <div className="space-y-3 pb-4 px-0 border-none border-b-0 min-w-0">
        {title && (
          <div className="flex items-center justify-between min-w-0 gap-2">
            <h3 className="text-lg font-semibold truncate min-w-0">{title}</h3>
            {showLoadMore && (
              <Button variant="outline" size="sm" onClick={handleLoadMore}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Load More
              </Button>
            )}
          </div>
        )}
        <div className="flex items-center gap-4 min-w-0 flex-wrap">
          <div className="relative flex-shrink-0 min-w-0 max-w-full sm:w-96">
            <Input
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex flex-wrap items-center gap-3 ml-auto flex-shrink-0">
            {dateFilterOptions.map((option) => (
              <div key={option.key} className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">{option.label}</span>
                <Input
                  type="date"
                  value={filters[option.key]}
                  onChange={(e) => handleDateFilterChange(option.key, e.target.value)}
                  className="w-40 h-9"
                />
              </div>
            ))}
          </div>
          {hasActiveFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={clearFilters}
              className="text-muted-foreground"
            >
              <X className="h-4 w-4 mr-1" />
              Clear filters
            </Button>
          )}
          {!title && showLoadMore && (
            <Button variant="outline" size="sm" onClick={handleLoadMore}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Load More
            </Button>
          )}
        </div>
        <div className="flex items-center justify-between min-w-0 gap-2 flex-wrap">
          <div className="flex items-center gap-2 flex-wrap min-w-0">
            {levelOptions.map((option) => (
              <Badge
                key={option.value}
                variant={filters.level === option.value ? 'default' : 'outline'}
                className="cursor-pointer transition-colors flex-shrink-0"
                onClick={() => handleLevelChange(option.value)}
              >
                {option.label}
              </Badge>
            ))}
          </div>
          <div className="flex items-center gap-2 flex-shrink-0">
            {actionButtons.map((button, index) => {
              const Icon = button.icon;
              return (
                <React.Fragment key={button.key}>
                  {index === 3 && <div className="w-px h-5 bg-border" />}
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={button.onClick}
                    disabled={button.disabled}
                    title={button.title}
                  >
                    <Icon className={`h-4 w-4 ${button.className || ''}`} />
                  </Button>
                </React.Fragment>
              );
            })}
          </div>
        </div>
      </div>
      <div className="p-0 border rounded-md overflow-hidden min-w-0">
        <div className="flex items-center gap-3 px-4 py-2 border-b bg-muted/30 text-xs font-medium text-muted-foreground uppercase tracking-wider min-w-0">
          {tableHeaderColumns.map((col) => (
            <div key={col.key} className={`${col.width} flex-shrink-0`}>
              {col.label}
            </div>
          ))}
        </div>
        {isLoading && logs.length === 0 ? (
          <div className="p-4 space-y-3">
            {loadingSkeletons.map((i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        ) : logs.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">
            <p>No logs available</p>
          </div>
        ) : (
          <div className="max-h-[600px] overflow-y-auto overflow-x-hidden">
            {logs.map((log) => (
              <div key={log.id}>
                <DeploymentLogRow
                  log={log}
                  isExpanded={isLogExpanded(log.id)}
                  onToggle={() => toggleLogExpansion(log.id)}
                  isDense={isDense}
                />
                {isLogExpanded(log.id) && (
                  <div
                    className={cn(
                      'px-4 bg-muted/20 border-b border-border/50 min-w-0 overflow-x-hidden',
                      isDense ? 'pb-2 pt-1' : 'pb-4 pt-2'
                    )}
                  >
                    <div className="ml-7 min-w-0">
                      <div
                        className={cn(
                          'font-mono bg-muted/50 rounded border break-words whitespace-pre-wrap overflow-wrap-anywhere',
                          isDense ? 'text-xs p-2' : 'text-sm p-3'
                        )}
                      >
                        {log.message}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </CardWrapper>
  );
}

interface DeploymentLogRowProps {
  log: FormattedLogEntry;
  isExpanded: boolean;
  onToggle: () => void;
  isDense?: boolean;
}

const levelStyles: Record<LogLevel, string> = {
  error: 'bg-red-500/10 text-red-500 border-red-500/20',
  warn: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
  info: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  debug: 'bg-gray-500/10 text-gray-500 border-gray-500/20'
};

function DeploymentLogRow({ log, isExpanded, onToggle, isDense }: DeploymentLogRowProps) {
  return (
    <div
      className={cn(
        'flex items-start gap-3 px-4 cursor-pointer transition-colors min-w-0',
        'hover:bg-muted/50 border-b border-border/50',
        isExpanded && 'bg-muted/30',
        isDense ? 'py-1' : 'py-3'
      )}
      onClick={onToggle}
    >
      <div className="flex-shrink-0 mt-0.5">
        <ChevronRight
          className={cn(
            'text-muted-foreground transition-transform duration-200',
            isExpanded && 'rotate-90',
            isDense ? 'h-3 w-3' : 'h-4 w-4'
          )}
        />
      </div>
      <Badge
        variant="outline"
        className={cn(
          'justify-center flex-shrink-0',
          levelStyles[log.level],
          isDense ? 'text-[10px] px-1 py-0 w-12' : 'text-xs px-2 py-0 w-14'
        )}
      >
        {log.level.toUpperCase()}
      </Badge>
      <div className={cn('flex-shrink-0', isDense ? 'w-40' : 'w-44')}>
        <span className={cn('font-mono text-muted-foreground', isDense ? 'text-xs' : 'text-sm')}>
          {log.formattedTime}
        </span>
      </div>
      <div className="flex-1 min-w-0">
        <span
          className={cn(
            isDense ? 'text-xs' : 'text-sm',
            isExpanded
              ? 'whitespace-pre-wrap break-words'
              : 'break-words line-clamp-1 overflow-hidden'
          )}
        >
          {log.message}
        </span>
      </div>
    </div>
  );
}

export default DeploymentLogsTable;
