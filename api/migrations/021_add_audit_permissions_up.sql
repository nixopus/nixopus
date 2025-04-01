INSERT INTO permissions (id, name, description, resource) VALUES
(uuid_generate_v4(), 'create', 'Create audit logs', 'audit'),
(uuid_generate_v4(), 'read', 'Read audit logs', 'audit'),
(uuid_generate_v4(), 'update', 'Update audit logs', 'audit'),
(uuid_generate_v4(), 'delete', 'Delete audit logs', 'audit');

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), admin_role.id, permissions.id
FROM admin_role, permissions
WHERE permissions.resource = 'audit';

WITH viewer_role AS (
    SELECT id FROM roles WHERE name = 'viewer'
),
read_permissions AS (
    SELECT id FROM permissions 
    WHERE name = 'read' 
    AND resource = 'audit'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), viewer_role.id, read_permissions.id
FROM viewer_role, read_permissions;

WITH member_role AS (
    SELECT id FROM roles WHERE name = 'member'
),
member_permissions AS (
    SELECT id FROM permissions 
    WHERE name = 'read'
    AND resource = 'audit'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), member_role.id, member_permissions.id
FROM member_role, member_permissions; 