import { User } from './user';

export interface Permission {
  id: string;
  name: string;
  description: string;
  resource: string;
}

export interface Role {
  id: string;
  name: string;
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
  deleted_at: string | null;
}

export interface UserOrganization {
  id: string;
  organization: Organization;
  role: {
    id: string;
    name: string;
    description: string;
  };
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
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

export interface RemoveUserFromOrganizationRequest {
  user_id: string;
  organization_id: string;
}

export interface OrganizationUsers {
  id: string;
  user_id: string;
  organization_id: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  user: User;
  roles: string[];
  permissions: string[];
}

export interface UpdateOrganizationDetailsRequest {
  id: string;
  name: string;
  description: string;
}

export type UserTypes = 'admin' | 'member' | 'viewer';

export interface CreateUserRequest {
  email: string;
  password: string;
  username: string;
  type: UserTypes;
  organization: string;
}

export interface UpdateUserRoleRequest {
  user_id: string;
  organization_id: string;
  role: string;
}
