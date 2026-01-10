'use client';

import React, { useMemo } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { CircleCheck, CircleAlert, CircleX } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DeploymentHealthChartProps {
  deploymentsByStatus: Record<string, number>;
  totalDeployments: number;
  successRate: number;
}

const COLORS = {
  deployed: '#10b981',
  building: '#3b82f6',
  deploying: '#f59e0b',
  cloning: '#8b5cf6',
  failed: '#ef4444'
};

const STATUS_LABELS: Record<string, string> = {
  deployed: 'Deployed',
  building: 'Building',
  deploying: 'Deploying',
  cloning: 'Cloning',
  failed: 'Failed'
};

export function DeploymentHealthChart({
  deploymentsByStatus,
  totalDeployments,
  successRate
}: DeploymentHealthChartProps) {
  const { t } = useTranslation();

  const chartData = useMemo(() => {
    return Object.entries(deploymentsByStatus)
      .filter(([_, count]) => count > 0)
      .map(([status, count]) => ({
        name: STATUS_LABELS[status] || status,
        value: count,
        color: COLORS[status as keyof typeof COLORS] || '#6b7280'
      }));
  }, [deploymentsByStatus]);

  const getStatusIcon = () => {
    if (successRate >= 80) return <CircleCheck className="h-4 w-4 text-emerald-500" />;
    if (successRate >= 50) return <CircleAlert className="h-4 w-4 text-amber-500" />;
    return <CircleX className="h-4 w-4 text-red-500" />;
  };

  const getHealthStatus = () => {
    if (successRate >= 80)
      return {
        text: t('selfHost.monitoring.chart.healthStatus.healthy'),
        color: 'text-emerald-500'
      };
    if (successRate >= 50)
      return { text: t('selfHost.monitoring.chart.healthStatus.warning'), color: 'text-amber-500' };
    return { text: t('selfHost.monitoring.chart.healthStatus.critical'), color: 'text-red-500' };
  };

  const healthStatus = getHealthStatus();

  if (totalDeployments === 0) {
    return (
      <Card className="h-full">
        <CardHeader>
          <CardTitle>{t('selfHost.monitoring.chart.title')}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <div className="h-32 w-32 rounded-full border-4 border-dashed border-muted flex items-center justify-center mb-4">
            <span className="text-3xl font-bold text-muted-foreground/50">0</span>
          </div>
          <p className="font-medium">{t('selfHost.monitoring.chart.noData')}</p>
          <p className="text-sm text-muted-foreground/60 mt-1">
            {t('selfHost.monitoring.chart.noDataDescription')}
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>{t('selfHost.monitoring.chart.title')}</CardTitle>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-muted">
            {getStatusIcon()}
            <span className={cn('text-sm font-medium', healthStatus.color)}>
              {healthStatus.text}
            </span>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col lg:flex-row items-center gap-8">
          <div className="relative w-64 h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={chartData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={2}
                  dataKey="value"
                  strokeWidth={0}
                >
                  {chartData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  content={({ active, payload }) => {
                    if (active && payload && payload.length) {
                      const data = payload[0].payload;
                      return (
                        <div className="bg-popover border rounded-lg shadow-lg px-3 py-2">
                          <p className="font-medium">{data.name}</p>
                          <p className="text-sm text-muted-foreground">
                            {data.value} {t('selfHost.monitoring.chart.deployments')}
                          </p>
                        </div>
                      );
                    }
                    return null;
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-4xl font-bold">{successRate}%</span>
              <span className="text-sm text-muted-foreground">
                {t('selfHost.monitoring.chart.successRate')}
              </span>
            </div>
          </div>

          <div className="flex-1 grid grid-cols-2 gap-4 w-full lg:w-auto">
            {Object.entries(deploymentsByStatus).map(([status, count]) => (
              <div key={status} className="flex items-center gap-3 p-3 rounded-lg bg-muted/50">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: COLORS[status as keyof typeof COLORS] || '#6b7280' }}
                />
                <div>
                  <p className="text-sm font-medium capitalize">
                    {STATUS_LABELS[status] || status}
                  </p>
                  <p className="text-2xl font-bold">{count}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
