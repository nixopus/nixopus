'use client';

import { useRouter } from 'next/navigation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import {
  useRemoveContainerMutation,
  useStartContainerMutation,
  useStopContainerMutation,
  useGetContainersQuery,
  Container,
  ContainerGroup
} from '@/redux/services/container/containerApi';
import { useFeatureFlags } from '@/packages/hooks/shared/features_provider';
import { usePruneBuildCacheMutation } from '@/redux/services/container/imagesApi';
import { usePruneImagesMutation } from '@/redux/services/container/imagesApi';
import { SelectOption } from '@/components/ui/select-wrapper';

export function useContainers() {
  const { t } = useTranslation();
  const router = useRouter();
  // Params state for pagination, sorting, and search
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [sortConfig, setSortConfig] = useState<{
    key: 'name' | 'status';
    direction: 'asc' | 'desc';
  }>({
    key: 'name',
    direction: 'asc'
  });

  const { data, isLoading, refetch, isFetching } = useGetContainersQuery(
    {
      page,
      page_size: pageSize,
      search,
      sort_by: sortConfig.key,
      sort_order: sortConfig.direction
    },
    { refetchOnMountOrArgChange: true }
  );
  const [startContainer] = useStartContainerMutation();
  const [stopContainer] = useStopContainerMutation();
  const [removeContainer] = useRemoveContainerMutation();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [containerToDelete, setContainerToDelete] = useState<string | null>(null);
  const [showPruneImagesConfirm, setShowPruneImagesConfirm] = useState(false);
  const [showPruneBuildCacheConfirm, setShowPruneBuildCacheConfirm] = useState(false);
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  const [pruneImages] = usePruneImagesMutation();
  const [pruneBuildCache] = usePruneBuildCacheMutation();

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await refetch();
    } finally {
      setIsRefreshing(false);
    }
  };

  // Debounce search input to avoid fetching on every keystroke
  useEffect(() => {
    const handle = setTimeout(() => {
      setSearch(searchInput);
      setPage(1);
    }, 300);
    return () => clearTimeout(handle);
  }, [searchInput]);

  const handleSearchChange = (value: string) => {
    setSearchInput(value);
    setPage(1);
  };

  const handleSortChange = (key: 'name' | 'status', direction: 'asc' | 'desc') => {
    setSortConfig({ key, direction });
    setPage(1);
  };

  const sortOptions: SelectOption[] = [
    { value: 'name_asc', label: t('containers.sortOptions.name') + ' (A-Z)' },
    { value: 'name_desc', label: t('containers.sortOptions.nameDesc') },
    { value: 'status_asc', label: t('containers.sortOptions.status') + ' (A-Z)' },
    { value: 'status_desc', label: t('containers.sortOptions.statusDesc') }
  ];

  // Keep previous data to avoid page flash on param changes
  const [lastData, setLastData] = useState<
    | {
        containers?: Container[];
        groups?: ContainerGroup[];
        ungrouped?: Container[];
        total_count: number;
        group_count?: number;
        page: number;
        page_size: number;
      }
    | undefined
  >(undefined);
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (data) {
      setLastData(data);
      if (!initialized) setInitialized(true);
    }
  }, [data]);

  const effectiveData = data ?? lastData;

  // Flatten groups and ungrouped for backward compatibility and stats
  const containers = useMemo(() => {
    const allContainers: Container[] = [];
    if (effectiveData?.groups) {
      for (const group of effectiveData.groups) {
        allContainers.push(...group.containers);
      }
    }
    if (effectiveData?.ungrouped) {
      allContainers.push(...effectiveData.ungrouped);
    }
    // Fallback to containers array for backward compatibility
    if (allContainers.length === 0 && effectiveData?.containers) {
      return effectiveData.containers;
    }
    return allContainers;
  }, [effectiveData]);

  const totalCount = effectiveData?.total_count ?? 0;
  // Paginate by groups, so total pages = group_count / page_size
  const groupCount = effectiveData?.group_count ?? 0;
  const totalPages =
    groupCount > 0
      ? Math.max(1, Math.ceil(groupCount / pageSize))
      : Math.max(1, Math.ceil(totalCount / pageSize));

  const runningCount = useMemo(
    () => containers.filter((c) => c.status === 'running').length,
    [containers]
  );
  const stoppedCount = useMemo(
    () => containers.filter((c) => c.status !== 'running').length,
    [containers]
  );

  const handleContainerAction = async (
    containerId: string,
    action: 'start' | 'stop' | 'remove'
  ) => {
    try {
      switch (action) {
        case 'start':
          await startContainer(containerId).unwrap();
          toast.success(t(`containers.${action}_success`));
          break;
        case 'stop':
          await stopContainer(containerId).unwrap();
          toast.success(t(`containers.${action}_success`));
          break;
        case 'remove':
          setContainerToDelete(containerId);
          break;
      }
    } catch (error) {
      toast.error(t(`containers.${action}_error`));
    }
  };

  const handleDeleteConfirm = async () => {
    if (!containerToDelete) return;
    try {
      await removeContainer(containerToDelete).unwrap();
      toast.success(t('containers.remove_success'));
      setContainerToDelete(null);
    } catch (error) {
      toast.error(t('containers.remove_error'));
    }
  };

  const handlePruneImages = async () => {
    try {
      await pruneImages({
        dangling: true
      }).unwrap();
      toast.success(t('containers.prune_images_success'));
    } catch (error) {
      toast.error(t('containers.prune_images_error'));
    }
  };

  const handlePruneBuildCache = async () => {
    try {
      await pruneBuildCache({
        all: true
      }).unwrap();
      toast.success(t('containers.prune_build_cache_success'));
    } catch (error) {
      toast.error(t('containers.prune_build_cache_error'));
    }
  };

  return {
    containers,
    groups: effectiveData?.groups ?? [],
    ungrouped: effectiveData?.ungrouped ?? [],
    isLoading,
    isFetching,
    initialized,
    refetch,
    handleRefresh,
    handleContainerAction,
    handleDeleteConfirm,
    handlePruneImages,
    handlePruneBuildCache,
    showPruneImagesConfirm,
    showPruneBuildCacheConfirm,
    setShowPruneImagesConfirm,
    setShowPruneBuildCacheConfirm,
    isFeatureFlagsLoading,
    isRefreshing,
    isFeatureEnabled,
    t,
    router,
    containerToDelete,
    setContainerToDelete,
    page,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    totalCount,
    runningCount,
    stoppedCount,
    searchInput,
    handleSearchChange,
    sortConfig,
    handleSortChange,
    sortOptions
  };
}
