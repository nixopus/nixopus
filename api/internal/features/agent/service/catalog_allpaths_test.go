package service

// Comprehensive CI-safe coverage for every operation listed in the api-catalog.
//
// Two test functions:
//  1. TestCatalogAllOps_CorrectFormats — every operation's well-formed call passes schema.
//  2. TestCatalogAllOps_MockLLMProbe  — scripted mock LLM returns the correct call for
//     each operation and we assert the recorder captures schema-valid arguments.
//
// Intentionally no real LLM or database is required; these run in plain `go test`.

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// catalogOp is one testable API catalog entry.
type catalogOp struct {
	name      string         // subtest display name
	method    string         // HTTP method
	path      string         // full example path (path params substituted with placeholders)
	body      map[string]any // optional request body
	skipProbe bool           // skip agent-loop probe (complex multi-step operations)
}

// allCatalogOps is the complete table of every operation in catalog.go.
// Path-param placeholders like {app_uuid} are replaced with example values.
var allCatalogOps = []catalogOp{

	// ─── Applications ──────────────────────────────────────────────────────────
	{name: "list_applications", method: "GET", path: "/api/v1/deploy/applications"},
	{name: "list_applications_paged", method: "GET", path: "/api/v1/deploy/applications?page=1&page_size=20"},
	{name: "get_application", method: "GET", path: "/api/v1/deploy/application?id=app-uuid-123"},
	{name: "list_deployments", method: "GET", path: "/api/v1/deploy/application/deployments?id=app-uuid-123"},
	{name: "list_deployments_paged", method: "GET", path: "/api/v1/deploy/application/deployments?id=app-uuid-123&page=1&page_size=10"},
	{name: "get_deployment_by_id", method: "GET", path: "/api/v1/deploy/application/deployments/dep-uuid-456"},
	{name: "get_deployment_logs", method: "GET", path: "/api/v1/deploy/application/deployments/dep-uuid-456/logs"},
	{name: "get_deployment_logs_filtered", method: "GET", path: "/api/v1/deploy/application/deployments/dep-uuid-456/logs?level=ERROR&page=1"},
	{name: "get_application_logs", method: "GET", path: "/api/v1/deploy/application/logs/app-uuid-123"},
	{name: "create_application", method: "POST", path: "/api/v1/deploy/application",
		body: map[string]any{"repository": "owner/repo", "name": "my-app", "port": 3000}},
	{name: "deploy_from_template", method: "POST", path: "/api/v1/deploy/application/template",
		body: map[string]any{"template_id": "tmpl-001"}},
	{name: "update_application", method: "PUT", path: "/api/v1/deploy/application",
		body: map[string]any{"id": "app-uuid-123", "port": 8080}},
	{name: "delete_application", method: "DELETE", path: "/api/v1/deploy/application",
		body: map[string]any{"id": "app-uuid-123"}},
	{name: "redeploy_application", method: "POST", path: "/api/v1/deploy/application/redeploy",
		body: map[string]any{"id": "app-uuid-123"}},
	{name: "restart_deployment", method: "POST", path: "/api/v1/deploy/application/restart",
		body: map[string]any{"id": "dep-uuid-456"}},
	{name: "rollback_deployment", method: "POST", path: "/api/v1/deploy/application/rollback",
		body: map[string]any{"id": "app-uuid-123"}},
	{name: "cancel_deployment", method: "POST", path: "/api/v1/deploy/application/cancel-deployment",
		body: map[string]any{"deployment_id": "dep-uuid-456"}},
	{name: "recover_application", method: "POST", path: "/api/v1/deploy/application/recover",
		body: map[string]any{"application_id": "app-uuid-123"}},
	{name: "update_labels", method: "PUT", path: "/api/v1/deploy/application/labels?id=app-uuid-123",
		body: map[string]any{"labels": []string{"prod", "v2"}}},
	{name: "add_domain_to_app", method: "POST", path: "/api/v1/deploy/application/domains?id=app-uuid-123",
		body: map[string]any{"domain": "myapp.example.com"}},
	{name: "remove_domain_from_app", method: "DELETE", path: "/api/v1/deploy/application/domains?id=app-uuid-123",
		body: map[string]any{"domain": "myapp.example.com"}},
	{name: "list_compose_services", method: "GET", path: "/api/v1/deploy/application/compose-services?id=app-uuid-123"},
	{name: "preview_compose", method: "POST", path: "/api/v1/deploy/application/preview-compose",
		body: map[string]any{"repository": "owner/repo"}},
	{name: "get_app_servers", method: "GET", path: "/api/v1/deploy/application/servers?id=app-uuid-123"},
	{name: "set_app_servers", method: "PUT", path: "/api/v1/deploy/application/servers",
		body: map[string]any{"application_id": "app-uuid-123", "server_ids": []string{"srv-001"}}},

	// ─── Projects ──────────────────────────────────────────────────────────────
	{name: "create_project", method: "POST", path: "/api/v1/deploy/application/project",
		body: map[string]any{"name": "my-project", "repository": "owner/repo"}},
	{name: "deploy_project", method: "POST", path: "/api/v1/deploy/application/project/deploy",
		body: map[string]any{"id": "app-uuid-123"}},
	{name: "duplicate_project", method: "POST", path: "/api/v1/deploy/application/project/duplicate",
		body: map[string]any{"id": "app-uuid-123"}},
	{name: "get_project_family", method: "GET", path: "/api/v1/deploy/application/project/family?family_id=fam-uuid-789"},
	{name: "list_family_environments", method: "GET", path: "/api/v1/deploy/application/project/family/environments?family_id=fam-uuid-789"},
	{name: "add_project_to_family", method: "POST", path: "/api/v1/deploy/application/project/add-to-family",
		body: map[string]any{"project_id": "app-uuid-123", "family_id": "fam-uuid-789"}},

	// ─── Deploy Artifacts ──────────────────────────────────────────────────────
	{name: "list_artifacts", method: "GET", path: "/api/v1/deploy/artifacts?application_id=app-uuid-123"},
	{name: "download_artifact", method: "GET", path: "/api/v1/deploy/artifacts/dep-uuid-456/download"},
	{name: "delete_artifact", method: "DELETE", path: "/api/v1/deploy/artifacts/dep-uuid-456"},

	// ─── Domains ───────────────────────────────────────────────────────────────
	{name: "list_domains", method: "GET", path: "/api/v1/domain"},
	{name: "list_domains_by_type", method: "GET", path: "/api/v1/domain?type=custom"},
	{name: "generate_subdomain", method: "GET", path: "/api/v1/domain/generate"},
	{name: "create_custom_domain", method: "POST", path: "/api/v1/domain/custom",
		body: map[string]any{"name": "custom.example.com"}},
	{name: "delete_custom_domain", method: "DELETE", path: "/api/v1/domain/custom",
		body: map[string]any{"id": "dom-uuid-123"}},
	{name: "verify_domain", method: "POST", path: "/api/v1/domain/verify",
		body: map[string]any{"id": "dom-uuid-123"}},
	{name: "dns_check", method: "GET", path: "/api/v1/domain/dns-check?id=dom-uuid-123"},

	// ─── GitHub Connectors ─────────────────────────────────────────────────────
	{name: "create_github_connector", method: "POST", path: "/api/v1/github-connector",
		body: map[string]any{"app_id": "123", "client_id": "abc", "client_secret": "sec", "pem": "pem", "slug": "my-app", "webhook_secret": "wh"}},
	{name: "update_github_connector", method: "PUT", path: "/api/v1/github-connector",
		body: map[string]any{"connector_id": "conn-uuid-123", "installation_id": "inst-456"}},
	{name: "delete_github_connector", method: "DELETE", path: "/api/v1/github-connector",
		body: map[string]any{"id": "conn-uuid-123"}},
	{name: "list_github_connectors", method: "GET", path: "/api/v1/github-connector/all"},
	{name: "list_github_repositories", method: "GET", path: "/api/v1/github-connector/repositories"},
	{name: "list_github_branches", method: "POST", path: "/api/v1/github-connector/repository/branches",
		body: map[string]any{"repository_name": "owner/repo"}},

	// ─── Containers ────────────────────────────────────────────────────────────
	{name: "list_containers", method: "GET", path: "/api/v1/container"},
	{name: "list_containers_filtered", method: "GET", path: "/api/v1/container?status=running&page=1"},
	{name: "get_container", method: "GET", path: "/api/v1/container/container-id-abc"},
	{name: "container_logs", method: "POST", path: "/api/v1/container/container-id-abc/logs",
		body: map[string]any{"id": "container-id-abc", "tail": 100}},
	{name: "start_container", method: "POST", path: "/api/v1/container/container-id-abc/start"},
	{name: "stop_container", method: "POST", path: "/api/v1/container/container-id-abc/stop"},
	{name: "restart_container", method: "POST", path: "/api/v1/container/container-id-abc/restart"},
	{name: "remove_container", method: "DELETE", path: "/api/v1/container/container-id-abc"},
	{name: "update_container_resources", method: "PUT", path: "/api/v1/container/container-id-abc/resources",
		body: map[string]any{"memory": "512m"}},
	{name: "list_images", method: "POST", path: "/api/v1/container/images",
		body: map[string]any{"all": true}},
	{name: "prune_build_cache", method: "POST", path: "/api/v1/container/prune/build-cache"},
	{name: "prune_images", method: "POST", path: "/api/v1/container/prune/images"},

	// ─── Machines ──────────────────────────────────────────────────────────────
	{name: "list_machines", method: "GET", path: "/api/v1/machines"},
	{name: "register_machine", method: "POST", path: "/api/v1/machines",
		body: map[string]any{"name": "my-server", "host": "1.2.3.4"}},
	{name: "verify_machine_ssh", method: "POST", path: "/api/v1/machines/srv-uuid-123/verify"},
	{name: "rename_machine", method: "PATCH", path: "/api/v1/machines/srv-uuid-123/rename",
		body: map[string]any{"name": "new-name"}},
	{name: "delete_machine", method: "DELETE", path: "/api/v1/machines/srv-uuid-123"},
	{name: "ssh_status_all", method: "GET", path: "/api/v1/machines/ssh/status"},
	{name: "ssh_status_one", method: "GET", path: "/api/v1/machines/srv-uuid-123/ssh/status"},
	{name: "set_default_machine", method: "PUT", path: "/api/v1/machines/srv-uuid-123/set-default"},
	{name: "machine_stats", method: "GET", path: "/api/v1/machines/stats"},
	{name: "machine_exec", method: "POST", path: "/api/v1/machines/exec",
		body: map[string]any{"command": "df -h"}},
	{name: "machine_status", method: "GET", path: "/api/v1/machines/status"},
	{name: "restart_machine", method: "POST", path: "/api/v1/machines/restart"},
	{name: "pause_machine", method: "POST", path: "/api/v1/machines/pause"},
	{name: "resume_machine", method: "POST", path: "/api/v1/machines/resume"},
	{name: "machine_metrics", method: "GET", path: "/api/v1/machines/metrics?from=2024-01-01&to=2024-01-02"},
	{name: "machine_metrics_summary", method: "GET", path: "/api/v1/machines/metrics/summary?from=2024-01-01"},
	{name: "machine_events", method: "GET", path: "/api/v1/machines/events?limit=50"},
	{name: "machine_plans", method: "GET", path: "/api/v1/machines/plans"},
	{name: "select_machine_plan", method: "POST", path: "/api/v1/machines/plan/select",
		body: map[string]any{"plan_id": "plan-basic"}},
	{name: "machine_billing", method: "GET", path: "/api/v1/machines/billing"},
	{name: "backup_schedule", method: "GET", path: "/api/v1/machines/backup/schedule"},
	{name: "update_backup_schedule", method: "PUT", path: "/api/v1/machines/backup/schedule",
		body: map[string]any{"cron": "0 3 * * *"}},
	{name: "list_backups", method: "GET", path: "/api/v1/machines/backups"},
	{name: "trigger_backup", method: "POST", path: "/api/v1/machines/backup"},
	{name: "provision_trial", method: "POST", path: "/api/v1/machines/trial/provision"},
	{name: "trial_status", method: "GET", path: "/api/v1/machines/trial/status/sess-uuid-123"},

	// ─── System ────────────────────────────────────────────────────────────────
	{name: "health_check", method: "GET", path: "/api/v1/health"},
	{name: "check_for_updates", method: "GET", path: "/api/v1/update/check"},
	{name: "trigger_update", method: "POST", path: "/api/v1/update"},
	{name: "audit_logs", method: "GET", path: "/api/v1/audit/logs"},
	{name: "audit_logs_filtered", method: "GET", path: "/api/v1/audit/logs?resource_type=deployment&page=1"},
	{name: "list_feature_flags", method: "GET", path: "/api/v1/feature-flags"},
	{name: "check_feature_flag", method: "GET", path: "/api/v1/feature-flags/check?feature_name=agent_enabled"},
	{name: "update_feature_flag", method: "PUT", path: "/api/v1/feature-flags",
		body: map[string]any{"name": "agent_enabled", "enabled": true}},

	// ─── MCP ───────────────────────────────────────────────────────────────────
	{name: "mcp_catalog", method: "GET", path: "/api/v1/mcp/catalog"},
	{name: "list_mcp_servers", method: "GET", path: "/api/v1/mcp/servers"},
	{name: "add_mcp_server", method: "POST", path: "/api/v1/mcp/servers",
		body: map[string]any{"name": "my-mcp", "provider": "custom"}},
	{name: "update_mcp_server", method: "PUT", path: "/api/v1/mcp/servers/mcp-uuid-123",
		body: map[string]any{"name": "updated-mcp"}},
	{name: "delete_mcp_server", method: "DELETE", path: "/api/v1/mcp/servers",
		body: map[string]any{"id": "mcp-uuid-123"}},
	{name: "test_mcp_connection", method: "POST", path: "/api/v1/mcp/servers/test",
		body: map[string]any{"server_id": "mcp-uuid-123"}},
	{name: "discover_mcp_tools", method: "GET", path: "/api/v1/mcp/internal/tools"},
	{name: "list_enabled_mcp_servers", method: "GET", path: "/api/v1/mcp/internal/servers"},
	{name: "call_mcp_tool", method: "POST", path: "/api/v1/mcp/internal/tools/call",
		body: map[string]any{"server_id": "mcp-uuid-123", "tool_name": "search"}},

	// ─── Extensions ────────────────────────────────────────────────────────────
	{name: "list_extensions", method: "GET", path: "/api/v1/extensions"},
	{name: "list_extension_categories", method: "GET", path: "/api/v1/extensions/categories"},
	{name: "get_extension_by_id", method: "GET", path: "/api/v1/extensions/ext-uuid-123"},
	{name: "get_extension_by_extension_id", method: "GET", path: "/api/v1/extensions/by-extension-id/github-actions"},

	// ─── Notifications ─────────────────────────────────────────────────────────
	{name: "send_notification_slack", method: "POST", path: "/api/v1/notification/send",
		body: map[string]any{"channel": "slack", "message": "hello"}},
	{name: "send_notification_email", method: "POST", path: "/api/v1/notification/send",
		body: map[string]any{"channel": "email", "message": "hello", "subject": "Alert", "to": "user@example.com"}},
	{name: "get_notification_prefs", method: "GET", path: "/api/v1/notification/preferences"},
	{name: "update_notification_prefs", method: "PATCH", path: "/api/v1/notification/preferences",
		body: map[string]any{"slack_enabled": true}},
	{name: "get_smtp_config", method: "GET", path: "/api/v1/notification/smtp?id=org-uuid-123"},
	{name: "create_smtp_config", method: "POST", path: "/api/v1/notification/smtp",
		body: map[string]any{"host": "smtp.example.com", "port": 587}},
	{name: "update_smtp_config", method: "PUT", path: "/api/v1/notification/smtp",
		body: map[string]any{"host": "smtp.example.com"}},
	{name: "delete_smtp_config", method: "DELETE", path: "/api/v1/notification/smtp",
		body: map[string]any{"id": "smtp-uuid-123"}},
	{name: "get_webhook_notification", method: "GET", path: "/api/v1/notification/webhook/deployment"},
	{name: "create_webhook_notification", method: "POST", path: "/api/v1/notification/webhook",
		body: map[string]any{"url": "https://hooks.example.com/notify", "events": []string{"deploy"}}},
	{name: "update_webhook_notification", method: "PUT", path: "/api/v1/notification/webhook",
		body: map[string]any{"id": "wh-uuid-123", "url": "https://hooks.example.com/v2"}},
	{name: "delete_webhook_notification", method: "DELETE", path: "/api/v1/notification/webhook",
		body: map[string]any{"id": "wh-uuid-123"}},

	// ─── Health Checks ─────────────────────────────────────────────────────────
	{name: "create_health_check", method: "POST", path: "/api/v1/healthcheck",
		body: map[string]any{"application_id": "app-uuid-123", "url": "https://myapp.com/health"}},
	{name: "get_health_checks", method: "GET", path: "/api/v1/healthcheck?application_id=app-uuid-123"},
	{name: "update_health_check", method: "PUT", path: "/api/v1/healthcheck",
		body: map[string]any{"id": "hc-uuid-123", "interval_seconds": 60}},
	{name: "delete_health_check", method: "DELETE", path: "/api/v1/healthcheck?application_id=app-uuid-123"},
	{name: "toggle_health_check", method: "PATCH", path: "/api/v1/healthcheck/toggle",
		body: map[string]any{"id": "hc-uuid-123", "enabled": true}},
	{name: "health_check_results", method: "GET", path: "/api/v1/healthcheck/results?application_id=app-uuid-123"},
	{name: "health_check_stats", method: "GET", path: "/api/v1/healthcheck/stats?application_id=app-uuid-123"},
}

