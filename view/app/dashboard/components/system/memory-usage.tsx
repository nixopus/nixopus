'use client';

import React from 'react';
import { BarChart } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

interface MemoryUsageCardProps {
  systemStats: SystemStatsType;
}

const formatGB = (value: number) => `${value.toFixed(2)}`;

const MemoryUsageCard: React.FC<MemoryUsageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  const { memory } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <BarChart className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.memory.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 sm:space-y-3">
          <div className="w-full h-2 bg-gray-200 rounded-full">
            <div
              className={`h-2 rounded-full bg-primary`}
              style={{ width: `${memory.percentage}%` }}
            />
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
              {t('dashboard.memory.used').replace('{value}', formatGB(memory.used))}
            </TypographyMuted>
            <TypographyMuted className="text-xs truncate max-w-[60px] sm:max-w-[80px]">
              {t('dashboard.memory.percentage').replace('{value}', memory.percentage.toFixed(1))}
            </TypographyMuted>
            <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
              {t('dashboard.memory.total').replace('{value}', formatGB(memory.total))}
            </TypographyMuted>
          </div>
          <TypographyMuted className="text-xs mt-1 sm:mt-2 line-clamp-2 sm:line-clamp-none break-all">
            {memory.rawInfo}
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
        <div className="space-y-2 sm:space-y-3">
          <Skeleton className="w-full h-2 rounded-full" />

          <div className="flex justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-10" />
            <Skeleton className="h-4 w-20" />
          </div>

          <Skeleton className="h-4 w-full mt-1 sm:mt-2" />
          <Skeleton className="h-4 w-2/3" />
        </div>
      </CardContent>
    </Card>
  );
}
