'use client';

import React from 'react';
import { HardDrive } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '../hooks/use_monitor';

interface DiskUsageCardProps {
  systemStats: SystemStatsType;
}

const DiskUsageCard: React.FC<DiskUsageCardProps> = ({ systemStats }) => {
  if (!systemStats) return null;

  const { disk } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <HardDrive className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          Disk Usage
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 sm:space-y-3">
          <div className="w-full h-2 bg-gray-200 rounded-full">
            <div
              className={`h-2 rounded-full ${disk.percentage > 80 ? 'bg-red-500' : 'bg-green-500'}`}
              style={{ width: `${disk.percentage}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>{disk.used} GB used</span>
            <span>{disk.percentage}%</span>
            <span>{disk.total} GB total</span>
          </div>
          <div className="text-xs font-mono text-muted-foreground mt-1 sm:mt-2">
            <table className="min-w-full">
              <thead>
                <tr>
                  <th className="text-left pr-1 sm:pr-2">Mount</th>
                  <th className="text-right pr-1 sm:pr-2">Size</th>
                  <th className="text-right pr-1 sm:pr-2">Used</th>
                  <th className="text-right">%</th>
                </tr>
              </thead>
              <tbody className="text-xxs sm:text-xs">
                {disk.allMounts.map((mount, index) => (
                  <tr key={index}>
                    <td className="text-left pr-1 sm:pr-2 truncate max-w-[60px] sm:max-w-none">
                      {mount.mountPoint}
                    </td>
                    <td className="text-right pr-1 sm:pr-2">{mount.size}</td>
                    <td className="text-right pr-1 sm:pr-2">{mount.used}</td>
                    <td className="text-right">{mount.capacity}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default DiskUsageCard;
