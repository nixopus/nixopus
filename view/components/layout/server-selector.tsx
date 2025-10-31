'use client';

import React, { useEffect } from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Server } from '@/redux/types/server';
import {
  useGetAllServersQuery,
  useGetActiveServerQuery,
  useUpdateServerStatusMutation
} from '@/redux/services/settings/serversApi';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { setActiveServer, clearActiveServer } from '@/redux/features/server/serverSlice';

import { authApi } from '@/redux/services/users/authApi';
import { userApi } from '@/redux/services/users/userApi';
import { notificationApi } from '@/redux/services/settings/notificationApi';
import { domainsApi } from '@/redux/services/settings/domainsApi';
import { serversApi } from '@/redux/services/settings/serversApi';
import { GithubConnectorApi } from '@/redux/services/connector/githubConnectorApi';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { fileManagersApi } from '@/redux/services/file-manager/fileManagersApi';
import { auditApi } from '@/redux/services/audit';
import { FeatureFlagsApi } from '@/redux/services/feature-flags/featureFlagsApi';
import { containerApi } from '@/redux/services/container/containerApi';
import { imagesApi } from '@/redux/services/container/imagesApi';
import { Skeleton } from '@/components/ui/skeleton';
import { ServerIcon } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/hooks/use-translation';

interface ServerSelectorProps {
  className?: string;
}

const getStatusLabel = (status: string, t: (key: string) => string) => {
  switch (status) {
    case 'active':
      return t('servers.selector.status.active');
    case 'maintenance':
      return t('servers.selector.status.maintenance');
    case 'inactive':
    default:
      return t('servers.selector.status.inactive');
  }
};

export function ServerSelector({ className }: ServerSelectorProps) {
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const activeServer = useAppSelector((state) => state.server.activeServer);

  const { data: serverResponse, isLoading: isServersLoading } = useGetAllServersQuery({
    page: 1,
    page_size: 100,
    sort_by: 'name',
    sort_order: 'asc'
  });

  const { data: apiActiveServer, isLoading: isActiveServerLoading } = useGetActiveServerQuery();

  const [updateServerStatus] = useUpdateServerStatusMutation();

  const servers = serverResponse?.servers || [];

  useEffect(() => {
    if (apiActiveServer && (!activeServer || activeServer.id !== apiActiveServer.id)) {
      dispatch(setActiveServer(apiActiveServer));
    } else if (!apiActiveServer && activeServer) {
      dispatch(clearActiveServer());
    }
  }, [apiActiveServer, activeServer, dispatch]);

  const invalidateAllCaches = () => {
    try {
      dispatch(authApi.util.invalidateTags([{ type: 'Authentication', id: 'LIST' }]));
      dispatch(userApi.util.invalidateTags([{ type: 'User', id: 'LIST' }]));
      dispatch(notificationApi.util.invalidateTags([{ type: 'Notification', id: 'LIST' }]));
      dispatch(domainsApi.util.invalidateTags([{ type: 'Domains', id: 'LIST' }]));
      dispatch(serversApi.util.invalidateTags([{ type: 'Servers', id: 'LIST' }, { type: 'Servers', id: 'ACTIVE' }]));
      dispatch(GithubConnectorApi.util.invalidateTags([{ type: 'GithubConnector', id: 'LIST' }]));
      dispatch(deployApi.util.invalidateTags([{ type: 'Deploy', id: 'LIST' }, { type: 'Applications', id: 'LIST' }]));
      dispatch(fileManagersApi.util.invalidateTags([{ type: 'FileListAll', id: 'LIST' }]));
      dispatch(auditApi.util.invalidateTags([{ type: 'AuditLogs', id: 'LIST' }]));
      dispatch(FeatureFlagsApi.util.invalidateTags([{ type: 'FeatureFlags', id: 'LIST' }]));
      dispatch(containerApi.util.invalidateTags([{ type: 'Container', id: 'LIST' }]));
      dispatch(imagesApi.util.invalidateTags(['Images']));
    } catch (error) {
      console.error('Failed to invalidate cache:', error);
    }
  };

  const handleServerChange = async (value: string) => {
    if (value === 'default') {
      if (activeServer) {
        try {
          await updateServerStatus({
            id: activeServer.id,
            status: 'inactive'
          }).unwrap();
          dispatch(clearActiveServer());
          // hack to invalidate caches instantly..
          invalidateAllCaches();
          invalidateAllCaches();
        } catch (error) {
          console.error('Failed to deactivate server:', error);
        }
      }
    } else {
      const selectedServer = servers.find((server: Server) => server.id === value);
      if (selectedServer) {
        try {
          await updateServerStatus({
            id: selectedServer.id,
            status: 'active'
          }).unwrap();
          // hack to invalidate caches instantly..
          invalidateAllCaches();
          invalidateAllCaches();
        } catch (error) {
          console.error('Failed to activate server:', error);
        }
      }
    }
  };

  if (isServersLoading || isActiveServerLoading) {
    return <Skeleton className="h-9 w-[200px]" />;
  }

  if (!servers.length) {
    return null;
  }

  return (
    <Select
      value={activeServer?.id || 'default'}
      onValueChange={handleServerChange}
    >
      <SelectTrigger className={cn('w-[200px]', className)}>
        <div className="flex items-center gap-2">
          <ServerIcon className="h-4 w-4" />
          <SelectValue placeholder={t('servers.selector.placeholder')}>
            {activeServer ? (
              <span>{activeServer.name}</span>
            ) : (
              t('servers.selector.default')
            )}
          </SelectValue>
        </div>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="default">
          <div className="flex items-center gap-2">
            <span>{t('servers.selector.default')}</span>
          </div>
        </SelectItem>
        {servers.map((server: Server) => (
          <SelectItem key={server.id} value={server.id}>
            <div className="flex items-center gap-2">
              <span>{server.name}</span>
              <span className="text-xs text-muted-foreground">{getStatusLabel(server.status, t)}</span>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
