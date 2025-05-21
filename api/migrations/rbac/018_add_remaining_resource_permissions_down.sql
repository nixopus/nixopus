DELETE FROM role_permissions 
WHERE permission_id IN (
    SELECT id FROM permissions 
    WHERE resource IN ('domain', 'github-connector', 'notification', 'file-manager', 'deploy')
);

DELETE FROM permissions 
WHERE resource IN ('domain', 'github-connector', 'notification', 'file-manager', 'deploy'); 