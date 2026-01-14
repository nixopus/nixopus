'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { ArrowRight, GitBranch } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { ApplicationDeployment } from '@/redux/types/applications';
import { TypographyMuted, TypographySmall } from '@/components/ui/typography';
import { cn } from '@/lib/utils';
import { getDeploymentStatusIcon, getDeploymentStatusBadgeClasses } from '@/packages/utils/colors';
import { formatRelativeDate } from '@/packages/utils/format-date';
import { useDeploymentsWidget } from '@/packages/hooks/dashboard/use-deployment-stats';
import { useDeploymentStats } from '@/packages/hooks/dashboard/use-deployment-stats';
import { BarChart3 } from 'lucide-react';
import { ChartStyle } from '@/components/ui/chart';
import { ChartContainer } from '@/components/ui/chart';
import { PieChart } from 'recharts';
import { ChartTooltip } from '@/components/ui/chart';
import { ChartTooltipContent } from '@/components/ui/chart';
import { Pie } from 'recharts';
import { Sector } from 'recharts';
import { Label } from 'recharts';
import { PieSectorDataItem } from 'recharts/types/polar/Pie';

export interface DeploymentsWidgetProps {
  deploymentsData: ApplicationDeployment[];
}

export const DeploymentsWidget: React.FC<DeploymentsWidgetProps> = ({ deploymentsData }) => {
  const router = useRouter();
  const { deploymentItems, isEmpty } = useDeploymentsWidget(deploymentsData);

  const cardActions = (
    <Button variant="outline" size="sm" onClick={() => router.push('/self-host')}>
      <ArrowRight className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
      View All
    </Button>
  );

  const handleDeploymentClick = (deployment: ApplicationDeployment) => {
    if (deployment.application_id) {
      router.push(
        `/self-host/application/${deployment.application_id}/deployments/${deployment.id}`
      );
    }
  };

  if (isEmpty) {
    return (
      <CardWrapper title="Latest Deployments" icon={GitBranch} compact actions={cardActions}>
        <div className="text-center py-8">
          <TypographyMuted>No deployments found</TypographyMuted>
        </div>
      </CardWrapper>
    );
  }

  return (
    <CardWrapper title="Latest Deployments" icon={GitBranch} compact actions={cardActions}>
      {deploymentItems.map((item) => (
        <div
          key={item.deployment.id}
          className={cn(
            'relative flex items-start gap-4 py-3 px-2 rounded-lg hover:bg-muted/50 transition-colors cursor-pointer',
            !item.isLast && 'border-b border-border'
          )}
          onClick={() => handleDeploymentClick(item.deployment)}
        >
          {!item.isLast && <div className="absolute left-[11px] top-8 bottom-0 w-0.5 bg-border" />}

          <div
            className={cn(
              'relative z-10 h-6 w-6 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5',
              item.statusColor
            )}
          >
            {getDeploymentStatusIcon(item.status)}
          </div>

          <div className="flex-1 min-w-0 space-y-1">
            <div className="flex items-center justify-between gap-2">
              <TypographySmall className="font-semibold truncate">
                {item.applicationName}
              </TypographySmall>
              <TypographyMuted className="text-xs flex-shrink-0">
                {formatRelativeDate(item.deployment.created_at)}
              </TypographyMuted>
            </div>

            {item.metadataItems.length > 0 && (
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                {item.metadataItems.map((meta, idx) => (
                  <span key={idx} className={cn('flex items-center gap-1', meta.className)}>
                    {'icon' in meta && <meta.icon className="h-3 w-3" />}
                    <span>{meta.content}</span>
                  </span>
                ))}
              </div>
            )}

            {item.deployment.status && (
              <span
                className={cn(
                  'inline-block text-xs px-2 py-0.5 rounded-full',
                  getDeploymentStatusBadgeClasses(item.status)
                )}
              >
                {item.status}
              </span>
            )}
          </div>
        </div>
      ))}
    </CardWrapper>
  );
};

