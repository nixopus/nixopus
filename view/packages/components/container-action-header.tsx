'use client';

import React from 'react';
import { RefreshCw, Trash2, Loader2, Scissors } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { AnyPermissionGuard } from '@/packages/components/rbac';
import { translationKey } from '@/packages/hooks/shared/use-translation';

interface ActionHeaderProps {
  handleRefresh: () => Promise<void>;
  isRefreshing: boolean;
  isFetching: boolean;
  t: (key: translationKey, params?: Record<string, string>) => string;
  setShowPruneImagesConfirm: React.Dispatch<React.SetStateAction<boolean>>;
  setShowPruneBuildCacheConfirm: React.Dispatch<React.SetStateAction<boolean>>;
}

export function ActionHeader({
  handleRefresh,
  isRefreshing,
  isFetching,
  t,
  setShowPruneImagesConfirm,
  setShowPruneBuildCacheConfirm
}: ActionHeaderProps) {
  return (
    <div className="flex items-center gap-2">
      <Button
        onClick={handleRefresh}
        variant="outline"
        size="sm"
        disabled={isRefreshing || isFetching}
      >
        {isRefreshing || isFetching ? (
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
        ) : (
          <RefreshCw className="mr-2 h-4 w-4" />
        )}
        {t('containers.refresh')}
      </Button>
      <AnyPermissionGuard
        permissions={['container:delete']}
        loadingFallback={<Skeleton className="h-9 w-20" />}
      >
        <Button variant="outline" size="sm" onClick={() => setShowPruneImagesConfirm(true)}>
          <Trash2 className="mr-2 h-4 w-4" />
          {t('containers.prune_images')}
        </Button>
        <Button variant="outline" size="sm" onClick={() => setShowPruneBuildCacheConfirm(true)}>
          <Scissors className="mr-2 h-4 w-4" />
          {t('containers.prune_build_cache')}
        </Button>
      </AnyPermissionGuard>
    </div>
  );
}
