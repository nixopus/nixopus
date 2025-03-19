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

export const DeploymentStatusChart = ({
  deployments = []
}: {
  deployments: ApplicationDeployment[];
}) => {
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
      { status: 'Building', value: counts.building },
      { status: 'Deployed', value: counts.deployed },
      { status: 'Deploying', value: counts.deploying },
      { status: 'Failed', value: counts.failed }
    ];
  }, [deployments]);

  const statusChartConfig = {
    value: {
      label: 'Count',
      color: 'hsl(var(--color-desktop))'
    }
  } satisfies ChartConfig;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Deployment Stats</CardTitle>
        <CardDescription>Current status of all deployments</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={statusChartConfig}>
          <BarChart accessibilityLayer data={statusCounts}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="status" type="category" tickLine={false} axisLine={false} width={100} />
            <YAxis type="number" tickLine={false} axisLine={false} />
            <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="dashed" />} />
            <Bar dataKey="value" className='fill-secondary' radius={4} />
          </BarChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="text-sm text-muted-foreground">
        Total deployments: {deployments.length}
      </CardFooter>
    </Card>
  );
};
