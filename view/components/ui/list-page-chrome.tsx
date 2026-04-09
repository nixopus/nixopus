'use client';

import React from 'react';
import { Button, Skeleton, CardWrapper } from '@nixopus/ui';
import { LayoutGrid, List, RotateCw } from 'lucide-react';

export const DATA_TABLE_CLASS = [
  '[&_thead_tr]:border-b [&_thead_tr]:border-border',
  '[&_th]:uppercase [&_th]:text-[11px] [&_th]:tracking-wider [&_th]:font-medium',
  '[&_th]:text-muted-foreground [&_th]:px-5 [&_th]:h-10',
  '[&_tbody]:bg-foreground/3',
  '[&_td]:px-5 [&_td]:py-4',
  '[&_tbody_tr]:border-b [&_tbody_tr]:border-border',
  '[&_tbody_tr:last-child]:border-0',
  '[&_tbody_tr]:transition-colors [&_tbody_tr:hover]:bg-foreground/5'
].join(' ');

export const LIST_GRID_CLASS =
  'grid grid-cols-1 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4';

interface ViewToggleProps {
  viewMode: 'grid' | 'table';
  onViewChange: (mode: 'grid' | 'table') => void;
}

export function ViewToggle({ viewMode, onViewChange }: ViewToggleProps) {
  return (
    <div className="flex items-center border rounded-md">
      <Button
        variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
        size="icon"
        className="h-8 w-8 rounded-r-none border-0"
        onClick={() => onViewChange('grid')}
        aria-label="Grid view"
      >
        <LayoutGrid className="h-4 w-4" />
      </Button>
      <Button
        variant={viewMode === 'table' ? 'secondary' : 'ghost'}
        size="icon"
        className="h-8 w-8 rounded-l-none border-0"
        onClick={() => onViewChange('table')}
        aria-label="Table view"
      >
        <List className="h-4 w-4" />
      </Button>
    </div>
  );
}

interface RefreshButtonProps {
  onClick: () => void;
  isFetching?: boolean;
  ariaLabel?: string;
}

export function RefreshButton({ onClick, isFetching, ariaLabel = 'Refresh' }: RefreshButtonProps) {
  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={onClick}
      className="h-8 w-8"
      aria-label={ariaLabel}
    >
      <RotateCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
    </Button>
  );
}

interface ListToolbarProps {
  left: React.ReactNode;
  children?: React.ReactNode;
}

export function ListToolbar({ left, children }: ListToolbarProps) {
  return (
    <div className="flex items-center justify-between flex-wrap gap-3 min-w-0">
      <div className="flex items-center gap-3 min-w-0 flex-1">{left}</div>
      <div className="flex items-center gap-2 shrink-0">{children}</div>
    </div>
  );
}

export const CARD_CLASS =
  'group relative rounded-md p-5 w-full overflow-hidden flex min-h-32 md:min-h-44 h-44 flex-col border border-border/50 transition-colors cursor-pointer';
export const CARD_HEADER_CLASS = 'w-full min-w-0 pb-0 px-0';

export function CardSkeleton({ count = 6 }: { count?: number }) {
  return (
    <div className={LIST_GRID_CLASS}>
      {Array.from({ length: count }).map((_, i) => (
        <CardWrapper
          key={i}
          className="group relative rounded-md p-5 w-full overflow-hidden flex min-h-32 md:min-h-44 h-44 flex-col border border-border/50"
          header={
            <div className="flex-1 min-w-0 w-full space-y-1.5">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-3.5 w-24" />
            </div>
          }
          headerClassName={CARD_HEADER_CLASS}
          contentClassName="flex flex-col flex-1"
        >
          <div className="flex-1" />
          <div className="flex items-center gap-3 pt-1.5 border-t border-border/50">
            <Skeleton className="h-3.5 w-16" />
            <Skeleton className="h-3.5 w-20" />
          </div>
        </CardWrapper>
      ))}
    </div>
  );
}
