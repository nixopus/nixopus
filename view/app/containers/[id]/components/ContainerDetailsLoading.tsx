'use client';

import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/packages/layouts/page-layout';

function ContainerDetailsLoading() {
  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-4">
          <Skeleton className="w-12 h-12 rounded-xl" />
          <div className="space-y-2">
            <Skeleton className="h-7 w-48" />
            <div className="flex items-center gap-2">
              <Skeleton className="h-5 w-24" />
              <Skeleton className="h-5 w-16" />
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-9 w-20" />
          <Skeleton className="h-9 w-20" />
          <Skeleton className="h-9 w-24" />
        </div>
      </div>

      <div className="border-b mb-6">
        <div className="flex gap-2">
          <Skeleton className="h-10 w-28" />
          <Skeleton className="h-10 w-20" />
          <Skeleton className="h-10 w-24" />
          <Skeleton className="h-10 w-20" />
        </div>
      </div>

      <div className="space-y-10">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-8 w-20" />
              <Skeleton className="h-4 w-24" />
            </div>
          ))}
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-32" />
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[1, 2, 3].map((i) => (
              <div key={i} className="space-y-3">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-8 w-8 rounded-lg" />
                  <Skeleton className="h-4 w-16" />
                </div>
                <Skeleton className="h-2 w-full rounded-full" />
                <Skeleton className="h-6 w-20" />
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-40" />
          <div className="flex flex-wrap gap-4">
            <Skeleton className="h-14 w-48 rounded-xl" />
            <Skeleton className="h-14 w-40 rounded-xl" />
            <Skeleton className="h-14 w-36 rounded-xl" />
          </div>
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-36" />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-12">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="flex items-start gap-3 py-2">
                <Skeleton className="h-4 w-4 mt-1" />
                <div className="space-y-1.5">
                  <Skeleton className="h-3 w-16" />
                  <Skeleton className="h-5 w-40" />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-14 w-full rounded-xl" />
        </div>
      </div>
    </PageLayout>
  );
}

export default ContainerDetailsLoading;
