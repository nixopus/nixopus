CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE github_connectors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    pem TEXT NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    webhook_secret VARCHAR(255) NOT NULL,
    installation_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_github_connectors_user_id ON github_connectors(user_id);
CREATE INDEX idx_github_connectors_slug ON github_connectors(slug);
CREATE INDEX idx_github_connectors_app_id ON github_connectors(app_id);
