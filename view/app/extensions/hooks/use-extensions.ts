'use client';

import { useState } from 'react';
import { Extension, ExtensionListParams, SortDirection, ExtensionSortField } from '@/redux/types/extension';
import { useGetExtensionsQuery } from '@/redux/services/extensions/extensionsApi';

export function useExtensions() {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{ key: ExtensionSortField; direction: SortDirection }>({
    key: 'name',
    direction: 'asc'
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(12);

  const queryParams: ExtensionListParams = {
    search: searchTerm || undefined,
    sort_by: sortConfig.key,
    sort_dir: sortConfig.direction,
    page: currentPage,
    page_size: itemsPerPage
  };

  const {
    data: response,
    isLoading,
    error: apiError
  } = useGetExtensionsQuery(queryParams);

  const extensions = response?.extensions || [];
  const totalPages = response?.total_pages || 0;
  const totalExtensions = response?.total || 0;

  const handleSearchChange = (value: string) => {
    setSearchTerm(value);
    setCurrentPage(1); // Reset to first page when searching
  };

  const handleSortChange = (key: ExtensionSortField, direction: SortDirection) => {
    setSortConfig({ key, direction });
    setCurrentPage(1); // Reset to first page when sorting
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleInstall = (extension: Extension) => {
    // TODO: Implement installation logic
    console.log('Installing extension:', extension.name);
  };

  const handleViewDetails = (extension: Extension) => {
    console.log('Viewing details for extension:', extension.name);
  };

  const error = apiError ? 'Failed to load extensions' : null;

  return {
    extensions,
    isLoading,
    error,
    searchTerm,
    sortConfig,
    currentPage,
    totalPages,
    totalExtensions,
    handleSearchChange,
    handleSortChange,
    handlePageChange,
    handleInstall,
    handleViewDetails
  };
}
