'use client';

import React, { useEffect, useState } from 'react';
import GitHubAppSetup from './components/github-connector/github-app-setup';
import ListRepositories from './components/github-repositories/list-repositories';
import AppItem, { AppItemSkeleton } from './components/application';
import useGetDeployedApplications from './hooks/use_get_deployed_applications';
import PaginationWrapper from '@/components/pagination';
import { DahboardUtilityHeader } from '@/components/dashboard-page-header';
import { Application } from '@/redux/types/applications';
import { Button } from '@/components/ui/button';
import { useRouter, useSearchParams } from 'next/navigation';
import { useUpdateGithubConnectorMutation } from '@/redux/services/connector/githubConnectorApi';
import { Alert, AlertDescription } from '@/components/ui/alert';

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

  const router = useRouter();
  const searchParams = useSearchParams();
  const [updateGithubConnector, { isLoading: isUpdatingConnector }] =
    useUpdateGithubConnectorMutation();
  const [installationSuccess, setInstallationSuccess] = useState(false);
  const [inGitHubFlow, setInGitHubFlow] = useState(false);
  const code = searchParams.get('code');
  const installationId = searchParams.get('installation_id');
  const showApplications = applications?.length > 0 || isLoadingApplications;

  useEffect(() => {
    if (code) {
      setInGitHubFlow(true);
    }
  }, [code]);

  useEffect(() => {
    if (installationId) {
      const githubConnector = async () => {
        try {
          await updateGithubConnector({
            installation_id: installationId
          });
          setInstallationSuccess(true);
          await GetGithubConnectors();
          setInGitHubFlow(false);
          router.push('/self-host/create');
        } catch (error) {
          console.error('Failed to update GitHub connector:', error);
          setInGitHubFlow(false);
        }
      };
      githubConnector();
    }
  }, [installationId, router, GetGithubConnectors, updateGithubConnector]);

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
            label="Deployed Applications"
            className="mt-5 mb-5"
          />
          <Button
            className="mb-4 w-max flex justify-self-end mt-4"
            onClick={() => {
              router.push('/self-host/create');
            }}
          >
            Create
          </Button>
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
