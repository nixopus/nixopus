'use client';

import React from 'react';
import { RefreshCw, Trash2, Loader2, Scissors, LayoutGrid, List, Box, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import ContainersLoading from './components/skeleton';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FeatureNames } from '@/types/feature-flags';
import { Skeleton } from '@/components/ui/skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import useContainerList from './hooks/use-container-list';
import PageLayout from '@/components/layout/page-layout';
import ContainersTable from './components/table';
import PaginationWrapper from '@/components/ui/pagination';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { ContainerCard } from './components/card';
import { cn } from '@/lib/utils';

export default function ContainersPage() {
  const [viewMode, setViewMode] = React.useState<'table' | 'card'>(() => {
    if (typeof window !== 'undefined') {
      const existing = window.localStorage.getItem('containers_view');
      return (existing as 'table' | 'card') || 'table';
    }
    return 'table';
  });

  const {
    containers,
    isLoading,
    isFetching,
    initialized,
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
    setShowPruneBuildCacheConfirm,
    page,
    setPage,
    totalPages,
    totalCount,
    pageSize,
    setPageSize,
    searchInput,
    setSearchInput,
    sortBy,
    sortOrder,
    handleSort
  } = useContainerList();

  if (!initialized && isLoading) {
    return <ContainersLoading />;
  }

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureContainer)) {
    return <DisabledFeature />;
  }

  const runningCount = containers.filter((c) => c.status === 'running').length;
  const stoppedCount = containers.filter((c) => c.status !== 'running').length;

  return (
    <ResourceGuard resource="container" action="read" loadingFallback={<ContainersLoading />}>
      <PageLayout maxWidth="full" padding="md" spacing="lg" className="relative z-10">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{t('containers.title')}</h1>
            <p className="text-muted-foreground mt-1">{t('containers.description')}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              onClick={handleRefresh}
              variant="outline"
              size="sm"
              disabled={isRefreshing || isFetching}
            >
              {isRefreshing || isFetching ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="mr-2 h-4 w-4" />
              )}
              {t('containers.refresh')}
            </Button>
            <AnyPermissionGuard
              permissions={['container:delete']}
              loadingFallback={<Skeleton className="h-9 w-20" />}
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

        {totalCount > 0 && (
          <div className="flex items-center gap-6 mb-6">
            <StatPill value={totalCount} label="Total" />
            <StatPill value={runningCount} label="Running" color="emerald" />
            <StatPill value={stoppedCount} label="Stopped" color="zinc" />
          </div>
        )}

        <div className="flex items-center gap-3 mb-6">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search containers..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex items-center gap-2 ml-auto">
            <SelectWrapper
              value={String(pageSize)}
              onValueChange={(v) => {
                const num = parseInt(v, 10);
                setPageSize(num);
                setPage(1);
              }}
              options={[
                { value: '10', label: '10 per page' },
                { value: '20', label: '20 per page' },
                { value: '50', label: '50 per page' }
              ]}
              placeholder="Page size"
              className="w-[130px]"
            />
            <div className="hidden sm:flex items-center border rounded-lg p-0.5">
              <button
                onClick={() => {
                  setViewMode('table');
                  if (typeof window !== 'undefined')
                    window.localStorage.setItem('containers_view', 'table');
                }}
                className={cn(
                  'p-2 rounded-md transition-colors',
                  viewMode === 'table' ? 'bg-muted' : 'hover:bg-muted/50'
                )}
              >
                <List className="h-4 w-4" />
              </button>
              <button
                onClick={() => {
                  setViewMode('card');
                  if (typeof window !== 'undefined')
                    window.localStorage.setItem('containers_view', 'card');
                }}
                className={cn(
                  'p-2 rounded-md transition-colors',
                  viewMode === 'card' ? 'bg-muted' : 'hover:bg-muted/50'
                )}
              >
                <LayoutGrid className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>

        {containers.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
            <Box className="h-16 w-16 mb-4 opacity-20" />
            <p className="text-lg font-medium">{t('containers.no_containers')}</p>
            <p className="text-sm mt-1">No containers match your search criteria</p>
          </div>
        ) : viewMode === 'card' ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
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
        ) : (
          <ContainersTable
            containersData={containers}
            sortBy={sortBy}
            sortOrder={sortOrder}
            onSort={handleSort}
            onAction={handleContainerAction}
          />
        )}

        {totalPages > 1 && (
          <div className="mt-6 flex items-center justify-center">
            <PaginationWrapper currentPage={page} totalPages={totalPages} onPageChange={setPage} />
          </div>
        )}

        <AnyPermissionGuard permissions={['container:delete']} loadingFallback={null}>
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

function StatPill({
  value,
  label,
  color
}: {
  value: number;
  label: string;
  color?: 'emerald' | 'zinc';
}) {
  return (
    <div className="flex items-center gap-2">
      {color && (
        <span
          className={cn(
            'w-2 h-2 rounded-full',
            color === 'emerald' ? 'bg-emerald-500' : 'bg-zinc-500'
          )}
        />
      )}
      <span className="text-xl font-bold">{value}</span>
      <span className="text-sm text-muted-foreground">{label}</span>
    </div>
  );
}