// ─── Tier 1: Schema validation for every operation ───────────────────────────

// TestCatalogAllOps_CorrectFormats verifies that for every catalog operation,
// the correctly-formed nixopus_api call passes JSON schema validation.
func TestCatalogAllOps_CorrectFormats(t *testing.T) {
	schema := nixopusAPISchemaFull()

	for _, op := range allCatalogOps {
		t.Run(op.name, func(t *testing.T) {
			args := buildCatalogArgs(op)
			require.NoError(t, llm.ValidateToolArgs(args, schema),
				"operation %q: correctly-formed call must pass schema\nargs: %s", op.name, string(args))
		})
	}
}

// TestCatalogAllOps_PathPrefix verifies every catalog operation uses a path
// starting with /api/v1/.
func TestCatalogAllOps_PathPrefix(t *testing.T) {
	for _, op := range allCatalogOps {
		t.Run(op.name, func(t *testing.T) {
			assert.True(t,
				len(op.path) > 0 && op.path[:8] == "/api/v1/",
				"operation %q: path must start with /api/v1/, got %q", op.name, op.path)
		})
	}
}

// TestCatalogAllOps_MethodInEnum verifies every catalog operation uses an
// allowed HTTP method.
func TestCatalogAllOps_MethodInEnum(t *testing.T) {
	allowed := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}
	for _, op := range allCatalogOps {
		t.Run(op.name, func(t *testing.T) {
			assert.True(t, allowed[op.method],
				"operation %q: method %q is not in allowed set", op.name, op.method)
		})
	}
}

