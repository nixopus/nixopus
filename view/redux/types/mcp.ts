export interface ProviderField {
  key: string;
  label: string;
  required: boolean;
  header_name?: string;
  header_prefix?: string;
  is_query_param?: boolean;
  sensitive: boolean;
}

export interface MCPProvider {
  id: string;
  name: string;
  description: string;
  logo_url: string;
  fields: ProviderField[];
}

export interface MCPServer {
  id: string;
  org_id: string;
  provider_id: string;
  name: string;
  credentials: Record<string, string>;
  custom_url?: string;
  url: string;
  enabled: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateMCPServerRequest {
  provider_id: string;
  name: string;
  credentials: Record<string, string>;
  custom_url?: string;
  enabled: boolean;
}

export interface UpdateMCPServerRequest {
  id: string;
  name: string;
  credentials: Record<string, string>;
  custom_url?: string;
  enabled: boolean;
}

export interface DeleteMCPServerRequest {
  id: string;
}

export interface TestMCPServerRequest {
  provider_id: string;
  credentials: Record<string, string>;
  custom_url?: string;
}

export interface TestMCPServerResult {
  ok: boolean;
  error?: string;
}
