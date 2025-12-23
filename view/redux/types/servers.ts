export type ServerStatus = 'online' | 'offline' | 'connecting' | 'error';
export type ServerRole = 'manager' | 'worker' | 'standalone';
export type ServerHealth = 'healthy' | 'degraded' | 'critical' | 'unknown';

export interface ServerHealthMetrics {
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  container_count: number;
  is_online: boolean;
  last_checked_at: string;
}

export interface Server {
  id: string;
  name: string;
  host: string;
  port: number;
  ssh_user: string;
  docker_socket: string;
  organization_id: string;
  status: ServerStatus;
  role: ServerRole;
  health: ServerHealth;
  labels: string[];
  swarm_token?: string;
  cluster_id?: string;
  created_at: string;
  updated_at: string;
  metrics?: ServerHealthMetrics;
}

export interface ClusterInfo {
  id: string;
  name: string;
  manager_count: number;
  worker_count: number;
  is_initialized: boolean;
  created_at: string;
}

export interface CreateServerRequest {
  name: string;
  host: string;
  port: number;
  ssh_user: string;
  ssh_private_key?: string;
  ssh_password?: string;
  docker_socket: string;
  labels?: string[];
}

export interface UpdateServerRequest {
  name?: string;
  host?: string;
  port?: number;
  ssh_user?: string;
  ssh_private_key?: string;
  ssh_password?: string;
  docker_socket?: string;
  labels?: string[];
}

export interface TestConnectionResponse {
  success: boolean;
  message: string;
  ssh_connected: boolean;
  docker_connected: boolean;
}

export interface ClusterJoinToken {
  worker_token: string;
  manager_token: string;
}
