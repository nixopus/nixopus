'use client';

import React from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { cn } from '@/lib/utils';

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface SelectWrapperProps {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  triggerClassName?: string;
  contentClassName?: string;
  emptyMessage?: string;
  searchable?: boolean;
  clearable?: boolean;
  onClear?: () => void;
  loading?: boolean;
  error?: boolean;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'outline' | 'ghost';
}

export function SelectWrapper({
  value,
  defaultValue,
  onValueChange,
  options,
  placeholder = 'Select an option',
  disabled = false,
  className,
  triggerClassName,
  contentClassName,
  emptyMessage = 'No options available',
  searchable = false,
  clearable = false,
  onClear,
  loading = false,
  error = false,
  size = 'md',
  variant = 'default'
}: SelectWrapperProps) {
  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'h-8 px-2 text-xs';
      case 'lg':
        return 'h-12 px-4 text-base';
      default:
        return 'h-10 px-3 text-sm';
    }
  };

  const getVariantClasses = () => {
    switch (variant) {
      case 'outline':
        return 'border-2';
      case 'ghost':
        return 'border-0 bg-transparent';
      default:
        return 'border';
    }
  };

  const getErrorClasses = () => {
    return error ? 'border-destructive focus:border-destructive' : '';
  };

  const handleValueChange = (newValue: string) => {
    if (onValueChange) {
      onValueChange(newValue);
    }
  };

  const renderEmptyState = () => {
    if (loading) {
      return (
        <div className="flex items-center justify-center p-4">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
          <span className="ml-2 text-sm text-muted-foreground">Loading...</span>
        </div>
      );
    }

    return (
      <div className="flex items-center justify-center p-4">
        <span className="text-sm text-muted-foreground">{emptyMessage}</span>
      </div>
    );
  };

  return (
    <div className={cn('relative', className)}>
      <Select
        value={value}
        defaultValue={defaultValue}
        onValueChange={handleValueChange}
        disabled={disabled || loading}
      >
        <SelectTrigger
          className={cn(
            'w-full',
            getSizeClasses(),
            getVariantClasses(),
            getErrorClasses(),
            triggerClassName
          )}
        >
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent className={contentClassName}>
          {options.length === 0 ? (
            renderEmptyState()
          ) : (
            <>
              {clearable && value && onClear && (
                <SelectItem
                  value=""
                  onSelect={onClear}
                  className="text-muted-foreground"
                >
                  Clear selection
                </SelectItem>
              )}
              {options.map((option) => (
                <SelectItem
                  key={option.value}
                  value={option.value}
                  disabled={option.disabled}
                >
                  {option.label}
                </SelectItem>
              ))}
            </>
          )}
        </SelectContent>
      </Select>
    </div>
  );
}

export default SelectWrapper;
