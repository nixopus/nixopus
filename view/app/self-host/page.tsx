'use client';
import React from 'react';
import GitHubAppSetup from './components/github-connector/github-app-setup';
import ListRepositories from './components/github-repositories/list-repositories';
import AppItem, { AppItemSkeleton } from './components/application';
import useGetDeployedApplications from './hooks/use_get_deployed_applications';
import PaginationWrapper from '@/components/pagination';
import { DahboardUtilityHeader } from '@/components/dashboard-page-header';
import { Application } from '@/redux/types/applications';
import { Button } from '@/components/ui/button';

function page() {
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

  const renderContent = () => {
    if (inGitHubFlow) {
      return <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />;
    }

    if (!showApplications) {
      if (!connectors?.length) {
        return <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />;
      } else {
        return <ListRepositories />;
      }
    } else {
      return (
        <>
          <DahboardUtilityHeader<Application>
            searchTerm={searchTerm}
            handleSearchChange={handleSearchChange}
            sortConfig={sortConfig}
            onSortChange={onSortChange}
            sortOptions={sortOptions}
            label="Applications"
            className="mt-5 mb-5 justify-between items-center"
            children={
              <Button
                className="mb-4 w-max flex justify-self-end mt-4"
                onClick={() => {
                  router.push('/self-host/create');
                }}
              >
                Create
              </Button>
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
    }
  };

  return <div className="container mx-auto py-6 space-y-8 max-w-4xl">{renderContent()}</div>;
}

export default page;
