'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useGetHealthCheckQuery } from '@/redux/services/deploy/healthcheckApi';
import { Application } from '@/redux/types/applications';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { HealthCheckDialog } from './health-check-dialog';
import { HealthCheckChart } from './health-check-chart';
import { useState } from 'react';
import { useHealthCheckWebSocket } from '@/packages/hooks/applications/use-health-check-websocket';

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
