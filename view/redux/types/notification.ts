export interface SMTPConfig {
  id: string;
  host: string;
  port: number;
  username: string;
  password: string;
  from_email: string;
  from_name: string;
  security: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  user_id: string;
}


export interface CreateSMTPConfigRequest {
  host: string;
  port: number;
  username: string;
  password: string;
  from_name: string;
  from_email: string;
}

export interface UpdateSMTPConfigRequest extends CreateSMTPConfigRequest {
  id: string;
}