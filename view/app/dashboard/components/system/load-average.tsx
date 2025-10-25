'use client';

import React from 'react';
import { Activity } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { BarChartComponent, BarChartDataItem } from '@/components/ui/bar-chart-component';

interface LoadAverageCardProps {
  systemStats: SystemStatsType | null;
}

const LoadAverageCard: React.FC<LoadAverageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  
  if (!systemStats) {
    return <LoadAverageCardSkeleton />;
  }
  
  const { load } = systemStats;

  // Prepare data for bar chart with distinct vibrant colors
  const chartData: BarChartDataItem[] = [
    {
      name: '1 min',
      value: load.oneMin,
      fill: '#3b82f6' // Bright blue
    },
    {
      name: '5 min',
      value: load.fiveMin,
      fill: '#10b981' // Bright green
    },
    {
      name: '15 min',
      value: load.fifteenMin,
      fill: '#f59e0b' // Bright amber/orange
    }
  ];

  const chartConfig = {
    oneMin: {
      label: '1 min',
      color: '#3b82f6'
    },
    fiveMin: {
      label: '5 min',
      color: '#10b981'
    },
    fifteenMin: {
      label: '15 min',
      color: '#f59e0b'
    }
  };

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Activity className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.load.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <BarChartComponent
          data={chartData}
          chartConfig={chartConfig}
          height="h-[180px]"
          yAxisLabel="Load"
          xAxisLabel="Time Period"
          showAxisLabels={true}
        />
        
        {/* Summary Stats with Color Indicators */}
        <div className="mt-4 grid grid-cols-3 gap-2 text-center">
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: '#3b82f6' }} />
              <TypographyMuted className="text-xs">1 min</TypographyMuted>
            </div>
            <TypographySmall className="text-sm font-bold">{load.oneMin.toFixed(2)}</TypographySmall>
          </div>
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: '#10b981' }} />
              <TypographyMuted className="text-xs">5 min</TypographyMuted>
            </div>
            <TypographySmall className="text-sm font-bold">{load.fiveMin.toFixed(2)}</TypographySmall>
          </div>
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: '#f59e0b' }} />
              <TypographyMuted className="text-xs">15 min</TypographyMuted>
            </div>
            <TypographySmall className="text-sm font-bold">{load.fifteenMin.toFixed(2)}</TypographySmall>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default LoadAverageCard;

export function LoadAverageCardSkeleton() {
  const { t } = useTranslation();

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Activity className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.load.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Skeleton className="h-[180px] w-full rounded-lg" />
        <div className="mt-4 grid grid-cols-3 gap-2 text-center">
          <div>
            <TypographyMuted className="text-xs">1 min</TypographyMuted>
            <Skeleton className="h-4 w-12 mx-auto mt-1" />
          </div>
          <div>
            <TypographyMuted className="text-xs">5 min</TypographyMuted>
            <Skeleton className="h-4 w-12 mx-auto mt-1" />
          </div>
          <div>
            <TypographyMuted className="text-xs">15 min</TypographyMuted>
            <Skeleton className="h-4 w-12 mx-auto mt-1" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
