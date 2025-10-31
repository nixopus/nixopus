CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION generate_server_name() RETURNS VARCHAR(255) AS $$
DECLARE
    prefixes TEXT[] := ARRAY['web', 'api', 'db', 'app', 'srv', 'node', 'host', 'prod', 'dev', 'stage'];
    suffixes TEXT[] := ARRAY['alpha', 'beta', 'gamma', 'delta', 'epsilon', 'zeta', 'eta', 'theta', 'iota', 'kappa'];
    numbers TEXT[] := ARRAY['01', '02', '03', '04', '05', '06', '07', '08', '09', '10'];
    prefix TEXT;
    suffix TEXT;
    number TEXT;
BEGIN
    prefix := prefixes[1 + floor(random() * array_length(prefixes, 1))::int];
    suffix := suffixes[1 + floor(random() * array_length(suffixes, 1))::int];
    number := numbers[1 + floor(random() * array_length(numbers, 1))::int];
    RETURN prefix || '-' || suffix || '-' || number;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL DEFAULT generate_server_name(),
    description TEXT DEFAULT 'Auto Generated Server',
    host VARCHAR(255) NOT NULL DEFAULT 'localhost',
    port INT NOT NULL DEFAULT 22,
    username VARCHAR(255) NOT NULL DEFAULT 'root',
    ssh_password VARCHAR(255),
    ssh_private_key_path VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT ssh_auth_check CHECK (
        (ssh_password IS NOT NULL AND ssh_private_key_path IS NULL) OR
        (ssh_password IS NULL AND ssh_private_key_path IS NOT NULL)
    ),
    CONSTRAINT check_server_status CHECK (status IN ('active', 'inactive', 'maintenance'))
);

CREATE INDEX idx_servers_user_id ON servers(user_id);
CREATE INDEX idx_servers_organization_id ON servers(organization_id);
CREATE UNIQUE INDEX idx_servers_org_name_unique
  ON servers(organization_id, name)
  WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_servers_org_host_port_unique
  ON servers(organization_id, host, port)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_servers_active_per_org 
ON servers(organization_id) 
WHERE status = 'active' AND deleted_at IS NULL;

CREATE OR REPLACE FUNCTION ensure_single_active_server() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'active' AND NEW.deleted_at IS NULL THEN
        UPDATE servers 
        SET status = 'inactive', updated_at = CURRENT_TIMESTAMP
        WHERE organization_id = NEW.organization_id 
          AND id != NEW.id 
          AND status = 'active' 
          AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_active_server
    BEFORE INSERT OR UPDATE ON servers
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_active_server();

CREATE INDEX idx_servers_status ON servers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_servers_org_status ON servers(organization_id, status) WHERE deleted_at IS NULL;
