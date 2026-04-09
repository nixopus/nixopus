export type ProvisionStep =
  | 'INITIALIZING'
  | 'CREATING_CONTAINER'
  | 'SETUP_NETWORKING'
  | 'INSTALLING_DEPENDENCIES'
  | 'CONFIGURING_SSH'
  | 'SETUP_SSH_FORWARDING'
  | 'VERIFYING_SSH'
  | 'COMPLETED';

export interface ServerProvision {
  id: string;
  user_id: string;
  organization_id: string;
  ssh_key_id: string;
  step?: ProvisionStep;
  status?: 'NOT_STARTED' | 'PROVISIONING' | 'ACTIVE' | 'FAILED';
  lxd_container_name: string | null;
  subdomain: string | null;
  domain: string | null;
  created_at: string;
  updated_at: string;
}

export interface Server {
  id: string;
  organization_id: string;
  name: string;
  host: string | null;
  user: string | null;
  port: number;
  is_active: boolean;
  is_default: boolean;
  total_vcpu: number;
  total_ram_mb: number;
  total_disk_gb: number;
  created_at: string;
  updated_at: string;
  provision: ServerProvision | null;
}

export interface GetServersResponse {
  servers: Server[];
  total_count: number;
  page: number;
  page_size: number;
  sort_by: string;
  sort_order: string;
  search: string;
  status: string;
  is_active: boolean | null;
}

export interface GetServersParams {
  page?: number;
  page_size?: number;
  search?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
  status?: 'NOT_STARTED' | 'PROVISIONING' | 'ACTIVE' | 'FAILED';
  is_active?: boolean;
}

export interface CreateMachineRequest {
  name: string;
  host: string;
  port?: number;
  user?: string;
}

export interface CreateMachineResponse {
  id: string;
  name: string;
  host: string;
  port: number;
  user: string;
  public_key: string;
}

export interface ProvisionMachineRequest {
  name?: string;
}

export interface ProvisionStatusResponse {
  provision_id: string;
  step: string;
  status: string;
  error?: string;
}

export interface MachineVerifyResponse {
  status: string;
}

export interface MachineSshStatusResponse {
  is_active: boolean;
  last_used_at?: string;
}

export interface DeleteMachineResponse {
  status: string;
}
