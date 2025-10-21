'use client';

import React, { } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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
import useActivities, { ActivityListProps, getActionColor, resourceTypeOptions } from './hooks/use-activities';

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

function ActivityList({
  title = 'Team Activities',
  showFilters = true,
  pageSize = 10
}: ActivityListProps) {
  const {
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
    currentPage,

  } = useActivities()

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
