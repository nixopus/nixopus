package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"

	"github.com/nixopus/nixopus/api/internal/storage"
)

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { db.Close() })

	// Manually create tables with SQLite-compatible schemas
	stmts := []string{
		`CREATE TABLE applications (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, port INTEGER NOT NULL DEFAULT 3000,
			environment TEXT NOT NULL DEFAULT 'production', proxy_server TEXT NOT NULL DEFAULT 'caddy',
			build_variables TEXT NOT NULL DEFAULT '', environment_variables TEXT NOT NULL DEFAULT '',
			build_pack TEXT NOT NULL DEFAULT 'dockerfile', repository TEXT NOT NULL DEFAULT '',
			branch TEXT NOT NULL DEFAULT 'main', pre_run_command TEXT NOT NULL DEFAULT '',
			post_run_command TEXT NOT NULL DEFAULT '', dockerfile_path TEXT NOT NULL DEFAULT 'Dockerfile',
			base_path TEXT NOT NULL DEFAULT '/', user_id TEXT NOT NULL, organization_id TEXT NOT NULL,
			family_id TEXT, labels TEXT, source TEXT NOT NULL DEFAULT 'github',
			is_live_deployment INTEGER NOT NULL DEFAULT 0, template_id TEXT NOT NULL DEFAULT '',
			routing_strategy TEXT NOT NULL DEFAULT 'single',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE application_status (
			id TEXT PRIMARY KEY, application_id TEXT NOT NULL, status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE application_domains (
			id TEXT PRIMARY KEY, application_id TEXT NOT NULL, domain TEXT NOT NULL,
			compose_service_id TEXT, port INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE domains (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, user_id TEXT NOT NULL,
			organization_id TEXT NOT NULL, type TEXT NOT NULL DEFAULT 'system',
			status TEXT NOT NULL DEFAULT 'active', verification_token TEXT,
			dns_provider TEXT, target_subdomain TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
		)`,
		`CREATE TABLE ssh_keys (
			id TEXT PRIMARY KEY, organization_id TEXT NOT NULL, name TEXT NOT NULL,
			description TEXT, host TEXT, proxy_host TEXT, "user" TEXT, port INTEGER DEFAULT 22,
			public_key TEXT, private_key_encrypted TEXT, password_encrypted TEXT,
			key_type TEXT DEFAULT 'rsa', key_size INTEGER DEFAULT 4096, fingerprint TEXT,
			auth_method TEXT NOT NULL DEFAULT 'key', is_active INTEGER NOT NULL DEFAULT 1,
			is_default INTEGER NOT NULL DEFAULT 0, last_used_at DATETIME, deleted_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE member (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL, organization_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member', created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE github_connectors (
			id TEXT PRIMARY KEY, app_id TEXT NOT NULL DEFAULT '', slug TEXT NOT NULL,
			pem TEXT NOT NULL DEFAULT '', client_id TEXT NOT NULL DEFAULT '',
			client_secret TEXT NOT NULL DEFAULT '', webhook_secret TEXT NOT NULL DEFAULT '',
			installation_id TEXT NOT NULL DEFAULT '', user_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
		)`,
		`CREATE TABLE mcp_servers (
			id TEXT PRIMARY KEY, org_id TEXT NOT NULL, provider_id TEXT NOT NULL,
			name TEXT NOT NULL, credentials TEXT, custom_url TEXT,
			enabled INTEGER NOT NULL DEFAULT 1, created_by TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		require.NoError(t, err)
	}

	return db
}

func testLogger() logger.Logger {
	return logger.NewLogger()
}

func TestInjectUserContext_InvalidOrgID(t *testing.T) {
	svc := &AgentService{logger: testLogger()}
	result := svc.injectUserContext(context.Background(), "not-a-uuid")
	assert.Equal(t, "", result)
}

func TestInjectUserContext_Empty(t *testing.T) {
	db := setupTestDB(t)

	svc := &AgentService{
		store:  &storage.Store{DB: db},
		logger: testLogger(),
	}

	orgID := uuid.New().String()
	result := svc.injectUserContext(context.Background(), orgID)
	assert.Equal(t, "", result)
}

func TestInjectUserContext_WithApps(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	orgID := uuid.New()
	userID := uuid.New()
	appID := uuid.New()
	statusID := uuid.New()

	app := &shared_types.Application{
		ID: appID, Name: "my-app", Port: 3000, Environment: "production",
		ProxyServer: "caddy", BuildPack: "dockerfile", Repository: "github.com/org/repo",
		Branch: "main", DockerfilePath: "Dockerfile", BasePath: "/",
		UserID: userID, OrganizationID: orgID, Source: "github",
	}
	_, err := db.NewInsert().Model(app).Exec(ctx)
	require.NoError(t, err)

	status := &shared_types.ApplicationStatus{ID: statusID, ApplicationID: appID, Status: "running"}
	_, err = db.NewInsert().Model(status).Exec(ctx)
	require.NoError(t, err)

	svc := &AgentService{
		store:  &storage.Store{DB: db},
		logger: testLogger(),
	}

	result := svc.injectUserContext(ctx, orgID.String())
	assert.Contains(t, result, "[user-context]")
	assert.Contains(t, result, "[/user-context]")
	assert.Contains(t, result, "my-app")
	assert.Contains(t, result, "status:running")
	assert.Contains(t, result, "port:3000")
	assert.Contains(t, result, "branch:main")
}

func TestInjectUserContext_WithDomains(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	orgID := uuid.New()
	userID := uuid.New()
	domainID := uuid.New()

	domain := &shared_types.Domain{
		ID: domainID, Name: "example.com", UserID: userID,
		OrganizationID: orgID, Type: "system", Status: "active",
	}
	_, err := db.NewInsert().Model(domain).Exec(ctx)
	require.NoError(t, err)

	svc := &AgentService{
		store:  &storage.Store{DB: db},
		logger: testLogger(),
	}

	result := svc.injectUserContext(ctx, orgID.String())
	assert.Contains(t, result, "domains: example.com")
	assert.Contains(t, result, "type:system")
}

func TestInjectUserContext_WithServers(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	orgID := uuid.New()
	serverID := uuid.New()
	host := "1.2.3.4"

	server := &shared_types.SSHKey{
		ID: serverID, OrganizationID: orgID, Name: "prod-server", Host: &host,
	}
	_, err := db.NewInsert().Model(server).Exec(ctx)
	require.NoError(t, err)

	svc := &AgentService{
		store:  &storage.Store{DB: db},
		logger: testLogger(),
	}

	result := svc.injectUserContext(ctx, orgID.String())
	assert.Contains(t, result, "servers: prod-server")
	assert.Contains(t, result, "host:1.2.3.4")
}

func TestInjectUserContext_WithConnectors(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	orgID := uuid.New()
	userID := uuid.New()
	connectorID := uuid.New()
	memberID := uuid.New()

	member := &shared_types.OrganizationUsers{
		ID: memberID, UserID: userID, OrganizationID: orgID, Role: "owner",
	}
	_, err := db.NewInsert().Model(member).Exec(ctx)
	require.NoError(t, err)

	connector := &shared_types.GithubConnector{
		ID: connectorID, Slug: "my-github-app", UserID: userID,
		AppID: "123", Pem: "x", ClientID: "x", ClientSecret: "x",
		WebhookSecret: "x", InstallationID: "x",
	}
	_, err = db.NewInsert().Model(connector).Exec(ctx)
	require.NoError(t, err)

	svc := &AgentService{
		store:  &storage.Store{DB: db},
		logger: testLogger(),
	}

	result := svc.injectUserContext(ctx, orgID.String())
	assert.Contains(t, result, "connectors: my-github-app")
}

func TestInjectUserContext_WithMCPServers(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	orgID := uuid.New()
	serverID := uuid.New()
	userID := uuid.New()

	mcpSrv := &shared_types.MCPServer{
		ID: serverID, OrgID: orgID, ProviderID: "linear",
		Name: "Linear MCP", Enabled: true, CreatedBy: userID,
	}
	_, err := db.NewInsert().Model(mcpSrv).Exec(ctx)
	require.NoError(t, err)

	svc := &AgentService{
		store:  &storage.Store{DB: db},
		logger: testLogger(),
	}

	result := svc.injectUserContext(ctx, orgID.String())
	assert.Contains(t, result, "mcp_servers: Linear MCP")
	assert.Contains(t, result, "provider:linear")
}

// Ensure Application uses shared_types to verify compilation
var _ = shared_types.Application{}
