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
      started: 0,
      running: 0,
      stopped: 0,
      failed: 0
    };

    deployments.forEach((deployment) => {
      if (deployment.status && deployment.status.status) {
        counts[deployment.status.status as keyof typeof counts]++;
      }
    });

    return [
      { status: 'Started', value: counts.started },
      { status: 'Running', value: counts.running },
      { status: 'Stopped', value: counts.stopped },
      { status: 'Failed', value: counts.failed }
    ];
  }, [deployments]);

  const statusChartConfig = {
    value: {
      label: 'Count',
      color: 'hsl(var(--primary))'
    }
  } satisfies ChartConfig;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Deployment Status Distribution</CardTitle>
        <CardDescription>Current status of all deployments</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={statusChartConfig}>
          <BarChart accessibilityLayer data={statusCounts} layout="vertical">
            <CartesianGrid horizontal={false} />
            <XAxis type="number" tickLine={false} axisLine={false} />
            <YAxis dataKey="status" type="category" tickLine={false} axisLine={false} width={100} />
            <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="dashed" />} />
            <Bar dataKey="value" fill="var(--color-desktop)" radius={4} />
          </BarChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="text-sm text-muted-foreground">
        Total deployments: {deployments.length}
      </CardFooter>
    </Card>
  );
};
