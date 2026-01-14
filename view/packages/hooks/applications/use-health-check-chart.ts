'use client';

import * as React from 'react';
import { useGetHealthCheckResultsQuery } from '@/redux/services/deploy/healthcheckApi';
import { HealthCheckResult } from '@/redux/types/healthcheck';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import type { ChartConfig } from '@/components/ui/chart';
import { SelectOption } from '@/components/ui/select-wrapper';

interface UseHealthCheckChartProps {
  applicationId: string;
}

interface ChartDataPoint {
  date: string;
  responseTime: number;
  healthStatus: number;
  healthStatusHealthy: number | null;
  healthStatusUnhealthy: number | null;
  isHealthy: boolean;
  healthyCount: number;
  totalCount: number;
}

export function useHealthCheckChart({ applicationId }: UseHealthCheckChartProps) {
  const { t } = useTranslation();
  const [timeRange, setTimeRange] = React.useState('1h');

  const chartConfig = React.useMemo(
    () =>
      ({
        responseTime: {
          label: t('selfHost.monitoring.healthCheck.responseTime' as any) || 'Response Time',
          color: 'var(--primary)'
        },
        healthStatusHealthy: {
          label: t('selfHost.monitoring.healthCheck.healthStatus' as any) || 'Health Status',
          color: 'var(--primary)'
        },
        healthStatusUnhealthy: {
          label: t('selfHost.monitoring.healthCheck.healthStatus' as any) || 'Health Status',
          color: 'var(--destructive)'
        }
      }) satisfies ChartConfig,
    [t]
  );

  const { startTime, endTime } = React.useMemo(() => {
    const end = new Date();
    const start = new Date();

    switch (timeRange) {
      case '10m':
        start.setMinutes(start.getMinutes() - 10);
        break;
      case '1h':
        start.setHours(start.getHours() - 1);
        break;
      case '24h':
        start.setHours(start.getHours() - 24);
        break;
      case '7d':
        start.setDate(start.getDate() - 7);
        break;
      case '30d':
        start.setDate(start.getDate() - 30);
        break;
      case '90d':
        start.setDate(start.getDate() - 90);
        break;
      default:
        start.setHours(start.getHours() - 1);
    }

    return {
      startTime: start.toISOString(),
      endTime: end.toISOString()
    };
  }, [timeRange]);

  const {
    data: results,
    isLoading,
    isFetching
  } = useGetHealthCheckResultsQuery(
    {
      application_id: applicationId,
      limit: 1000,
      start_time: startTime,
      end_time: endTime
    },
    {
      skip: !applicationId,
      refetchOnMountOrArgChange: false,
      refetchOnFocus: false,
      refetchOnReconnect: false
    }
  );

  const chartData = React.useMemo((): ChartDataPoint[] => {
    if (!results || results.length === 0) return [];

    const intervalMinutes =
      timeRange === '10m'
        ? 1
        : timeRange === '1h'
          ? 5
          : timeRange === '24h'
            ? 30
            : timeRange === '7d'
              ? 60
              : timeRange === '30d'
                ? 240
                : timeRange === '90d'
                  ? 720
                  : 5;
    const grouped: Record<string, { responseTime: number[]; healthy: number; total: number }> = {};

    results.forEach((result: HealthCheckResult) => {
      const date = new Date(result.checked_at);
      const intervalKey = new Date(
        Math.floor(date.getTime() / (intervalMinutes * 60 * 1000)) * (intervalMinutes * 60 * 1000)
      ).toISOString();

      if (!grouped[intervalKey]) {
        grouped[intervalKey] = { responseTime: [], healthy: 0, total: 0 };
      }

      grouped[intervalKey].responseTime.push(result.response_time_ms);
      grouped[intervalKey].total++;
      if (result.status === 'healthy') {
        grouped[intervalKey].healthy++;
      }
    });

    return Object.entries(grouped)
      .map(([date, data]) => {
        const avgResponseTime = Math.round(
          data.responseTime.reduce((a, b) => a + b, 0) / data.responseTime.length
        );
        const healthPercentage = data.total > 0 ? (data.healthy / data.total) * 100 : 0;
        const isHealthy = healthPercentage === 100;
        return {
          date,
          responseTime: avgResponseTime,
          healthStatus: healthPercentage,
          healthStatusHealthy: isHealthy ? healthPercentage : null,
          healthStatusUnhealthy: !isHealthy ? healthPercentage : null,
          isHealthy,
          healthyCount: data.healthy,
          totalCount: data.total
        };
      })
      .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
  }, [results, timeRange]);

  const timeRangeOptions: SelectOption[] = React.useMemo(
    () => [
      {
        value: '10m',
        label: t('selfHost.monitoring.healthCheck.last10m' as any) || 'Last 10 minutes'
      },
      {
        value: '1h',
        label: t('selfHost.monitoring.healthCheck.last1h' as any) || 'Last 1 hour'
      },
      {
        value: '24h',
        label: t('selfHost.monitoring.healthCheck.last24h' as any) || 'Last 24 hours'
      },
      {
        value: '7d',
        label: t('selfHost.monitoring.healthCheck.last7d' as any) || 'Last 7 days'
      },
      {
        value: '30d',
        label: t('selfHost.monitoring.healthCheck.last30d' as any) || 'Last 30 days'
      },
      {
        value: '90d',
        label: t('selfHost.monitoring.healthCheck.last90d' as any) || 'Last 90 days'
      }
    ],
    [t]
  );

  const xAxisTickFormatter = React.useCallback(
    (value: string | number) => {
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
    },
    [timeRange]
  );

  const tooltipLabelFormatter = React.useCallback(
    (value: string | number) => {
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
    },
    [timeRange]
  );

  const tooltipFormatter = React.useCallback(
    (value: any, name: string | number, item: any, index: number, payload: any) => {
      const nameStr = String(name);
      if (nameStr === 'responseTime') {
        return [
          `${value}ms`,
          t('selfHost.monitoring.healthCheck.responseTime' as any) || 'Response Time'
        ];
      }
      if (nameStr === 'healthStatusHealthy' || nameStr === 'healthStatusUnhealthy') {
        const data = payload as any;
        const numValue = typeof value === 'number' ? value : parseFloat(String(value));
        if (numValue === null || isNaN(numValue)) return null;
        const status =
          nameStr === 'healthStatusHealthy'
            ? t('selfHost.monitoring.healthCheck.healthy' as any) || 'Healthy'
            : t('selfHost.monitoring.healthCheck.unhealthy' as any) || 'Unhealthy';
        return [
          `${numValue.toFixed(1)}% (${data.healthyCount}/${data.totalCount})`,
          `${t('selfHost.monitoring.healthCheck.healthStatus' as any) || 'Health Status'}: ${status}`
        ];
      }
      return null;
    },
    [t]
  );

  return {
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
    hasData: results && results.length > 0
  };
}
