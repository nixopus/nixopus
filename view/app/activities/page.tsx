'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { useSearchable } from '@/hooks/use-searchable';
import { SortOption } from '@/components/ui/sort-selector';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useGetActivitiesQuery } from '@/redux/services/audit';
import { formatDistanceToNow } from 'date-fns';
import { Loader2 } from 'lucide-react';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { DahboardUtilityHeader } from '@/components/layout/dashboard-page-header';
import PaginationWrapper from '@/components/ui/pagination';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { ActivityMessage } from '@/redux/types/audit';

export default function ActivitiesPage() {
  const { t } = useTranslation();

  return (
    <div className="">
      <ActivityList
        title={t('activities.list.title')}
        description={t('activities.list.description')}
        showFilters={true}
        pageSize={20}
      />
    </div>
  );
}

const getActionColor = (actionColor: string) => {
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

const resourceTypeOptions: { value: string; label: string }[] = [
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
  { value: 'audit', label: 'Audit' },
];

interface ActivityListProps {
  title?: string;
  description?: string;
  showFilters?: boolean;
  pageSize?: number;
}

function ActivityList({
  title = 'Team Activities',
  showFilters = true,
  pageSize = 10
}: ActivityListProps) {
  const { t } = useTranslation();
  const [resourceType, setResourceType] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const ITEMS_PER_PAGE = 20;

  const { data: activitiesData, isLoading, error } = useGetActivitiesQuery({
    page: 1,
    pageSize: 100,
    search: '',
    resource_type: resourceType === 'all' ? '' : resourceType,
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

  return (
    <div className="space-y-4">
      <DahboardUtilityHeader<ActivityMessage>
        searchTerm={searchTerm}
        handleSearchChange={handleSearchChange}
        sortConfig={sortConfig}
        onSortChange={onSortChange}
        sortOptions={sortOptions}
        label={title}
        searchPlaceHolder="Search activities..."
      >
        {showFilters && (
          <Select value={resourceType} onValueChange={handleResourceTypeChange}>
            <SelectTrigger className="w-full sm:w-48">
              <SelectValue placeholder="Filter by resource" />
            </SelectTrigger>
            <SelectContent>
              {resourceTypeOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </DahboardUtilityHeader>
      {isLoading ? (
        <div className="flex items-center justify-center p-8">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span className="ml-2">Loading activities...</span>
        </div>
      ) : error ? (
        <div className="p-4 text-red-600 text-center">
          Failed to load activities. Please try again.
        </div>
      ) : activities.length > 0 ? (
        <>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[50px]"></TableHead>
                  <TableHead>Message</TableHead>
                  <TableHead>Timestamp</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {activities.map((activity: ActivityMessage) => (
                  <TableRow key={activity.id}>
                    <TableCell>
                      <div className={`h-3 w-3 rounded-full ${getActionColor(activity.action_color)}`}></div>
                    </TableCell>
                    <TableCell className="max-w-md">
                      <TypographySmall className="text-foreground">
                        {activity.message}
                      </TypographySmall>
                    </TableCell>
                    <TableCell>
                      <TypographyMuted className="text-xs">
                        {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
                      </TypographyMuted>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {totalPages > 1 && (
            <div className="mt-6 flex justify-center">
              <PaginationWrapper
                currentPage={currentPage}
                totalPages={totalPages}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </>
      ) : (
        <div className="text-center text-muted-foreground py-8">
          {searchTerm || (resourceType && resourceType !== 'all') ? 'No activities found matching your filters.' : 'No activities yet.'}
        </div>
      )}
    </div>
  );
}
