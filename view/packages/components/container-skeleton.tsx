'use client';

import { Skeleton } from '@/components/ui/skeleton';
import PageLayout from '@/components/layout/page-layout';

export default function ContainersLoading() {
  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-9 w-24" />
          <Skeleton className="h-9 w-32" />
          <Skeleton className="h-9 w-36" />
        </div>
      </div>

      {/* Stats */}
      <div className="flex items-center gap-6 mb-6">
        <div className="flex items-center gap-2">
          <Skeleton className="h-7 w-8" />
          <Skeleton className="h-4 w-12" />
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-2 w-2 rounded-full" />
          <Skeleton className="h-7 w-6" />
          <Skeleton className="h-4 w-16" />
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-2 w-2 rounded-full" />
          <Skeleton className="h-7 w-6" />
          <Skeleton className="h-4 w-16" />
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-3 mb-6">
        <Skeleton className="h-10 w-80" />
        <div className="ml-auto flex items-center gap-2">
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-20" />
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border overflow-hidden">
        {/* Header */}
        <div className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 bg-muted/30">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-14" />
          <Skeleton className="h-4 w-12" />
          <div className="w-24" />
        </div>

        {/* Rows */}
        <div className="divide-y divide-border/50">
          {[1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 items-center"
            >
              <div className="flex items-center gap-3">
                <Skeleton className="h-10 w-10 rounded-lg" />
                <div className="space-y-1.5">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-20" />
                </div>
              </div>
              <div className="space-y-1.5">
                <Skeleton className="h-4 w-40" />
                <Skeleton className="h-3 w-24" />
              </div>
              <Skeleton className="h-6 w-20 rounded-full" />
              <div className="w-32 space-y-1">
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-3 w-20" />
              </div>
              <div className="w-24" />
            </div>
          ))}
        </div>
      </div>
    </PageLayout>
  );
}
