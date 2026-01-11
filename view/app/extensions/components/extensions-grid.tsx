'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { ExtensionCard } from './extension-card';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertCircle } from 'lucide-react';
import { Extension } from '@/redux/types/extension';

interface ExtensionsGridProps {
  extensions?: Extension[];
  isLoading?: boolean;
  error?: string;
  onInstall?: (extension: Extension) => void;
  onViewDetails?: (extension: Extension) => void;
}

function ExtensionsGrid({
  extensions = [],
  isLoading = false,
  error,
  onInstall,
  onViewDetails
}: ExtensionsGridProps) {
  const { t } = useTranslation();

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (isLoading) {
    return <ExtensionsGridSkeleton />;
  }

  if (extensions.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="mx-auto max-w-md">
          <div className="text-6xl mb-4">🔍</div>
          <h3 className="text-lg font-semibold mb-2">{t('extensions.noExtensions')}</h3>
          <p className="text-muted-foreground">
            Try adjusting your search or filters to find more extensions.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {extensions.map((extension) => (
        <ExtensionCard
          key={extension.id}
          extension={extension}
          onInstall={onInstall}
          onViewDetails={onViewDetails}
        />
      ))}
    </div>
  );
}

function ExtensionsGridSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className="group h-full transition-all duration-200 bg-card border-border p-6 rounded-xl"
        >
          <div className="space-y-4">
            <div className="flex items-start gap-4">
              <Skeleton className="h-12 w-12 rounded-full flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <Skeleton className="h-6 w-48 mb-2" />
                <Skeleton className="h-4 w-32" />
              </div>
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <div className="flex gap-2 pt-6">
              <Skeleton className="h-10 flex-1" />
              <Skeleton className="h-10 w-10" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

export default ExtensionsGrid;
