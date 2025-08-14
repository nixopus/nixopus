'use client';

import { RefreshCw, Play, StopCircle, Trash2, Loader2, Scissors } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { cn, isNixopusContainer } from '@/lib/utils';
import ContainersLoading from './skeleton';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FeatureNames } from '@/types/feature-flags';
import { Skeleton } from '@/components/ui/skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import useContainerList from './hooks/use-container-list';
import { TypographyH1, TypographyH2, TypographyMuted } from '@/components/ui/typography';
import { useTranslation } from '@/hooks/use-translation';

interface ContainerActionsProps {
  container: any;
  onAction: (id: string, action: 'start' | 'stop' | 'remove') => void;
}

const ContainerActions = ({ container, onAction }: ContainerActionsProps) => {
  const containerName: string = typeof container?.name === 'string' ? container.name : '';
  const isProtected = isNixopusContainer(containerName);

  return (
    <div className="flex gap-2">
      <ResourceGuard
        resource="container"
        action="update"
        loadingFallback={<Skeleton className="h-8 w-8" />}
      >
        {container.status !== 'running' && (
          <Button
            variant="ghost"
            size="icon"
            disabled={isProtected}
            onClick={(e) => {
              e.stopPropagation();
              onAction(container.id, 'start');
            }}
          >
            <Play className="h-4 w-4" />
          </Button>
        )}
        {container.status === 'running' && (
          <Button
            variant="ghost"
            size="icon"
            disabled={isProtected}
            onClick={(e) => {
              e.stopPropagation();
              onAction(container.id, 'stop');
            }}
          >
            <StopCircle className="h-4 w-4" />
          </Button>
        )}
      </ResourceGuard>
      <ResourceGuard
        resource="container"
        action="delete"
        loadingFallback={<Skeleton className="h-8 w-8" />}
      >
        <Button
          variant="ghost"
          size="icon"
          disabled={isProtected}
          onClick={(e) => {
            e.stopPropagation();
            onAction(container.id, 'remove');
          }}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </ResourceGuard>
    </div>
  );
};

interface ContainerInfoProps {
  container: any;
}

const ContainerInfo = ({ container }: ContainerInfoProps) => {
  const { t } = useTranslation();
  return (
    <div className="space-y-3">
      <div className="text-sm">
        <span className="font-medium">Ports:</span>
        <div className="flex flex-wrap gap-2 mt-1">
          {container?.ports?.length > 0 ? (
            container.ports.map((port: any) => (
              <Badge key={`${port.private_port}-${port.public_port}`} variant="outline">
                {port.public_port} → {port.private_port}
              </Badge>
            ))
          ) : (
            <span className="text-xs text-muted-foreground">{t("containers.no_ports_exposed")}</span>
          )}
        </div>
      </div>
      <div className="text-sm">
        <span className="font-medium">Memory:</span>
        <span className="ml-2">
          {`${(container.host_config.memory / (1024 * 1024)).toFixed(2)} MB`}
        </span>
      </div>
    </div>
  );
};

interface ContainerCardProps {
  container: any;
  onClick: () => void;
  getGradientFromName: (name: string) => string;
  onAction: (id: string, action: 'start' | 'stop' | 'remove') => void;
}

const ContainerCard = ({
  container,
  onClick,
  getGradientFromName,
  onAction
}: ContainerCardProps) => {
  return (
    <Card
      className={cn(
        'group relative overflow-hidden transition-all duration-300 hover:shadow-lg cursor-pointer h-full flex flex-col',
        getGradientFromName(container.name)
      )}
      onClick={onClick}
    >
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f1a_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f1a_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
      <div className="absolute inset-0 bg-gradient-to-br opacity-20 transition-opacity duration-300 group-hover:opacity-30" />
      <CardContent className="relative p-6 z-10 flex-1 flex flex-col">
        <div className="flex items-start justify-between mb-4">
          <div className="space-y-2 flex-1 min-w-0">
            <h3 className="text-xl font-semibold truncate">{container.name}</h3>
            <p className="text-sm text-muted-foreground truncate" title={container.image}>{container.image}</p>
            <Badge variant={container.status === 'running' ? 'default' : 'secondary'}>
              {container.status}
            </Badge>
          </div>
          <div className="flex-shrink-0 ml-4">
            <ContainerActions
              container={container}
              onAction={onAction}
            />
          </div>
        </div>
        <div className="mt-auto">
          <ContainerInfo container={container} />
        </div>
      </CardContent>
    </Card>
  );
};


