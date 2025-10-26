'use client';

import React from 'react';
import { HardDrive } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { SystemMetricCard } from './system-metric-card';

interface DiskUsageCardProps {
  systemStats: SystemStatsType | null;
}

interface MountData {
  mountPoint: string;
  size: string;
  used: string;
  capacity: string;
}

const DiskUsageCard: React.FC<DiskUsageCardProps> = ({ systemStats }) => {
  const { t } = useTranslation();
  const isLoading = !systemStats;

  const { disk } = systemStats || {
    disk: {
      percentage: 0,
      used: 0,
      total: 0,
      allMounts: []
    }
  };

  return (
    <SystemMetricCard
      title={t('dashboard.disk.title')}
      icon={HardDrive}
      isLoading={isLoading}
      skeletonContent={<DiskUsageCardSkeletonContent />}
    >
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
          <DiskMountsTable mounts={disk.allMounts} />
        </div>
      </div>
    </SystemMetricCard>
  );
};

export default DiskUsageCard;

function DiskMountsTable({ mounts }: { mounts: MountData[] }) {
  const { t } = useTranslation();

  const columns: TableColumn<MountData>[] = [
    {
      key: 'mount',
      title: t('dashboard.disk.table.headers.mount'),
      dataIndex: 'mountPoint',
      className: 'text-xs pr-1 sm:pr-2',
      render: (mountPoint) => (
        <TypographySmall className="text-xs">{mountPoint}</TypographySmall>
      )
    },
    {
      key: 'size',
      title: t('dashboard.disk.table.headers.size'),
      dataIndex: 'size',
      className: 'text-xs pr-1 sm:pr-2',
      render: (size) => (
        <TypographySmall className="text-xs">{size}</TypographySmall>
      )
    },
    {
      key: 'used',
      title: t('dashboard.disk.table.headers.used'),
      dataIndex: 'used',
      className: 'text-xs pr-1 sm:pr-2',
      render: (used) => (
        <TypographySmall className="text-xs">{used}</TypographySmall>
      )
    },
    {
      key: 'capacity',
      title: t('dashboard.disk.table.headers.percentage'),
      dataIndex: 'capacity',
      className: 'text-xs',
      render: (capacity) => (
        <TypographySmall className="text-xs">{capacity}</TypographySmall>
      )
    }
  ];

  return (
    <DataTable
      data={mounts}
      columns={columns}
      tableClassName="min-w-full overflow-x-hidden"
      showBorder={false}
      hoverable={false}
      striped={false}
    />
  );
}

function DiskUsageCardSkeletonContent() {
  const { t } = useTranslation();

  return (
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
  );
}

const DiskUsageCardSkeleton = () => {
  const { t } = useTranslation();

  return (
    <SystemMetricCard
      title={t('dashboard.disk.title')}
      icon={HardDrive}
      isLoading={true}
      skeletonContent={<DiskUsageCardSkeletonContent />}
    >
      <div />
    </SystemMetricCard>
  );
};
