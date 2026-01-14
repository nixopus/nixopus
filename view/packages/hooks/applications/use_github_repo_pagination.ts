'use client';
import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useGetAllGithubRepositoriesQuery } from '@/redux/services/connector/githubConnectorApi';
import { GithubRepository } from '@/redux/types/github';
import { SortOption } from '@/components/ui/sort-selector';
import { useAppSelector } from '@/redux/hooks';
import { useDebounce } from '@/packages/hooks/shared/use-debounce';

/**
 * @function useGithubRepoPagination
 * @description A hook for getting all the github repositories and connectors of the user.
 * It fetches the connectors and repositories and filters them using the search term and sort config.
 * It also paginates the filtered and sorted data.
 * @returns An object with the following properties:
 * - connectors: An array of all the github connectors of the user.
 * - GetGithubConnectors: A function to refetch the connectors.
 * - githubRepositories: An array of all the github repositories of the user.
 * - selectedRepository: The selected repository or null if no repository is selected.
 * - setSelectedRepository: A function to set the selected repository.
 * - filteredAndSortedApplications: An array of the filtered and sorted repositories.
 * - searchTerm: The search term used to filter the repositories.
 * - handleSearchChange: A function to update the search term.
 * - onSortChange: A function to update the sort config.
 * - sortOptions: An array of sort options.
 * - sortConfig: The current sort config.
 * - handlePageChange: A function to update the current page.
 * - currentPage: The current page number.
 * - totalPages: The total number of pages.
 * - paginatedApplications: An array of the paginated filtered and sorted repositories.
 */
function useGithubRepoPagination() {
  const [currentPage, setCurrentPage] = useState(1);
  const PAGE_SIZE = 10;
  const router = useRouter();
  const [selectedRepository, setSelectedRepository] = React.useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{
    key: keyof GithubRepository;
    direction: 'asc' | 'desc';
  }>({
    key: 'name',
    direction: 'asc'
  });

  // Debounce search term to avoid excessive API calls
  const debouncedSearchTerm = useDebounce(searchTerm, 300);

  // Get active connector ID from Redux
  const activeConnectorId = useAppSelector((state) => state.githubConnector.activeConnectorId);

  // Pass search term to API for server-side filtering
  const { data, isLoading, isFetching } = useGetAllGithubRepositoriesQuery({
    page: currentPage,
    page_size: PAGE_SIZE,
    connector_id: activeConnectorId || undefined,
    search: debouncedSearchTerm || undefined
  });
  const isSearching = isFetching;

  // Server side pagination and search, the API already returns filtered and paginated results
  const paginatedApplications = data?.repositories || [];
  const filteredAndSortedApplications = paginatedApplications;

  const totalPages = React.useMemo(
    () => Math.max(1, Math.ceil((data?.total_count || 0) / PAGE_SIZE)),
    [data?.total_count]
  );

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  };

  const handleSortChange = (key: keyof GithubRepository) => {
    setSortConfig((prev) => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  const handlePageChange = (pageNumber: number) => {
    setCurrentPage(pageNumber);
  };

  // Reset the current page when the search term or connector changes
  // Note: sortConfig is excluded since sorting is not yet implemented
  useEffect(() => {
    setCurrentPage(1);
  }, [debouncedSearchTerm, activeConnectorId]);

  const onSortChange = (newSort: SortOption<GithubRepository>) => {
    handleSortChange(newSort.value as keyof GithubRepository);
  };

  const sortOptions: SortOption<GithubRepository>[] = React.useMemo(
    () => [
      { value: 'name', label: 'Name', direction: 'asc' },
      { value: 'description', label: 'Description', direction: 'asc' },
      { value: 'stargazers_count', label: 'Stars', direction: 'desc' }
    ],
    []
  );

  const onSelectRepository = (repository: string) => {
    setSelectedRepository(repository);
    router.push(`/self-host/create/${repository}`);
  };

  return {
    githubRepositories: data?.repositories,
    selectedRepository,
    setSelectedRepository,
    filteredAndSortedApplications,
    searchTerm,
    handleSearchChange,
    onSortChange,
    sortOptions,
    sortConfig,
    handlePageChange,
    currentPage,
    totalPages,
    paginatedApplications,
    isLoading,
    isFetching,
    isSearching,
    onSelectRepository
  };
}

export default useGithubRepoPagination;
