'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Extension,
  ExtensionListParams,
  SortDirection,
  ExtensionSortField,
  ExtensionCategory
} from '@/redux/types/extension';
import {
  useGetExtensionsQuery,
  useRunExtensionMutation,
  useCancelExecutionMutation,
  useGetExtensionCategoriesQuery
} from '@/redux/services/extensions/extensionsApi';

export function useExtensions() {
  const router = useRouter();
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{
    key: ExtensionSortField;
    direction: SortDirection;
  }>({
    key: 'name',
    direction: 'asc'
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(9);
  const [runModalOpen, setRunModalOpen] = useState(false);
  const [selectedExtension, setSelectedExtension] = useState<Extension | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<ExtensionCategory | null>(null);

  const queryParams: ExtensionListParams = {
    search: searchTerm || undefined,
    category: selectedCategory || undefined,
    sort_by: sortConfig.key,
    sort_dir: sortConfig.direction,
    page: currentPage,
    page_size: itemsPerPage
  };

  const { data: response, isLoading, error: apiError } = useGetExtensionsQuery(queryParams);

  const { data: categories = [] } = useGetExtensionCategoriesQuery();

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

  const handleCategoryChange = (value: string | null) => {
    setSelectedCategory((value as ExtensionCategory) || null);
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleInstall = (extension: Extension) => {
    setSelectedExtension(extension);
    setRunModalOpen(true);
  };

  const handleViewDetails = (extension: Extension) => {
    router.push(`/extensions/${extension.id}`);
  };

  const error = apiError ? 'Failed to load extensions' : null;

  const [runExtensionMutation] = useRunExtensionMutation();
  const [cancelExecutionMutation] = useCancelExecutionMutation();

  const handleRun = async (values: Record<string, unknown>) => {
    if (!selectedExtension) return;
    const exec = await runExtensionMutation({
      extensionId: selectedExtension.extension_id,
      body: { variables: values }
    }).unwrap();
    setRunModalOpen(false);
    router.push(`/extensions/${selectedExtension.id}?exec=${exec.id}&openLogs=1`);
  };

  const handleCancel = async (executionId: string) => {
    await cancelExecutionMutation({ executionId });
  };

  return {
    extensions,
    isLoading,
    error,
    categories,
    searchTerm,
    sortConfig,
    currentPage,
    totalPages,
    totalExtensions,
    handleSearchChange,
    handleSortChange,
    selectedCategory,
    handleCategoryChange,
    handlePageChange,
    handleInstall,
    handleViewDetails,
    handleRun,
    handleCancel,
    runModalOpen,
    setRunModalOpen,
    selectedExtension
  };
}
