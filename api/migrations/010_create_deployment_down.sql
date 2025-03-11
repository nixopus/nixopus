DROP INDEX IF EXISTS idx_application_logs_deployment_id;
DROP INDEX IF EXISTS idx_application_deployment_status_deployment_id;
DROP INDEX IF EXISTS idx_application_deployment_application_id;

ALTER TABLE application_logs 
DROP COLUMN IF EXISTS application_deployment_id;

DROP TABLE IF EXISTS application_deployment_status;
DROP TABLE IF EXISTS application_deployment;