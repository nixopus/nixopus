'use client';

import React from 'react';
import { Activity } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

interface LoadAverageCardProps {
  systemStats: SystemStatsType;
}

const LoadAverageCard: React.FC<LoadAverageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  const { load } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Activity className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.load.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <TypographyMuted>{t('dashboard.load.labels.oneMin')}</TypographyMuted>
            <TypographySmall>{load.oneMin.toFixed(2)}</TypographySmall>
          </div>
          <div className="flex items-center justify-between">
            <TypographyMuted>{t('dashboard.load.labels.fiveMin')}</TypographyMuted>
            <TypographySmall>{load.fiveMin.toFixed(2)}</TypographySmall>
          </div>
          <div className="flex items-center justify-between">
            <TypographyMuted>{t('dashboard.load.labels.fifteenMin')}</TypographyMuted>
            <TypographySmall>{load.fifteenMin.toFixed(2)}</TypographySmall>
          </div>
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
    <TypographyMuted className="text-xs sm:text-sm">{label}</TypographyMuted>
    <div className="flex items-center">
      <div className="w-20 sm:w-32 h-2 bg-gray-200 rounded-full mr-1 sm:mr-2">
        <div
          className={`h-2 rounded-full ${value > 80 ? 'bg-destructive' : 'bg-primary'}`}
          style={{ width: `${Math.min(value * 25, 100)}%` }}
        />
      </div>
      <TypographySmall className="text-xs sm:text-sm">{value.toFixed(2)}</TypographySmall>
    </div>
  </div>
);

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
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <TypographyMuted>{t('dashboard.load.labels.oneMin')}</TypographyMuted>
            <Skeleton className="h-4 w-12" />
          </div>
          <div className="flex items-center justify-between">
            <TypographyMuted>{t('dashboard.load.labels.fiveMin')}</TypographyMuted>
            <Skeleton className="h-4 w-12" />
          </div>
          <div className="flex items-center justify-between">
            <TypographyMuted>{t('dashboard.load.labels.fifteenMin')}</TypographyMuted>
            <Skeleton className="h-4 w-12" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

interface LoadBarSkeletonProps {
  label: string;
}

const LoadBarSkeleton: React.FC<LoadBarSkeletonProps> = ({ label }) => (
  <div className="flex justify-between items-center">
    <TypographyMuted className="text-xs sm:text-sm">{label}</TypographyMuted>
    <div className="flex items-center">
      <Skeleton className="w-20 sm:w-32 h-2 rounded-full mr-1 sm:mr-2" />
      <Skeleton className="w-8 h-4" />
    </div>
  </div>
);
