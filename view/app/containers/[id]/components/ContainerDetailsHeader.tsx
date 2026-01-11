'use client';

import { Play, StopCircle, Trash2, RotateCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import SubPageHeader from '@/components/ui/sub-page-header';
import { Container } from '@/redux/services/container/containerApi';
import { translationKey } from '@/hooks/use-translation';

interface ContainerDetailsHeaderProps {
  container: Container;
  isLoading: boolean;
  isProtected: boolean;
  handleContainerAction: (action: 'start' | 'stop' | 'restart' | 'remove') => void;
  t: (key: translationKey, params?: Record<string, string>) => string;
}

export function ContainerDetailsHeader({
  container,
  isLoading,
  isProtected,
  handleContainerAction,
  t
}: ContainerDetailsHeaderProps) {
  const statusColor =
    container.status === 'running'
      ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20'
      : container.status === 'exited'
        ? 'bg-red-500/10 text-red-500 border-red-500/20'
        : 'bg-amber-500/10 text-amber-500 border-amber-500/20';

  const statusIconBg =
    container.status === 'running'
      ? 'bg-emerald-500/10'
      : container.status === 'exited'
        ? 'bg-red-500/10'
        : 'bg-amber-500/10';

  const statusDotColor =
    container.status === 'running'
      ? 'bg-emerald-500 animate-pulse'
      : container.status === 'exited'
        ? 'bg-red-500'
        : 'bg-amber-500';

  const icon = (
    <div className={cn('w-12 h-12 rounded-xl flex items-center justify-center', statusIconBg)}>
      <div className={cn('w-3 h-3 rounded-full', statusDotColor)} />
    </div>
  );

  const metadata = (
    <div className="flex items-center gap-2">
      <code className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded">
        {container.id.slice(0, 12)}
      </code>
      <Badge variant="outline" className={cn('text-xs', statusColor)}>
        {container.status}
      </Badge>
    </div>
  );

  const actions = (
    <>
      <ResourceGuard
        resource="container"
        action="update"
        loadingFallback={<Skeleton className="h-9 w-24" />}
      >
        {container.status !== 'running' ? (
          <Button
            variant="default"
            size="sm"
            onClick={() => handleContainerAction('start')}
            disabled={isLoading || isProtected}
            className="bg-emerald-600 hover:bg-emerald-700"
          >
            <Play className="mr-2 h-4 w-4" />
            {t('containers.start')}
          </Button>
        ) : (
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleContainerAction('stop')}
              disabled={isLoading || isProtected}
            >
              <StopCircle className="mr-2 h-4 w-4" />
              {t('containers.stop')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleContainerAction('restart')}
              disabled={isLoading || isProtected}
            >
              <RotateCw className="mr-2 h-4 w-4" />
              {t('containers.restart')}
            </Button>
          </>
        )}
      </ResourceGuard>
      <ResourceGuard
        resource="container"
        action="delete"
        loadingFallback={<Skeleton className="h-9 w-20" />}
      >
        <Button
          variant="outline"
          size="sm"
          onClick={() => handleContainerAction('remove')}
          disabled={isLoading || isProtected}
          className="text-red-500 hover:text-red-600 hover:bg-red-500/10 border-red-500/20"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          {t('containers.remove')}
        </Button>
      </ResourceGuard>
    </>
  );

  return <SubPageHeader icon={icon} title={container.name} metadata={metadata} actions={actions} />;
}
