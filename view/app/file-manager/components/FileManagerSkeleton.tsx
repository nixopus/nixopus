import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

const FileManagerSkeleton: React.FC = () => {
  return (
    <div className="mx-auto max-w-7xl p-6">
      <div className="mb-6 flex flex-col justify-between sm:flex-row">
        <Skeleton className="mb-4 h-8 w-48 sm:mb-0" />
        <Skeleton className="h-8 w-64" />
      </div>
      <div className="my-6 flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
        <Skeleton className="h-6 w-64" />
        <div className="flex items-center gap-4">
          <Skeleton className="h-8 w-24" />
          <Skeleton className="h-8 w-24" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 sm:gap-6 md:grid-cols-4 lg:grid-cols-5">
        {[...Array(10)].map((_, index) => (
          <div key={index} className="flex flex-col items-center">
            <Skeleton className="mb-2 h-16 w-16" />
            <Skeleton className="h-4 w-20" />
          </div>
        ))}
      </div>
    </div>
  );
};

export default FileManagerSkeleton;
