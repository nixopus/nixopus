'use client';

import React from 'react';
import { Search, ChevronsUpDown, RefreshCw, X, Calendar, Rows3, Rows4 } from 'lucide-react';
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
    setIsDense
  } = useDeploymentLogsViewer({ id, isDeployment });

  return (
    <Card className="border-0 shadow-none">
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
      />
      <CardContent className="p-0 border rounded-md overflow-hidden">
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
  onDenseChange
}: LogsHeaderProps) {
  const hasActiveFilters =
    filters.startDate || filters.endDate || filters.level !== 'all' || searchTerm;

  return (
    <CardHeader className="space-y-3 pb-4 px-0 border-none border-b-0">
      {title && (
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">{title}</h3>
          <LoadMoreButton
            currentPage={currentPage}
            totalPages={totalPages}
            onLoadMore={onLoadMore}
          />
        </div>
      )}
      <div className="flex items-center gap-4">
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
      <div className="flex items-center justify-between">
        <LevelFilter
          value={filters.level}
          onChange={(v) => onFiltersChange({ ...filters, level: v })}
        />
        <div className="flex items-center gap-2">
          <ExpandCollapseButton onExpandAll={onExpandAll} onCollapseAll={onCollapseAll} />
          <DenseToggle isDense={isDense} onChange={onDenseChange} />
        </div>
      </div>
    </CardHeader>
  );
}

function SearchInput({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <div className="relative flex-shrink-0 w-96">
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
    <div className="flex flex-wrap items-center gap-3 ml-auto">
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
    <div className="flex items-center gap-2">
      {levelOptions.map((option) => (
        <Badge
          key={option.value}
          variant={value === option.value ? 'default' : 'outline'}
          className="cursor-pointer transition-colors"
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
    <div className="flex items-center gap-3 px-4 py-2 border-b bg-muted/30 text-xs font-medium text-muted-foreground uppercase tracking-wider">
      <div className="w-4" />
      <div className="w-14">Level</div>
      <div className="w-44">Timestamp</div>
      <div className="flex-1">Message</div>
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
    <div className="max-h-[600px] overflow-y-auto">
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
