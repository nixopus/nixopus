import { cn } from '@/lib/utils';
import React from 'react';

interface DashboardPageHeaderProps {
  className?: string;
  label: string;
  description: string;
}

function DashboardPageHeader({ className, label, description }: DashboardPageHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between space-y-2', className)}>
      <span className="">
        <h2 className="text-2xl font-bold tracking-tight">{label}</h2>
        <p className="text-muted-foreground">{description}</p>
      </span>
    </div>
  );
}

export default DashboardPageHeader;
