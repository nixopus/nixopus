INSERT INTO feature_flags (id, organization_id, feature_name, is_enabled, created_at, updated_at)
SELECT 
    uuid_generate_v4(),
    o.id,
    'extensions',
    true,
    NOW(),
    NOW()
FROM organizations o
WHERE NOT EXISTS (
    SELECT 1 FROM feature_flags ff 
    WHERE ff.organization_id = o.id 
    AND ff.feature_name = 'extensions'
);
