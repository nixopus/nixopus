'use client';

import { StatBlock } from './stat-block';
import { useTranslation } from '@/hooks/use-translation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity } from 'lucide-react';

interface DeploymentOverviewProps {
  totalDeployments: number;
  successfulDeployments: number;
  failedDeployments: number;
  currentStatus?: string;
}

export function DeploymentOverview({
  totalDeployments,
  successfulDeployments,
  failedDeployments,
  currentStatus
}: DeploymentOverviewProps) {
  const { t } = useTranslation();

  const getStatusColor = (status?: string): 'emerald' | 'red' | 'amber' | 'blue' | 'purple' => {
    switch (status) {
      case 'deployed':
        return 'emerald';
      case 'failed':
        return 'red';
      case 'building':
        return 'amber';
      case 'deploying':
        return 'blue';
      case 'cloning':
        return 'purple';
      default:
        return 'amber';
    }
  };

  const isActive = currentStatus === 'deployed';

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5 text-muted-foreground" />
          <CardTitle>{t('selfHost.monitoring.overview.title')}</CardTitle>
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
          <StatBlock
            value={currentStatus || t('selfHost.monitoring.overview.noDeployment')}
            label={t('selfHost.monitoring.overview.status')}
            color={getStatusColor(currentStatus)}
            pulse={isActive}
          />
          <StatBlock
            value={totalDeployments}
            label={t('selfHost.monitoring.overview.totalDeployments')}
            sublabel={t('selfHost.monitoring.overview.allTime')}
          />
          <StatBlock
            value={successfulDeployments}
            label={t('selfHost.monitoring.overview.successful')}
            color="emerald"
          />
          <StatBlock
            value={failedDeployments}
            label={t('selfHost.monitoring.overview.failed')}
            color={failedDeployments > 0 ? 'red' : undefined}
          />
        </div>
      </CardContent>
    </Card>
  );
}