export interface DeploymentStatsWidgetProps {
  deploymentsData: ApplicationDeployment[];
}

export const DeploymentStatsWidget: React.FC<DeploymentStatsWidgetProps> = ({
  deploymentsData
}) => {
  const { stats, pieData, chartConfig, activeStatus, activeIndex, activeData, setActiveStatus } =
    useDeploymentStats(deploymentsData);

  const id = 'deployment-stats-pie';

  if (stats.total === 0 || pieData.length === 0) {
    return (
      <CardWrapper title="Deployment Stats" icon={BarChart3} compact>
        <div className="flex items-center justify-center h-full min-h-[200px]">
          <TypographyMuted>No deployment data available</TypographyMuted>
        </div>
      </CardWrapper>
    );
  }

  return (
    <CardWrapper title="Deployment Stats" icon={BarChart3} compact>
      <div className="flex flex-col h-full" data-chart={id}>
        <ChartStyle id={id} config={chartConfig} />
        <div className="flex flex-1 justify-center items-center py-4">
          <ChartContainer
            id={id}
            config={chartConfig}
            className="mx-auto aspect-square w-full max-w-[280px]"
          >
            <PieChart>
              <ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel />} />
              <Pie
                data={pieData}
                dataKey="count"
                nameKey="status"
                innerRadius={60}
                outerRadius={90}
                strokeWidth={5}
                activeIndex={activeIndex >= 0 ? activeIndex : undefined}
                activeShape={({ outerRadius = 0, ...props }: PieSectorDataItem) => (
                  <g>
                    <Sector {...props} outerRadius={outerRadius + 10} />
                    <Sector
                      {...props}
                      outerRadius={outerRadius + 25}
                      innerRadius={outerRadius + 12}
                    />
                  </g>
                )}
                onClick={(_, index) => {
                  if (pieData[index]) {
                    setActiveStatus(pieData[index].status);
                  }
                }}
              >
                <Label
                  content={({ viewBox }) => {
                    if (viewBox && 'cx' in viewBox && 'cy' in viewBox && activeData) {
                      return (
                        <text
                          x={viewBox.cx}
                          y={viewBox.cy}
                          textAnchor="middle"
                          dominantBaseline="middle"
                        >
                          <tspan
                            x={viewBox.cx}
                            y={viewBox.cy}
                            className="fill-foreground text-3xl font-bold"
                          >
                            {activeData.count.toLocaleString()}
                          </tspan>
                          <tspan
                            x={viewBox.cx}
                            y={(viewBox.cy || 0) + 24}
                            className="fill-muted-foreground text-sm"
                          >
                            {chartConfig[activeData.status as keyof typeof chartConfig]?.label ||
                              'Deployments'}
                          </tspan>
                        </text>
                      );
                    }
                    return null;
                  }}
                />
              </Pie>
            </PieChart>
          </ChartContainer>
        </div>
        <div className="space-y-1.5 pt-2">
          {pieData.map((item) => {
            const config = chartConfig[item.status as keyof typeof chartConfig];
            const isActive = activeStatus === item.status;
            return (
              <div
                key={item.status}
                className={`flex items-center justify-between p-1.5 rounded-lg cursor-pointer transition-colors ${
                  isActive ? 'bg-muted' : 'hover:bg-muted/50'
                }`}
                onClick={() => setActiveStatus(item.status)}
              >
                <div className="flex items-center gap-2">
                  <span
                    className="flex h-2.5 w-2.5 shrink-0 rounded-full"
                    style={{
                      backgroundColor: `var(--color-${item.status})`
                    }}
                  />
                  <TypographyMuted className="text-xs">
                    {config?.label || item.status}
                  </TypographyMuted>
                </div>
                <TypographySmall className="font-semibold text-xs">{item.count}</TypographySmall>
              </div>
            );
          })}
        </div>
      </div>
    </CardWrapper>
  );
};
