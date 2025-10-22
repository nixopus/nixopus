import React from 'react';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';

export interface SortOption<T> {
  value: keyof T;
  label: string;
  direction: 'asc' | 'desc';
}

interface SortSelectProps<T> {
  options: SortOption<T>[];
  currentSort: SortOption<T>;
  onSortChange: (newSort: SortOption<T>) => void;
  placeholder?: string;
  className?: string;
}

export function SortSelect<T>({
  options,
  currentSort,
  onSortChange,
  placeholder = 'Sort by',
  className = 'w-full sm:w-[180px]'
}: SortSelectProps<T>) {
  const selectOptions: SelectOption[] = options.map((option) => ({
    value: `${option.value as string}_${option.direction}`,
    label: option.label
  }));

  const handleValueChange = (value: string) => {
    const [key, direction] = value.split('_');
    const newSort = options.find(
      (option) => option.value === key && option.direction === direction
    );
    if (newSort) {
      onSortChange(newSort);
    }
  };

  return (
    <SelectWrapper
      value={`${currentSort.value as string}_${currentSort.direction}`}
      onValueChange={handleValueChange}
      options={selectOptions}
      placeholder={placeholder}
      className={className}
    />
  );
}
