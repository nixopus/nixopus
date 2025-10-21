export type ExtensionCategory =
  | 'Security'
  | 'Containers'
  | 'Database'
  | 'Web Server'
  | 'Maintenance'
  | 'Monitoring'
  | 'Storage'
  | 'Network'
  | 'Development'
  | 'Other';

export type ValidationStatus = 'not_validated' | 'valid' | 'invalid';

export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed';

export type ExtensionType = 'install' | 'run';

export interface ExtensionVariable {
  id: string;
  extension_id: string;
  variable_name: string;
  variable_type: string;
  description: string;
  default_value: unknown;
  is_required: boolean;
  validation_pattern: string;
  created_at: string;
}

export interface Extension {
  id: string;
  extension_id: string;
  parent_extension_id?: string;
  name: string;
  description: string;
  author: string;
  icon: string;
  category: ExtensionCategory;
  extension_type: ExtensionType;
  version: string;
  is_verified: boolean;
  yaml_content: string;
  parsed_content: string;
  content_hash: string;
  validation_status: ValidationStatus;
  validation_errors: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  variables?: ExtensionVariable[];
}

export interface ExtensionExecution {
  id: string;
  extension_id: string;
  server_hostname: string;
  variable_values: string;
  status: ExecutionStatus;
  started_at: string;
  completed_at?: string;
  exit_code: number;
  error_message: string;
  execution_log: string;
  created_at: string;
  extension?: Extension;
  log_seq?: number;
}

export interface ExecutionStep {
  id: string;
  execution_id: string;
  step_name: string;
  phase: string;
  step_order: number;
  started_at: string;
  completed_at?: string;
  status: ExecutionStatus;
  exit_code: number;
  output: string;
  created_at: string;
}

export interface ExtensionLog {
  id: string;
  execution_id: string;
  step_id?: string;
  level: string;
  message: string;
  data: unknown;
  sequence: number;
  created_at: string;
}

export interface ListLogsResponse {
  logs: ExtensionLog[];
  next_after: number;
}

export type SortDirection = 'asc' | 'desc';

export type ExtensionSortField =
  | 'name'
  | 'author'
  | 'category'
  | 'is_verified'
  | 'created_at'
  | 'updated_at';

export interface ExtensionListParams {
  category?: ExtensionCategory;
  type?: ExtensionType;
  search?: string;
  sort_by?: ExtensionSortField;
  sort_dir?: SortDirection;
  page?: number;
  page_size?: number;
}

export interface ExtensionListResponse {
  extensions: Extension[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