// TestCatalogAllOps_ReadOpsUseGET verifies that all "list" and "get" operations
// use GET and never POST/PUT/DELETE.
func TestCatalogAllOps_ReadOpsUseGET(t *testing.T) {
	readOps := []string{
		"list_applications", "get_application",
		"list_deployments", "get_deployment_by_id", "get_deployment_logs", "get_application_logs",
		"list_artifacts", "download_artifact",
		"list_domains", "generate_subdomain", "dns_check",
		"list_github_connectors", "list_github_repositories",
		"list_containers", "get_container",
		"list_machines", "ssh_status_all", "ssh_status_one", "machine_stats",
		"machine_status", "machine_metrics", "machine_metrics_summary",
		"machine_events", "machine_plans", "machine_billing",
		"backup_schedule", "list_backups", "trial_status",
		"health_check", "check_for_updates", "audit_logs", "list_feature_flags", "check_feature_flag",
		"mcp_catalog", "list_mcp_servers", "discover_mcp_tools", "list_enabled_mcp_servers",
		"list_extensions", "list_extension_categories", "get_extension_by_id", "get_extension_by_extension_id",
		"get_notification_prefs", "get_smtp_config", "get_webhook_notification",
		"get_health_checks", "health_check_results", "health_check_stats",
		"list_compose_services", "get_app_servers",
		"get_project_family", "list_family_environments",
	}

	opMap := make(map[string]catalogOp, len(allCatalogOps))
	for _, op := range allCatalogOps {
		opMap[op.name] = op
	}

	for _, name := range readOps {
		t.Run(name, func(t *testing.T) {
			op, ok := opMap[name]
			require.True(t, ok, "read op %q not found in allCatalogOps — update the list", name)
			assert.Equal(t, "GET", op.method,
				"read operation %q must use GET, not %q", name, op.method)
		})
	}
}

