'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import {
  Server,
  ClusterInfo,
  CreateServerRequest,
  ServerStatus,
  ServerRole,
  ServerHealth
} from '@/redux/types/servers';
import {
  MOCK_SERVERS,
  MOCK_CLUSTER,
  testConnection as mockTestConnection
} from '../mocks/servers-mock-data';

interface UseServersReturn {
  servers: Server[];
  cluster: ClusterInfo | null;
  isLoading: boolean;
  selectedServer: Server | null;
  setSelectedServer: (server: Server | null) => void;
  viewMode: 'list' | 'topology';
  setViewMode: (mode: 'list' | 'topology') => void;
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  filteredServers: Server[];
  stats: {
    total: number;
    online: number;
    offline: number;
    managers: number;
    workers: number;
  };
  handleAddServer: (data: CreateServerRequest) => Promise<void>;
  handleUpdateServer: (id: string, data: Partial<Server>) => Promise<void>;
  handleDeleteServer: (id: string) => Promise<void>;
  handleTestConnection: (id: string) => Promise<{ success: boolean; message: string }>;
  handleInitCluster: (serverId: string) => Promise<void>;
  handleJoinCluster: (serverId: string, role: ServerRole) => Promise<void>;
  handleLeaveCluster: (serverId: string) => Promise<void>;
  isAddDialogOpen: boolean;
  setIsAddDialogOpen: (open: boolean) => void;
  isDeleteDialogOpen: boolean;
  setIsDeleteDialogOpen: (open: boolean) => void;
  serverToDelete: Server | null;
  setServerToDelete: (server: Server | null) => void;
  isTestingConnection: boolean;
  t: ReturnType<typeof useTranslation>['t'];
}

