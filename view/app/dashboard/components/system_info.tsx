'use client';

import React from 'react';
import { Server } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '../hooks/use_monitor';
import { Skeleton } from '@/components/ui/skeleton';

interface SystemInfoCardProps {
  systemStats: SystemStatsType;
}

const SystemInfoCard: React.FC<SystemInfoCardProps> = ({ systemStats }) => {
  const { load } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Server className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          System Information
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1 sm:space-y-2">
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">OS Type</span>
            <span className="text-xs sm:text-sm font-medium truncate max-w-[60%] text-right">
              {systemStats.os_type}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">CPU</span>
            <span className="text-xs sm:text-sm font-medium truncate max-w-[60%] text-right">
              {systemStats.cpu_info}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">Uptime</span>
            <span className="text-xs sm:text-sm font-medium truncate max-w-[60%] text-right">
              {load.uptime}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">Last Updated</span>
            <span className="text-xs sm:text-sm font-medium truncate max-w-[60%] text-right">
              {new Date(systemStats.timestamp).toLocaleTimeString()}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default SystemInfoCard;

export const SystemInfoCardSkeleton: React.FC = () => {
  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Server className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          System Information
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1 sm:space-y-2">
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">OS Type</span>
            <Skeleton className="h-4 w-24" />
          </div>
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">CPU</span>
            <Skeleton className="h-4 w-32" />
          </div>
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">Uptime</span>
            <Skeleton className="h-4 w-20" />
          </div>
          <div className="flex justify-between">
            <span className="text-xs sm:text-sm text-muted-foreground">Last Updated</span>
            <Skeleton className="h-4 w-24" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
