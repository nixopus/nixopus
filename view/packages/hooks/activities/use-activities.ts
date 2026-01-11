'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { useSearchable } from '@/hooks/use-searchable';
import { SortOption } from '@/components/ui/sort-selector';
import { useGetActivitiesQuery } from '@/redux/services/audit';
import { ActivityMessage } from '@/redux/types/audit';

export const getActionColor = (actionColor: string) => {
  switch (actionColor) {
    case 'green':
      return 'bg-green-500';
    case 'blue':
      return 'bg-blue-500';
    case 'red':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
};

export const resourceTypeOptions: { value: string; label: string }[] = [
  { value: 'all', label: 'All Resources' },
  { value: 'user', label: 'Users' },
  { value: 'organization', label: 'Organization' },
  { value: 'application', label: 'Applications' },
  { value: 'deployment', label: 'Deployments' },
  { value: 'domain', label: 'Domains' },
  { value: 'file_manager', label: 'Files' },
  { value: 'container', label: 'Containers' },
  { value: 'role', label: 'Roles' },
  { value: 'permission', label: 'Permissions' },
  { value: 'feature_flag', label: 'Feature Flags' },
  { value: 'notification', label: 'Notifications' },
  { value: 'integration', label: 'Integrations' },
  { value: 'terminal', label: 'Terminal' },
  { value: 'audit', label: 'Audit' }
];

export interface ActivityListProps {
  title?: string;
  description?: string;
  showFilters?: boolean;
  pageSize?: number;
}

function useActivities() {
  const { t } = useTranslation();
  const [resourceType, setResourceType] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const ITEMS_PER_PAGE = 20;

  const {
    data: activitiesData,
    isLoading,
    error
  } = useGetActivitiesQuery({
    page: 1,
    pageSize: 100,
    search: '',
    resource_type: resourceType === 'all' ? '' : resourceType
  });

  const {
    filteredAndSortedData: filteredAndSortedActivities,
    searchTerm,
    handleSearchChange,
    handleSortChange,
    sortConfig
  } = useSearchable<ActivityMessage>(
    activitiesData?.activities || [],
    ['message', 'actor', 'resource', 'action'],
    { key: 'timestamp', direction: 'desc' }
  );

  const totalPages = Math.ceil((filteredAndSortedActivities?.length || 0) / ITEMS_PER_PAGE);

  const paginatedActivities = useMemo(
    () =>
      filteredAndSortedActivities?.slice(
        (currentPage - 1) * ITEMS_PER_PAGE,
        currentPage * ITEMS_PER_PAGE
      ) || [],
    [currentPage, filteredAndSortedActivities, ITEMS_PER_PAGE]
  );

  const handlePageChange = (pageNumber: number) => {
    setCurrentPage(pageNumber);
  };

  const handleResourceTypeChange = (value: string) => {
    setResourceType(value);
    setCurrentPage(1);
  };

  useEffect(() => {
    setCurrentPage(1);
  }, [searchTerm, sortConfig]);

  const onSortChange = (newSort: SortOption<ActivityMessage>) => {
    handleSortChange(newSort.value as keyof ActivityMessage);
  };

  const sortOptions: SortOption<ActivityMessage>[] = React.useMemo(
    () => [
      { label: 'Timestamp', value: 'timestamp', direction: 'desc' },
      { label: 'Message', value: 'message', direction: 'asc' },
      { label: 'Actor', value: 'actor', direction: 'asc' },
      { label: 'Resource', value: 'resource', direction: 'asc' },
      { label: 'Action', value: 'action', direction: 'asc' }
    ],
    []
  );

  const activities = paginatedActivities;

  return {
    resourceType,
    activities,
    sortOptions,
    onSortChange,
    handleResourceTypeChange,
    handlePageChange,
    handleSearchChange,
    isLoading,
    error,
    totalPages,
    sortConfig,
    searchTerm,
    currentPage
  };
}

export default useActivities;
