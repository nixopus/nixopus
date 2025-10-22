ALTER TABLE extension_executions
ADD COLUMN IF NOT EXISTS log_seq BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS extension_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID NOT NULL REFERENCES extension_executions(id) ON DELETE CASCADE,
    step_id UUID REFERENCES execution_steps(id) ON DELETE SET NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    sequence BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_extension_logs_exec_seq ON extension_logs(execution_id, sequence);
CREATE INDEX IF NOT EXISTS idx_extension_logs_exec_created ON extension_logs(execution_id, created_at);

