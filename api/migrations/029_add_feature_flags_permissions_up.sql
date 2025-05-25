INSERT INTO permissions (id, name, resource, description, created_at, updated_at)
VALUES 
    (uuid_generate_v4(), 'read', 'feature_flags', 'Read feature flags', NOW(), NOW()),
    (uuid_generate_v4(), 'update', 'feature_flags', 'Update feature flags', NOW(), NOW());

INSERT INTO role_permissions (id, role_id, permission_id, created_at, updated_at)
SELECT 
    uuid_generate_v4(),
    r.id,
    p.id,
    NOW(),
    NOW()
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin' AND p.resource = 'feature_flags'; 