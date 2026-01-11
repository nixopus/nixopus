'use client';

import React, { useState, useCallback } from 'react';
import {
  Search,
  ChevronsUpDown,
  RefreshCw,
  X,
  Calendar,
  Rows3,
  Rows4,
  Copy,
  Download,
  Check,
  Loader2
} from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';
import { Card, CardHeader, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import {
  useDeploymentLogsViewer,
  LogFilters,
  LogLevel
} from '../../hooks/use_deployment_logs_viewer';
import { DeploymentLogRow } from './DeploymentLogRow';
import { DeploymentLogDetails } from './DeploymentLogDetails';

interface DeploymentLogsTableProps {
  id: string;
  isDeployment?: boolean;
  title?: string;
}

export function DeploymentLogsTable({ id, isDeployment = false, title }: DeploymentLogsTableProps) {
  const { t } = useTranslation();
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

  const [isCopied, setIsCopied] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleCopyLogs = useCallback(async () => {
    const logText = logs
      .map((log) => `[${log.formattedTime}] [${log.level.toUpperCase()}] ${log.message}`)
      .join('\n');
    if (!logText) {
      toast.error(t('selfHost.logs.copyEmpty'));
      return;
    }
    try {
      await navigator.clipboard.writeText(logText);
      setIsCopied(true);
      toast.success(t('selfHost.logs.copySuccess'));
      setTimeout(() => setIsCopied(false), 2000);
    } catch {
      toast.error(t('selfHost.logs.copyError'));
    }
  }, [logs, t]);

  const handleDownloadLogs = useCallback(() => {
    const logText = logs
      .map((log) => `[${log.formattedTime}] [${log.level.toUpperCase()}] ${log.message}`)
      .join('\n');
    if (!logText) {
      toast.error(t('selfHost.logs.downloadEmpty'));
      return;
    }
    const blob = new Blob([logText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${isDeployment ? 'deployment' : 'application'}-${id}-logs-${new Date().toISOString().split('T')[0]}.log`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success(t('selfHost.logs.downloadSuccess'));
  }, [logs, id, isDeployment, t]);

  const handleRefreshLogs = useCallback(async () => {
    setIsRefreshing(true);
    try {
      await refreshLogs();
    } finally {
      setIsRefreshing(false);
    }
  }, [refreshLogs]);

  return (
    <Card className="border-0 shadow-none overflow-x-hidden">
      <LogsHeader
        title={title}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        onExpandAll={expandAll}
        onCollapseAll={collapseAll}
        currentPage={currentPage}
        totalPages={totalPages}
        onLoadMore={() => setCurrentPage(currentPage + 1)}
        filters={filters}
        onFiltersChange={setFilters}
        onClearFilters={clearFilters}
        isDense={isDense}
        onDenseChange={setIsDense}
        onCopyLogs={handleCopyLogs}
        onDownloadLogs={handleDownloadLogs}
        isCopied={isCopied}
        hasLogs={logs.length > 0}
        onRefresh={handleRefreshLogs}
        isRefreshing={isRefreshing}
      />
      <CardContent className="p-0 border rounded-md overflow-hidden min-w-0">
        <TableHeader />
        <LogsList
          logs={logs}
          isLoading={isLoading}
          isLogExpanded={isLogExpanded}
          onToggle={toggleLogExpansion}
          isDense={isDense}
        />
      </CardContent>
    </Card>
  );
}

interface LogsHeaderProps {
  title?: string;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  onExpandAll: () => void;
  onCollapseAll: () => void;
  currentPage: number;
  totalPages: number;
  onLoadMore: () => void;
  filters: LogFilters;
  onFiltersChange: (filters: LogFilters) => void;
  onClearFilters: () => void;
  isDense: boolean;
  onDenseChange: (isDense: boolean) => void;
  onCopyLogs: () => void;
  onDownloadLogs: () => void;
  isCopied: boolean;
  hasLogs: boolean;
  onRefresh: () => void;
  isRefreshing: boolean;
}

function LogsHeader({
  title,
  searchTerm,
  onSearchChange,
  onExpandAll,
  onCollapseAll,
  currentPage,
  totalPages,
  onLoadMore,
  filters,
  onFiltersChange,
  onClearFilters,
  isDense,
  onDenseChange,
  onCopyLogs,
  onDownloadLogs,
  isCopied,
  hasLogs,
  onRefresh,
  isRefreshing
}: LogsHeaderProps) {
  const hasActiveFilters =
    filters.startDate || filters.endDate || filters.level !== 'all' || searchTerm;

  return (
    <CardHeader className="space-y-3 pb-4 px-0 border-none border-b-0 min-w-0">
      {title && (
        <div className="flex items-center justify-between min-w-0 gap-2">
          <h3 className="text-lg font-semibold truncate min-w-0">{title}</h3>
          <LoadMoreButton
            currentPage={currentPage}
            totalPages={totalPages}
            onLoadMore={onLoadMore}
          />
        </div>
      )}
      <div className="flex items-center gap-4 min-w-0 flex-wrap">
        <SearchInput value={searchTerm} onChange={onSearchChange} />
        <DateFiltersRow filters={filters} onFiltersChange={onFiltersChange} />
        {hasActiveFilters && <ClearFiltersButton onClick={onClearFilters} />}
        {!title && (
          <LoadMoreButton
            currentPage={currentPage}
            totalPages={totalPages}
            onLoadMore={onLoadMore}
          />
        )}
      </div>
      <div className="flex items-center justify-between min-w-0 gap-2 flex-wrap">
        <LevelFilter
          value={filters.level}
          onChange={(v) => onFiltersChange({ ...filters, level: v })}
        />
        <div className="flex items-center gap-2 flex-shrink-0">
          <RefreshButton onRefresh={onRefresh} isRefreshing={isRefreshing} />
          <CopyDownloadButtons
            onCopyLogs={onCopyLogs}
            onDownloadLogs={onDownloadLogs}
            isCopied={isCopied}
            hasLogs={hasLogs}
          />
          <div className="w-px h-5 bg-border" />
          <ExpandCollapseButton onExpandAll={onExpandAll} onCollapseAll={onCollapseAll} />
          <DenseToggle isDense={isDense} onChange={onDenseChange} />
        </div>
      </div>
    </CardHeader>
  );
}

function SearchInput({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <div className="relative flex-shrink-0 min-w-0 max-w-full sm:w-96">
      <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
      <Input
        placeholder="Search logs..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="pl-10"
      />
    </div>
  );
}

function DateFiltersRow({
  filters,
  onFiltersChange
}: {
  filters: LogFilters;
  onFiltersChange: (filters: LogFilters) => void;
}) {
  return (
    <div className="flex flex-wrap items-center gap-3 ml-auto flex-shrink-0">
      <DateFilter
        label="From"
        value={filters.startDate}
        onChange={(v) => onFiltersChange({ ...filters, startDate: v })}
      />
      <DateFilter
        label="To"
        value={filters.endDate}
        onChange={(v) => onFiltersChange({ ...filters, endDate: v })}
      />
    </div>
  );
}

function DateFilter({
  label,
  value,
  onChange
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <Calendar className="h-4 w-4 text-muted-foreground" />
      <span className="text-sm text-muted-foreground">{label}</span>
      <Input
        type="date"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-40 h-9"
      />
    </div>
  );
}

const levelOptions: { value: LogLevel | 'all'; label: string }[] = [
  { value: 'all', label: 'All Levels' },
  { value: 'error', label: 'Error' },
  { value: 'warn', label: 'Warning' },
  { value: 'info', label: 'Info' },
  { value: 'debug', label: 'Debug' }
];

function LevelFilter({
  value,
  onChange
}: {
  value: LogLevel | 'all';
  onChange: (v: LogLevel | 'all') => void;
}) {
  return (
    <div className="flex items-center gap-2 flex-wrap min-w-0">
      {levelOptions.map((option) => (
        <Badge
          key={option.value}
          variant={value === option.value ? 'default' : 'outline'}
          className="cursor-pointer transition-colors flex-shrink-0"
          onClick={() => onChange(option.value)}
        >
          {option.label}
        </Badge>
      ))}
    </div>
  );
}

function ClearFiltersButton({ onClick }: { onClick: () => void }) {
  return (
    <Button variant="ghost" size="sm" onClick={onClick} className="text-muted-foreground">
      <X className="h-4 w-4 mr-1" />
      Clear filters
    </Button>
  );
}

function ExpandCollapseButton({
  onExpandAll,
  onCollapseAll
}: {
  onExpandAll: () => void;
  onCollapseAll: () => void;
}) {
  const [expanded, setExpanded] = React.useState(false);

  const handleClick = () => {
    if (expanded) {
      onCollapseAll();
    } else {
      onExpandAll();
    }
    setExpanded(!expanded);
  };

  return (
    <Button variant="outline" size="icon" onClick={handleClick} title="Expand/Collapse all">
      <ChevronsUpDown className="h-4 w-4" />
    </Button>
  );
}

function RefreshButton({
  onRefresh,
  isRefreshing
}: {
  onRefresh: () => void;
  isRefreshing: boolean;
}) {
  return (
    <Button
      variant="outline"
      size="icon"
      onClick={onRefresh}
      disabled={isRefreshing}
      title="Refresh logs"
    >
      {isRefreshing ? (
        <Loader2 className="h-4 w-4 animate-spin" />
      ) : (
        <RefreshCw className="h-4 w-4" />
      )}
    </Button>
  );
}

function DenseToggle({ isDense, onChange }: { isDense: boolean; onChange: (v: boolean) => void }) {
  return (
    <Button
      variant="outline"
      size="icon"
      onClick={() => onChange(!isDense)}
      title={isDense ? 'Normal view' : 'Dense view'}
    >
      {isDense ? <Rows3 className="h-4 w-4" /> : <Rows4 className="h-4 w-4" />}
    </Button>
  );
}

function CopyDownloadButtons({
  onCopyLogs,
  onDownloadLogs,
  isCopied,
  hasLogs
}: {
  onCopyLogs: () => void;
  onDownloadLogs: () => void;
  isCopied: boolean;
  hasLogs: boolean;
}) {
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="outline"
        size="icon"
        onClick={onCopyLogs}
        disabled={!hasLogs}
        title="Copy logs"
      >
        {isCopied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
      </Button>
      <Button
        variant="outline"
        size="icon"
        onClick={onDownloadLogs}
        disabled={!hasLogs}
        title="Download logs"
      >
        <Download className="h-4 w-4" />
      </Button>
    </div>
  );
}

function LoadMoreButton({
  currentPage,
  totalPages,
  onLoadMore
}: {
  currentPage: number;
  totalPages: number;
  onLoadMore: () => void;
}) {
  if (currentPage >= totalPages) return null;

  return (
    <Button variant="outline" size="sm" onClick={onLoadMore}>
      <RefreshCw className="h-4 w-4 mr-2" />
      Load More
    </Button>
  );
}

function TableHeader() {
  return (
    <div className="flex items-center gap-3 px-4 py-2 border-b bg-muted/30 text-xs font-medium text-muted-foreground uppercase tracking-wider min-w-0">
      <div className="w-4 flex-shrink-0" />
      <div className="w-14 flex-shrink-0">Level</div>
      <div className="w-44 flex-shrink-0">Timestamp</div>
      <div className="flex-1 min-w-0">Message</div>
    </div>
  );
}

interface LogsListProps {
  logs: ReturnType<typeof useDeploymentLogsViewer>['logs'];
  isLoading: boolean;
  isLogExpanded: (id: string) => boolean;
  onToggle: (id: string) => void;
  isDense: boolean;
}

function LogsList({ logs, isLoading, isLogExpanded, onToggle, isDense }: LogsListProps) {
  if (isLoading && logs.length === 0) {
    return <LoadingState />;
  }

  if (logs.length === 0) {
    return <EmptyState />;
  }

  return (
    <div className="max-h-[600px] overflow-y-auto overflow-x-hidden">
      {logs.map((log) => (
        <LogItem
          key={log.id}
          log={log}
          isExpanded={isLogExpanded(log.id)}
          onToggle={() => onToggle(log.id)}
          isDense={isDense}
        />
      ))}
    </div>
  );
}

function LogItem({
  log,
  isExpanded,
  onToggle,
  isDense
}: {
  log: ReturnType<typeof useDeploymentLogsViewer>['logs'][0];
  isExpanded: boolean;
  onToggle: () => void;
  isDense: boolean;
}) {
  return (
    <div>
      <DeploymentLogRow log={log} isExpanded={isExpanded} onToggle={onToggle} isDense={isDense} />
      {isExpanded && <DeploymentLogDetails log={log} isDense={isDense} />}
    </div>
  );
}

function LoadingState() {
  return (
    <div className="p-4 space-y-3">
      {[1, 2, 3, 4, 5].map((i) => (
        <Skeleton key={i} className="h-12 w-full" />
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="p-8 text-center text-muted-foreground">
      <p>No logs available</p>
    </div>
  );
}

export default DeploymentLogsTable;
