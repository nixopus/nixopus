DROP INDEX IF EXISTS idx_health_check_results_health_check_id;
DROP INDEX IF EXISTS idx_health_check_results_checked_at;
DROP INDEX IF EXISTS idx_health_checks_application_id;
DROP INDEX IF EXISTS idx_health_checks_organization_id;
DROP INDEX IF EXISTS idx_health_checks_enabled;
DROP INDEX IF EXISTS idx_health_checks_last_checked_at;

DROP TABLE IF EXISTS health_check_results;
DROP TABLE IF EXISTS health_checks;

