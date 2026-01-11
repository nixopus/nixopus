import { cn } from '@/lib/utils';
import React from 'react';
import { TypographyH1, TypographyMuted } from '@/components/ui/typography';

export interface MainPageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  label: string;
  description?: string;
  badge?: React.ReactNode;
  actions?: React.ReactNode;
  children?: React.ReactNode;
  highlightLabel?: boolean;
}

export function MainPageHeader({
  className,
  label,
  description,
  badge,
  actions,
  children,
  highlightLabel = true,
  ...props
}: MainPageHeaderProps) {
  return (
    <div
      data-slot="main-page-header"
      className={cn(
        'flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between mb-8',
        className
      )}
      {...props}
    >
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <TypographyH1 className={cn(highlightLabel && 'text-primary')}>{label}</TypographyH1>
          {badge && (
            <div data-slot="main-page-header-badge" className="shrink-0">
              {badge}
            </div>
          )}
        </div>
        {description && (
          <div data-slot="main-page-header-description" className="mt-1">
            <TypographyMuted>{description}</TypographyMuted>
          </div>
        )}
        {children && (
          <div data-slot="main-page-header-content" className={cn(description && 'mt-2')}>
            {children}
          </div>
        )}
      </div>
      {actions && (
        <div
          data-slot="main-page-header-actions"
          className="flex items-center gap-2 shrink-0 sm:ml-4"
        >
          {actions}
        </div>
      )}
    </div>
  );
}

export default MainPageHeader;
