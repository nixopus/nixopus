INSERT INTO permissions (id, name, description, resource) 
SELECT uuid_generate_v4(), 'create', 'Create extensions', 'extension'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE name = 'create' AND resource = 'extension'
);

INSERT INTO permissions (id, name, description, resource) 
SELECT uuid_generate_v4(), 'read', 'Read extension information', 'extension'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE name = 'read' AND resource = 'extension'
);

INSERT INTO permissions (id, name, description, resource) 
SELECT uuid_generate_v4(), 'update', 'Update extension information', 'extension'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE name = 'update' AND resource = 'extension'
);

INSERT INTO permissions (id, name, description, resource) 
SELECT uuid_generate_v4(), 'delete', 'Delete extensions', 'extension'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE name = 'delete' AND resource = 'extension'
);

INSERT INTO permissions (id, name, description, resource) 
SELECT uuid_generate_v4(), 'install', 'Install extensions', 'extension'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE name = 'install' AND resource = 'extension'
);

INSERT INTO permissions (id, name, description, resource) 
SELECT uuid_generate_v4(), 'uninstall', 'Uninstall extensions', 'extension'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE name = 'uninstall' AND resource = 'extension'
);

-- Assign all extensions permissions to admin role
WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
),
extensions_permissions AS (
    SELECT id FROM permissions WHERE resource = 'extension'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), admin_role.id, extensions_permissions.id
FROM admin_role, extensions_permissions
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions rp 
    WHERE rp.role_id = admin_role.id 
    AND rp.permission_id = extensions_permissions.id
);

-- Assign read permission to viewer role
WITH viewer_role AS (
    SELECT id FROM roles WHERE name = 'viewer'
),
read_permission AS (
    SELECT id FROM permissions 
    WHERE name = 'read' AND resource = 'extension'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), viewer_role.id, read_permission.id
FROM viewer_role, read_permission
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions rp 
    WHERE rp.role_id = viewer_role.id 
    AND rp.permission_id = read_permission.id
);

-- Assign read, install, and uninstall permissions to member role
WITH member_role AS (
    SELECT id FROM roles WHERE name = 'member'
),
member_permissions AS (
    SELECT id FROM permissions 
    WHERE name IN ('read', 'install', 'uninstall') 
    AND resource = 'extension'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), member_role.id, member_permissions.id
FROM member_role, member_permissions
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions rp 
    WHERE rp.role_id = member_role.id 
    AND rp.permission_id = member_permissions.id
);
