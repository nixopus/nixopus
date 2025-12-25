'use client';

import React from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';
import {
  useGetAllGithubConnectorQuery,
  useDeleteGithubConnectorMutation,
  useUpdateGithubConnectorMutation
} from '@/redux/services/connector/githubConnectorApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import { setActiveConnectorId } from '@/redux/features/github-connector/githubConnectorSlice';

function useGithubConnectorSettings() {
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const [isSettingsModalOpen, setIsSettingsModalOpen] = React.useState(false);
  const [isDeleting, setIsDeleting] = React.useState<string | null>(null);
  const [isUpdating, setIsUpdating] = React.useState<string | null>(null);

  const activeConnectorId = useAppSelector(
    (state) => state.githubConnector.activeConnectorId
  );

  const {
    data: connectors,
    isLoading: isLoadingConnectors,
    refetch: refetchConnectors
  } = useGetAllGithubConnectorQuery();

  const [deleteConnector, { isLoading: isDeletingConnector }] = useDeleteGithubConnectorMutation();
  const [updateConnector, { isLoading: isUpdatingConnector }] = useUpdateGithubConnectorMutation();

  // Initialize active connector if not set and connectors are available
  React.useEffect(() => {
    if (!activeConnectorId && connectors && connectors.length > 0) {
      dispatch(setActiveConnectorId(connectors[0].id));
    }
  }, [activeConnectorId, connectors, dispatch]);

  const activeConnector = React.useMemo(() => {
    if (!activeConnectorId || !connectors) {
      return null;
    }
    return connectors.find((c) => c.id === activeConnectorId) || null;
  }, [activeConnectorId, connectors]);

  const handleSetActiveConnector = React.useCallback(
    (connectorId: string) => {
      dispatch(setActiveConnectorId(connectorId));
      // Invalidate cache to refetch repositories with new connector
      dispatch(GithubConnectorApi.util.invalidateTags([{ type: 'GithubConnector', id: 'LIST' }]));
      toast.success(t('selfHost.connectorSettings.actions.switch.success' as any));
    },
    [dispatch, t]
  );

  const handleDeleteConnector = React.useCallback(
    async (connectorId: string) => {
      if (!connectors || connectors.length <= 1) {
        toast.error(t('selfHost.connectorSettings.actions.delete.error.lastConnector' as any));
        return;
      }

      setIsDeleting(connectorId);
      try {
        await deleteConnector(connectorId).unwrap();
        toast.success(t('selfHost.connectorSettings.actions.delete.success' as any));

        // If deleted connector was active, switch to first available
        if (activeConnectorId === connectorId) {
          const remainingConnectors = connectors.filter((c) => c.id !== connectorId);
          if (remainingConnectors.length > 0) {
            handleSetActiveConnector(remainingConnectors[0].id);
          }
        }

        await refetchConnectors();
      } catch (error) {
        console.error('Failed to delete connector:', error);
        toast.error(t('selfHost.connectorSettings.actions.delete.error.generic' as any));
      } finally {
        setIsDeleting(null);
      }
    },
    [connectors, activeConnectorId, deleteConnector, refetchConnectors, handleSetActiveConnector, t]
  );

  const handleUpdateConnector = React.useCallback(
    async (connectorId: string, installationId: string) => {
      setIsUpdating(connectorId);
      try {
        await updateConnector({ 
          installation_id: installationId,
          connector_id: connectorId 
        }).unwrap();
        toast.success(t('selfHost.connectorSettings.actions.update.success' as any));
        await refetchConnectors();
        // Invalidate cache to refetch repositories
        dispatch(GithubConnectorApi.util.invalidateTags([{ type: 'GithubConnector', id: 'LIST' }]));
      } catch (error) {
        console.error('Failed to update connector:', error);
        toast.error(t('selfHost.connectorSettings.actions.update.error' as any));
      } finally {
        setIsUpdating(null);
      }
    },
    [updateConnector, refetchConnectors, dispatch, t]
  );

  const handleResetConnector = React.useCallback(
    async (connectorId: string) => {
      const connector = connectors?.find((c) => c.id === connectorId);
      if (connector?.installation_id) {
        // Open GitHub installation settings in new tab
        window.open(
          `https://github.com/settings/installations/${connector.installation_id}`,
          '_blank',
          'noopener,noreferrer'
        );
        toast.info(t('selfHost.connectorSettings.actions.reset.info' as any));
      } else {
        toast.error(t('selfHost.connectorSettings.actions.reset.error' as any));
      }
    },
    [connectors, t]
  );

  const openSettingsModal = React.useCallback(() => {
    setIsSettingsModalOpen(true);
  }, []);

  const closeSettingsModal = React.useCallback(() => {
    setIsSettingsModalOpen(false);
  }, []);

  return {
    connectors: connectors || [],
    activeConnector,
    activeConnectorId,
    isLoadingConnectors,
    isSettingsModalOpen,
    isDeleting,
    isUpdating,
    isDeletingConnector,
    isUpdatingConnector,
    handleSetActiveConnector,
    handleDeleteConnector,
    handleUpdateConnector,
    handleResetConnector,
    openSettingsModal,
    closeSettingsModal,
    refetchConnectors,
    setIsSettingsModalOpen
  };
}

export default useGithubConnectorSettings;

