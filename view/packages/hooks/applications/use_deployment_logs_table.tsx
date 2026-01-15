import { useState, useCallback, useMemo } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { LogLevel, FormattedLogEntry, LogFilters } from './use_deployment_logs_viewer';
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

interface UseDeploymentLogsTableProps {
  id: string;
  isDeployment?: boolean;
  logs: FormattedLogEntry[];
  refreshLogs: () => Promise<void>;
  searchTerm: string;
  filters: LogFilters;
  currentPage: number;
  totalPages: number;
  isDense: boolean;
  expandAll: () => void;
  collapseAll: () => void;
  setCurrentPage: (page: number) => void;
  setFilters: (filters: LogFilters) => void;
  clearFilters: () => void;
  setIsDense: (isDense: boolean) => void;
}

const levelOptions: { value: LogLevel | 'all'; label: string }[] = [
  { value: 'all', label: 'All Levels' },
  { value: 'error', label: 'Error' },
  { value: 'warn', label: 'Warning' },
  { value: 'info', label: 'Info' },
  { value: 'debug', label: 'Debug' }
];

const dateFilterOptions = [
  { key: 'startDate', label: 'From' },
  { key: 'endDate', label: 'To' }
] as const;

const tableHeaderColumns = [
  { key: 'expand', width: 'w-4', label: '' },
  { key: 'level', width: 'w-14', label: 'Level' },
  { key: 'timestamp', width: 'w-44', label: 'Timestamp' },
  { key: 'message', width: 'flex-1', label: 'Message' }
];

const loadingSkeletonCount = 5;

export function useDeploymentLogsTable({
  id,
  isDeployment = false,
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
}: UseDeploymentLogsTableProps) {
  const { t } = useTranslation();
  const [isCopied, setIsCopied] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);

  const logText = useMemo(
    () =>
      logs
        .map((log) => `[${log.formattedTime}] [${log.level.toUpperCase()}] ${log.message}`)
        .join('\n'),
    [logs]
  );

  const hasActiveFilters = useMemo(
    () => filters.startDate || filters.endDate || filters.level !== 'all' || searchTerm,
    [filters, searchTerm]
  );

  const showLoadMore = useMemo(() => currentPage < totalPages, [currentPage, totalPages]);

  const handleCopyLogs = useCallback(async () => {
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
  }, [logText, t]);

  const handleDownloadLogs = useCallback(() => {
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
  }, [logText, id, isDeployment, t]);

  const handleRefreshLogs = useCallback(async () => {
    setIsRefreshing(true);
    try {
      await refreshLogs();
    } finally {
      setIsRefreshing(false);
    }
  }, [refreshLogs]);

  const handleExpandCollapse = useCallback(() => {
    if (isExpanded) {
      collapseAll();
    } else {
      expandAll();
    }
    setIsExpanded(!isExpanded);
  }, [isExpanded, expandAll, collapseAll]);

  const handleLoadMore = useCallback(() => {
    setCurrentPage(currentPage + 1);
  }, [currentPage, setCurrentPage]);

  const handleLevelChange = useCallback(
    (level: LogLevel | 'all') => {
      setFilters({ ...filters, level });
    },
    [filters, setFilters]
  );

  const handleDateFilterChange = useCallback(
    (key: 'startDate' | 'endDate', value: string) => {
      setFilters({ ...filters, [key]: value });
    },
    [filters, setFilters]
  );

  const handleDenseToggle = useCallback(() => {
    setIsDense(!isDense);
  }, [isDense, setIsDense]);

  const loadingSkeletons = useMemo(
    () => Array.from({ length: loadingSkeletonCount }, (_, i) => i + 1),
    []
  );

  const actionButtons = useMemo(
    () => [
      {
        key: 'refresh',
        icon: isRefreshing ? Loader2 : RefreshCw,
        onClick: handleRefreshLogs,
        disabled: isRefreshing,
        title: 'Refresh logs',
        className: isRefreshing ? 'animate-spin' : ''
      },
      {
        key: 'copy',
        icon: isCopied ? Check : Copy,
        onClick: handleCopyLogs,
        disabled: !logs.length,
        title: 'Copy logs',
        className: isCopied ? 'text-green-500' : ''
      },
      {
        key: 'download',
        icon: Download,
        onClick: handleDownloadLogs,
        disabled: !logs.length,
        title: 'Download logs'
      },
      {
        key: 'expandCollapse',
        icon: ChevronsUpDown,
        onClick: handleExpandCollapse,
        title: 'Expand/Collapse all'
      },
      {
        key: 'dense',
        icon: isDense ? Rows3 : Rows4,
        onClick: handleDenseToggle,
        title: isDense ? 'Normal view' : 'Dense view'
      }
    ],
    [
      isRefreshing,
      isCopied,
      logs.length,
      isDense,
      handleRefreshLogs,
      handleCopyLogs,
      handleDownloadLogs,
      handleExpandCollapse,
      handleDenseToggle
    ]
  );

  return {
    isCopied,
    isRefreshing,
    handleCopyLogs,
    handleDownloadLogs,
    handleRefreshLogs,
    handleLoadMore,
    handleLevelChange,
    handleDateFilterChange,
    handleDenseToggle,
    hasLogs: logs.length > 0,
    hasActiveFilters,
    showLoadMore,
    levelOptions,
    dateFilterOptions,
    tableHeaderColumns,
    loadingSkeletons,
    actionButtons
  };
}
