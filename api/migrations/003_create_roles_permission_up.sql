CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_roles_name ON roles(name);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(name, resource)
);

CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource ON permissions(resource);

CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

INSERT INTO roles (id, name, description) VALUES
(uuid_generate_v4(), 'admin', 'Administrator with full access'),
(uuid_generate_v4(), 'member', 'Regular organization member'),
(uuid_generate_v4(), 'viewer', 'Read-only access to resources');

INSERT INTO permissions (id, name, description, resource) VALUES
(uuid_generate_v4(), 'create', 'Create users', 'user'),
(uuid_generate_v4(), 'read', 'Read user information', 'user'),
(uuid_generate_v4(), 'update', 'Update user information', 'user'),
(uuid_generate_v4(), 'delete', 'Delete users', 'user'),

(uuid_generate_v4(), 'create', 'Create organizations', 'organization'),
(uuid_generate_v4(), 'read', 'Read organization information', 'organization'),
(uuid_generate_v4(), 'update', 'Update organization information', 'organization'),
(uuid_generate_v4(), 'delete', 'Delete organizations', 'organization'),

(uuid_generate_v4(), 'create', 'Create roles', 'role'),
(uuid_generate_v4(), 'read', 'Read role information', 'role'),
(uuid_generate_v4(), 'update', 'Update role information', 'role'),
(uuid_generate_v4(), 'delete', 'Delete roles', 'role'),

(uuid_generate_v4(), 'create', 'Create permissions', 'permission'),
(uuid_generate_v4(), 'read', 'Read permission information', 'permission'),
(uuid_generate_v4(), 'update', 'Update permission information', 'permission'),
(uuid_generate_v4(), 'delete', 'Delete permissions', 'permission');

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), admin_role.id, permissions.id
FROM admin_role, permissions;

WITH viewer_role AS (
    SELECT id FROM roles WHERE name = 'viewer'
),
read_permissions AS (
    SELECT id FROM permissions WHERE name = 'read'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), viewer_role.id, read_permissions.id
FROM viewer_role, read_permissions;

WITH member_role AS (
    SELECT id FROM roles WHERE name = 'member'
),
member_permissions AS (
    SELECT id FROM permissions WHERE name IN ('read', 'update') AND resource IN ('user', 'organization')
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), member_role.id, member_permissions.id
FROM member_role, member_permissions;