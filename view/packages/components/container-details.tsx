'use client';

import { Play, StopCircle, Trash2, RotateCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ResourceGuard } from '@/packages/components/rbac';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import SubPageHeader from '@/components/ui/sub-page-header';
import PageLayout from '@/packages/layouts/page-layout';
import { Container } from '@/redux/services/container/containerApi';
import { translationKey } from '@/packages/hooks/shared/use-translation';

// ============================================================================
// Container Details Header
// ============================================================================

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

// ============================================================================
// Container Details Loading
// ============================================================================

export default function ContainerDetailsLoading() {
  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-4">
          <Skeleton className="w-12 h-12 rounded-xl" />
          <div className="space-y-2">
            <Skeleton className="h-7 w-48" />
            <div className="flex items-center gap-2">
              <Skeleton className="h-5 w-24" />
              <Skeleton className="h-5 w-16" />
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-9 w-20" />
          <Skeleton className="h-9 w-20" />
          <Skeleton className="h-9 w-24" />
        </div>
      </div>

      <div className="border-b mb-6">
        <div className="flex gap-2">
          <Skeleton className="h-10 w-28" />
          <Skeleton className="h-10 w-20" />
          <Skeleton className="h-10 w-24" />
          <Skeleton className="h-10 w-20" />
        </div>
      </div>

      <div className="space-y-10">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-8 w-20" />
              <Skeleton className="h-4 w-24" />
            </div>
          ))}
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-32" />
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[1, 2, 3].map((i) => (
              <div key={i} className="space-y-3">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-8 w-8 rounded-lg" />
                  <Skeleton className="h-4 w-16" />
                </div>
                <Skeleton className="h-2 w-full rounded-full" />
                <Skeleton className="h-6 w-20" />
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-40" />
          <div className="flex flex-wrap gap-4">
            <Skeleton className="h-14 w-48 rounded-xl" />
            <Skeleton className="h-14 w-40 rounded-xl" />
            <Skeleton className="h-14 w-36 rounded-xl" />
          </div>
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-36" />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-12">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="flex items-start gap-3 py-2">
                <Skeleton className="h-4 w-4 mt-1" />
                <div className="space-y-1.5">
                  <Skeleton className="h-3 w-16" />
                  <Skeleton className="h-5 w-40" />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-14 w-full rounded-xl" />
        </div>
      </div>
    </PageLayout>
  );
}
