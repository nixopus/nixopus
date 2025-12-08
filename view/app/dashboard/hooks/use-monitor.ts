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
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const isInitializedRef = useRef(false);

  const startMonitoring = useCallback(() => {
    if (!isReady) {
      console.log('WebSocket not ready, skipping monitoring start');
      return;
    }

    console.log('Starting dashboard monitoring');
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
    console.log('Stopping dashboard monitoring');
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

        if (parsedMessage.action === 'get_containers' && parsedMessage.data) {
          setContainersData(parsedMessage.data);
          setLastError(null);
        } else if (parsedMessage.action === 'get_system_stats' && parsedMessage.data) {
          setSystemStats(parsedMessage.data);
          setLastError(null);
        } else if (parsedMessage.action === 'error') {
          setLastError(parsedMessage.error || 'Unknown error occurred');
          // Retry after error
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
      console.log('WebSocket ready, initializing monitoring');
      startMonitoring();
      isInitializedRef.current = true;
    }

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [isReady, startMonitoring]);

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
    startMonitoring,
    stopMonitoring
  };
}

export default use_monitor;
