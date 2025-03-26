'use client';
import { useWebSocket } from '@/hooks/socket_provider';
import { useEffect, useState } from 'react';

export type ContainerData = {
  Id: string;
  Names: string[];
  Image: string;
  ImageID: string;
  Command: string;
  Created: number;
  Ports: Port[];
  SizeRw?: number;
  SizeRootFs?: number;
  Labels: { [key: string]: string };
  State: string;
  Status: string;
  HostConfig: {
    NetworkMode?: string;
    Annotations?: { [key: string]: string };
  };
};

interface Port {
  IP?: string;
  PrivatePort: number;
  PublicPort?: number;
  Type: string;
}

export interface MemoryStats {
  used: number;
  total: number;
  percentage: number;
  rawInfo: string;
}

export interface LoadStats {
  oneMin: number;
  fiveMin: number;
  fifteenMin: number;
  uptime: string;
}

export interface DiskMount {
  filesystem: string;
  size: string;
  used: string;
  avail: string;
  capacity: string;
  mountPoint: string;
}

export interface DiskStats {
  total: number;
  used: number;
  available: number;
  percentage: number;
  mountPoint: string;
  allMounts: DiskMount[];
}

export interface SystemStatsType {
  os_type: string;
  cpu_info: string;
  memory: MemoryStats;
  load: LoadStats;
  disk: DiskStats;
  timestamp: number;
}

function use_monitor() {
  const { sendJsonMessage, message, isReady } = useWebSocket();
  const [containersData, setContainersData] = useState<ContainerData[]>([]);
  const [systemStats, setSystemStats] = useState<SystemStatsType | null>(null);

  useEffect(() => {
    if (message) {
      try {
        const parsedMessage =
          typeof message === 'string' && message.startsWith('{') ? JSON.parse(message) : message;

        if (parsedMessage.action === 'get_containers' && parsedMessage.data) {
          setContainersData(parsedMessage.data);
        } else if (parsedMessage.action === 'get_system_stats' && parsedMessage.data) {
          setSystemStats(parsedMessage.data);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    }
  }, [message]);

  const startMonitoring = () => {
    sendJsonMessage({
      action: 'dashboard_monitor',
      data: {
        interval: 10,
        operations: ['get_containers', 'get_system_stats']
      }
    });
  };

  const stopMonitoring = () => {
    sendJsonMessage({
      action: 'stop_dashboard_monitor'
    });
  };

  useEffect(() => {
    startMonitoring();
    return () => {
      stopMonitoring();
    };
  }, [isReady]);

  return {
    containersData,
    systemStats
  };
}

export default use_monitor;
