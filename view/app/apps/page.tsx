'use client';
import React from 'react';
import GitHubAppSetup, { ManagedGitHubAppSetup } from '@/packages/components/github-connector';
import { ListRepositories } from '@/packages/components/github-repositories';
import AppItem, { AppItemSkeleton } from '../../packages/components/application';
import useGetDeployedApplications from '../../packages/hooks/applications/use_get_deployed_applications';
import { PaginationWrapper } from '@nixopus/ui';
import { SearchBar } from '@nixopus/ui';
import { SortSelect } from '@/components/ui/sort-selector';
import { Application } from '@/redux/types/applications';
import { Button } from '@nixopus/ui';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';
import { TypographyH2, TypographyMuted } from '@nixopus/ui';
import { Plus } from 'lucide-react';
import { LabelFilter } from '@/components/ui/label-filter';
import { MainPageHeader } from '@nixopus/ui';
import { Skeleton } from '@nixopus/ui';
import { SSHBanner } from '@/packages/components/dashboard';

function page() {
  const { t } = useTranslation();
  const {
    connectors,
    GetGithubConnectors,
    isLoading,
    applications,
    isLoadingApplications,
    searchTerm,
    handleSearchChange,
    onSortChange,
    sortOptions,
    sortConfig,
    handlePageChange,
    currentPage,
    totalPages,
    inGitHubFlow,
    showApplications,
    router,
    labelFilter,
    selfHosted
  } = useGetDeployedApplications();

  if (selfHosted === null || isLoadingApplications || isLoading) {
    return (
      <PageLayout maxWidth="7xl" padding="md" spacing="lg">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-9 w-32 rounded-md" />
        </div>
        <div className="flex items-center justify-between gap-4">
          <Skeleton className="h-10 flex-1 max-w-sm rounded-md" />
          <Skeleton className="h-10 w-36 rounded-md" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-4 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <AppItemSkeleton key={i} />
          ))}
        </div>
      </PageLayout>
    );
  }

  const isManagedMode = !selfHosted;

  const isShowingGitHubSetup = isManagedMode
    ? inGitHubFlow || (!showApplications && !connectors?.length)
    : inGitHubFlow || (!showApplications && !inGitHubFlow && !connectors?.length);
  const isShowingRepositories =
    !showApplications && !inGitHubFlow && connectors?.length && connectors.length > 0;

  const renderContent = () => {
    return (
      <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
        {isManagedMode ? (
          <>
            {(inGitHubFlow || (!showApplications && !connectors?.length)) && (
              <ManagedGitHubAppSetup />
            )}
            {!inGitHubFlow && !showApplications && connectors?.length && connectors.length > 0 && (
              <ListRepositories />
            )}
          </>
        ) : (
          <>
            {inGitHubFlow && <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />}
            {!showApplications && !inGitHubFlow && (
              <>
                {!connectors?.length ? (
                  <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />
                ) : (
                  <ListRepositories />
                )}
              </>
            )}
          </>
        )}
      </AnyPermissionGuard>
    );
  };

  return (
    <ResourceGuard
      resource="deploy"
      action="read"
      loadingFallback={<Skeleton />}
      fallback={
        <div className="flex h-full items-center justify-center">
          <div className="text-center">
            <TypographyH2>{t('selfHost.page.accessDenied.title')}</TypographyH2>
            <TypographyMuted>{t('selfHost.page.accessDenied.description')}</TypographyMuted>
          </div>
        </div>
      }
    >
      <PageLayout maxWidth="7xl" padding="md" spacing="lg">
        {!isShowingGitHubSetup && (
          <MainPageHeader
            label={
              isShowingRepositories ? t('selfHost.repositories.title') : t('selfHost.page.title')
            }
            highlightLabel={false}
            actions={
              showApplications ? (
                <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
                  <Button onClick={() => router.push('/apps/create')} className="gap-2">
                    <Plus className="h-4 w-4" />
                    {t('selfHost.page.createButton')}
                  </Button>
                </AnyPermissionGuard>
              ) : undefined
            }
          />
        )}

        <SSHBanner />

        {renderContent()}

        {showApplications && !inGitHubFlow && (
          <>
            <div className="flex items-center justify-between gap-4 flex-wrap mb-4">
              <div className="flex-1 min-w-[220px]">
                <SearchBar
                  searchTerm={searchTerm}
                  handleSearchChange={handleSearchChange}
                  label={t('selfHost.page.search.placeholder')}
                />
              </div>
              <SortSelect<Application>
                options={sortOptions}
                currentSort={{
                  value: sortConfig.key,
                  direction: sortConfig.direction,
                  label:
                    sortOptions.find(
                      (option) =>
                        option.value === sortConfig.key && option.direction === sortConfig.direction
                    )?.label || ''
                }}
                onSortChange={onSortChange}
                placeholder="Sort by"
              />
            </div>

            {labelFilter.availableLabels.length > 0 && (
              <div className="mb-6">
                <LabelFilter
                  availableLabels={labelFilter.availableLabels}
                  selectedLabels={labelFilter.selectedLabels}
                  onToggle={labelFilter.toggleLabel}
                  onClear={labelFilter.clearFilters}
                />
              </div>
            )}

            {isLoading || isLoadingApplications ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-4 gap-4">
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-4 gap-4">
                  {applications &&
                    applications.map((app: any) => <AppItem key={app.id} {...app} />)}
                </div>

                {totalPages > 1 && (
                  <div className="mt-8 flex justify-center">
                    <PaginationWrapper
                      currentPage={currentPage}
                      totalPages={totalPages}
                      onPageChange={handlePageChange}
                    />
                  </div>
                )}
              </>
            )}
          </>
        )}
      </PageLayout>
    </ResourceGuard>
  );
}

export default page;
