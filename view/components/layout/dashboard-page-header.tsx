import { cn } from '@/lib/utils';
import React from 'react';
import { SortOption, SortSelect } from '@/components/ui/sort-selector';
import { SearchBar } from '@/components/ui/search-bar';
import { SortConfig } from '@/hooks/use-searchable';
import { TypographyH1, TypographyH2, TypographyMuted } from '@/components/ui/typography';

interface DashboardPageHeaderProps {
  className?: string;
  label: string;
  description: string;
}

function DashboardPageHeader({ className, label, description }: DashboardPageHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between space-y-2', className)}>
      <span className="">
        <TypographyH1 className="text-primary">{label}</TypographyH1>
        <TypographyMuted>{description}</TypographyMuted>
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
      <TypographyH2 className="text-primary">{label}</TypographyH2>
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
