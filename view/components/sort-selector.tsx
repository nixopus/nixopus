import React from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';

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
  return (
    <Select
      onValueChange={(value) => {
        const [key, direction] = value.split('_');
        const newSort = options.find(
          (option) => option.value === key && option.direction === direction
        );
        if (newSort) {
          onSortChange(newSort);
        }
      }}
      value={`${currentSort.value as string}_${currentSort.direction}`}
    >
      <SelectTrigger className={className}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {options.map((option) => (
          <SelectItem
            key={`${option.value as string}_${option.direction}`}
            value={`${option.value as string}_${option.direction}`}
          >
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
