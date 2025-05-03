'use client';
import React from 'react';
import GitHubAppSetup from './components/github-connector/github-app-setup';
import ListRepositories from './components/github-repositories/list-repositories';
import AppItem, { AppItemSkeleton } from './components/application';
import useGetDeployedApplications from './hooks/use_get_deployed_applications';
import PaginationWrapper from '@/components/ui/pagination';
import { DahboardUtilityHeader } from '@/components/layout/dashboard-page-header';
import { Application } from '@/redux/types/applications';
import { Button } from '@/components/ui/button';
import { useAppSelector } from '@/redux/hooks';
import { hasPermission } from '@/lib/permission';
import { useTranslation } from '@/hooks/use-translation';
import { useFeatureFlags } from '@/hooks/features_provider';
import Skeleton from '../file-manager/components/skeleton/Skeleton';
import { FeatureNames } from '@/types/feature-flags';
import DisabledFeature from '@/components/features/disabled-feature';

function page() {
  const { t } = useTranslation();
  const user = useAppSelector((state) => state.auth.user);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
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

  const canRead = hasPermission(user, 'deploy', 'read', activeOrg?.id);
  const canCreate = hasPermission(user, 'deploy', 'create', activeOrg?.id);

  if (!canRead) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">{t('selfHost.page.accessDenied.title')}</h2>
          <p className="text-muted-foreground">{t('selfHost.page.accessDenied.description')}</p>
        </div>
      </div>
    );
  }

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureSelfHosted)) {
    return (
      <DisabledFeature />
    );
  }

  const renderContent = () => {
    if (canCreate) {
      if (inGitHubFlow) {
        return <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />;
      }

      if (!showApplications) {
        if (!connectors?.length) {
          return <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />;
        } else {
          return <ListRepositories />;
        }
      }
    }

    return (
      <>
        <DahboardUtilityHeader<Application>
          searchTerm={searchTerm}
          handleSearchChange={handleSearchChange}
          sortConfig={sortConfig}
          onSortChange={onSortChange}
          sortOptions={sortOptions}
          label={t('selfHost.page.title')}
          className="mt-5 mb-5 justify-between items-center"
          children={
            canCreate && (
              <Button
                className="mb-4 w-max flex justify-self-end mt-4"
                onClick={() => {
                  router.push('/self-host/create');
                }}
              >
                {t('selfHost.page.createButton')}
              </Button>
            )
          }
        />
        {isLoading || isLoadingApplications ? (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <AppItemSkeleton />
            <AppItemSkeleton />
            <AppItemSkeleton />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              {applications && applications.map((app: any) => <AppItem key={app.id} {...app} />)}
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
    );
  };

  return <div className="container mx-auto py-6 space-y-8 max-w-4xl">{renderContent()}</div>;
}

export default page;
