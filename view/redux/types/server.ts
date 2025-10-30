export type ServerStatus = 'active' | 'inactive' | 'maintenance';

export interface Server {
  id: string;
  name: string;
  description: string;
  host: string;
  port: number;
  username: string;
  ssh_password?: string;
  ssh_private_key_path?: string;
  status: ServerStatus;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  user_id: string;
  organization_id: string;
}

export interface CreateServerRequest {
  name: string;
  description: string;
  host: string;
  port: number;
  username: string;
  ssh_password?: string;
  ssh_private_key_path?: string;
  status?: ServerStatus;
}

export interface UpdateServerRequest {
  id: string;
  name: string;
  description: string;
  host: string;
  port: number;
  username: string;
  ssh_password?: string;
  ssh_private_key_path?: string;
  status?: ServerStatus;
}

export interface UpdateServerStatusRequest {
  id: string;
  status: ServerStatus;
}

export interface DeleteServerRequest {
  id: string;
}

export interface CreateServerResponse {
  id: string;
}

export interface GetServersRequest {
  page?: number;
  page_size?: number;
  search?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface Pagination {
  current_page: number;
  page_size: number;
  total_pages: number;
  total_items: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ServerListResponse {
  servers: Server[];
  pagination: Pagination;
}

export enum AuthenticationType {
  PASSWORD = 'password',
  PRIVATE_KEY = 'private_key'
}