// TestCatalogAllOps_MutatingOpsNotGET verifies that mutating/destructive
// operations are not incorrectly assigned GET.
func TestCatalogAllOps_MutatingOpsNotGET(t *testing.T) {
	mutatingOps := []string{
		"create_application", "update_application", "delete_application",
		"redeploy_application", "restart_deployment", "rollback_deployment",
		"cancel_deployment", "recover_application",
		"create_github_connector", "delete_github_connector",
		"start_container", "stop_container", "restart_container", "remove_container",
		"restart_machine", "pause_machine", "resume_machine",
		"trigger_backup", "trigger_update",
		"send_notification_slack", "send_notification_email",
		"create_health_check", "delete_health_check",
	}

	opMap := make(map[string]catalogOp, len(allCatalogOps))
	for _, op := range allCatalogOps {
		opMap[op.name] = op
	}

	for _, name := range mutatingOps {
		t.Run(name, func(t *testing.T) {
			op, ok := opMap[name]
			require.True(t, ok, "mutating op %q not found in allCatalogOps", name)
			assert.NotEqual(t, "GET", op.method,
				"mutating operation %q must not use GET, actual method: %q", name, op.method)
		})
	}
}

// ─── Tier 2: Old format rejection for all ops ─────────────────────────────────

// TestCatalogAllOps_OldFormatAlwaysFails verifies that the old `operation+params`
// format is rejected by schema validation for every catalog operation.
func TestCatalogAllOps_OldFormatAlwaysFails(t *testing.T) {
	schema := nixopusAPISchemaFull()

	for _, op := range allCatalogOps {
		t.Run(op.name, func(t *testing.T) {
			oldArgs, _ := json.Marshal(map[string]any{
				"operation": op.name,
				"params":    map[string]any{"id": "some-uuid"},
			})
			err := llm.ValidateToolArgs(oldArgs, schema)
			assert.Error(t, err,
				"old format for %q must fail — field 'method' is missing", op.name)
		})
	}
}

