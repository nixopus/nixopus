DROP INDEX IF EXISTS idx_extension_logs_exec_created;
DROP INDEX IF EXISTS idx_extension_logs_exec_seq;
DROP TABLE IF EXISTS extension_logs;
ALTER TABLE extension_executions DROP COLUMN IF EXISTS log_seq;

