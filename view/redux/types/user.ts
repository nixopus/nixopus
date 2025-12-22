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

export interface TwoFactorSetupResponse {
  secret: string;
  qr_code: string;
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

export interface UpdateCheckResponse {
  current_version: string;
  latest_version: string;
  update_available: boolean;
  last_checked: string;
  environment: string;
}
