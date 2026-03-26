-- api/migrations/001_create_mcp_servers_up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS mcp_servers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL,
    provider_id     TEXT NOT NULL,
    name            TEXT NOT NULL,
    credentials     JSONB NOT NULL DEFAULT '{}',
    custom_url      TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_mcp_servers_org_id ON mcp_servers(org_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_mcp_servers_org_name ON mcp_servers(org_id, name) WHERE deleted_at IS NULL;
