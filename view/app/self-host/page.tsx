'use client';
import React from 'react';
import GitHubAppSetup from './components/github-connector/github-app-setup';
import ListRepositories from './components/github-repositories/list-repositories';
import AppItem, { AppItemSkeleton } from './components/application';
import useGetDeployedApplications from './hooks/use_get_deployed_applications';
import PaginationWrapper from '@/components/ui/pagination';
import { SearchBar } from '@/components/ui/search-bar';
import { SortSelect } from '@/components/ui/sort-selector';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import { Application } from '@/redux/types/applications';
import { Button } from '@/components/ui/button';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import Skeleton from '../file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/types/feature-flags';
import DisabledFeature from '@/components/features/disabled-feature';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import PageLayout from '@/components/layout/page-layout';
import { TypographyH1, TypographyH2, TypographyMuted } from '@/components/ui/typography';

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
    router
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
      <PageLayout maxWidth="6xl" padding="md" spacing="lg">
        {!isShowingGitHubSetup && (
          <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
            <span>
              <TypographyH1>
                {isShowingRepositories
                  ? t('selfHost.repositories.title')
                  : t('selfHost.page.title')}
              </TypographyH1>
              <TypographyMuted>
                {isShowingRepositories
                  ? t('selfHost.repositories.search.placeholder')
                  : t('selfHost.page.description')}
              </TypographyMuted>
            </span>
            {showApplications && (
              <AnyPermissionGuard permissions={['deploy:create']} loadingFallback={null}>
                <Button
                  onClick={() => {
                    router.push('/self-host/create');
                  }}
                >
                  {t('selfHost.page.createButton')}
                </Button>
              </AnyPermissionGuard>
            )}
          </div>
        )}

        {renderContent()}

        {showApplications && (
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

            {isLoading || isLoadingApplications ? (
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <AppItemSkeleton />
                <AppItemSkeleton />
                <AppItemSkeleton />
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
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
