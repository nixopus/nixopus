DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'feature_flags'
);

DELETE FROM permissions
WHERE resource = 'feature_flags'; 