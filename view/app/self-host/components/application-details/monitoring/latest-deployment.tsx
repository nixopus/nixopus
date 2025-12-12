'use client';

import { Box, Clock, GitCommit, Container } from 'lucide-react';
import { formatDistanceToNow, format } from 'date-fns';
import { InfoLine } from './info-line';
import { SectionLabel } from './section-label';
import { StatusIndicator } from './status-indicator';
import { useTranslation } from '@/hooks/use-translation';
import { ApplicationDeployment, Status } from '@/redux/types/applications';
import { Card, CardContent } from '@/components/ui/card';

interface LatestDeploymentProps {
  deployment?: ApplicationDeployment;
}

export function LatestDeployment({ deployment }: LatestDeploymentProps) {
  const { t } = useTranslation();

  if (!deployment) {
    return (
      <Card className="h-full border-dashed">
        <CardContent className="flex flex-col items-center justify-center h-full py-8 text-muted-foreground">
          <Box className="h-10 w-10 mb-3 opacity-30" />
          <p className="font-medium">{t('selfHost.monitoring.latestDeployment.noDeployment')}</p>
          <p className="text-sm text-muted-foreground/60 mt-1 text-center">
            {t('selfHost.monitoring.latestDeployment.noDeploymentDescription')}
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="h-full">
      <CardContent className="pt-6 h-full flex flex-col">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/10">
              <Container className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">
                {deployment.container_name || t('selfHost.monitoring.latestDeployment.deployment')}
              </p>
              <p className="text-xs text-muted-foreground font-mono">
                {deployment.id.slice(0, 8)}...
              </p>
            </div>
          </div>
          <StatusIndicator status={deployment.status?.status as Status} />
        </div>

        <SectionLabel>{t('selfHost.monitoring.latestDeployment.title')}</SectionLabel>

        <div className="flex flex-col gap-y-2 mt-4 flex-1">
          {deployment.commit_hash && (
            <InfoLine
              icon={GitCommit}
              label={t('selfHost.monitoring.latestDeployment.commitHash')}
              value={deployment.commit_hash}
              displayValue={deployment.commit_hash.slice(0, 7)}
              mono
              copyable
            />
          )}
          {deployment.container_id && (
            <InfoLine
              icon={Box}
              label={t('selfHost.monitoring.latestDeployment.containerId')}
              value={deployment.container_id}
              displayValue={deployment.container_id.slice(0, 12) + '...'}
              mono
              copyable
            />
          )}
          {deployment.container_image && (
            <InfoLine
              icon={Box}
              label={t('selfHost.monitoring.latestDeployment.image')}
              value={deployment.container_image}
              copyable
            />
          )}
          <InfoLine
            icon={Clock}
            label={t('selfHost.monitoring.latestDeployment.deployedAt')}
            value={formatDistanceToNow(new Date(deployment.created_at), { addSuffix: true })}
            sublabel={format(new Date(deployment.created_at), 'PPpp')}
          />
        </div>
      </CardContent>
    </Card>
  );
}
