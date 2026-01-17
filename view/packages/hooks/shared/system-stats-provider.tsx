'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useWebSocket } from '@/packages/hooks/shared/socket-provider';
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
  const { message } = useWebSocket();
  const [systemStats, setSystemStats] = useState<SystemStatsType | null>(null);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
          setIsMonitoring(true);
        } else if (parsedMessage.action === 'dashboard_monitor_started') {
          setIsMonitoring(true);
          setError(null);
        } else if (parsedMessage.action === 'dashboard_monitor_stopped') {
          setIsMonitoring(false);
        } else if (parsedMessage.action === 'error') {
          setError(parsedMessage.error || 'Unknown error occurred');
          setIsMonitoring(false);
        }
      } catch (err) {
        console.error('Error parsing WebSocket message:', err);
        setError('Failed to parse message');
      }
    }
  }, [message]);

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
