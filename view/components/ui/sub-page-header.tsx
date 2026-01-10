import { cn } from '@/lib/utils';
import React from 'react';

export interface SubPageHeaderProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  icon?: React.ReactNode;
  title: React.ReactNode;
  metadata?: React.ReactNode;
  actions?: React.ReactNode;
  children?: React.ReactNode;
}

export function SubPageHeader({
  className,
  icon,
  title,
  metadata,
  actions,
  children,
  ...props
}: SubPageHeaderProps) {
  return (
    <div
      data-slot="sub-page-header"
      className={cn(
        'flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6',
        className
      )}
      {...props}
    >
      <div className="flex items-center gap-4 flex-1 min-w-0">
        {icon && (
          <div data-slot="sub-page-header-icon" className="shrink-0">
            {icon}
          </div>
        )}
        <div className="flex-1 min-w-0">
          <div className="text-2xl font-bold tracking-tight">{title}</div>
          {metadata && (
            <div data-slot="sub-page-header-metadata" className="mt-1">
              {metadata}
            </div>
          )}
          {children && (
            <div
              data-slot="sub-page-header-content"
              className={cn(metadata && 'mt-2', !metadata && 'mt-1')}
            >
              {children}
            </div>
          )}
        </div>
      </div>
      {actions && (
        <div data-slot="sub-page-header-actions" className="flex items-center gap-2 shrink-0">
          {actions}
        </div>
      )}
    </div>
  );
}

export default SubPageHeader;
