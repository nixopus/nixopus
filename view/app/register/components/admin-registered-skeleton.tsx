'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

export const AdminRegisteredSkeleton = () => {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className="flex flex-col gap-6">
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <Skeleton className="h-8 w-48" />
                    <Skeleton className="mt-4 h-4 w-64" />
                  </div>
                  <div className="space-y-4">
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-16" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-20" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-32" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <Skeleton className="h-10 w-full" />
                    <Skeleton className="mx-auto h-4 w-48" />
                  </div>
                </div>
              </div>
              <div className="bg-muted relative hidden md:block">
                <Skeleton className="absolute inset-0 h-full w-full" />
              </div>
            </CardContent>
          </Card>
          <Skeleton className="mx-auto h-4 w-64" />
        </div>
      </div>
    </div>
  );
};
