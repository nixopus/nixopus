'use client';
import { useWebSocket } from '@/hooks/socket-provider';
import { ContainerData, SystemStatsType } from '@/redux/types/monitor';
import { useEffect, useState, useRef, useCallback } from 'react';

function use_monitor() {
  const { sendJsonMessage, message, isReady } = useWebSocket();
  const [containersData, setContainersData] = useState<ContainerData[]>([]);
  const [systemStats, setSystemStats] = useState<SystemStatsType | null>(null);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [lastError, setLastError] = useState<string | null>(null);
  const [isLoadingInitialData, setIsLoadingInitialData] = useState(true);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const isInitializedRef = useRef(false);
  const hasReceivedContainersRef = useRef(false);
  const hasReceivedSystemStatsRef = useRef(false);

  const startMonitoring = useCallback(() => {
    if (!isReady) return;

    sendJsonMessage({
      action: 'dashboard_monitor',
      data: {
        interval: 10,
        operations: ['get_containers', 'get_system_stats']
      }
    });
    setIsMonitoring(true);
    setLastError(null);
  }, [isReady, sendJsonMessage]);

  const stopMonitoring = useCallback(() => {
    sendJsonMessage({
      action: 'stop_dashboard_monitor'
    });
    setIsMonitoring(false);
  }, [sendJsonMessage]);

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (message) {
      try {
        const parsedMessage =
          typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

        if (parsedMessage.topic != 'dashboard_monitor') {
          return;
        }

        if (parsedMessage.action === 'get_containers' && parsedMessage.data !== undefined) {
          const containers = Array.isArray(parsedMessage.data) ? parsedMessage.data : [];
          setContainersData(containers);
          setLastError(null);
          hasReceivedContainersRef.current = true;
          setIsLoadingInitialData(false);
        } else if (parsedMessage.action === 'get_system_stats' && parsedMessage.data) {
          setSystemStats(parsedMessage.data);
          setLastError(null);
          hasReceivedSystemStatsRef.current = true;
          setIsLoadingInitialData(false);
        } else if (parsedMessage.action === 'error') {
          setLastError(parsedMessage.error || 'Unknown error occurred');
          if (!hasReceivedContainersRef.current && !hasReceivedSystemStatsRef.current) {
            setIsLoadingInitialData(false);
          }

          setTimeout(() => {
            if (isReady) {
              startMonitoring();
            }
          }, 5000);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
        setLastError('Failed to parse message');
      }
    }
  }, [message, isReady, startMonitoring]);

  // Initialize monitoring when WebSocket is ready
  useEffect(() => {
    if (isReady && !isInitializedRef.current) {
      startMonitoring();
      isInitializedRef.current = true;

      const loadingTimeoutId = setTimeout(() => {
        setIsLoadingInitialData(false);
      }, 3000);

      const containerCheckTimeoutId = setTimeout(() => {
        if (!hasReceivedContainersRef.current) {
          startMonitoring();
        }
      }, 2000);

      return () => {
        clearTimeout(loadingTimeoutId);
        clearTimeout(containerCheckTimeoutId);
      };
    }

    if (!isReady && isInitializedRef.current) {
      isInitializedRef.current = false;
      setIsMonitoring(false);
      hasReceivedContainersRef.current = false;
      hasReceivedSystemStatsRef.current = false;
      setIsLoadingInitialData(true);
    }

    // Cleanup: stop monitoring when component unmounts
    return () => {
      if (isInitializedRef.current) {
        stopMonitoring();
        isInitializedRef.current = false;
        hasReceivedContainersRef.current = false;
        hasReceivedSystemStatsRef.current = false;
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [isReady, startMonitoring, stopMonitoring]);

  // Retry monitoring on error
  useEffect(() => {
    if (isReady && !isMonitoring && lastError) {
      console.log('Retrying monitoring after error');
      reconnectTimeoutRef.current = setTimeout(() => {
        startMonitoring();
      }, 5000);
    }

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [isReady, isMonitoring, lastError, startMonitoring]);

  return {
    containersData,
    systemStats,
    isMonitoring,
    lastError,
    isLoadingInitialData,
    startMonitoring,
    stopMonitoring
  };
}

export default use_monitor;