export default function ContainersPage() {
  const {
    containers,
    isLoading,
    handleRefresh,
    handleContainerAction,
    handleDeleteConfirm,
    handlePruneImages,
    handlePruneBuildCache,
    showPruneImagesConfirm,
    showPruneBuildCacheConfirm,
    isFeatureFlagsLoading,
    isRefreshing,
    isFeatureEnabled,
    t,
    router,
    containerToDelete,
    setContainerToDelete,
    getGradientFromName,
    setShowPruneImagesConfirm,
    setShowPruneBuildCacheConfirm
  } = useContainerList();

  if (isLoading) {
    return <ContainersLoading />;
  }

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureContainer)) {
    return <DisabledFeature />;
  }

  // TODO: Add pagination for containers listing

  return (
    <ResourceGuard
      resource="container"
      action="read"
      loadingFallback={<ContainersLoading />}
    >
      <div className="min-h-screen w-full overflow-x-hidden">
        <div className="relative w-full">
          <div className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 py-6 relative z-10">
            <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
              <span>
                <TypographyH1 className="text-2xl font-bold">{t('containers.title')}</TypographyH1>
                <TypographyMuted>{t('containers.description')}</TypographyMuted>
              </span>
              <div className="flex items-center gap-2 flex-wrap">
                <Button onClick={handleRefresh} variant="outline" size="sm" disabled={isRefreshing}>
                  {isRefreshing ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <RefreshCw className="mr-2 h-4 w-4" />
                  )}
                  {t('containers.refresh')}
                </Button>
                <AnyPermissionGuard
                  permissions={['container:delete']}
                  loadingFallback={<Skeleton className="h-8 w-20" />}
                >
                  <Button variant="outline" size="sm" onClick={() => setShowPruneImagesConfirm(true)}>
                    <Trash2 className="mr-2 h-4 w-4" />
                    {t('containers.prune_images')}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowPruneBuildCacheConfirm(true)}
                  >
                    <Scissors className="mr-2 h-4 w-4" />
                    {t('containers.prune_build_cache')}
                  </Button>
                </AnyPermissionGuard>
              </div>
            </div>
            {containers.length === 0 && (
              <div className="flex justify-center items-center h-full">
                <TypographyH2 className="text-muted-foreground">
                  {t('containers.no_containers')}
                </TypographyH2>
              </div>
            )}
            {containers.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3  gap-4 md:gap-6">
                {containers.map((container) => (
                  <ContainerCard
                    key={container.id}
                    container={container}
                    onClick={() => router.push(`/containers/${container.id}`)}
                    getGradientFromName={getGradientFromName}
                    onAction={handleContainerAction}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
        <AnyPermissionGuard
          permissions={['container:delete']}
          loadingFallback={null}
        >
          <DeleteDialog
            title={t('containers.deleteDialog.title')}
            description={t('containers.deleteDialog.description')}
            onConfirm={handleDeleteConfirm}
            open={!!containerToDelete}
            onOpenChange={(open) => !open && setContainerToDelete(null)}
            variant="destructive"
            confirmText={t('containers.deleteDialog.confirm')}
            cancelText={t('containers.deleteDialog.cancel')}
            icon={Trash2}
          />
          <DeleteDialog
            title={t('containers.pruneImagesDialog.title')}
            description={t('containers.pruneImagesDialog.description')}
            onConfirm={handlePruneImages}
            open={showPruneImagesConfirm}
            onOpenChange={setShowPruneImagesConfirm}
            variant="destructive"
            confirmText={t('containers.pruneImagesDialog.confirm')}
            cancelText={t('containers.pruneImagesDialog.cancel')}
            icon={Trash2}
          />
          <DeleteDialog
            title={t('containers.pruneBuildCacheDialog.title')}
            description={t('containers.pruneBuildCacheDialog.description')}
            onConfirm={handlePruneBuildCache}
            open={showPruneBuildCacheConfirm}
            onOpenChange={setShowPruneBuildCacheConfirm}
            variant="destructive"
            confirmText={t('containers.pruneBuildCacheDialog.confirm')}
            cancelText={t('containers.pruneBuildCacheDialog.cancel')}
            icon={Scissors}
          />
        </AnyPermissionGuard>
      </div>
    </ResourceGuard>
  );
}
