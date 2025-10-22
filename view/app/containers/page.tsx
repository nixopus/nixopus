'use client';

import React from 'react';
import { RefreshCw, Trash2, Loader2, Scissors, Grid, List } from 'lucide-react';
import { Button } from '@/components/ui/button';
import ContainersLoading from './components/skeleton';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FeatureNames } from '@/types/feature-flags';
import { Skeleton } from '@/components/ui/skeleton';
import DisabledFeature from '@/components/features/disabled-feature';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import useContainerList from './hooks/use-container-list';
import { TypographyH1, TypographyH2, TypographyMuted } from '@/components/ui/typography';
import PageLayout from '@/components/layout/page-layout';
import ContainersTable from './components/table';
import PaginationWrapper from '@/components/ui/pagination';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';
import { SearchBar } from '@/components/ui/search-bar';
import { ContainerCard } from './components/card';

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
    search,
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

  return (
    <ResourceGuard resource="container" action="read" loadingFallback={<ContainersLoading />}>
      <PageLayout maxWidth="6xl" padding="md" spacing="lg" className="relative z-10">
        <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
          <span>
            <TypographyH1 className="text-2xl font-bold">{t('containers.title')}</TypographyH1>
          </span>
          <div className="flex items-center gap-2 flex-wrap">
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

        <div className="flex items-center justify-between gap-4 flex-wrap mb-4">
          <div className="flex-1 min-w-[220px]">
            <SearchBar
              searchTerm={searchInput}
              handleSearchChange={(e) => setSearchInput(e.target.value)}
              label={t('common.searchFiles')}
            />
          </div>
          <div className="flex items-center gap-2">
            <SelectWrapper
              value={String(pageSize)}
              onValueChange={(v) => {
                const num = parseInt(v, 10);
                setPageSize(num);
                setPage(1);
              }}
              options={[
                { value: '10', label: '10' },
                { value: '20', label: '20' },
                { value: '50', label: '50' },
                { value: '100', label: '100' }
              ]}
              placeholder="Page size"
              className="w-[110px]"
            />
            <div className="hidden sm:flex items-center gap-2 ml-2">
              <Button
                variant="outline"
                size="icon"
                onClick={() => {
                  const next = viewMode === 'table' ? 'card' : 'table';
                  setViewMode(next);
                  if (typeof window !== 'undefined')
                    window.localStorage.setItem('containers_view', next);
                }}
              >
                {viewMode === 'table' ? <Grid className="h-4 w-4" /> : <List className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </div>
        {containers.length === 0 && (
          <div className="flex justify-center items-center h-full">
            <TypographyH2 className="text-muted-foreground">
              {t('containers.no_containers')}
            </TypographyH2>
          </div>
        )}
        {viewMode === 'card' ? (
          <>
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
          </>
        ) : (
          <ContainersTable
            containersData={containers}
            sortBy={sortBy}
            sortOrder={sortOrder}
            onSort={handleSort}
            onAction={handleContainerAction}
          />
        )}

        {totalCount > 0 && (
          <div className="mt-4 flex items-center justify-between flex-wrap gap-2">
            <TypographyMuted>{totalCount} containers</TypographyMuted>
            {totalPages > 1 && (
              <PaginationWrapper
                currentPage={page}
                totalPages={totalPages}
                onPageChange={setPage}
              />
            )}
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
