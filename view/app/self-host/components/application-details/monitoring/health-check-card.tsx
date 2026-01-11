'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  useGetHealthCheckQuery,
  useGetHealthCheckStatsQuery
} from '@/redux/services/deploy/healthcheckApi';
import { Application } from '@/redux/types/applications';
import { Heart, HeartOff, Settings, AlertCircle } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { HealthCheckDialog } from './health-check-dialog';
import { useState } from 'react';
import { useHealthCheckWebSocket } from '../../../hooks/use-health-check-websocket';

interface HealthCheckCardProps {
  application: Application;
}

export function HealthCheckCard({ application }: HealthCheckCardProps) {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data: healthCheck, isLoading: isLoadingCheck } = useGetHealthCheckQuery(application.id, {
    skip: !application.id
  });

  useHealthCheckWebSocket({ applicationId: application.id });

  const { data: stats, isLoading: isLoadingStats } = useGetHealthCheckStatsQuery(
    { application_id: application.id },
    { skip: !healthCheck || !application.id }
  );

  if (isLoadingCheck) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{t('selfHost.monitoring.healthCheck.title' as any)}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-muted-foreground">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  if (!healthCheck) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{t('selfHost.monitoring.healthCheck.title' as any)}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              {t('selfHost.monitoring.healthCheck.notConfigured' as any)}
            </p>
            <Button onClick={() => setDialogOpen(true)}>
              {t('selfHost.monitoring.healthCheck.enable' as any)}
            </Button>
          </div>
        </CardContent>
        <HealthCheckDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          application={application}
        />
      </Card>
    );
  }

  const isHealthy = healthCheck.consecutive_fails < healthCheck.failure_threshold;
  const statusColor = isHealthy ? 'emerald' : 'red';
  const statusText = isHealthy
    ? t('selfHost.monitoring.healthCheck.healthy' as any)
    : t('selfHost.monitoring.healthCheck.unhealthy' as any);

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>{t('selfHost.monitoring.healthCheck.title' as any)}</CardTitle>
            <Button variant="ghost" size="sm" onClick={() => setDialogOpen(true)}>
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              {isHealthy ? (
                <Heart className="h-5 w-5 text-emerald-500" />
              ) : (
                <HeartOff className="h-5 w-5 text-red-500" />
              )}
              <Badge variant={statusColor === 'emerald' ? 'default' : 'destructive'}>
                {statusText}
              </Badge>
              {!healthCheck.enabled && (
                <Badge variant="secondary">
                  {t('selfHost.monitoring.healthCheck.disabled' as any)}
                </Badge>
              )}
            </div>

            {stats && (
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="text-muted-foreground">
                    {t('selfHost.monitoring.healthCheck.uptime' as any)}
                  </div>
                  <div className="text-lg font-semibold">{stats.uptime_percentage.toFixed(1)}%</div>
                </div>
                <div>
                  <div className="text-muted-foreground">
                    {t('selfHost.monitoring.healthCheck.avgResponseTime' as any)}
                  </div>
                  <div className="text-lg font-semibold">{stats.average_response_time_ms}ms</div>
                </div>
              </div>
            )}

            {healthCheck.last_checked_at && (
              <div className="text-xs text-muted-foreground">
                {t('selfHost.monitoring.healthCheck.lastChecked' as any)}:{' '}
                {new Date(healthCheck.last_checked_at).toLocaleString()}
              </div>
            )}

            {!isHealthy && healthCheck.last_error_message && (
              <div className="mt-4 p-3 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-md">
                <div className="flex items-start gap-2">
                  <AlertCircle className="h-4 w-4 text-red-600 dark:text-red-400 mt-0.5 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-semibold text-red-800 dark:text-red-200 mb-1">
                      {t('selfHost.monitoring.healthCheck.error' as any) || 'Error'}
                    </div>
                    <div className="text-xs text-red-700 dark:text-red-300 break-words">
                      {healthCheck.last_error_message}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
      <HealthCheckDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        application={application}
        healthCheck={healthCheck}
      />
    </>
  );
}
