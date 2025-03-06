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
    GetApplications,
    isLoadingApplications,
    searchTerm,
    handleSearchChange,
    onSortChange,
    sortOptions,
    sortConfig,
    handlePageChange,
    currentPage,
    totalPages
  } = useGetDeployedApplications();

  const showApplications = applications?.length > 0 || isLoadingApplications;

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      {!showApplications ? (
        connectors?.length === 0 ? (
          <GitHubAppSetup GetGithubConnectors={GetGithubConnectors} />
        ) : (
          <ListRepositories />
        )
      ) : (
        <>
          <DahboardUtilityHeader<Application>
            searchTerm={searchTerm}
            handleSearchChange={handleSearchChange}
            sortConfig={sortConfig}
            onSortChange={onSortChange}
            sortOptions={sortOptions}
            label="Deployed Applications"
            className="mt-5 mb-5"
          />
          <Button className="mb-4 w-max flex justify-self-end mt-4" onClick={() => { }}>Create</Button>
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

              {
                totalPages > 1 && (
                  <div className="mt-8 flex justify-center">
                    <PaginationWrapper
                      currentPage={currentPage}
                      totalPages={totalPages}
                      onPageChange={handlePageChange}
                    />
                  </div>
                )
              }
            </>
          )}
        </>
      )}
    </div>
  );
}

export default page;