// ─── Tier 2: Mock-LLM agent probe for all ops ─────────────────────────────────

// TestCatalogAllOps_MockLLMProbe runs an agent loop for each catalog operation
// using a scripted mock LLM. Verifies:
//  1. The recorder captures the nixopus_api call.
//  2. The captured args pass schema validation.
//  3. The method and path prefix match expectations.
func TestCatalogAllOps_MockLLMProbe(t *testing.T) {
	nixopusSrv := newMockNixopusServer(t, map[string]any{"status": "success", "data": nil})
	defer nixopusSrv.Close()

	schema := nixopusAPISchemaFull()

	for _, op := range allCatalogOps {
		op := op
		t.Run(op.name, func(t *testing.T) {
			correctArgs := string(buildCatalogArgs(op))

			mockLLM := newScriptedLLM(t, []llm.Response{
				toolCallResponse("c1", "nixopus_api", correctArgs),
				textResponse(fmt.Sprintf("Done: %s completed.", op.name)),
			})
			defer mockLLM.Close()

			rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, schema)
			_, err := agent.Run(ctxWithBase(nixopusSrv.URL), op.name)
			require.NoError(t, err)

			calls := rec.CallsFor("nixopus_api")
			require.Len(t, calls, 1, "expected exactly one nixopus_api call for %q", op.name)

			assertSchemaValid(t, calls[0].Args, schema, "nixopus_api")

			var captured struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			}
			require.NoError(t, json.Unmarshal(calls[0].Args, &captured))
			assert.Equal(t, op.method, captured.Method,
				"method mismatch for %q: expected %s got %s", op.name, op.method, captured.Method)
			assert.Equal(t, op.path, captured.Path,
				"path mismatch for %q: expected %s got %s", op.name, op.path, captured.Path)
		})
	}
}

// ─── Coverage summary ─────────────────────────────────────────────────────────

// TestCatalogCoverage_Count ensures the table stays in sync with the catalog.
// Update the expected count whenever a new operation is added to catalog.go.
func TestCatalogCoverage_Count(t *testing.T) {
	const expectedMinOps = 90 // update when new operations are added
	assert.GreaterOrEqual(t, len(allCatalogOps), expectedMinOps,
		"allCatalogOps has %d entries but catalog has at least %d operations — add missing ops",
		len(allCatalogOps), expectedMinOps)
	t.Logf("catalog coverage: %d operations tested", len(allCatalogOps))
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// buildCatalogArgs builds a json.RawMessage nixopus_api call from a catalogOp.
func buildCatalogArgs(op catalogOp) json.RawMessage {
	payload := map[string]any{
		"method": op.method,
		"path":   op.path,
	}
	if len(op.body) > 0 {
		payload["body"] = op.body
	}
	b, _ := json.Marshal(payload)
	return b
}
