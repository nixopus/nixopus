import { useEffect, useMemo, useRef, useState } from 'react';

function useLogViewer({
  logs,
  title,
  description,
  onRefresh,
  currentPage,
  setCurrentPage
}: LogViewerProps) {
  const [autoScroll, setAutoScroll] = useState(true);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLevel, setSelectedLevel] = useState<string>('all');
  const [timeRange, setTimeRange] = useState<string>('all');
  const [markers, setMarkers] = useState<any[]>([]);
  const [currentSearchIndex, setCurrentSearchIndex] = useState(0);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const refreshTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const [selectedContainer, setSelectedContainer] = useState<string>('all');
  const editorRef = useRef<any>(null);

  const filteredLogs = useMemo(() => {
    let result = logs?.split('\n');

    if (selectedLevel !== 'all') {
      result = result.filter((line) => line.toLowerCase().includes(selectedLevel.toLowerCase()));
    }

    if (selectedContainer !== 'all') {
      result = result.filter((line) => line.includes(selectedContainer));
    }

    if (timeRange !== 'all') {
      const now = new Date();
      let cutoffTime: Date;

      switch (timeRange) {
        case '5m':
          cutoffTime = new Date(now.getTime() - 5 * 60 * 1000);
          break;
        case '15m':
          cutoffTime = new Date(now.getTime() - 15 * 60 * 1000);
          break;
        case '1h':
          cutoffTime = new Date(now.getTime() - 60 * 60 * 1000);
          break;
        case '24h':
          cutoffTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
          break;
        default:
          cutoffTime = new Date(0);
      }

      result = result.filter((line) => {
        try {
          const datePart = line.split(' ')[0] + ' ' + line.split(' ')[1];
          const logDate = new Date(datePart);
          return !isNaN(logDate.getTime()) && logDate >= cutoffTime;
        } catch {
          return true;
        }
      });
    }

    return result.join('\n');
  }, [logs, selectedLevel, timeRange, selectedContainer]);

  const handleRefresh = async () => {
    if (isRefreshing || !onRefresh) return;

    setIsRefreshing(true);
    try {
      await onRefresh();
    } catch (error) {
      console.error('Failed to refresh logs:', error);
    } finally {
      refreshTimeoutRef.current = setTimeout(() => {
        setIsRefreshing(false);
      }, 1000);
    }
  };

  useEffect(() => {
    return () => {
      if (refreshTimeoutRef.current) {
        clearTimeout(refreshTimeoutRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (searchTerm && editorRef.current) {
      const lines = filteredLogs.split('\n');
      const searchResults: any[] = [];

      lines.forEach((line, index) => {
        if (line.toLowerCase().includes(searchTerm.toLowerCase())) {
          searchResults.push({
            startRow: index,
            startCol: line.toLowerCase().indexOf(searchTerm.toLowerCase()),
            endRow: index,
            endCol: line.toLowerCase().indexOf(searchTerm.toLowerCase()) + searchTerm.length,
            className: 'search-result',
            type: 'text'
          });
        }
      });

      setMarkers(searchResults);
      if (searchResults.length > 0) {
        editorRef.current.editor.scrollToRow(searchResults[currentSearchIndex]?.startRow);
      }
    } else {
      setMarkers([]);
    }
  }, [searchTerm, filteredLogs, currentSearchIndex]);

  useEffect(() => {
    if (autoScroll && editorRef.current && !searchTerm) {
      const lastRow = editorRef.current.editor.session.getLength() - 1;
      editorRef.current.editor.scrollToRow(lastRow);
    }
  }, [filteredLogs, autoScroll, searchTerm]);

  const handleEditorLoad = (editor: any) => {
    editorRef.current = { editor };
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  const handleDownload = () => {
    const blob = new Blob([filteredLogs], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs-${new Date().toISOString()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const navigateSearch = (direction: 'next' | 'prev') => {
    if (markers.length === 0) return;

    let newIndex;
    if (direction === 'next') {
      newIndex = (currentSearchIndex + 1) % markers.length;
    } else {
      newIndex = (currentSearchIndex - 1 + markers.length) % markers.length;
    }
    setCurrentSearchIndex(newIndex);
  };

  return {
    filteredLogs,
    handleRefresh,
    isRefreshing,
    handleEditorLoad,
    toggleFullscreen,
    isFullscreen,
    handleDownload,
    navigateSearch,
    currentSearchIndex,
    markers,
    searchTerm,
    setSearchTerm,
    setSelectedLevel,
    selectedLevel,
    setTimeRange,
    timeRange,
    autoScroll,
    setAutoScroll
  };
}

export default useLogViewer;

export interface LogViewerProps {
  logs: string;
  title?: string;
  description?: string;
  onRefresh?: () => void;
  currentPage?: number;
  setCurrentPage: (page: number) => void;
}

export interface LogLevel {
  label: string;
  value: string;
  color: string;
}

export const LOG_LEVELS: LogLevel[] = [
  { label: 'ERROR', value: 'error', color: 'text-red-500' },
  { label: 'WARN', value: 'warn', color: 'text-yellow-500' },
  { label: 'INFO', value: 'info', color: 'text-blue-500' },
  { label: 'DEBUG', value: 'debug', color: 'text-gray-500' }
];
