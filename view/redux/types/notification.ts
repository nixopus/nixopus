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
  organization_id: string;
}

export interface UpdateSMTPConfigRequest extends CreateSMTPConfigRequest {
  id: string;
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
  type: string;
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
