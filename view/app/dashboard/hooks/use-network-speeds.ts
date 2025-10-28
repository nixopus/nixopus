'use client';

import { useState, useEffect, useRef } from 'react';
import { SystemStatsType } from '@/redux/types/monitor';

const formatBytes = (bytes: number, perSecond = false): string => {
  const unit = perSecond ? '/s' : '';
  if (bytes < 1024) return `${bytes.toFixed(1)}B${unit}`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB${unit}`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)}MB${unit}`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)}GB${unit}`;
};

export function useNetworkSpeeds(systemStats: SystemStatsType | null) {
  const [networkSpeeds, setNetworkSpeeds] = useState({
    downloadSpeed: '0 B/s',
    uploadSpeed: '0 B/s'
  });

  const previousNetworkStatsRef = useRef<{
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

    if (previousNetworkStatsRef.current) {
      const timeDiff = Math.max(0.001, (now - previousNetworkStatsRef.current.timestamp) / 1000);
      const bytesDownloaded = Math.max(0, bytesRecv - previousNetworkStatsRef.current.bytesRecv);
      const bytesUploaded = Math.max(0, bytesSent - previousNetworkStatsRef.current.bytesSent);

      setNetworkSpeeds({
        downloadSpeed: formatBytes(bytesDownloaded / timeDiff, true),
        uploadSpeed: formatBytes(bytesUploaded / timeDiff, true)
      });
    } else {
      setNetworkSpeeds({
        downloadSpeed: '0 B/s',
        uploadSpeed: '0 B/s'
      });
    }

    previousNetworkStatsRef.current = {
      bytesRecv,
      bytesSent,
      timestamp: now
    };
  }, [systemStats]);

  return { networkSpeeds };
}
