'use client';

import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export const NetworkCardSkeletonContent: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full space-y-4">
      {/* 2-column grid for download/upload */}
      <div className="grid grid-cols-2 gap-4 w-full">
        {/* Download column */}
        <div className="flex flex-col items-center text-center">
          <Skeleton className="h-8 w-8 rounded-full mb-2" /> {/* ArrowDownCircle icon */}
          <Skeleton className="h-3 w-14 mb-1" /> {/* "Download" label */}
          <Skeleton className="h-7 w-16" /> {/* text-2xl speed value */}
        </div>
        {/* Upload column */}
        <div className="flex flex-col items-center text-center">
          <Skeleton className="h-8 w-8 rounded-full mb-2" /> {/* ArrowUpCircle icon */}
          <Skeleton className="h-3 w-12 mb-1" /> {/* "Upload" label */}
          <Skeleton className="h-7 w-16" /> {/* text-2xl speed value */}
        </div>
      </div>
      {/* Total download/upload row */}
      <div className="flex gap-4 text-xs">
        <Skeleton className="h-3 w-16" /> {/* ↓ total download */}
        <Skeleton className="h-3 w-16" /> {/* ↑ total upload */}
      </div>
    </div>
  );
};
