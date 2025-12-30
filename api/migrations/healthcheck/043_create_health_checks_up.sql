CREATE TABLE IF NOT EXISTS health_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,

    endpoint TEXT NOT NULL DEFAULT '/',
    method TEXT NOT NULL DEFAULT 'GET',
    expected_status_codes INTEGER[] NOT NULL DEFAULT '{200}',
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    
    interval_seconds INTEGER NOT NULL DEFAULT 60,

    failure_threshold INTEGER NOT NULL DEFAULT 3,
    success_threshold INTEGER NOT NULL DEFAULT 1,

    headers JSONB DEFAULT '{}',
    body TEXT,
    
    consecutive_fails INTEGER NOT NULL DEFAULT 0,
    last_checked_at TIMESTAMPTZ,
    
    retention_days INTEGER NOT NULL DEFAULT 30,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(application_id)
);

CREATE TABLE IF NOT EXISTS health_check_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    health_check_id UUID NOT NULL REFERENCES health_checks(id) ON DELETE CASCADE,
    
    status TEXT NOT NULL,
    response_time_ms INTEGER,
    status_code INTEGER,
    error_message TEXT,
    
    checked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_health_check_results_health_check_id ON health_check_results(health_check_id);
CREATE INDEX IF NOT EXISTS idx_health_check_results_checked_at ON health_check_results(checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_health_checks_application_id ON health_checks(application_id);
CREATE INDEX IF NOT EXISTS idx_health_checks_organization_id ON health_checks(organization_id);
CREATE INDEX IF NOT EXISTS idx_health_checks_enabled ON health_checks(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_health_checks_last_checked_at ON health_checks(last_checked_at);

ALTER TABLE health_checks ADD CONSTRAINT check_valid_method 
    CHECK (method IN ('GET', 'POST', 'HEAD'));

ALTER TABLE health_check_results ADD CONSTRAINT check_valid_status 
    CHECK (status IN ('healthy', 'unhealthy', 'timeout', 'error'));


ALTER TABLE health_checks ADD CONSTRAINT check_timeout_range 
    CHECK (timeout_seconds >= 5 AND timeout_seconds <= 120);

ALTER TABLE health_checks ADD CONSTRAINT check_interval_range 
    CHECK (interval_seconds >= 30 AND interval_seconds <= 3600);

ALTER TABLE health_checks ADD CONSTRAINT check_failure_threshold_range 
    CHECK (failure_threshold >= 1 AND failure_threshold <= 10);

ALTER TABLE health_checks ADD CONSTRAINT check_success_threshold_range 
    CHECK (success_threshold >= 1 AND success_threshold <= 10);

ALTER TABLE health_checks ADD CONSTRAINT check_retention_days_range 
    CHECK (retention_days >= 1 AND retention_days <= 365);

