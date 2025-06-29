import { useRouter } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import React, { useState } from 'react';
import { toast } from 'sonner';
import {
  useRemoveContainerMutation,
  useStartContainerMutation,
  useStopContainerMutation
} from '@/redux/services/container/containerApi';
import { useGetContainersQuery } from '@/redux/services/container/containerApi';
import { useFeatureFlags } from '@/hooks/features_provider';
import { usePruneBuildCacheMutation } from '@/redux/services/container/imagesApi';
import { usePruneImagesMutation } from '@/redux/services/container/imagesApi';

function useContainerList() {
  const { t } = useTranslation();
  const router = useRouter();
  const { data: containers = [], isLoading, refetch } = useGetContainersQuery();
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

  const getGradientFromName = (name: string) => {
    const colors = [
      'from-blue-500/20 to-purple-500/20',
      'from-green-500/20 to-teal-500/20',
      'from-yellow-500/20 to-orange-500/20',
      'from-red-500/20 to-pink-500/20',
      'from-indigo-500/20 to-violet-500/20',
      'from-emerald-500/20 to-cyan-500/20'
    ];
    const index = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length;
    return colors[index];
  };

  return {
    containers,
    isLoading,
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
    getGradientFromName
  };
}

export default useContainerList;
