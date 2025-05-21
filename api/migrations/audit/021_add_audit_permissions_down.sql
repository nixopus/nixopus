DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'audit'
);
DELETE FROM permissions WHERE resource = 'audit'; 