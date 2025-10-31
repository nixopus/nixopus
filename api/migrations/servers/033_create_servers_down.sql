DROP TRIGGER IF EXISTS trigger_ensure_single_active_server ON servers;
DROP FUNCTION IF EXISTS ensure_single_active_server();

DROP INDEX IF EXISTS idx_servers_org_status;
DROP INDEX IF EXISTS idx_servers_status;
DROP INDEX IF EXISTS idx_servers_active_per_org;
ALTER TABLE IF EXISTS servers DROP CONSTRAINT IF EXISTS check_server_status;

DROP INDEX IF EXISTS idx_servers_org_host_port_unique;
DROP INDEX IF EXISTS idx_servers_org_name_unique;
DROP INDEX IF EXISTS idx_servers_organization_id;
DROP INDEX IF EXISTS idx_servers_user_id;

DROP TABLE IF EXISTS servers;
DROP FUNCTION IF EXISTS generate_server_name();
