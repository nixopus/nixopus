'use client';
import React from 'react';
import GitHubAppSetup from '@/packages/components/github-connector';
import { ListRepositories } from '@/packages/components/github-repositories';
import AppItem, { AppItemSkeleton } from '../../packages/components/application';
import useGetDeployedApplications from '../../packages/hooks/applications/use_get_deployed_applications';
import PaginationWrapper from '@/components/ui/pagination';
import { SearchBar } from '@/components/ui/search-bar';
import { SortSelect } from '@/components/ui/sort-selector';
import { Application } from '@/redux/types/applications';
import { Button } from '@/components/ui/button';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useFeatureFlags } from '@/packages/hooks/shared/features_provider';
import Skeleton from '../../packages/deprecated/file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/packages/types/feature-flags';
import DisabledFeature from '@/packages/components/rbac';
import { ResourceGuard, AnyPermissionGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';
import { TypographyH2, TypographyMuted } from '@/components/ui/typography';
import { Plus } from 'lucide-react';
import { LabelFilter } from '@/components/ui/label-filter';
import MainPageHeader from '@/components/ui/main-page-header';

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
    labelFilter
  } = useGetDeployedApplications();
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureSelfHosted)) {
    return <DisabledFeature />;
  }

  const isShowingGitHubSetup =
    inGitHubFlow || (!showApplications && !inGitHubFlow && !connectors?.length);
  const isShowingRepositories =
    !showApplications && !inGitHubFlow && connectors?.length && connectors.length > 0;

  const renderContent = () => {
    return (
      <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
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
      <PageLayout maxWidth="full" padding="md" spacing="lg">
        {!isShowingGitHubSetup && (
          <MainPageHeader
            label={
              isShowingRepositories ? t('selfHost.repositories.title') : t('selfHost.page.title')
            }
            description={
              isShowingRepositories
                ? t('selfHost.repositories.search.placeholder')
                : t('selfHost.page.description')
            }
            actions={
              showApplications ? (
                <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
                  <Button onClick={() => router.push('/self-host/create')} className="gap-2">
                    <Plus className="h-4 w-4" />
                    {t('selfHost.page.createButton')}
                  </Button>
                </AnyPermissionGuard>
              ) : undefined
            }
          />
        )}

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
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
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
