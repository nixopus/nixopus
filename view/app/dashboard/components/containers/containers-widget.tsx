'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Package, ArrowRight } from 'lucide-react';
import { ContainerData } from '@/redux/types/monitor';
import { useTranslation } from '@/hooks/use-translation';
import { TypographySmall } from '@/components/ui/typography';
import { useRouter } from 'next/navigation';
import ContainersTable from './container-table';

interface ContainersWidgetProps {
  containersData: ContainerData[];
}

const ContainersWidget: React.FC<ContainersWidgetProps> = ({ containersData }) => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-xs sm:text-sm font-bold flex items-center">
          <Package className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{t('dashboard.containers.title')}</TypographySmall>
        </CardTitle>
        <Button variant="outline" size="sm" onClick={() => router.push('/containers')}>
          <ArrowRight className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          {t('dashboard.containers.viewAll')}
        </Button>
      </CardHeader>
      <CardContent>
        <ContainersTable containersData={containersData} />
      </CardContent>
    </Card>
  );
};

export default ContainersWidget;

