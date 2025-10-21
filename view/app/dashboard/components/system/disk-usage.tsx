'use client';

import React from 'react';
import { HardDrive } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import {
  Table,
  TableBody,
  TableRow,
  TableCell,
  TableHead,
  TableHeader
} from '@/components/ui/table';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

interface DiskUsageCardProps {
  systemStats: SystemStatsType | null;
}

const DiskUsageCard: React.FC<DiskUsageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();

  if (!systemStats) {
    return <DiskUsageCardSkeleton />;
  }

  const { disk } = systemStats;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-bold flex items-center">
          <HardDrive className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.disk.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 sm:space-y-3">
          <div className="w-full h-2 bg-gray-200 rounded-full">
            <div
              className={`h-2 rounded-full bg-primary`}
              style={{ width: `${disk.percentage}%` }}
            />
          </div>
          <div className="flex justify-between">
            <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
              {t('dashboard.disk.used').replace('{value}', disk.used.toFixed(2))}
            </TypographyMuted>
            <TypographyMuted className="text-xs truncate max-w-[60px] sm:max-w-[80px]">
              {t('dashboard.disk.percentage').replace('{value}', disk.percentage.toFixed(1))}
            </TypographyMuted>
            <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
              {t('dashboard.disk.total').replace('{value}', disk.total.toFixed(2))}
            </TypographyMuted>
          </div>
          <div className="text-xs font-mono mt-1 sm:mt-2">
            <Table className="min-w-full overflow-x-hidden">
              <TableHeader>
                <TableRow>
                  <TableHead className="text-left pr-1 sm:pr-2">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.mount')}
                    </TypographySmall>
                  </TableHead>
                  <TableHead className="text-left pr-1 sm:pr-2">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.size')}
                    </TypographySmall>
                  </TableHead>
                  <TableHead className="text-left pr-1 sm:pr-2">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.used')}
                    </TypographySmall>
                  </TableHead>
                  <th className="text-left">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.percentage')}
                    </TypographySmall>
                  </th>
                </TableRow>
              </TableHeader>
              <TableBody>
                {disk.allMounts.map((mount, index) => (
                  <TableRow key={index} className="border-0">
                    <TableCell>
                      <TypographySmall className="text-xs">{mount.mountPoint}</TypographySmall>
                    </TableCell>
                    <TableCell>
                      <TypographySmall className="text-xs">{mount.size}</TypographySmall>
                    </TableCell>
                    <TableCell>
                      <TypographySmall className="text-xs">{mount.used}</TypographySmall>
                    </TableCell>
                    <TableCell>
                      <TypographySmall className="text-xs">{mount.capacity}</TypographySmall>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default DiskUsageCard;

const DiskUsageCardSkeleton = () => {
  const { t } = useTranslation();

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <HardDrive className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.disk.title')}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 sm:space-y-3">
          <div className="w-full h-2 bg-gray-200 rounded-full">
            <div className="h-2 rounded-full bg-gray-400" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-3 w-20" />
            <Skeleton className="h-3 w-10" />
            <Skeleton className="h-3 w-20" />
          </div>
          <div className="text-xs font-mono mt-1 sm:mt-2 overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr>
                  <th className="text-left pr-1 sm:pr-2">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.mount')}
                    </TypographySmall>
                  </th>
                  <th className="text-right pr-1 sm:pr-2">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.size')}
                    </TypographySmall>
                  </th>
                  <th className="text-right pr-1 sm:pr-2">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.used')}
                    </TypographySmall>
                  </th>
                  <th className="text-right">
                    <TypographySmall className="text-xs">
                      {t('dashboard.disk.table.headers.percentage')}
                    </TypographySmall>
                  </th>
                </tr>
              </thead>
              <tbody className="text-xxs sm:text-xs">
                <tr>
                  <td className="text-left pr-1 sm:pr-2">
                    <Skeleton className="h-3 w-10" />
                  </td>
                  <td className="text-right pr-1 sm:pr-2">
                    <Skeleton className="h-3 w-10" />
                  </td>
                  <td className="text-right pr-1 sm:pr-2">
                    <Skeleton className="h-3 w-10" />
                  </td>
                  <td className="text-right">
                    <Skeleton className="h-3 w-10" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
