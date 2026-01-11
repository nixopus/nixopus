'use client';

import React from 'react';
import { ArrowRight, Package } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { DataTable } from '@/components/ui/data-table';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRouter } from 'next/navigation';
import { ContainersWidgetProps } from '@/packages/types/containers';

export const ContainersWidget: React.FC<ContainersWidgetProps> = ({ containersData, columns }) => {
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
