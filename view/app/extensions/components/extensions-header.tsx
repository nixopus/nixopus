'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { Input } from '@/components/ui/input';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';
import { Search } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { ExtensionSortField, SortDirection } from '@/redux/types/extension';

interface ExtensionsHeaderProps {
  searchTerm?: string;
  onSearchChange?: (value: string) => void;
  sortConfig?: { key: ExtensionSortField; direction: SortDirection };
  onSortChange?: (key: ExtensionSortField, direction: SortDirection) => void;
  isLoading?: boolean;
}

function ExtensionsHeader({
  searchTerm = '',
  onSearchChange,
  sortConfig,
  onSortChange,
  isLoading = false
}: ExtensionsHeaderProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return <ExtensionsHeaderSkeleton />;
  }

  const sortOptions: SelectOption[] = [
    { value: 'name_asc', label: t('extensions.sortOptions.name') + ' (A-Z)' },
    { value: 'name_desc', label: t('extensions.sortOptions.name') + ' (Z-A)' }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">
            {t('extensions.exploreExtensionsTitle')}
          </h2>
        </div>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder={t('extensions.searchPlaceholder')}
              value={searchTerm}
              onChange={(e) => onSearchChange?.(e.target.value)}
              className="pl-10 w-full sm:w-[300px]"
            />
          </div>
          <SelectWrapper
            value={sortConfig ? `${sortConfig.key}_${sortConfig.direction}` : 'name_asc'}
            onValueChange={(value) => {
              const [key, direction] = value.split('_') as [ExtensionSortField, SortDirection];
              onSortChange?.(key, direction);
            }}
            options={sortOptions}
            placeholder={t('extensions.sortBy')}
            className="w-full sm:w-[180px]"
          />
        </div>
      </div>
    </div>
  );
}

function ExtensionsHeaderSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <Skeleton className="h-8 w-48" />
        </div>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <Skeleton className="h-10 w-full sm:w-[300px]" />
          <Skeleton className="h-10 w-full sm:w-[180px]" />
        </div>
      </div>
    </div>
  );
}

export default ExtensionsHeader;
