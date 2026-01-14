export type HealthCheck = {
  id: string;
  application_id: string;
  organization_id: string;
  enabled: boolean;
  endpoint: string;
  method: 'GET' | 'POST' | 'HEAD';
  expected_status_codes: number[];
  timeout_seconds: number;
  interval_seconds: number;
  failure_threshold: number;
  success_threshold: number;
  headers?: Record<string, string>;
  body?: string;
  consecutive_fails: number;
  last_checked_at?: string;
  last_error_message?: string;
  retention_days: number;
  created_at: string;
  updated_at: string;
};

export type HealthCheckResult = {
  id: string;
  health_check_id: string;
  status: 'healthy' | 'unhealthy' | 'timeout' | 'error';
  response_time_ms: number;
  status_code?: number;
  error_message?: string;
  checked_at: string;
};

export type HealthCheckStats = {
  total_checks: number;
  successful_checks: number;
  failed_checks: number;
  avg_response_time_ms: number;
  uptime_percentage: number;
  period?: string;
  last_status?: string;
  last_checked_at?: string;
};

export type CreateHealthCheckRequest = {
  application_id: string;
  endpoint?: string;
  method?: 'GET' | 'POST' | 'HEAD';
  expected_status_codes?: number[];
  timeout_seconds?: number;
  interval_seconds?: number;
  failure_threshold?: number;
  success_threshold?: number;
  headers?: Record<string, string>;
  body?: string;
  retention_days?: number;
};

export type UpdateHealthCheckRequest = {
  application_id: string;
  endpoint?: string;
  method?: 'GET' | 'POST' | 'HEAD';
  expected_status_codes?: number[];
  timeout_seconds?: number;
  interval_seconds?: number;
  failure_threshold?: number;
  success_threshold?: number;
  headers?: Record<string, string>;
  body?: string;
  retention_days?: number;
};

export type ToggleHealthCheckRequest = {
  application_id: string;
  enabled: boolean;
};

export type GetHealthCheckResultsRequest = {
  application_id: string;
  limit?: number;
  start_time?: string;
  end_time?: string;
};

export type GetHealthCheckStatsRequest = {
  application_id: string;
  period?: string;
};
