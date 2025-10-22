'use client';

import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { TypographyMuted } from '@/components/ui/typography';
import { cn } from '@/lib/utils';

export interface TableColumn<T = any> {
  key: string;
  title: string;
  dataIndex?: keyof T;
  render?: (value: any, record: T, index: number) => React.ReactNode;
  width?: string;
  className?: string;
  sortable?: boolean;
  align?: 'left' | 'center' | 'right';
}

export interface SortConfig {
  field: string;
  order: 'asc' | 'desc';
}

export interface DataTableProps<T = any> {
  data: T[];
  columns: TableColumn<T>[];
  loading?: boolean;
  emptyMessage?: string;
  emptyStateComponent?: React.ReactNode;
  onRowClick?: (record: T, index: number) => void;
  onSort?: (field: string) => void;
  sortConfig?: SortConfig;
  className?: string;
  tableClassName?: string;
  containerClassName?: string;
  rowClassName?: string | ((record: T, index: number) => string);
  loadingRows?: number;
  showBorder?: boolean;
  hoverable?: boolean;
  striped?: boolean;
}

export function DataTable<T = any>({
  data,
  columns,
  loading = false,
  emptyMessage = 'No data available',
  emptyStateComponent,
  onRowClick,
  onSort,
  sortConfig,
  className,
  tableClassName,
  containerClassName,
  rowClassName,
  loadingRows = 5,
  showBorder = true,
  hoverable = true,
  striped = false
}: DataTableProps<T>) {
  const handleRowClick = (record: T, index: number) => {
    if (onRowClick) {
      onRowClick(record, index);
    }
  };

  const handleSort = (field: string) => {
    if (onSort) {
      onSort(field);
    }
  };

  const getRowClassName = (record: T, index: number) => {
    const baseClasses = [];
    
    if (hoverable && onRowClick) {
      baseClasses.push('cursor-pointer');
    }
    
    if (striped && index % 2 === 0) {
      baseClasses.push('bg-muted/25');
    }
    
    if (!striped) {
      baseClasses.push('border-0');
    }
    
    if (typeof rowClassName === 'function') {
      baseClasses.push(rowClassName(record, index));
    } else if (rowClassName) {
      baseClasses.push(rowClassName);
    }
    
    return baseClasses.join(' ');
  };

  const renderLoadingRows = () => {
    return Array.from({ length: loadingRows }).map((_, index) => (
      <TableRow key={index}>
        {columns.map((column) => (
          <TableCell key={column.key}>
            <Skeleton className="h-4 w-full" />
          </TableCell>
        ))}
      </TableRow>
    ));
  };

  const renderEmptyState = () => {
    if (emptyStateComponent) {
      return (
        <TableRow>
          <TableCell colSpan={columns.length} className="text-center py-8">
            {emptyStateComponent}
          </TableCell>
        </TableRow>
      );
    }

    return (
      <TableRow>
        <TableCell colSpan={columns.length} className="text-center py-8">
          <TypographyMuted>{emptyMessage}</TypographyMuted>
        </TableCell>
      </TableRow>
    );
  };

  const renderCellContent = (column: TableColumn<T>, record: T, index: number) => {
    if (column.render) {
      return column.render(
        column.dataIndex ? record[column.dataIndex] : undefined,
        record,
        index
      );
    }

    const value = column.dataIndex ? record[column.dataIndex] : undefined;
    return value !== undefined && value !== null ? String(value) : '-';
  };

  const getSortIcon = (column: TableColumn<T>) => {
    if (!column.sortable || !onSort) return null;

    const isActive = sortConfig?.field === column.key;
    const isAsc = sortConfig?.order === 'asc';

    return (
      <span className="ml-1 opacity-40">
        {isActive ? (
          <span className="opacity-100">
            {isAsc ? '↑' : '↓'}
          </span>
        ) : (
          <span>↕</span>
        )}
      </span>
    );
  };

  return (
    <div className={cn('relative w-full overflow-x-auto', containerClassName)}>
      <Table className={cn(
        'w-full caption-bottom text-sm',
        showBorder && 'border',
        tableClassName
      )}>
        <TableHeader>
          <TableRow>
            {columns.map((column) => (
              <TableHead
                key={column.key}
                className={cn(
                  column.width && `w-[${column.width}]`,
                  column.align === 'center' && 'text-center',
                  column.align === 'right' && 'text-right',
                  column.sortable && onSort && 'cursor-pointer select-none',
                  column.className
                )}
                onClick={() => column.sortable && handleSort(column.key)}
              >
                <div className={cn(
                  "flex items-center gap-1",
                  column.align === 'center' && 'justify-center',
                  column.align === 'right' && 'justify-end'
                )}>
                  <span>{column.title}</span>
                  {getSortIcon(column)}
                </div>
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            renderLoadingRows()
          ) : data.length === 0 ? (
            renderEmptyState()
          ) : (
            data.map((record, index) => (
              <TableRow
                key={index}
                className={getRowClassName(record, index)}
                onClick={() => handleRowClick(record, index)}
              >
                {columns.map((column) => (
                  <TableCell
                    key={column.key}
                    className={cn(
                      column.align === 'center' && 'text-center',
                      column.align === 'right' && 'text-right',
                      column.className
                    )}
                  >
                    {renderCellContent(column, record, index)}
                  </TableCell>
                ))}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

export default DataTable;
