'use client';

import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export const NetworkCardSkeletonContent: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full space-y-4">
      <Skeleton className="h-16 w-full" />
      <Skeleton className="h-16 w-full" />
      <Skeleton className="h-8 w-3/4" />
    </div>
  );
};
