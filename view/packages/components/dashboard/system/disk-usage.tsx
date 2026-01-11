'use client';

import React from 'react';
import { SystemStatsType } from '@/redux/types/monitor';
import { HardDrive } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { DataTable, TableColumn } from '@/components/ui/data-table';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '@/packages/hooks/dashboard/use-system-metric';
import { formatPercentage } from '../utils/utils';
import { DEFAULT_METRICS } from '../utils/constants';
import { DiskUsageCardSkeletonContent } from './skeletons/disk-usage';

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
  const {
    data: disk,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.disk,
    defaultData: DEFAULT_METRICS.disk
  });

  return (
    <SystemMetricCard
      title={t('dashboard.disk.title')}
      icon={HardDrive}
      isLoading={isLoading}
      skeletonContent={<DiskUsageCardSkeletonContent />}
    >
      <div className="space-y-2 sm:space-y-3">
        <div className="w-full h-2 bg-gray-200 rounded-full">
          <div className={`h-2 rounded-full bg-primary`} style={{ width: `${disk.percentage}%` }} />
        </div>
        <div className="flex justify-between">
          <TypographyMuted className="text-xs truncate max-w-[80px] sm:max-w-[100px]">
            {t('dashboard.disk.used').replace('{value}', disk.used.toFixed(2))}
          </TypographyMuted>
          <TypographyMuted className="text-xs truncate max-w-[60px] sm:max-w-[80px]">
            {t('dashboard.disk.percentage').replace('{value}', formatPercentage(disk.percentage))}
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
      render: (mountPoint) => <TypographySmall className="text-xs">{mountPoint}</TypographySmall>
    },
    {
      key: 'size',
      title: t('dashboard.disk.table.headers.size'),
      dataIndex: 'size',
      className: 'text-xs pr-1 sm:pr-2',
      render: (size) => <TypographySmall className="text-xs">{size}</TypographySmall>
    },
    {
      key: 'used',
      title: t('dashboard.disk.table.headers.used'),
      dataIndex: 'used',
      className: 'text-xs pr-1 sm:pr-2',
      render: (used) => <TypographySmall className="text-xs">{used}</TypographySmall>
    },
    {
      key: 'capacity',
      title: t('dashboard.disk.table.headers.percentage'),
      dataIndex: 'capacity',
      className: 'text-xs',
      render: (capacity) => <TypographySmall className="text-xs">{capacity}</TypographySmall>
    }
  ];

  return (
    <div
      className="max-h-[300px] overflow-y-auto overflow-x-hidden scrollbar-accessible"
      role="region"
      aria-label={`${t('dashboard.disk.table.headers.mount')} table with ${mounts.length} ${mounts.length === 1 ? 'mount point' : 'mount points'}`}
      aria-live="polite"
      tabIndex={0}
    >
      <DataTable
        data={mounts}
        columns={columns}
        tableClassName="min-w-full"
        containerClassName="overflow-x-hidden"
        showBorder={false}
        hoverable={false}
        striped={false}
      />
    </div>
  );
}
