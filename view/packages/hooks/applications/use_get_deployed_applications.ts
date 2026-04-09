import { SortOption } from '@/components/ui/sort-selector';
import { useSearchable } from '@/packages/hooks/shared/use-searchable';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import {
  useGetAllGithubConnectorQuery,
  useUpdateGithubConnectorMutation,
  useCreateGithubConnectorMutation
} from '@/redux/services/connector/githubConnectorApi';
import { useGetApplicationsQuery } from '@/redux/services/deploy/applicationsApi';
import { Application } from '@/redux/types/applications';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';
import { setActiveConnectorId } from '@/redux/features/github-connector/githubConnectorSlice';
import { useLabelFilter } from './use_label_filter';
import { getSelfHosted } from '@/redux/conf';

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
    isLoading: isLoadingApplications,
    isFetching: isFetchingApplications
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
    ['name', 'domains', 'environment', 'updated_at', 'build_pack', 'port'],
    { key: 'name', direction: 'asc' }
  );

  const labelFilter = useLabelFilter(filteredAndSortedApplications);

  const totalPages = applications?.total_count ? Math.ceil(applications.total_count / limit) : 1;
  const ITEMS_PER_PAGE = React.useMemo(() => 9, []);
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  const paginatedApplications = React.useMemo(
    () =>
      labelFilter.filteredApplications.slice(
        (currentPage - 1) * ITEMS_PER_PAGE,
        currentPage * ITEMS_PER_PAGE
      ),
    [currentPage, labelFilter.filteredApplications, ITEMS_PER_PAGE]
  );

  const handlePageChange = (pageNumber: number) => {
    setCurrentPage(pageNumber);
  };

  // Reset the current page when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [searchTerm, sortConfig, labelFilter.selectedLabels]);

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
      { label: 'Domains', value: 'domains', direction: 'asc' },
      { label: 'Environment', value: 'environment', direction: 'asc' },
      { label: 'Updated At', value: 'updated_at', direction: 'asc' },
      { label: 'Build Pack', value: 'build_pack', direction: 'asc' },
      { label: 'Port', value: 'port', direction: 'asc' }
    ],
    []
  );

  const router = useRouter();
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  const [updateGithubConnector, { isLoading: isUpdatingConnector }] =
    useUpdateGithubConnectorMutation();
  const [createGithubConnector] = useCreateGithubConnectorMutation();
  const [inGitHubFlow, setInGitHubFlow] = useState(false);
  const [pendingConnectorId, setPendingConnectorId] = useState<string | null>(null);
  const processingInstallRef = useRef(false);
  const [selfHosted, setSelfHosted] = useState<boolean | null>(null);
  const [hasLoadedOnce, setHasLoadedOnce] = useState(false);

  useEffect(() => {
    getSelfHosted().then(setSelfHosted);
  }, []);

  useEffect(() => {
    if (!isLoadingApplications) {
      setHasLoadedOnce(true);
    }
  }, [isLoadingApplications]);

  const activeConnectorId = useAppSelector((state) => state.githubConnector.activeConnectorId);
  const code = searchParams.get('code');
  const installationId = searchParams.get('installation_id');
  const githubSetup = searchParams.get('github_setup');
  const connectorIdParam = searchParams.get('connector_id');
  const showApplications = hasLoadedOnce
    ? paginatedApplications?.length > 0
    : isLoadingApplications;

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
      const params = new URLSearchParams(searchParams.toString());
      params.delete('github_setup');
      const search = params.toString();
      router.replace(search ? `${pathname}?${search}` : pathname, { scroll: false });
    }
  }, [githubSetup, inGitHubFlow, router, pathname, searchParams]);

  useEffect(() => {
    if (!installationId || isLoading || processingInstallRef.current) return;
    processingInstallRef.current = true;

    const cleanupUrlParams = () => {
      const params = new URLSearchParams(searchParams.toString());
      params.delete('installation_id');
      params.delete('connector_id');
      const search = params.toString();
      router.replace(search ? `${pathname}?${search}` : pathname, { scroll: false });
    };

    const handleInstallationCallback = async () => {
      try {
        const hasConnectors = connectors && connectors.length > 0;

        if (!hasConnectors) {
          await createGithubConnector({ installation_id: installationId });
          await GetGithubConnectors();
        } else {
          let connectorIdToUpdate = pendingConnectorId;

          if (!connectorIdToUpdate) {
            const newConnector = connectors.find(
              (c) => !c.installation_id || c.installation_id.trim() === ''
            );
            if (newConnector) {
              connectorIdToUpdate = newConnector.id;
            } else if (connectors.length === 1) {
              connectorIdToUpdate = connectors[0].id;
            } else {
              connectorIdToUpdate = activeConnectorId || connectors[0].id;
            }
          }

          const updatePayload: { installation_id: string; connector_id?: string } = {
            installation_id: installationId
          };

          if (connectors.length > 1) {
            if (!connectorIdToUpdate) {
              throw new Error('connector_id is required when multiple connectors exist');
            }
            updatePayload.connector_id = connectorIdToUpdate;
          } else if (connectorIdToUpdate) {
            updatePayload.connector_id = connectorIdToUpdate;
          }

          await updateGithubConnector(updatePayload);
          await GetGithubConnectors();

          if (connectorIdToUpdate) {
            dispatch(setActiveConnectorId(connectorIdToUpdate));
          }
        }

        setInGitHubFlow(false);
        setPendingConnectorId(null);
        cleanupUrlParams();
        router.push('/apps');
      } catch (error) {
        console.error('Failed to handle GitHub installation callback:', error);
        setInGitHubFlow(false);
        setPendingConnectorId(null);
        cleanupUrlParams();
      } finally {
        processingInstallRef.current = false;
      }
    };

    handleInstallationCallback();
  }, [
    installationId,
    isLoading,
    pendingConnectorId,
    router,
    pathname,
    searchParams,
    GetGithubConnectors,
    updateGithubConnector,
    createGithubConnector,
    connectors,
    activeConnectorId,
    dispatch
  ]);

  return {
    connectors,
    GetGithubConnectors,
    isLoading,
    applications: paginatedApplications,
    GetApplications,
    isLoadingApplications,
    isFetchingApplications,
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
    inGitHubFlow,
    labelFilter,
    selfHosted
  };
}

export default useGetDeployedApplications;
