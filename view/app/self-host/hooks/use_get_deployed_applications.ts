import { SortOption } from '@/components/ui/sort-selector';
import { useSearchable } from '@/hooks/use-searchable';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import {
  useGetAllGithubConnectorQuery,
  useUpdateGithubConnectorMutation
} from '@/redux/services/connector/githubConnectorApi';
import { useGetApplicationsQuery } from '@/redux/services/deploy/applicationsApi';
import { Application } from '@/redux/types/applications';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { useEffect, useState } from 'react';
import { setActiveConnectorId } from '@/redux/features/github-connector/githubConnectorSlice';

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
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

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

  useEffect(() => {
    if (activeOrg) {
      GetApplications();
    }
  }, [activeOrg]);

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
  const dispatch = useAppDispatch();
  const [updateGithubConnector, { isLoading: isUpdatingConnector }] =
    useUpdateGithubConnectorMutation();
  const [inGitHubFlow, setInGitHubFlow] = useState(false);
  const [pendingConnectorId, setPendingConnectorId] = useState<string | null>(null);
  const activeConnectorId = useAppSelector(
    (state) => state.githubConnector.activeConnectorId
  );
  const code = searchParams.get('code');
  const installationId = searchParams.get('installation_id');
  const githubSetup = searchParams.get('github_setup');
  const connectorIdParam = searchParams.get('connector_id');
  const showApplications = paginatedApplications?.length > 0 || isLoadingApplications;

  useEffect(() => {
    if (code || githubSetup === 'true') {
      setInGitHubFlow(true);
    }
  }, [code, githubSetup]);

  // Track connector ID from URL parameter or find newly created connector
  useEffect(() => {
    if (connectorIdParam) {
      setPendingConnectorId(connectorIdParam);
    } else if (connectors && connectors.length > 0 && inGitHubFlow) {
      // Find the connector without installation_id (newly created)
      const newConnector = connectors.find(
        (c) => !c.installation_id || c.installation_id.trim() === ''
      );
      if (newConnector) {
        setPendingConnectorId(newConnector.id);
      }
    }
  }, [connectors, connectorIdParam, inGitHubFlow]);

  // Initialize active connector if not set and connectors are available
  useEffect(() => {
    if (!activeConnectorId && connectors && connectors.length > 0) {
      dispatch(setActiveConnectorId(connectors[0].id));
    }
  }, [activeConnectorId, connectors, dispatch]);

  // Clean up github_setup query parameter after setting the flow
  useEffect(() => {
    if (githubSetup === 'true' && inGitHubFlow) {
      // Remove the query parameter from URL without reloading
      const newUrl = new URL(window.location.href);
      newUrl.searchParams.delete('github_setup');
      router.replace(newUrl.pathname + newUrl.search, { scroll: false });
    }
  }, [githubSetup, inGitHubFlow, router]);

  useEffect(() => {
    if (installationId) {
      const githubConnector = async () => {
        try {
          // Determine which connector to update
          let connectorIdToUpdate = pendingConnectorId;
          
          // If no pending connector ID, try to find one without installation_id
          if (!connectorIdToUpdate && connectors && connectors.length > 0) {
            const newConnector = connectors.find(
              (c) => !c.installation_id || c.installation_id.trim() === ''
            );
            if (newConnector) {
              connectorIdToUpdate = newConnector.id;
            } else if (connectors.length === 1) {
              connectorIdToUpdate = connectors[0].id;
            } else {
              // Multiple connectors exist and we can't determine which one
              // Use active connector as fallback
              connectorIdToUpdate = activeConnectorId || connectors[0].id;
            }
          }

          // When multiple connectors exist, connector_id is required
          const updatePayload: { installation_id: string; connector_id?: string } = {
            installation_id: installationId
          };
          
          if (connectors && connectors.length > 1) {
            if (!connectorIdToUpdate) {
              throw new Error('connector_id is required when multiple connectors exist');
            }
            updatePayload.connector_id = connectorIdToUpdate;
          } else if (connectorIdToUpdate) {
            // Include connector_id even for single connector for clarity
            updatePayload.connector_id = connectorIdToUpdate;
          }

          await updateGithubConnector(updatePayload);
          await GetGithubConnectors();
          
          // Set the updated connector as active
          if (connectorIdToUpdate) {
            dispatch(setActiveConnectorId(connectorIdToUpdate));
          }
          
          setInGitHubFlow(false);
          setPendingConnectorId(null);
          // Clean up URL parameters
          const newUrl = new URL(window.location.href);
          newUrl.searchParams.delete('installation_id');
          newUrl.searchParams.delete('connector_id');
          router.replace(newUrl.pathname + newUrl.search, { scroll: false });
          router.push('/self-host');
        } catch (error) {
          console.error('Failed to update GitHub connector:', error);
          setInGitHubFlow(false);
          setPendingConnectorId(null);
        }
      };
      githubConnector();
    }
  }, [installationId, pendingConnectorId, router, GetGithubConnectors, updateGithubConnector, connectors, activeConnectorId, dispatch]);

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
