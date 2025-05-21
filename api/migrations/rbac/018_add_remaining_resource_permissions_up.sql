INSERT INTO permissions (id, name, description, resource) VALUES
(uuid_generate_v4(), 'create', 'Create domains', 'domain'),
(uuid_generate_v4(), 'read', 'Read domain information', 'domain'),
(uuid_generate_v4(), 'update', 'Update domain information', 'domain'),
(uuid_generate_v4(), 'delete', 'Delete domains', 'domain'),

(uuid_generate_v4(), 'create', 'Create github connectors', 'github-connector'),
(uuid_generate_v4(), 'read', 'Read github connector information', 'github-connector'),
(uuid_generate_v4(), 'update', 'Update github connector information', 'github-connector'),
(uuid_generate_v4(), 'delete', 'Delete github connectors', 'github-connector'),

(uuid_generate_v4(), 'create', 'Create notifications', 'notification'),
(uuid_generate_v4(), 'read', 'Read notification information', 'notification'),
(uuid_generate_v4(), 'update', 'Update notification information', 'notification'),
(uuid_generate_v4(), 'delete', 'Delete notifications', 'notification'),

(uuid_generate_v4(), 'create', 'Create files and directories', 'file-manager'),
(uuid_generate_v4(), 'read', 'Read files and directories', 'file-manager'),
(uuid_generate_v4(), 'update', 'Update files and directories', 'file-manager'),
(uuid_generate_v4(), 'delete', 'Delete files and directories', 'file-manager'),

(uuid_generate_v4(), 'create', 'Create deployments', 'deploy'),
(uuid_generate_v4(), 'read', 'Read deployment information', 'deploy'),
(uuid_generate_v4(), 'update', 'Update deployment information', 'deploy'),
(uuid_generate_v4(), 'delete', 'Delete deployments', 'deploy');

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), admin_role.id, permissions.id
FROM admin_role, permissions
WHERE permissions.resource IN ('domain', 'github-connector', 'notification', 'file-manager', 'deploy');

WITH viewer_role AS (
    SELECT id FROM roles WHERE name = 'viewer'
),
read_permissions AS (
    SELECT id FROM permissions 
    WHERE name = 'read' 
    AND resource IN ('domain', 'github-connector', 'notification', 'file-manager', 'deploy')
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), viewer_role.id, read_permissions.id
FROM viewer_role, read_permissions;

WITH member_role AS (
    SELECT id FROM roles WHERE name = 'member'
),
member_permissions AS (
    SELECT id FROM permissions 
    WHERE name IN ('read', 'update') 
    AND resource IN ('domain', 'github-connector', 'notification', 'file-manager', 'deploy')
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), member_role.id, member_permissions.id
FROM member_role, member_permissions; 