export function useServers(): UseServersReturn {
  const { t } = useTranslation();

  // State
  const [servers, setServers] = useState<Server[]>([]);
  const [cluster, setCluster] = useState<ClusterInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedServer, setSelectedServer] = useState<Server | null>(null);
  const [viewMode, setViewMode] = useState<'list' | 'topology'>('topology');
  const [searchQuery, setSearchQuery] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<Server | null>(null);
  const [isTestingConnection, setIsTestingConnection] = useState(false);

  // Load initial data
  useEffect(() => {
    const loadData = async () => {
      try {
        // TODO: Replace with actual API calls
        // const serversData = await serversApi.getServers().unwrap();
        // const clusterData = await serversApi.getCluster().unwrap();

        // Using mock data for now
        await new Promise((resolve) => setTimeout(resolve, 800));
        setServers(MOCK_SERVERS);
        setCluster(MOCK_CLUSTER);
      } catch {
        toast.error('Failed to load servers');
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, []);

  // Filtered servers based on search query
  const filteredServers = useMemo(() => {
    if (!searchQuery.trim()) return servers;

    const query = searchQuery.toLowerCase();
    return servers.filter(
      (server) =>
        server.name.toLowerCase().includes(query) ||
        server.host.toLowerCase().includes(query) ||
        server.labels.some((label) => label.toLowerCase().includes(query))
    );
  }, [servers, searchQuery]);

  // Calculate stats
  const stats = useMemo(
    () => ({
      total: servers.length,
      online: servers.filter((s) => s.status === 'online').length,
      offline: servers.filter((s) => s.status === 'offline' || s.status === 'error').length,
      managers: servers.filter((s) => s.role === 'manager').length,
      workers: servers.filter((s) => s.role === 'worker').length
    }),
    [servers]
  );

  // Handlers
  const handleAddServer = useCallback(
    async (data: CreateServerRequest) => {
      try {
        // TODO: Replace with actual API call
        // const newServer = await serversApi.createServer(data).unwrap();

        const newServer: Server = {
          id: `srv-${Date.now()}`,
          ...data,
          organization_id: 'org-001',
          status: 'connecting',
          role: 'standalone',
          health: 'unknown',
          labels: data.labels || [],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };

        setServers((prev) => [...prev, newServer]);
        setIsAddDialogOpen(false);
        toast.success(t('servers.messages.addSuccess'));
      } catch (error) {
        toast.error(t('servers.messages.addError'));
      }
    },
    [t]
  );

  const handleUpdateServer = useCallback(
    async (id: string, data: Partial<Server>) => {
      try {
        // TODO: Replace with actual API call
        // await serversApi.updateServer({ id, ...data }).unwrap();

        setServers((prev) =>
          prev.map((server) =>
            server.id === id ? { ...server, ...data, updated_at: new Date().toISOString() } : server
          )
        );
        toast.success(t('servers.messages.updateSuccess'));
      } catch (error) {
        toast.error(t('servers.messages.updateError'));
      }
    },
    [t]
  );

  const handleDeleteServer = useCallback(
    async (id: string) => {
      try {
        // TODO: Replace with actual API call
        // await serversApi.deleteServer(id).unwrap();

        setServers((prev) => prev.filter((server) => server.id !== id));
        setIsDeleteDialogOpen(false);
        setServerToDelete(null);
        toast.success(t('servers.messages.deleteSuccess'));
      } catch (error) {
        toast.error(t('servers.messages.deleteError'));
      }
    },
    [t]
  );

  const handleTestConnection = useCallback(
    async (id: string) => {
      setIsTestingConnection(true);
      try {
        // TODO: Replace with actual API call
        // const result = await serversApi.testConnection(id).unwrap();

        const result = await mockTestConnection(id);

        if (result.success) {
          toast.success(t('servers.messages.connectionSuccess'));
          // Update server status
          setServers((prev) =>
            prev.map((server) =>
              server.id === id
                ? { ...server, status: 'online' as ServerStatus, health: 'healthy' as ServerHealth }
                : server
            )
          );
        } else {
          toast.error(result.message);
        }

        return result;
      } catch (error) {
        const errorResult = { success: false, message: t('servers.messages.connectionError') };
        toast.error(errorResult.message);
        return errorResult;
      } finally {
        setIsTestingConnection(false);
      }
    },
    [t]
  );

  const handleInitCluster = useCallback(
    async (serverId: string) => {
      try {
        // TODO: Replace with actual API call
        // await serversApi.initCluster(serverId).unwrap();

        const server = servers.find((s) => s.id === serverId);
        if (server) {
          setServers((prev) =>
            prev.map((s) =>
              s.id === serverId
                ? { ...s, role: 'manager' as ServerRole, cluster_id: 'cluster-001' }
                : s
            )
          );
          setCluster({
            id: 'cluster-001',
            name: 'New Cluster',
            manager_count: 1,
            worker_count: 0,
            is_initialized: true,
            created_at: new Date().toISOString()
          });
        }
        toast.success(t('servers.messages.clusterInitSuccess'));
      } catch (error) {
        toast.error(t('servers.messages.clusterInitError'));
      }
    },
    [servers, t]
  );

  const handleJoinCluster = useCallback(
    async (serverId: string, role: ServerRole) => {
      try {
        // TODO: Replace with actual API call
        // await serversApi.joinCluster({ serverId, role }).unwrap();

        setServers((prev) =>
          prev.map((s) => (s.id === serverId ? { ...s, role, cluster_id: 'cluster-001' } : s))
        );

        if (cluster) {
          setCluster({
            ...cluster,
            manager_count: cluster.manager_count + (role === 'manager' ? 1 : 0),
            worker_count: cluster.worker_count + (role === 'worker' ? 1 : 0)
          });
        }
        toast.success(t('servers.messages.joinSuccess'));
      } catch (error) {
        toast.error(t('servers.messages.joinError'));
      }
    },
    [cluster, t]
  );

  const handleLeaveCluster = useCallback(
    async (serverId: string) => {
      try {
        // TODO: Replace with actual API call
        // await serversApi.leaveCluster(serverId).unwrap();

        const server = servers.find((s) => s.id === serverId);
        setServers((prev) =>
          prev.map((s) =>
            s.id === serverId
              ? { ...s, role: 'standalone' as ServerRole, cluster_id: undefined }
              : s
          )
        );

        if (cluster && server) {
          setCluster({
            ...cluster,
            manager_count: cluster.manager_count - (server.role === 'manager' ? 1 : 0),
            worker_count: cluster.worker_count - (server.role === 'worker' ? 1 : 0)
          });
        }
        toast.success(t('servers.messages.leaveSuccess'));
      } catch (error) {
        toast.error(t('servers.messages.leaveError'));
      }
    },
    [cluster, servers, t]
  );

  return {
    servers,
    cluster,
    isLoading,
    selectedServer,
    setSelectedServer,
    viewMode,
    setViewMode,
    searchQuery,
    setSearchQuery,
    filteredServers,
    stats,
    handleAddServer,
    handleUpdateServer,
    handleDeleteServer,
    handleTestConnection,
    handleInitCluster,
    handleJoinCluster,
    handleLeaveCluster,
    isAddDialogOpen,
    setIsAddDialogOpen,
    isDeleteDialogOpen,
    setIsDeleteDialogOpen,
    serverToDelete,
    setServerToDelete,
    isTestingConnection,
    t
  };
}
