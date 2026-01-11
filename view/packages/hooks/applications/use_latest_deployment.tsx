import { useMemo } from 'react';
import { formatDistanceToNow, format } from 'date-fns';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ApplicationDeployment, Status } from '@/redux/types/applications';
import { Box, Clock, GitCommit, Container } from 'lucide-react';
import { StatusIndicator } from '@/packages/components/application-details';

interface UseLatestDeploymentProps {
  deployment?: ApplicationDeployment;
}

export function useLatestDeployment({ deployment }: UseLatestDeploymentProps) {
  const { t } = useTranslation();

  const emptyStateContent = useMemo(
    () => (
      <div className="flex flex-col items-center justify-center h-full py-8 text-muted-foreground">
        <Box className="h-10 w-10 mb-3 opacity-30" />
        <p className="font-medium">{t('selfHost.monitoring.latestDeployment.noDeployment')}</p>
        <p className="text-sm text-muted-foreground/60 mt-1 text-center">
          {t('selfHost.monitoring.latestDeployment.noDeploymentDescription')}
        </p>
      </div>
    ),
    [t]
  );

  const headerContent = useMemo(() => {
    if (!deployment) return null;

    return (
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
    );
  }, [deployment, t]);

  const infoLines = useMemo(() => {
    if (!deployment) return [];

    const lines = [];

    if (deployment.commit_hash) {
      lines.push({
        key: 'commit',
        icon: GitCommit,
        label: t('selfHost.monitoring.latestDeployment.commitHash'),
        value: deployment.commit_hash,
        displayValue: deployment.commit_hash.slice(0, 7),
        mono: true,
        copyable: true
      });
    }

    if (deployment.container_id) {
      lines.push({
        key: 'containerId',
        icon: Box,
        label: t('selfHost.monitoring.latestDeployment.containerId'),
        value: deployment.container_id,
        displayValue: deployment.container_id.slice(0, 12) + '...',
        mono: true,
        copyable: true
      });
    }

    if (deployment.container_image) {
      lines.push({
        key: 'image',
        icon: Box,
        label: t('selfHost.monitoring.latestDeployment.image'),
        value: deployment.container_image,
        mono: false,
        copyable: true
      });
    }

    lines.push({
      key: 'deployedAt',
      icon: Clock,
      label: t('selfHost.monitoring.latestDeployment.deployedAt'),
      value: formatDistanceToNow(new Date(deployment.created_at), { addSuffix: true }),
      sublabel: format(new Date(deployment.created_at), 'PPpp'),
      mono: false,
      copyable: false
    });

    return lines;
  }, [deployment, t]);

  return {
    emptyStateContent,
    headerContent,
    infoLines,
    hasDeployment: !!deployment
  };
}
