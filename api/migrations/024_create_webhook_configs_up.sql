CREATE TABLE IF NOT EXISTS webhook_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    webhook_url TEXT NOT NULL,
    channel_id VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    webhook_secret TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_configs_user_id ON webhook_configs(user_id);
CREATE INDEX IF NOT EXISTS idx_webhook_configs_organization_id ON webhook_configs(organization_id);
CREATE INDEX IF NOT EXISTS idx_webhook_configs_type ON webhook_configs(type); 