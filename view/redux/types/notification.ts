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
  organization_id: string;
}

export interface CreateSMTPConfigRequest {
  host: string;
  port: number;
  username: string;
  password: string;
  from_name: string;
  from_email: string;
  organization_id: string;
}

export interface UpdateSMTPConfigRequest {
  id: string;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  from_name?: string;
  from_email?: string;
  organization_id: string;
}

export interface PreferenceType {
  id: string;
  label: string;
  description: string;
  enabled: boolean;
}

export interface CategoryPreferences {
  category: 'activity' | 'security' | 'update';
  preferences: PreferenceType[];
}

export interface UpdatePreferenceRequest {
  category: 'activity' | 'security' | 'update';
  type:
    | 'password-changes'
    | 'security-alerts'
    | 'team-updates'
    | 'login-alerts'
    | 'product-updates'
    | 'newsletter'
    | 'marketing';
  enabled: boolean;
}

export interface GetPreferencesResponse {
  activity: PreferenceType[];
  security: PreferenceType[];
  update: PreferenceType[];
}

export interface PreferenceItem {
  id: string;
  preference_id: string;
  category: string;
  type: string;
  enabled: boolean;
}

export interface WebhookConfig {
  id: string;
  type: 'slack' | 'discord';
  webhook_url: string;
  webhook_secret?: string;
  channel_id: string;
  is_active: boolean;
  user_id: string;
  organization_id: string;
  created_at: string;
  updated_at: string;
}

export interface CreateWebhookConfigRequest {
  type: 'slack' | 'discord';
  webhook_url: string;
  webhook_secret?: string;
  channel_id: string;
}

export interface UpdateWebhookConfigRequest {
  type: 'slack' | 'discord';
  webhook_url?: string;
  webhook_secret?: string;
  channel_id?: string;
  is_active?: boolean;
}

export interface DeleteWebhookConfigRequest {
  type: 'slack' | 'discord';
}

export interface GetWebhookConfigRequest {
  type: 'slack' | 'discord';
}

export interface SMTPFormData {
  smtp_host: string;
  smtp_port: string;
  smtp_username: string;
  smtp_password: string;
  smtp_from_email: string;
  smtp_from_name: string;
}
