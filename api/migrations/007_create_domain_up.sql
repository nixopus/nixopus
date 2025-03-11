CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL
);

CREATE INDEX idx_domains_user_id ON domains(user_id);
CREATE INDEX idx_domains_name ON domains(name);
CREATE UNIQUE INDEX idx_domains_name_unique ON domains(name) WHERE deleted_at IS NULL;
