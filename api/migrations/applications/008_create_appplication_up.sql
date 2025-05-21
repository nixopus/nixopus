DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'environment') THEN
        CREATE TYPE environment AS ENUM ('development', 'staging', 'production');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'build_pack') THEN
        CREATE TYPE build_pack AS ENUM ('dockerfile', 'docker-compose', 'static');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status') THEN
        CREATE TYPE status AS ENUM ('started', 'running', 'stopped', 'failed', 'cloning', 'building', 'deploying', 'deployed');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    port INTEGER NOT NULL,
    environment environment NOT NULL,
    build_variables TEXT NOT NULL,
    environment_variables TEXT NOT NULL,
    build_pack build_pack NOT NULL,
    repository TEXT NOT NULL,
    branch TEXT NOT NULL,
    pre_run_command TEXT NOT NULL,
    post_run_command TEXT NOT NULL,
    domain_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS application_status (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL,
    status status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS application_logs (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL,
    log TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_application_status_application_id ON application_status(application_id);
CREATE INDEX IF NOT EXISTS idx_application_logs_application_id ON application_logs(application_id);
CREATE INDEX IF NOT EXISTS idx_applications_domain_id ON applications(domain_id);
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications(user_id);
