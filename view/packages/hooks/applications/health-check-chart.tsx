'use client';

import * as React from 'react';
import { Area, CartesianGrid, XAxis, YAxis, Line, ComposedChart } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent
} from '@/components/ui/chart';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@/components/ui/button';
import { Settings } from 'lucide-react';
import { useHealthCheckChart } from '@/packages/hooks/applications/use-health-check-chart';

interface HealthCheckChartProps {
  applicationId: string;
  setDialogOpen: (open: boolean) => void;
  dialogOpen: boolean;
}

export function HealthCheckChart({ applicationId, setDialogOpen }: HealthCheckChartProps) {
  const { t } = useTranslation();
  const { chartData, chartConfig, timeRange, setTimeRange, isLoading, results, hasData } =
    useHealthCheckChart({ applicationId });

  if (isLoading && !results) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>
            {t('selfHost.monitoring.healthCheck.history' as any) || 'Health Check History'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-muted-foreground">
            {t('selfHost.monitoring.healthCheck.loadingChartData' as any) ||
              'Loading chart data...'}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!hasData) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>
            {t('selfHost.monitoring.healthCheck.history' as any) || 'Health Check History'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-muted-foreground">
            {t('selfHost.monitoring.healthCheck.noData' as any) || 'No historical data available'}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="pt-0">
      <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
        <div className="grid flex-1 gap-1">
          <CardTitle>
            {t('selfHost.monitoring.healthCheck.history' as any) || 'Health Check History'}
          </CardTitle>
          <CardDescription>
            {t('selfHost.monitoring.healthCheck.historyDescription' as any) ||
              'Response time and health status over time'}
          </CardDescription>
        </div>
        <Button variant="ghost" size="sm" onClick={() => setDialogOpen(true)}>
          <Settings className="h-4 w-4" />
        </Button>
        <Select value={timeRange} onValueChange={setTimeRange}>
          <SelectTrigger
            className="hidden w-[160px] rounded-lg sm:ml-auto sm:flex"
            aria-label={
              t('selfHost.monitoring.healthCheck.selectTimeRange' as any) || 'Select time range'
            }
          >
            <SelectValue
              placeholder={t('selfHost.monitoring.healthCheck.last1h' as any) || 'Last 1 hour'}
            />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            <SelectItem value="10m" className="rounded-lg">
              {t('selfHost.monitoring.healthCheck.last10m' as any) || 'Last 10 minutes'}
            </SelectItem>
            <SelectItem value="1h" className="rounded-lg">
              {t('selfHost.monitoring.healthCheck.last1h' as any) || 'Last 1 hour'}
            </SelectItem>
            <SelectItem value="24h" className="rounded-lg">
              {t('selfHost.monitoring.healthCheck.last24h' as any) || 'Last 24 hours'}
            </SelectItem>
            <SelectItem value="7d" className="rounded-lg">
              {t('selfHost.monitoring.healthCheck.last7d' as any) || 'Last 7 days'}
            </SelectItem>
            <SelectItem value="30d" className="rounded-lg">
              {t('selfHost.monitoring.healthCheck.last30d' as any) || 'Last 30 days'}
            </SelectItem>
            <SelectItem value="90d" className="rounded-lg">
              {t('selfHost.monitoring.healthCheck.last90d' as any) || 'Last 90 days'}
            </SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer config={chartConfig} className="aspect-auto h-[250px] w-full">
          <ComposedChart data={chartData}>
            <defs>
              <linearGradient id="fillResponseTime" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.8} />
                <stop offset="95%" stopColor="var(--primary)" stopOpacity={0.1} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
              tickFormatter={(value) => {
                const date = new Date(value);
                if (timeRange === '10m' || timeRange === '1h') {
                  return date.toLocaleTimeString('en-US', {
                    hour: 'numeric',
                    minute: '2-digit'
                  });
                }
                if (timeRange === '24h') {
                  return date.toLocaleTimeString('en-US', {
                    hour: 'numeric',
                    minute: '2-digit'
                  });
                }
                return date.toLocaleDateString('en-US', {
                  month: 'short',
                  day: 'numeric'
                });
              }}
            />
            <YAxis
              yAxisId="left"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => `${value}ms`}
            />
            <YAxis
              yAxisId="right"
              orientation="right"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              domain={[0, 100]}
              tickFormatter={(value) => `${value}%`}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  labelFormatter={(value) => {
                    const date = new Date(value);
                    if (timeRange === '10m' || timeRange === '1h' || timeRange === '24h') {
                      return date.toLocaleString('en-US', {
                        month: 'short',
                        day: 'numeric',
                        hour: 'numeric',
                        minute: '2-digit'
                      });
                    }
                    return date.toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                      hour: 'numeric',
                      minute: '2-digit'
                    });
                  }}
                  formatter={(value, name, props) => {
                    if (name === 'responseTime') {
                      return [
                        `${value}ms`,
                        t('selfHost.monitoring.healthCheck.responseTime' as any) || 'Response Time'
                      ];
                    }
                    if (name === 'healthStatusHealthy' || name === 'healthStatusUnhealthy') {
                      const data = props.payload as any;
                      const numValue =
                        typeof value === 'number' ? value : parseFloat(String(value));
                      if (numValue === null || isNaN(numValue)) return null;
                      const status =
                        name === 'healthStatusHealthy'
                          ? t('selfHost.monitoring.healthCheck.healthy' as any) || 'Healthy'
                          : t('selfHost.monitoring.healthCheck.unhealthy' as any) || 'Unhealthy';
                      return [
                        `${numValue.toFixed(1)}% (${data.healthyCount}/${data.totalCount})`,
                        `${t('selfHost.monitoring.healthCheck.healthStatus' as any) || 'Health Status'}: ${status}`
                      ];
                    }
                    return null;
                  }}
                  indicator="dot"
                />
              }
            />
            <Area
              yAxisId="left"
              dataKey="responseTime"
              type="natural"
              fill="url(#fillResponseTime)"
              stroke="var(--primary)"
              strokeWidth={2}
            />
            <Line
              yAxisId="right"
              dataKey="healthStatusHealthy"
              type="natural"
              stroke="var(--primary)"
              strokeWidth={2}
              strokeDasharray="5 5"
              dot={false}
              connectNulls={false}
            />
            <Line
              yAxisId="right"
              dataKey="healthStatusUnhealthy"
              type="natural"
              stroke="var(--destructive)"
              strokeWidth={2}
              strokeDasharray="5 5"
              dot={false}
              connectNulls={false}
            />
            <ChartLegend content={<ChartLegendContent />} />
          </ComposedChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
