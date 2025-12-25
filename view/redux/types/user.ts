import { Organization, OrganizationUsers } from './orgs';

export interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  type: string;
  organization_users: OrganizationUsers[];
  is_verified: boolean;
  is_email_verified: boolean;
  two_factor_enabled: boolean;
  two_factor_secret: string;
  created_at: string;
  updated_at: string;
  organizations: Organization[];
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
  temp_token?: string;
}

export interface RefreshTokenPayload {
  refresh_token: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface TwoFactorSetupResponse {
  secret: string;
  qr_code: string;
}

export interface TwoFactorLoginPayload {
  email: string;
  password: string;
  code: string;
}

export interface UserSettings {
  id: string;
  user_id: string;
  font_family: string;
  font_size: number;
  theme: string;
  language: string;
  auto_update: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpdateFontRequest {
  font_family: string;
  font_size: number;
}

export interface UpdateThemeRequest {
  theme: string;
}

export interface UpdateLanguageRequest {
  language: string;
}

export interface UpdateAutoUpdateRequest {
  auto_update: boolean;
}

export interface UpdateAvatarRequest {
  avatarData: string;
}

export interface UserPreferencesData {
  debug_mode: boolean;
  show_api_error_details: boolean;
  terminal_scrollback?: number;
  terminal_font_size?: number;
  terminal_cursor_style?: 'bar' | 'block' | 'underline';
  terminal_cursor_blink?: boolean;
  terminal_line_height?: number;
  terminal_cursor_width?: number;
  terminal_tab_stop_width?: number;
  terminal_font_family?: string;
  terminal_font_weight?: 'normal' | 'bold';
  terminal_letter_spacing?: number;
}

export interface UserPreferences {
  id: string;
  user_id: string;
  preferences: UserPreferencesData;
  created_at: string;
  updated_at: string;
}

export interface OrganizationSettingsData {
  websocket_reconnect_attempts: number;
  websocket_reconnect_interval: number;
  api_retry_attempts: number;
  disable_api_cache: boolean;
  container_log_tail_lines?: number;
  container_default_restart_policy?: 'no' | 'always' | 'on-failure' | 'unless-stopped';
  container_stop_timeout?: number;
  container_auto_prune_dangling_images?: boolean;
  container_auto_prune_build_cache?: boolean;
}

export interface OrganizationSettings {
  id: string;
  organization_id: string;
  settings: OrganizationSettingsData;
  created_at: string;
  updated_at: string;
}

export interface UpdateCheckResponse {
  current_version: string;
  latest_version: string;
  update_available: boolean;
  last_checked: string;
  environment: string;
}
