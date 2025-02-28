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