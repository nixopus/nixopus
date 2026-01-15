'use client';

import * as React from 'react';
import { Area, CartesianGrid, XAxis, YAxis, Line, ComposedChart } from 'recharts';
import { CardWrapper } from '@/components/ui/card-wrapper';
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent
} from '@/components/ui/chart';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Button } from '@/components/ui/button';
import { Settings } from 'lucide-react';
import { useHealthCheckChart } from '@/packages/hooks/applications/use-health-check-chart';
import { useState } from 'react';
import { useGetHealthCheckQuery } from '@/redux/services/deploy/healthcheckApi';
import { useHealthCheckWebSocket } from '@/packages/hooks/applications/use-health-check-websocket';
import { useHealthCheckDialog } from '@/packages/hooks/applications/use-health-check-dialog';
import {
  HealthCheckChartProps,
  HealthCheckCardProps,
  HealthCheckDialogProps
} from '@/redux/types/applications';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

export function HealthCheckChart({ applicationId, setDialogOpen }: HealthCheckChartProps) {
  const { t } = useTranslation();
  const {
    chartData,
    chartConfig,
    timeRange,
    setTimeRange,
    timeRangeOptions,
    xAxisTickFormatter,
    tooltipLabelFormatter,
    tooltipFormatter,
    isLoading,
    results,
    hasData
  } = useHealthCheckChart({ applicationId });

  if (isLoading && !results) {
    return (
      <CardWrapper
        title={t('selfHost.monitoring.healthCheck.history' as any) || 'Health Check History'}
      >
        <div className="text-muted-foreground">
          {t('selfHost.monitoring.healthCheck.loadingChartData' as any) || 'Loading chart data...'}
        </div>
      </CardWrapper>
    );
  }

  if (!hasData) {
    return (
      <CardWrapper
        title={t('selfHost.monitoring.healthCheck.history' as any) || 'Health Check History'}
      >
        <div className="text-muted-foreground">
          {t('selfHost.monitoring.healthCheck.noData' as any) || 'No historical data available'}
        </div>
      </CardWrapper>
    );
  }

  return (
    <CardWrapper
      className="pt-0"
      header={
        <>
          <div className="grid flex-1 gap-1">
            <div className="text-sm font-bold">
              {t('selfHost.monitoring.healthCheck.history' as any) || 'Health Check History'}
            </div>
            <div className="text-sm text-muted-foreground">
              {t('selfHost.monitoring.healthCheck.historyDescription' as any) ||
                'Response time and health status over time'}
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <Button variant="ghost" size="sm" onClick={() => setDialogOpen(true)}>
              <Settings className="h-4 w-4" />
            </Button>
            <SelectWrapper
              value={timeRange}
              onValueChange={setTimeRange}
              options={timeRangeOptions}
              placeholder={t('selfHost.monitoring.healthCheck.last1h' as any) || 'Last 1 hour'}
              triggerClassName="w-[160px] rounded-lg sm:ml-auto"
              contentClassName="rounded-xl"
              className="hidden sm:block"
            />
          </div>
        </>
      }
      headerClassName="border-b py-5"
      contentClassName="px-2 pt-4 sm:px-6 sm:pt-6"
    >
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
            tickFormatter={xAxisTickFormatter}
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
                labelFormatter={tooltipLabelFormatter}
                formatter={tooltipFormatter as any}
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
    </CardWrapper>
  );
}

export function HealthCheckCard({ application }: HealthCheckCardProps) {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data: healthCheck, isLoading: isLoadingCheck } = useGetHealthCheckQuery(application.id, {
    skip: !application.id
  });

  useHealthCheckWebSocket({ applicationId: application.id });

  if (isLoadingCheck) {
    return (
      <CardWrapper title={t('selfHost.monitoring.healthCheck.title' as any)}>
        <div className="text-muted-foreground">Loading...</div>
      </CardWrapper>
    );
  }

  if (!healthCheck) {
    return (
      <>
        <CardWrapper title={t('selfHost.monitoring.healthCheck.title' as any)}>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              {t('selfHost.monitoring.healthCheck.notConfigured' as any)}
            </p>
            <Button onClick={() => setDialogOpen(true)}>
              {t('selfHost.monitoring.healthCheck.enable' as any)}
            </Button>
          </div>
        </CardWrapper>
        <HealthCheckDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          application={application}
        />
      </>
    );
  }

  return (
    <>
      {healthCheck && (
        <HealthCheckChart
          applicationId={application.id}
          setDialogOpen={setDialogOpen}
          dialogOpen={dialogOpen}
        />
      )}
      <HealthCheckDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        application={application}
        healthCheck={healthCheck}
      />
    </>
  );
}

export function HealthCheckDialog({
  open,
  onOpenChange,
  application,
  healthCheck
}: HealthCheckDialogProps) {
  const { t } = useTranslation();
  const {
    endpoint,
    setEndpoint,
    method,
    setMethod,
    methodOptions,
    intervalSeconds,
    setIntervalSeconds,
    handleIntervalSecondsBlur,
    timeoutSeconds,
    setTimeoutSeconds,
    handleTimeoutSecondsBlur,
    enabled,
    handleSubmit,
    handleDelete,
    handleToggle,
    dialogActions,
    isLoading
  } = useHealthCheckDialog({
    application,
    healthCheck,
    onSuccess: () => onOpenChange(false)
  });

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title={
        healthCheck
          ? t('selfHost.monitoring.healthCheck.editTitle' as any)
          : t('selfHost.monitoring.healthCheck.createTitle' as any)
      }
      description={t('selfHost.monitoring.healthCheck.description' as any)}
      actions={dialogActions}
      loading={isLoading}
    >
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>{t('selfHost.monitoring.healthCheck.endpoint' as any)}</Label>
          <Input
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder="/health"
          />
        </div>

        <div className="space-y-2">
          <Label>{t('selfHost.monitoring.healthCheck.method' as any)}</Label>
          <SelectWrapper
            value={method}
            onValueChange={(v) => setMethod(v as 'GET' | 'POST' | 'HEAD')}
            options={methodOptions}
            placeholder="Select method"
          />
        </div>

        <div className="space-y-2">
          <Label>{t('selfHost.monitoring.healthCheck.intervalSeconds' as any)}</Label>
          <Input
            type="number"
            value={intervalSeconds}
            onChange={(e) => setIntervalSeconds(e.target.value)}
            onBlur={handleIntervalSecondsBlur}
            min={30}
            max={3600}
          />
        </div>

        <div className="space-y-2">
          <Label>{t('selfHost.monitoring.healthCheck.timeoutSeconds' as any)}</Label>
          <Input
            type="number"
            value={timeoutSeconds}
            onChange={(e) => setTimeoutSeconds(e.target.value)}
            onBlur={handleTimeoutSecondsBlur}
            min={5}
            max={120}
          />
        </div>

        {healthCheck && (
          <div className="flex items-center justify-between">
            <Label>{t('selfHost.monitoring.healthCheck.enabled' as any)}</Label>
            <Switch checked={enabled} onCheckedChange={handleToggle} />
          </div>
        )}
      </div>
    </DialogWrapper>
  );
}
