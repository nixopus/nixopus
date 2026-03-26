import { cn } from '@/lib/utils';
import React from 'react';

interface PageLayoutProps {
  children: React.ReactNode;
  className?: string;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | '5xl' | '6xl' | '7xl' | 'full';
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
  spacing?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
}

const maxWidthClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  '2xl': 'max-w-2xl',
  '3xl': 'max-w-3xl',
  '4xl': 'max-w-4xl',
  '5xl': 'max-w-5xl',
  '6xl': 'max-w-6xl',
  '7xl': 'max-w-7xl',
  full: 'max-w-full'
};

const paddingClasses = {
  none: '',
  sm: 'py-2 sm:py-3 lg:py-4',
  md: 'py-4 sm:py-6 lg:py-8',
  lg: 'py-6 sm:py-8 lg:py-10',
  xl: 'py-8 sm:py-10 lg:py-12'
};

const spacingClasses = {
  none: '',
  sm: 'space-y-4',
  md: 'space-y-6',
  lg: 'space-y-8',
  xl: 'space-y-10'
};

function PageLayout({
  children,
  className,
  maxWidth = 'full',
  padding = 'md',
  spacing = 'lg'
}: PageLayoutProps) {
  return (
    <div
      className={cn(
        maxWidth === 'full' ? 'w-full px-4 sm:px-6 lg:px-8' : 'container mx-auto',
        paddingClasses[padding],
        spacingClasses[spacing],
        maxWidthClasses[maxWidth],
        className
      )}
    >
      {children}
    </div>
  );
}

export default PageLayout;
