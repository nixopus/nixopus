'use client';

import { RefreshCw, Trash2, Loader2, Scissors } from 'lucide-react';
import { Button } from '@/components/ui/button';
import ContainersLoading from './skeleton';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FeatureNames } from '@/types/feature-flags';
import { Skeleton } from '@/components/ui/skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import useContainerList from './hooks/use-container-list';
import { TypographyH1, TypographyH2, TypographyMuted } from '@/components/ui/typography';
import PageLayout from '@/components/layout/page-layout';
// import { ContainerCard } from './components/card';
import ContainersTable from './components/table';

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
      <PageLayout maxWidth="6xl" padding="md" spacing="lg" className="relative z-10">
        <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
          <span>
            <TypographyH1 className="text-2xl font-bold">{t('containers.title')}</TypographyH1>
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
        {/* {containers.length > 0 && (
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
        )} */}
        <ContainersTable containersData={containers} />
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
      </PageLayout>
    </ResourceGuard>
  );
}
