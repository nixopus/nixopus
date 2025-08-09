'use client';

import React from 'react';
import { Server } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

interface SystemInfoCardProps {
  systemStats: SystemStatsType;
}

const SystemInfoCard: React.FC<SystemInfoCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  const { load } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Server className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.system.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1 sm:space-y-2">
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.osType')}
            </TypographyMuted>
            <TypographySmall className="truncate max-w-[60%] text-right">
              {systemStats.os_type}
            </TypographySmall>
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.cpu')}
            </TypographyMuted>
            <TypographySmall className="truncate max-w-[60%] text-right">
              {systemStats.cpu_info}
            </TypographySmall>
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.uptime')}
            </TypographyMuted>
            <TypographySmall className="truncate max-w-[60%] text-right">
              {load.uptime}
            </TypographySmall>
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.lastUpdated')}
            </TypographyMuted>
            <TypographySmall className="truncate max-w-[60%] text-right">
              {new Date(systemStats.timestamp).toLocaleTimeString()}
            </TypographySmall>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default SystemInfoCard;

export function SystemInfoCardSkeleton() {
  const { t } = useTranslation();

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Server className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          <TypographySmall>{t('dashboard.system.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-1 sm:space-y-2">
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.osType')}
            </TypographyMuted>
            <Skeleton className="h-4 w-24" />
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.cpu')}
            </TypographyMuted>
            <Skeleton className="h-4 w-32" />
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.uptime')}
            </TypographyMuted>
            <Skeleton className="h-4 w-20" />
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs sm:text-sm">
              {t('dashboard.system.lastUpdated')}
            </TypographyMuted>
            <Skeleton className="h-4 w-24" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
