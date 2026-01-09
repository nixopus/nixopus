'use client';

import React from 'react';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { Button } from '@/components/ui/button';
import { Package, ArrowRight } from 'lucide-react';
import { ContainerData } from '@/redux/types/monitor';
import { useTranslation } from '@/hooks/use-translation';
import { useRouter } from 'next/navigation';
import { Skeleton } from '@/components/ui/skeleton';
import { DataTable, TableColumn } from '@/components/ui/data-table';

interface ContainersWidgetProps {
  containersData: ContainerData[];
  columns: TableColumn<ContainerData>[];
}

const ContainersWidget: React.FC<ContainersWidgetProps> = ({ containersData, columns }) => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <CardWrapper
      title={t('dashboard.containers.title')}
      icon={Package}
      compact
      actions={
        <Button variant="outline" size="sm" onClick={() => router.push('/containers')}>
          <ArrowRight className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
          {t('dashboard.containers.viewAll')}
        </Button>
      }
    >
      <DataTable
        data={containersData}
        columns={columns}
        emptyMessage={t('dashboard.containers.table.noContainers')}
        showBorder={false}
        hoverable={false}
      />
    </CardWrapper>
  );
};

export default ContainersWidget;

export const ContainersWidgetSkeleton: React.FC = () => {
  const { t } = useTranslation();

  return (
    <CardWrapper
      title={t('dashboard.containers.title')}
      icon={Package}
      compact
      actions={<Skeleton className="h-8 w-24" />}
    >
      <div className="border-b pb-2 mb-2">
        <div className="grid grid-cols-6 gap-4">
          {['h-4 w-8', 'h-4 w-12', 'h-4 w-12', 'h-4 w-12', 'h-4 w-10', 'h-4 w-14'].map(
            (className, idx) => (
              <Skeleton key={idx} className={className} />
            )
          )}
        </div>
      </div>
      <div className="space-y-3">
        {[0, 1, 2].map((i) => (
          <div key={i} className="grid grid-cols-6 gap-4 items-center">
            {[
              'h-4 w-16 font-mono',
              'h-4 w-24',
              'h-4 w-32',
              'h-5 w-16 rounded-full',
              'h-4 w-12',
              'h-4 w-16'
            ].map((className, idx) => (
              <Skeleton key={idx} className={className} />
            ))}
          </div>
        ))}
      </div>
    </CardWrapper>
  );
};
