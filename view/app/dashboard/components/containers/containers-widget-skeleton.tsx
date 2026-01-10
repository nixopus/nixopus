'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Package } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall } from '@/components/ui/typography';

export const ContainersWidgetSkeleton: React.FC = () => {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-xs sm:text-sm font-bold flex items-center">
          <Package className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.containers.title')}</TypographySmall>
        </CardTitle>
        <Skeleton className="h-8 w-24" /> {/* View All button */}
      </CardHeader>
      <CardContent>
        {/* Table header skeleton */}
        <div className="border-b pb-2 mb-2">
          <div className="grid grid-cols-6 gap-4">
            <Skeleton className="h-4 w-8" /> {/* ID */}
            <Skeleton className="h-4 w-12" /> {/* Name */}
            <Skeleton className="h-4 w-12" /> {/* Image */}
            <Skeleton className="h-4 w-12" /> {/* Status */}
            <Skeleton className="h-4 w-10" /> {/* Ports */}
            <Skeleton className="h-4 w-14" /> {/* Created */}
          </div>
        </div>
        {/* Table rows skeleton */}
        <div className="space-y-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="grid grid-cols-6 gap-4 items-center">
              <Skeleton className="h-4 w-16 font-mono" /> {/* ID value */}
              <Skeleton className="h-4 w-24" /> {/* Name value */}
              <Skeleton className="h-4 w-32" /> {/* Image value */}
              <Skeleton className="h-5 w-16 rounded-full" /> {/* Status badge */}
              <Skeleton className="h-4 w-12" /> {/* Ports value */}
              <Skeleton className="h-4 w-16" /> {/* Created value */}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
};
