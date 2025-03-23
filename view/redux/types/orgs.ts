import { User } from './user';

export interface Permission {
  id: string;
  name: string;
  description: string;
  resource: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface Role {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface RoleWithPermissions extends Role {
  permissions: Permission[];
}

export interface Organization {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface UserOrganization {
  organization: Organization;
  role: {
    id: string;
    name: string;
    description: string;
    created_at: string;
    updated_at: string;
    deleted_at?: string | null;
    permissions: Permission[];
  };
}

export interface UserOrganizationsResponse {
  user: {
    id: string;
  };
  organizations: UserOrganization[];
}

export interface CreateOrganizationRequest {
  name: string;
  description: string;
}

export interface AddUserToOrganizationRequest {
  user_id: string;
  organization_id: string;
  role_id: string;
}

export interface OrganizationUsers {
  id: string;
  user_id: string;
  organization_id: string;
  role_id: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  role: RoleWithPermissions;
  user: User;
}

export interface UpdateOrganizationDetailsRequest {
  id: string;
  name: string;
  description: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  username: string;
  type: UserTypes;
  organization: string;
}

export enum UserTypes {
  ADMIN = 'admin',
  MEMBER = 'member',
  VIEWER = 'viewer'
}