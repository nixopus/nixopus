import { SortOption } from '@/components/sort-selector';
import { useSearchable } from '@/hooks/use-searchable';
import {
  useGetAllGithubConnectorQuery,
  useUpdateGithubConnectorMutation
} from '@/redux/services/connector/githubConnectorApi';
import { useGetApplicationsQuery } from '@/redux/services/deploy/applicationsApi';
import { Application } from '@/redux/types/applications';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { useEffect, useState } from 'react';

/**
 * Hook to get the deployed applications.
 *
 * This hook fetches all the connectors and applications of the user.
 * It also paginates the applications and allows the user to search and sort them.
 *
 * @returns An object with the following properties:
 * - connectors: An array of all the github connectors of the user.
 * - GetGithubConnectors: A function to refetch the connectors.
 * - isLoading: A boolean indicating whether the connectors are being fetched.
 * - applications: An array of the paginated applications.
 * - GetApplications: A function to refetch the applications.
 * - isLoadingApplications: A boolean indicating whether the applications are being fetched.
 * - searchTerm: The search term used to filter the applications.
 * - handleSearchChange: A function to update the search term.
 * - onSortChange: A function to update the sort config.
 * - sortOptions: An array of sort options.
 * - sortConfig: The current sort config.
 * - handlePageChange: A function to update the current page.
 * - currentPage: The current page number.
 * - totalPages: The total number of pages.
 */
function useGetDeployedApplications() {
  const [limit, setLimit] = React.useState(10);
  const [currentPage, setCurrentPage] = useState(1);
  const {
    data: connectors,
    refetch: GetGithubConnectors,
    isLoading
  } = useGetAllGithubConnectorQuery();
  const {
    data: applications,
    refetch: GetApplications,
    isLoading: isLoadingApplications
  } = useGetApplicationsQuery({
    page: currentPage,
    limit
  });
  const {
    filteredAndSortedData: filteredAndSortedApplications,
    searchTerm,
    handleSearchChange,
    handleSortChange,
    sortConfig
  } = useSearchable<Application>(
    applications?.applications || [],
    ['name', 'domain', 'environment', 'updated_at', 'build_pack', 'port'],
    { key: 'name', direction: 'asc' }
  );
  const totalPages = applications?.total_count ? Math.ceil(applications.total_count / limit) : 1;
  const ITEMS_PER_PAGE = React.useMemo(() => 9, []);

  const paginatedApplications = React.useMemo(
    () =>
      filteredAndSortedApplications.slice(
        (currentPage - 1) * ITEMS_PER_PAGE,
        currentPage * ITEMS_PER_PAGE
      ),
    [currentPage, filteredAndSortedApplications]
  );

  const handlePageChange = (pageNumber: number) => {
    setCurrentPage(pageNumber);
  };

  // Reset the current page when the search term or sort config changes
  useEffect(() => {
    setCurrentPage(1);
  }, [searchTerm, sortConfig]);

  const onSortChange = (newSort: SortOption<Application>) => {
    handleSortChange(newSort.value as keyof Application);
  };

  const sortOptions: SortOption<Application>[] = React.useMemo(
    () => [
      { label: 'Name', value: 'name', direction: 'asc' },
      { label: 'Domain', value: 'domain', direction: 'asc' },
      { label: 'Environment', value: 'environment', direction: 'asc' },
      { label: 'Updated At', value: 'updated_at', direction: 'asc' },
      { label: 'Build Pack', value: 'build_pack', direction: 'asc' },
      { label: 'Port', value: 'port', direction: 'asc' }
    ],
    []
  );

  const router = useRouter();
  const searchParams = useSearchParams();
  const [updateGithubConnector, { isLoading: isUpdatingConnector }] =
    useUpdateGithubConnectorMutation();
  const [inGitHubFlow, setInGitHubFlow] = useState(false);
  const code = searchParams.get('code');
  const installationId = searchParams.get('installation_id');
  const showApplications = paginatedApplications?.length > 0 || isLoadingApplications;

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

  return {
    connectors,
    GetGithubConnectors,
    isLoading,
    applications: paginatedApplications,
    GetApplications,
    isLoadingApplications,
    searchTerm,
    handleSearchChange,
    onSortChange,
    sortOptions,
    sortConfig,
    handlePageChange,
    currentPage,
    totalPages,
    router,
    showApplications,
    inGitHubFlow
  };
}

export default useGetDeployedApplications;
