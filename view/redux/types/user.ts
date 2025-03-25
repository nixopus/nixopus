import { OrganizationUsers } from "./orgs";

export interface User {
  id: string;
  username: string;
  email: string;
  avatar: string;
  created_at: Date;
  updated_at: Date;
  deleted_at: Date | null;
  is_verified: boolean;
  reset_token: string;
  type: string;
  organization_users: OrganizationUsers[];
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

export interface RefreshTokenPayload {
  refresh_token: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}
