import { useCallback } from 'react';

export function useLogsActions(
  onLoadMore: () => void | Promise<void>,
  onRefresh?: () => void | Promise<void>,
  setIsLoadingMore?: (loading: boolean) => void,
  setIsRefreshing?: (loading: boolean) => void
) {
  const handleLoadMore = useCallback(async () => {
    if (setIsLoadingMore) setIsLoadingMore(true);
    try {
      await Promise.resolve(onLoadMore());
    } finally {
      if (setIsLoadingMore) setIsLoadingMore(false);
    }
  }, [onLoadMore, setIsLoadingMore]);

  const handleRefresh = useCallback(async () => {
    if (!onRefresh) return;
    if (setIsRefreshing) setIsRefreshing(true);
    try {
      await Promise.resolve(onRefresh());
    } finally {
      if (setIsRefreshing) setIsRefreshing(false);
    }
  }, [onRefresh, setIsRefreshing]);

  return {
    handleLoadMore,
    handleRefresh
  };
}
