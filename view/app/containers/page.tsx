'use client';

import React from 'react';
import { Trash2, Scissors, LayoutGrid, List, Box } from 'lucide-react';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FeatureNames } from '@/packages/types/feature-flags';
import { Skeleton } from '@/components/ui/skeleton';
import { useContainers } from '@/packages/hooks/containers/use-containers';
import { useViewMode } from '@/packages/hooks/containers/use-view-mode';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';
import PaginationWrapper from '@/components/ui/pagination';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { cn } from '@/lib/utils';
import MainPageHeader from '@/components/ui/main-page-header';
import { SearchBar } from '@/components/ui/search-bar';
import ContainersLoading from '@/packages/components/container-skeleton';
import ContainersTable from '@/packages/components/container-table';
import { ContainerCard, Action } from '@/packages/components/container-card';
import { StatPill } from '@/packages/components/container-stat-pill';
import { ActionHeader } from '@/packages/components/container-action-header';
import { translationKey } from '@/hooks/use-translation';
import DisabledFeature from '@/packages/components/rbac';

export default function ContainersPage() {
  const { viewMode, setViewMode } = useViewMode();

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
    setShowPruneImagesConfirm,
    setShowPruneBuildCacheConfirm,
    page,
    setPage,
    totalPages,
    totalCount,
    runningCount,
    stoppedCount,
    pageSize,
    setPageSize,
    searchInput,
    handleSearchChange,
    sortConfig,
    handleSortChange,
    sortOptions
  } = useContainers();

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
      <PageLayout maxWidth="full" padding="md" spacing="lg" className="relative z-10">
        <MainPageHeader
          label={t('containers.title')}
          description={t('containers.description')}
          actions={
            <ActionHeader
              handleRefresh={handleRefresh}
              isRefreshing={isRefreshing}
              isFetching={isFetching}
              t={t}
              setShowPruneImagesConfirm={setShowPruneImagesConfirm}
              setShowPruneBuildCacheConfirm={setShowPruneBuildCacheConfirm}
            />
          }
        />

        {totalCount > 0 && (
          <div className="flex items-center gap-6 mb-6">
            <StatPill value={totalCount} label="Total" />
            <StatPill value={runningCount} label="Running" color="emerald" />
            <StatPill value={stoppedCount} label="Stopped" color="zinc" />
          </div>
        )}

        <div className="flex items-center gap-3 mb-6">
          <div className="relative flex-1 max-w-md">
            <SearchBar
              searchTerm={searchInput}
              handleSearchChange={(e) => handleSearchChange(e.target.value)}
              label={t('containers.searchPlaceholder')}
            />
          </div>
          <div className="flex items-center gap-2 ml-auto">
            <SelectWrapper
              value={sortConfig ? `${sortConfig.key}_${sortConfig.direction}` : 'name_asc'}
              onValueChange={(value) => {
                const [key, direction] = value.split('_') as ['name' | 'status', 'asc' | 'desc'];
                handleSortChange(key, direction);
              }}
              options={sortOptions}
              placeholder={t('containers.sortBy')}
              className="w-full sm:w-[180px]"
            />
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
                onClick={() => setViewMode('table')}
                className={cn(
                  'p-2 rounded-md transition-colors',
                  viewMode === 'table' ? 'bg-muted' : 'hover:bg-muted/50'
                )}
              >
                <List className="h-4 w-4" />
              </button>
              <button
                onClick={() => setViewMode('card')}
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

        <div className="space-y-6">
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
                  onAction={handleContainerAction}
                />
              ))}
            </div>
          ) : (
            <ContainersTable
              containersData={containers}
              sortBy={sortConfig?.key || 'name'}
              sortOrder={sortConfig?.direction || 'asc'}
              onSort={(field) => {
                const currentKey = sortConfig?.key || 'name';
                const currentDir = sortConfig?.direction || 'asc';
                if (currentKey === field) {
                  handleSortChange(field, currentDir === 'asc' ? 'desc' : 'asc');
                } else {
                  handleSortChange(field, 'asc');
                }
              }}
              onAction={handleContainerAction}
            />
          )}

          {totalPages > 1 && (
            <div className="flex justify-center pt-6">
              <PaginationWrapper
                currentPage={page}
                totalPages={totalPages}
                onPageChange={setPage}
              />
            </div>
          )}
        </div>

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
