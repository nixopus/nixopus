DROP INDEX IF EXISTS idx_applications_org_server;
DROP INDEX IF EXISTS idx_applications_server_id;

ALTER TABLE applications
DROP COLUMN IF EXISTS server_id;
