DROP TRIGGER IF EXISTS applications_notify ON applications;
DROP TRIGGER IF EXISTS application_status_notify ON application_status;
DROP TRIGGER IF EXISTS application_logs_notify ON application_logs;
DROP TRIGGER IF EXISTS application_deployment_notify ON application_deployment;
DROP TRIGGER IF EXISTS application_deployment_status_notify ON application_deployment_status;
DROP FUNCTION IF EXISTS notify_application_change();