'use client';

import React, { createContext, useContext, useState, useRef, useEffect, ReactNode } from 'react';
import { useWebSocket } from '@/hooks/socket-provider';
import { SystemStatsType } from '@/redux/types/monitor';

interface SystemStatsContextType {
  systemStats: SystemStatsType | null;
  isMonitoring: boolean;
  error: string | null;
}

const SystemStatsContext = createContext<SystemStatsContextType>({
  systemStats: null,
  isMonitoring: false,
  error: null
});

export function SystemStatsProvider({ children }: { children: ReactNode }) {
  const { sendJsonMessage, message, isReady } = useWebSocket();
  const [systemStats, setSystemStats] = useState<SystemStatsType | null>(null);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isInitializedRef = useRef(false);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

  useEffect(() => {
    if (message) {
      try {
        const parsedMessage =
          typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

        if (parsedMessage.topic !== 'dashboard_monitor') {
          return;
        }

        if (parsedMessage.action === 'get_system_stats' && parsedMessage.data) {
          setSystemStats(parsedMessage.data);
          setError(null);
        } else if (parsedMessage.action === 'error') {
          setError(parsedMessage.error || 'Unknown error occurred');
          if (isMonitoring) {
            stopMonitoring();
            setTimeout(startMonitoring, 5000);
          }
        }
      } catch (err) {
        console.error('Error parsing WebSocket message:', err);
        setError('Failed to parse message');
      }
    }
  }, [message]);

  const startMonitoring = () => {
    if (!isMonitoring) {
      sendJsonMessage({
        action: 'dashboard_monitor',
        data: {
          interval: 10,
          operations: ['get_system_stats']
        }
      });
      setIsMonitoring(true);
      setError(null);
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
    if (isReady && !isMonitoring && error) {
      reconnectTimeoutRef.current = setTimeout(() => {
        startMonitoring();
      }, 5000);
    }
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [isReady, isMonitoring, error]);

  return (
    <SystemStatsContext.Provider value={{ systemStats, isMonitoring, error }}>
      {children}
    </SystemStatsContext.Provider>
  );
}

export function useSystemStats() {
  const context = useContext(SystemStatsContext);
  if (!context) {
    throw new Error('useSystemStats must be used within SystemStatsProvider');
  }
  return context;
}

