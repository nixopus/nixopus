'use client';

import React from 'react';
import { Activity } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '../hooks/use_monitor';

interface LoadAverageCardProps {
  systemStats: SystemStatsType;
}

const LoadAverageCard: React.FC<LoadAverageCardProps> = ({ systemStats }) => {
  if (!systemStats) return null;

  const { load } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Activity className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          Load Average
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1 sm:space-y-2">
          <LoadBar label="1 minute" value={load.oneMin} />
          <LoadBar label="5 minutes" value={load.fiveMin} />
          <LoadBar label="15 minutes" value={load.fifteenMin} />
        </div>
      </CardContent>
    </Card>
  );
};

interface LoadBarProps {
  label: string;
  value: number;
}

const LoadBar: React.FC<LoadBarProps> = ({ label, value }) => (
  <div className="flex justify-between items-center">
    <span className="text-xs sm:text-sm text-muted-foreground">{label}</span>
    <div className="flex items-center">
      <div className="w-20 sm:w-32 h-2 bg-gray-200 rounded-full mr-1 sm:mr-2">
        <div
          className="h-2 bg-blue-500 rounded-full"
          style={{ width: `${Math.min(value * 25, 100)}%` }}
        />
      </div>
      <span className="text-xs sm:text-sm font-medium">{value.toFixed(2)}</span>
    </div>
  </div>
);

export default LoadAverageCard;
