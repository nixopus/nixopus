import { cn } from '@/lib/utils';
import React from 'react';
import { SortOption, SortSelect } from './sort-selector';
import { SearchBar } from './search-bar';
import { SortConfig } from '@/hooks/use-searchable';

interface DashboardPageHeaderProps {
  className?: string;
  label: string;
  description: string;
}

function DashboardPageHeader({ className, label, description }: DashboardPageHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between space-y-2', className)}>
      <span className="">
        <h2 className="text-2xl font-bold tracking-tight">{label}</h2>
        <p className="text-muted-foreground">{description}</p>
      </span>
    </div>
  );
}

export default DashboardPageHeader;

interface DashboardUtilityHeaderProps<T> {
  className?: string;
  searchTerm: string;
  handleSearchChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  sortConfig: SortConfig<T>;
  onSortChange: (newSort: SortOption<T>) => void;
  sortOptions: SortOption<T>[];
  label: string;
  searchPlaceHolder?: string;
  children?: React.ReactNode;
}

export function DahboardUtilityHeader<T>({
  className,
  searchTerm,
  handleSearchChange,
  sortConfig,
  onSortChange,
  sortOptions,
  label,
  children,
  searchPlaceHolder = 'Search...'
}: DashboardUtilityHeaderProps<T>) {
  return (
    <div className={'space-y-6' + className}>
      <h1 className="text-3xl font-bold">{label}</h1>
      <div className="flex flex-col gap-4 sm:flex-row mt-4 justify-between items-center">
        <div className="flex-grow">
          <SearchBar
            searchTerm={searchTerm}
            handleSearchChange={handleSearchChange}
            label={searchPlaceHolder}
          />
        </div>
        <div className="flex gap-4 items-center">
          <SortSelect<T>
            options={sortOptions}
            currentSort={{
              value: sortConfig.key,
              direction: sortConfig.direction,
              label:
                sortOptions.find(
                  (option) =>
                    option.value === sortConfig.key && option.direction === sortConfig.direction
                )?.label || ''
            }}
            onSortChange={onSortChange}
            placeholder="Sort by"
          />
          {children}
        </div>
      </div>
    </div>
  );
}
