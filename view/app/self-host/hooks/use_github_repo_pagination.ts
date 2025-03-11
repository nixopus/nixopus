'use client';
import React, { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  useGetAllGithubRepositoriesQuery,
  useUpdateGithubConnectorMutation
} from '@/redux/services/connector/githubConnectorApi';
import { useSearchable } from '@/hooks/use-searchable';
import { GithubRepository } from '@/redux/types/github';
import { SortOption } from '@/components/sort-selector';

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
  const searchParams = useSearchParams();
  const { data: githubRepositories, isLoading } = useGetAllGithubRepositoriesQuery();
  const router = useRouter();
  const [selectedRepository, setSelectedRepository] = React.useState<string | null>(null);
  const {
    filteredAndSortedData: filteredAndSortedApplications,
    searchTerm,
    handleSearchChange,
    handleSortChange,
    sortConfig
  } = useSearchable<GithubRepository>(
    githubRepositories || [],
    ['name', 'description', 'stargazers_count'],
    { key: 'name', direction: 'asc' }
  );
  const [currentPage, setCurrentPage] = useState(1);
  const ITEMS_PER_PAGE = React.useMemo(() => 9, []);

  const paginatedApplications = React.useMemo(
    () =>
      filteredAndSortedApplications.slice(
        (currentPage - 1) * ITEMS_PER_PAGE,
        currentPage * ITEMS_PER_PAGE
      ),
    [currentPage, filteredAndSortedApplications]
  );

  const totalPages = React.useMemo(
    () => Math.ceil(filteredAndSortedApplications.length / ITEMS_PER_PAGE),
    [filteredAndSortedApplications]
  );

  const handlePageChange = (pageNumber: number) => {
    setCurrentPage(pageNumber);
  };

  // Reset the current page when the search term or sort config changes
  useEffect(() => {
    setCurrentPage(1);
  }, [searchTerm, sortConfig]);

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
    githubRepositories,
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
    onSelectRepository
  };
}

export default useGithubRepoPagination;
