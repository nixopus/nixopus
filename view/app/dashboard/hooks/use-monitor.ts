'use client';
import { useWebSocket } from '@/hooks/socket-provider';
import { ContainerData, SystemStatsType } from '@/redux/types/monitor';
import { useEffect, useState, useRef } from 'react';

function use_monitor() {
  const { sendJsonMessage, message, isReady } = useWebSocket();
  const [containersData, setContainersData] = useState<ContainerData[]>([]);
  const [systemStats, setSystemStats] = useState<SystemStatsType | null>(null);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [lastError, setLastError] = useState<string | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const isInitializedRef = useRef(false);

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
          if (isMonitoring) {
            stopMonitoring();
            setTimeout(startMonitoring, 5000);
          }
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
        setLastError('Failed to parse message');
      }
    }
  }, [message]);

  const startMonitoring = () => {
    if (!isMonitoring) {
      sendJsonMessage({
        action: 'dashboard_monitor',
        data: {
          interval: 10,
          operations: ['get_containers', 'get_system_stats']
        }
      });
      setIsMonitoring(true);
      setLastError(null);
    }
  };

  const stopMonitoring = () => {
    if (isMonitoring) {
      sendJsonMessage({
        action: 'stop_dashboard_monitor'
      });
      setIsMonitoring(false);
    }
  };

  useEffect(() => {
    if (isReady && !isInitializedRef.current) {
      startMonitoring();
      isInitializedRef.current = true;
    }
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [isReady]);

  useEffect(() => {
    if (isReady && !isMonitoring && lastError) {
      reconnectTimeoutRef.current = setTimeout(() => {
        startMonitoring();
      }, 5000);
    }
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [isReady, isMonitoring, lastError]);

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
