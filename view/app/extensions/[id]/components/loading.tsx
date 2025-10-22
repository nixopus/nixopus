'use client';

import PageLayout from '@/components/layout/page-layout';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { Info, Terminal } from 'lucide-react';

export default function Loading() {
  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Skeleton className="h-10 w-10 rounded" />
          <div>
            <Skeleton className="h-6 w-48 mb-2" />
            <Skeleton className="h-4 w-32" />
          </div>
        </div>
      </div>

      <div className="mt-6">
        <Tabs defaultValue="overview" className="w-full">
          <TabsList>
            <TabsTrigger value="overview">
              <Info className="mr-2 h-4 w-4" />
              <span>Overview</span>
            </TabsTrigger>
            <TabsTrigger value="executions">
              <Terminal className="mr-2 h-4 w-4" />
              <span>Executions</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="mt-6">
            <Skeleton className="h-40 w-full" />
          </TabsContent>

          <TabsContent value="executions" className="mt-6">
            <div className="rounded-md border overflow-hidden">
              <div className="grid grid-cols-12 bg-muted/50 px-3 py-2 text-xs font-medium text-muted-foreground">
                <div className="col-span-4">
                  <Skeleton className="h-4 w-24" />
                </div>
                <div className="col-span-2">
                  <Skeleton className="h-4 w-16" />
                </div>
                <div className="col-span-3">
                  <Skeleton className="h-4 w-20" />
                </div>
                <div className="col-span-3">
                  <Skeleton className="h-4 w-28" />
                </div>
              </div>
              <div className="divide-y">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="grid grid-cols-12 px-3 py-3 text-sm">
                    <div className="col-span-4">
                      <Skeleton className="h-4 w-40" />
                    </div>
                    <div className="col-span-2">
                      <Skeleton className="h-4 w-20" />
                    </div>
                    <div className="col-span-3">
                      <Skeleton className="h-4 w-32" />
                    </div>
                    <div className="col-span-3">
                      <Skeleton className="h-4 w-32" />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </PageLayout>
  );
}
