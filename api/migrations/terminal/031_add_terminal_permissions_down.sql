DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE resource = 'terminal'
);

DELETE FROM permissions
WHERE resource = 'terminal'; 