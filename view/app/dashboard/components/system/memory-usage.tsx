'use client';

import React from 'react';
import { BarChart } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { DoughnutChartComponent, DoughnutChartDataItem } from '@/components/ui/doughnut-chart-component';

interface MemoryUsageCardProps {
  systemStats: SystemStatsType | null;
}

const formatGB = (value: number) => `${value.toFixed(2)}`;

const MemoryUsageCard: React.FC<MemoryUsageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  
  if (!systemStats) {
    return <MemoryUsageCardSkeleton />;
  }
  
  const { memory } = systemStats;

  // Calculate free memory
  const freeMemory = memory.total - memory.used;

  // Prepare data for doughnut chart with distinct colors
  const chartData: DoughnutChartDataItem[] = [
    {
      name: 'Used',
      value: memory.used,
      fill: '#3b82f6' // Bright blue for used memory
    },
    {
      name: 'Free',
      value: freeMemory,
      fill: '#10b981' // Bright green for free memory
    }
  ];

  const chartConfig = {
    used: {
      label: 'Used Memory',
      color: '#3b82f6'
    },
    free: {
      label: 'Free Memory',
      color: '#10b981'
    }
  };

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <BarChart className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.memory.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Doughnut Chart */}
          <DoughnutChartComponent
            data={chartData}
            chartConfig={chartConfig}
            centerLabel={{
              value: `${memory.percentage.toFixed(1)}%`,
              subLabel: 'Used'
            }}
            innerRadius={60}
            outerRadius={80}
            maxHeight="max-h-[200px]"
          />

          {/* Summary Stats with Distinct Colors */}
          <div className="flex justify-between text-xs">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: '#3b82f6' }} />
              <TypographyMuted>
                Used: {formatGB(memory.used)} GB
              </TypographyMuted>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: '#10b981' }} />
              <TypographyMuted>
                Free: {formatGB(freeMemory)} GB
              </TypographyMuted>
            </div>
          </div>

          {/* Additional Info */}
          <TypographyMuted className="text-xs text-center">
            Total: {formatGB(memory.total)} GB
          </TypographyMuted>
        </div>
      </CardContent>
    </Card>
  );
};

export default MemoryUsageCard;

export function MemoryUsageCardSkeleton() {
  const { t } = useTranslation();

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <BarChart className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.memory.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <Skeleton className="mx-auto aspect-square max-h-[200px] w-[200px] rounded-full" />

          <div className="flex justify-between text-xs">
            <div className="flex items-center gap-2">
              <Skeleton className="h-3 w-3 rounded-sm" />
              <Skeleton className="h-4 w-20" />
            </div>
            <div className="flex items-center gap-2">
              <Skeleton className="h-3 w-3 rounded-sm" />
              <Skeleton className="h-4 w-20" />
            </div>
          </div>

          <Skeleton className="h-4 w-32 mx-auto" />
        </div>
      </CardContent>
    </Card>
  );
}
