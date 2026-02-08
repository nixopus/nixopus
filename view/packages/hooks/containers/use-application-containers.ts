'use client';

import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useMemo, useState } from 'react';
import { toast } from 'sonner';
import {
  useRemoveContainerMutation,
  useStartContainerMutation,
  useStopContainerMutation,
  useGetContainersQuery,
  Container
} from '@/redux/services/container/containerApi';

export function useApplicationContainers(applicationId: string) {
  const { t } = useTranslation();

  // Fetch all containers with grouping enabled
  const { data, isLoading, refetch, isFetching } = useGetContainersQuery(
    {
      page: 1,
      page_size: 100, // Large page size to get all containers for the app
      sort_by: 'name',
      sort_order: 'asc'
    },
    { refetchOnMountOrArgChange: true }
  );

  const [startContainer] = useStartContainerMutation();
  const [stopContainer] = useStopContainerMutation();
  const [removeContainer] = useRemoveContainerMutation();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [containerToDelete, setContainerToDelete] = useState<string | null>(null);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await refetch();
    } finally {
      setIsRefreshing(false);
    }
  };

  // Filter containers for this specific application
  const applicationGroup = useMemo(() => {
    return data?.groups?.find((group) => group.application_id === applicationId);
  }, [data?.groups, applicationId]);

  const containers = useMemo(() => {
    return applicationGroup?.containers ?? [];
  }, [applicationGroup]);

  const totalCount = containers.length;

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

  return {
    containers,
    isLoading,
    isFetching,
    refetch,
    handleRefresh,
    handleContainerAction,
    handleDeleteConfirm,
    isRefreshing,
    t,
    containerToDelete,
    setContainerToDelete,
    totalCount
  };
}
