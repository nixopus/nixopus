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
  sm: 'py-4',
  md: 'py-6',
  lg: 'py-8',
  xl: 'py-10'
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
  maxWidth = '6xl',
  padding = 'md',
  spacing = 'lg'
}: PageLayoutProps) {
  return (
    <div
      className={cn(
        'container mx-auto',
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
