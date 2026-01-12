'use client';

import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { CardWrapper } from '@/components/ui/card-wrapper';
import PageLayout from '@/packages/layouts/page-layout';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Package } from 'lucide-react';

export const ContainersWidgetSkeleton: React.FC = () => {
  const { t } = useTranslation();

  return (
    <CardWrapper
      title={t('dashboard.containers.title')}
      icon={Package}
      compact
      actions={<Skeleton className="h-8 w-24" />}
    >
      <div className="border-b pb-2 mb-2">
        <div className="grid grid-cols-6 gap-4">
          {['h-4 w-8', 'h-4 w-12', 'h-4 w-12', 'h-4 w-12', 'h-4 w-10', 'h-4 w-14'].map(
            (className, idx) => (
              <Skeleton key={idx} className={className} />
            )
          )}
        </div>
      </div>
      <div className="space-y-3">
        {[0, 1, 2].map((i) => (
          <div key={i} className="grid grid-cols-6 gap-4 items-center">
            {[
              'h-4 w-16 font-mono',
              'h-4 w-24',
              'h-4 w-32',
              'h-5 w-16 rounded-full',
              'h-4 w-12',
              'h-4 w-16'
            ].map((className, idx) => (
              <Skeleton key={idx} className={className} />
            ))}
          </div>
        ))}
      </div>
    </CardWrapper>
  );
};

export default function ContainerDetailsLoading() {
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

export function ImagesSectionSkeleton() {
  return (
    <div className="space-y-8">
      <div className="flex items-center gap-8">
        {[1, 2].map((i) => (
          <div key={i} className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-lg" />
            <div className="space-y-1.5">
              <Skeleton className="h-6 w-16" />
              <Skeleton className="h-3 w-20" />
            </div>
          </div>
        ))}
      </div>

      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-center gap-4 p-4 rounded-xl bg-muted/10">
            <Skeleton className="h-12 w-12 rounded-xl" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-3 w-24" />
            </div>
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-24" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function ContainersLoading() {
  return (
    <PageLayout maxWidth="full" padding="md" spacing="lg">
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

      <div className="flex items-center gap-3 mb-6">
        <Skeleton className="h-10 w-80" />
        <div className="ml-auto flex items-center gap-2">
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-20" />
        </div>
      </div>

      <div className="rounded-xl border overflow-hidden">
        <div className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 px-4 py-3 bg-muted/30">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-14" />
          <Skeleton className="h-4 w-12" />
          <div className="w-24" />
        </div>

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
