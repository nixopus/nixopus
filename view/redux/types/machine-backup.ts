export interface MachineBackup {
  id: string;
  user_id: string;
  organization_id: string;
  provision_id: string | null;
  machine_name: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  trigger: string;
  snapshot_path: string | null;
  s3_path: string | null;
  size_bytes: number;
  error: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface TriggerBackupResponse {
  status: string;
  message: string;
  request_id: string;
}

export interface BackupListParams {
  page?: number;
  page_size?: number;
  search?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
  status?: string;
  server_id?: string;
}

export interface BackupListResponseData {
  backups: MachineBackup[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface BackupListResponse {
  status: string;
  message: string;
  data: BackupListResponseData;
}

export interface BackupScheduleData {
  enabled: boolean;
  frequency: 'daily' | 'weekly';
  hour_utc: number;
  day_of_week: number;
  retention_count: number;
}

export interface BackupScheduleResponse {
  status: string;
  message: string;
  data: BackupScheduleData;
}
