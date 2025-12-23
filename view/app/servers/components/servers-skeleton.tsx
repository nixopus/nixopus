'use client';

import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import PageLayout from '@/components/layout/page-layout';

export function ServersPageSkeleton() {
  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div>
          <Skeleton className="h-8 w-32 mb-2" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-9 w-24" />
          <Skeleton className="h-9 w-32" />
        </div>
      </div>

      <div className="flex items-center gap-6 mb-6">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center gap-2">
            <Skeleton className="h-6 w-8" />
            <Skeleton className="h-4 w-16" />
          </div>
        ))}
      </div>

      <div className="flex items-center gap-3 mb-6">
        <Skeleton className="h-10 flex-1 max-w-md" />
        <Skeleton className="h-10 w-24" />
      </div>

      <Skeleton className="h-[500px] w-full rounded-xl" />
    </PageLayout>
  );
}

export function ServerCardSkeleton() {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start gap-3">
          <Skeleton className="h-10 w-10 rounded-lg" />
          <div className="flex-1">
            <Skeleton className="h-5 w-32 mb-2" />
            <Skeleton className="h-4 w-24" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex gap-2">
          <Skeleton className="h-5 w-20" />
          <Skeleton className="h-5 w-16" />
        </div>
        <div className="grid grid-cols-3 gap-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-16 rounded-lg" />
          ))}
        </div>
        <Skeleton className="h-4 w-28" />
        <div className="flex gap-2 pt-2 border-t">
          <Skeleton className="h-8 flex-1" />
          <Skeleton className="h-8 w-8" />
        </div>
      </CardContent>
    </Card>
  );
}

export function ServerListSkeleton() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <ServerCardSkeleton key={i} />
      ))}
    </div>
  );
}
