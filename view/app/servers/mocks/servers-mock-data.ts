import { Server, ClusterInfo, ServerHealthMetrics } from '@/redux/types/servers';

// Mock health metrics generator
const generateHealthMetrics = (isOnline: boolean): ServerHealthMetrics => ({
  cpu_usage: isOnline ? Math.floor(Math.random() * 60) + 10 : 0,
  memory_usage: isOnline ? Math.floor(Math.random() * 70) + 15 : 0,
  disk_usage: isOnline ? Math.floor(Math.random() * 50) + 20 : 0,
  container_count: isOnline ? Math.floor(Math.random() * 15) + 2 : 0,
  is_online: isOnline,
  last_checked_at: new Date().toISOString()
});

// Mock servers data - easily replaceable with API calls
export const MOCK_SERVERS: Server[] = [
  {
    id: 'srv-001',
    name: 'production-manager-01',
    host: '192.168.1.100',
    port: 22,
    ssh_user: 'root',
    docker_socket: '/var/run/docker.sock',
    organization_id: 'org-001',
    status: 'online',
    role: 'manager',
    health: 'healthy',
    labels: ['production', 'web', 'api'],
    cluster_id: 'cluster-001',
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-12-20T14:30:00Z',
    metrics: generateHealthMetrics(true)
  },
  {
    id: 'srv-002',
    name: 'production-worker-01',
    host: '192.168.1.101',
    port: 22,
    ssh_user: 'root',
    docker_socket: '/var/run/docker.sock',
    organization_id: 'org-001',
    status: 'online',
    role: 'worker',
    health: 'healthy',
    labels: ['production', 'worker'],
    cluster_id: 'cluster-001',
    created_at: '2024-02-10T09:00:00Z',
    updated_at: '2024-12-20T14:30:00Z',
    metrics: generateHealthMetrics(true)
  },
  {
    id: 'srv-003',
    name: 'production-worker-02',
    host: '192.168.1.102',
    port: 22,
    ssh_user: 'admin',
    docker_socket: '/var/run/docker.sock',
    organization_id: 'org-001',
    status: 'online',
    role: 'worker',
    health: 'degraded',
    labels: ['production', 'database'],
    cluster_id: 'cluster-001',
    created_at: '2024-03-05T11:00:00Z',
    updated_at: '2024-12-20T14:30:00Z',
    metrics: { ...generateHealthMetrics(true), cpu_usage: 85, memory_usage: 78 }
  },
  {
    id: 'srv-004',
    name: 'staging-server',
    host: '192.168.2.50',
    port: 22,
    ssh_user: 'deploy',
    docker_socket: '/var/run/docker.sock',
    organization_id: 'org-001',
    status: 'online',
    role: 'standalone',
    health: 'healthy',
    labels: ['staging', 'test'],
    created_at: '2024-04-20T08:00:00Z',
    updated_at: '2024-12-19T16:00:00Z',
    metrics: generateHealthMetrics(true)
  },
  {
    id: 'srv-005',
    name: 'dev-server',
    host: '192.168.3.10',
    port: 2222,
    ssh_user: 'developer',
    docker_socket: '/var/run/docker.sock',
    organization_id: 'org-001',
    status: 'offline',
    role: 'standalone',
    health: 'unknown',
    labels: ['development'],
    created_at: '2024-05-01T12:00:00Z',
    updated_at: '2024-12-18T10:00:00Z',
    metrics: generateHealthMetrics(false)
  },
  {
    id: 'srv-006',
    name: 'backup-worker',
    host: '192.168.1.200',
    port: 22,
    ssh_user: 'root',
    docker_socket: '/var/run/docker.sock',
    organization_id: 'org-001',
    status: 'connecting',
    role: 'worker',
    health: 'unknown',
    labels: ['backup', 'disaster-recovery'],
    cluster_id: 'cluster-001',
    created_at: '2024-06-15T14:00:00Z',
    updated_at: '2024-12-20T14:35:00Z',
    metrics: generateHealthMetrics(false)
  }
];

export const MOCK_CLUSTER: ClusterInfo = {
  id: 'cluster-001',
  name: 'Production Cluster',
  manager_count: 1,
  worker_count: 3,
  is_initialized: true,
  created_at: '2024-01-15T10:00:00Z'
};

// Helper to get servers - can be replaced with API call
export const getServers = (): Promise<Server[]> => {
  return new Promise((resolve) => {
    setTimeout(() => resolve(MOCK_SERVERS), 500);
  });
};

// Helper to get cluster info - can be replaced with API call
export const getClusterInfo = (): Promise<ClusterInfo | null> => {
  return new Promise((resolve) => {
    setTimeout(() => resolve(MOCK_CLUSTER), 300);
  });
};

// Helper to add server - can be replaced with API call
export const addServer = (
  server: Omit<Server, 'id' | 'created_at' | 'updated_at' | 'metrics'>
): Promise<Server> => {
  return new Promise((resolve) => {
    const newServer: Server = {
      ...server,
      id: `srv-${Date.now()}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metrics: generateHealthMetrics(server.status === 'online')
    };
    setTimeout(() => resolve(newServer), 500);
  });
};

// Helper to test connection - can be replaced with API call
export const testConnection = (
  serverId: string
): Promise<{ success: boolean; message: string }> => {
  return new Promise((resolve) => {
    const success = Math.random() > 0.2;
    setTimeout(
      () =>
        resolve({
          success,
          message: success
            ? 'Connection successful'
            : 'Connection failed: Unable to establish SSH connection'
        }),
      1500
    );
  });
};
