CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS application_deployment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS application_deployment_status (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_deployment_id UUID NOT NULL REFERENCES application_deployment(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE application_logs 
ADD COLUMN IF NOT EXISTS application_deployment_id UUID REFERENCES application_deployment(id) ON DELETE SET NULL;

CREATE INDEX idx_application_deployment_application_id ON application_deployment(application_id);
CREATE INDEX idx_application_deployment_status_deployment_id ON application_deployment_status(application_deployment_id);
CREATE INDEX idx_application_logs_deployment_id ON application_logs(application_deployment_id);