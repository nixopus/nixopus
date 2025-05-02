DELETE FROM role_permissions 
WHERE permission_id IN (
    SELECT id FROM permissions 
    WHERE resource = 'container'
);

DELETE FROM permissions 
WHERE resource = 'container'; 