'use client';

import { useState, useEffect, useRef } from 'react';
import { SystemStatsType } from '@/redux/types/monitor';

interface NetworkData {
  downloadSpeed: string;
  uploadSpeed: string;
  totalDownload: string;
  totalUpload: string;
  interfaces: Array<{
    name: string;
    downloadSpeed: string;
    uploadSpeed: string;
  }>;
}

interface UseNetworkProps {
  systemStats: SystemStatsType | null;
}

const formatBytes = (bytes: number, perSecond = false): string => {
  const unit = perSecond ? '/s' : '';
  if (bytes < 1024) return `${bytes.toFixed(2)} B${unit}`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB${unit}`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB${unit}`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB${unit}`;
};

export function useNetwork({ systemStats }: UseNetworkProps) {
  const [networkData, setNetworkData] = useState<NetworkData>({
    downloadSpeed: '0 B/s',
    uploadSpeed: '0 B/s',
    totalDownload: '0 B',
    totalUpload: '0 B',
    interfaces: []
  });

  const previousStatsRef = useRef<{
    bytesRecv: number;
    bytesSent: number;
    timestamp: number;
  } | null>(null);

  useEffect(() => {
    if (!systemStats?.network) {
      return;
    }

    const network = systemStats.network;
    const now = Date.now();
    const bytesRecv = network.totalBytesRecv || 0;
    const bytesSent = network.totalBytesSent || 0;

    if (previousStatsRef.current) {
      const timeDiff = Math.max(0.001, (now - previousStatsRef.current.timestamp) / 1000);
      const bytesDownloaded = Math.max(0, bytesRecv - previousStatsRef.current.bytesRecv);
      const bytesUploaded = Math.max(0, bytesSent - previousStatsRef.current.bytesSent);

      const downloadSpeed = bytesDownloaded / timeDiff;
      const uploadSpeed = bytesUploaded / timeDiff;

      setNetworkData({
        downloadSpeed: formatBytes(Math.max(0, downloadSpeed), true),
        uploadSpeed: formatBytes(Math.max(0, uploadSpeed), true),
        totalDownload: formatBytes(bytesRecv),
        totalUpload: formatBytes(bytesSent),
        interfaces: []
      });
    } else {
      setNetworkData({
        downloadSpeed: '0 B/s',
        uploadSpeed: '0 B/s',
        totalDownload: formatBytes(bytesRecv),
        totalUpload: formatBytes(bytesSent),
        interfaces: []
      });
    }

    previousStatsRef.current = {
      bytesRecv,
      bytesSent,
      timestamp: now
    };
  }, [systemStats]);

  return { networkData };
}
