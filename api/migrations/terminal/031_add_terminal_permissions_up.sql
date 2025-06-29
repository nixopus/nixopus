INSERT INTO permissions (id, name, description, resource) VALUES
(uuid_generate_v4(), 'create', 'Create terminal session', 'terminal'),
(uuid_generate_v4(), 'read', 'Read terminal', 'terminal'),
(uuid_generate_v4(), 'update', 'Update terminal session', 'terminal'),
(uuid_generate_v4(), 'delete', 'Delete terminal session', 'terminal');

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), admin_role.id, permissions.id
FROM admin_role, permissions
WHERE permissions.resource = 'terminal';

WITH member_role AS (
    SELECT id FROM roles WHERE name = 'member'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), member_role.id, permissions.id
FROM member_role, permissions
WHERE permissions.resource = 'terminal';

WITH viewer_role AS (
    SELECT id FROM roles WHERE name = 'viewer'
),
read_permissions AS (
    SELECT id FROM permissions 
    WHERE name = 'read' 
    AND resource = 'terminal'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), viewer_role.id, read_permissions.id
FROM viewer_role, read_permissions; 