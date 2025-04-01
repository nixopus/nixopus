import { useGetApplicationLogsQuery, useGetDeploymentLogsQuery } from '@/redux/services/deploy/applicationsApi';
import { useEffect, useMemo, useRef, useState } from 'react';
import { ApplicationLogs, ApplicationLogsResponse } from '@/redux/types/applications';
import { useApplicationWebSocket } from './use_application_websocket';
import { SOCKET_EVENTS } from '@/redux/api-conf';

export interface LogViewerProps {
  id: string;
  onRefresh?: () => void;
  currentPage?: number;
  setCurrentPage: (page: number) => void;
  isDeployment?: boolean;
}

function useLogViewer({
  id,
  onRefresh,
  currentPage = 1,
  setCurrentPage,
  isDeployment = false
}: LogViewerProps) {
  const [autoScroll, setAutoScroll] = useState(true);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [markers, setMarkers] = useState<any[]>([]);
  const [currentSearchIndex, setCurrentSearchIndex] = useState(0);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const refreshTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const editorRef = useRef<any>(null);
  const lastLogCountRef = useRef<number>(0);
  const [allLogs, setAllLogs] = useState<ApplicationLogs[]>([]);
  const { message } = useApplicationWebSocket(id);

  useEffect(() => {
    if (!message) return;

    try {
      const parsedMessage = JSON.parse(message);
      
      if (!parsedMessage?.topic || !parsedMessage?.data) return;
      
      if (parsedMessage.topic.includes(SOCKET_EVENTS.MONITOR_APPLICATION_DEPLOYMENT)) {
        const { action, table, data } = parsedMessage.data;
        
        if (!action || !table || !data) return;
        
        if ((action === "INSERT" || action === "UPDATE") && table === "application_logs") {
          setAllLogs(prevLogs => [...prevLogs, data]);
        }
      }
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
    }
  }, [message]);

  const { data: applicationLogsResponse } = useGetApplicationLogsQuery({
    id: id,
    page: currentPage,
    page_size: 100,
    search_term: searchTerm
  }, { skip: isDeployment || !id });

  const { data: deploymentLogsResponse } = useGetDeploymentLogsQuery({
    id: id,
    page: currentPage,
    page_size: 100,
    search_term: searchTerm
  }, { skip: !isDeployment || !id });

  const logsResponse = isDeployment ? deploymentLogsResponse : applicationLogsResponse;

  useEffect(() => {
    if (logsResponse?.logs) {
      if (currentPage === 1) {
        setAllLogs(logsResponse.logs);
      } else {
        setAllLogs(prevLogs => {
          const newLogs = logsResponse.logs.filter(newLog =>
            !prevLogs.some(prevLog => prevLog.id === newLog.id)
          );
          return [...newLogs, ...prevLogs];
        });
      }
    }
  }, [logsResponse, currentPage]);

  const filteredLogs = useMemo(() => {
    if (allLogs.length === 0) return '';
    const sortedLogs = [...allLogs].sort((a, b) =>
      new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    );

    return sortedLogs.map((log: ApplicationLogs) => {
      const date = new Date(log.created_at);
      const timestamp = date.toLocaleString();
      return `[${timestamp}] ${log.log}`;
    }).join('\n');
  }, [allLogs]);

  useEffect(() => {
    if (editorRef.current && allLogs.length > 0 && currentPage > 1) {
      editorRef.current.editor.scrollToRow(0);
    }
  }, [allLogs, currentPage]);

  useEffect(() => {
    if (editorRef.current && allLogs.length > 0 && currentPage === 1) {
      const lastRow = editorRef.current.editor.session.getLength() - 1;
      editorRef.current.editor.scrollToRow(lastRow);
    }
  }, [allLogs, currentPage]);

  useEffect(() => {
    if (searchTerm && editorRef.current) {
      const lines = filteredLogs.split('\n');
      const searchResults: any[] = [];

      lines.forEach((line: string, index: number) => {
        const regex = new RegExp(searchTerm, 'gi');
        let match;
        while ((match = regex.exec(line)) !== null) {
          searchResults.push({
            startRow: index,
            startCol: match.index,
            endRow: index,
            endCol: match.index + searchTerm.length,
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
      if (autoScroll && editorRef.current) {
        const lastRow = editorRef.current.editor.session.getLength() - 1;
        editorRef.current.editor.scrollToRow(lastRow);
      }
    }
  }, [searchTerm, filteredLogs, currentSearchIndex, autoScroll]);

  const handleEditorLoad = (editor: any) => {
    editorRef.current = { editor };
    const lastRow = editor.session.getLength() - 1;
    editor.scrollToRow(lastRow);
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

  const handleRefresh = async () => {
    if (isRefreshing) return;

    setIsRefreshing(true);
    try {
      setCurrentPage(1);
      if (onRefresh) {
        await onRefresh();
      }
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
    autoScroll,
    setAutoScroll
  };
}

export default useLogViewer;

export interface LogLevel {
  label: string;
  value: string;
  color: string;
}