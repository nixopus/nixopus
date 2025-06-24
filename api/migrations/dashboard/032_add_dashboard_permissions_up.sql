INSERT INTO permissions (id, name, description, resource) VALUES
(uuid_generate_v4(), 'read', 'Read dashboard information', 'dashboard');

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), admin_role.id, permissions.id
FROM admin_role, permissions
WHERE permissions.resource = 'dashboard';

WITH member_role AS (
    SELECT id FROM roles WHERE name = 'member'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), member_role.id, permissions.id
FROM member_role, permissions
WHERE permissions.resource = 'dashboard';

WITH viewer_role AS (
    SELECT id FROM roles WHERE name = 'viewer'
)
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT uuid_generate_v4(), viewer_role.id, permissions.id
FROM viewer_role, permissions
WHERE permissions.resource = 'dashboard'; 