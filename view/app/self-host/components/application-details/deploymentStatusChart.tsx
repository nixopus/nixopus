'use client';

import { ApplicationDeployment } from '@/redux/types/applications';
import React, { useMemo } from 'react';
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent
} from '@/components/ui/chart';
import { useTranslation } from '@/hooks/use-translation';

export const DeploymentStatusChart = ({
  deployments = []
}: {
  deployments: ApplicationDeployment[];
}) => {
  const { t } = useTranslation();

  const statusCounts = useMemo(() => {
    const counts = {
      failed: 0,
      building: 0,
      deploying: 0,
      deployed: 0
    };

    deployments.forEach((deployment) => {
      if (deployment.status && deployment.status.status) {
        counts[deployment.status.status as keyof typeof counts]++;
      }
    });

    return [
      { status: t('selfHost.deployments.chart.status.building'), value: counts.building },
      { status: t('selfHost.deployments.chart.status.deployed'), value: counts.deployed },
      { status: t('selfHost.deployments.chart.status.deploying'), value: counts.deploying },
      { status: t('selfHost.deployments.chart.status.failed'), value: counts.failed }
    ];
  }, [deployments, t]);

  const statusChartConfig = {
    value: {
      label: 'Count',
      color: 'hsl(var(--color-desktop))'
    }
  } satisfies ChartConfig;

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('selfHost.deployments.chart.title')}</CardTitle>
        <CardDescription>{t('selfHost.deployments.chart.description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={statusChartConfig}>
          <BarChart accessibilityLayer data={statusCounts}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="status" type="category" tickLine={false} axisLine={false} width={100} />
            <YAxis type="number" tickLine={false} axisLine={false} />
            <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="dashed" />} />
            <Bar dataKey="value" className="fill-secondary" radius={4} />
          </BarChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="text-sm text-muted-foreground">
        {t('selfHost.deployments.chart.total').replace('{count}', deployments.length.toString())}
      </CardFooter>
    </Card>
  );
};
