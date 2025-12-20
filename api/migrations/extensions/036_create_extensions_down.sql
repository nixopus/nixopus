DROP TRIGGER IF EXISTS trigger_extensions_updated_at ON extensions;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_execution_steps_execution;
DROP INDEX IF EXISTS idx_extension_executions_status;
DROP INDEX IF EXISTS idx_extension_executions_extension;
DROP INDEX IF EXISTS idx_extension_variables_extension;
DROP INDEX IF EXISTS idx_extensions_deleted_at;
DROP INDEX IF EXISTS idx_extensions_extension_id;
DROP INDEX IF EXISTS idx_extensions_created;
DROP INDEX IF EXISTS idx_extensions_validation_status;
DROP INDEX IF EXISTS idx_extensions_verified;
DROP INDEX IF EXISTS idx_extensions_category;

DROP TABLE IF EXISTS execution_steps;
DROP TABLE IF EXISTS extension_executions;
DROP TABLE IF EXISTS extension_variables;
DROP TABLE IF EXISTS extensions;

DROP TYPE IF EXISTS execution_status;
DROP TYPE IF EXISTS validation_status;
DROP TYPE IF EXISTS extension_category;
