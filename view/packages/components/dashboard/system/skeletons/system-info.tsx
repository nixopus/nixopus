'use client';

import React from 'react';
import { Server } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall } from '@/components/ui/typography';

export function SystemInfoCardSkeleton() {
  const { t } = useTranslation();

  return (
    <Card className="overflow-hidden h-full flex flex-col w-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-bold flex items-center">
          <Server className="h-4 w-4 mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.system.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[...Array(8)].map((_, index) => (
            <div key={index} className="flex items-start gap-3 p-2 rounded-lg">
              {/* Icon placeholder */}
              <Skeleton className="h-4 w-4 mt-0.5 rounded" />
              <div className="flex-1 min-w-0 space-y-1">
                {/* Label */}
                <Skeleton className="h-3 w-16" />
                {/* Value */}
                <Skeleton className="h-3 w-20" />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
