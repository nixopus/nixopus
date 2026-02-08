'use client';

import { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useGetContainerQuery,
  useStartContainerMutation,
  useStopContainerMutation,
  useRemoveContainerMutation
} from '@/redux/services/container/containerApi';

export function useInlineContainerDetail(containerId: string, onBack: () => void) {
  const { t } = useTranslation();
  const { data: container, isLoading } = useGetContainerQuery(containerId);
  const [startContainer] = useStartContainerMutation();
  const [stopContainer] = useStopContainerMutation();
  const [removeContainer] = useRemoveContainerMutation();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  const handleContainerAction = async (action: 'start' | 'stop' | 'remove' | 'restart') => {
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
        case 'restart':
          await stopContainer(containerId).unwrap();
          await startContainer(containerId).unwrap();
          toast.success(t('containers.restart_success'));
          break;
        case 'remove':
          setIsDeleteDialogOpen(true);
          break;
      }
    } catch (error) {
      if (action === 'restart') {
        toast.error(t('containers.restart_error'));
      } else {
        toast.error(t(`containers.${action}_error`));
      }
    }
  };

  const handleDeleteConfirm = async () => {
    try {
      await removeContainer(containerId).unwrap();
      toast.success(t('containers.remove_success'));
      onBack();
    } catch (error) {
      toast.error(t('containers.remove_error'));
    }
  };

  return {
    container,
    isLoading,
    containerId,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen,
    handleContainerAction,
    handleDeleteConfirm,
    t
  };
}
