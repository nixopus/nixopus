//go:build agent_probe

package service

// Live agent probe tests — require a real LLM API key.
//
// Run with:
//   LLM_API_KEY=<key> go test -tags agent_probe ./internal/features/agent/service/... \
//     -run TestAgentLiveProbe -v -timeout 120s
//
// The "nonce pattern" (inspired by OpenClaw's read probe):
//   1. Inject known data into a mock Nixopus/GitHub backend
//   2. Ask the agent a question only answerable via tool call
//   3. Assert the agent retrieved the nonce (proving it called the tool)
//   4. Assert args were schema-valid (proving no hallucination)
//
// Skipped automatically if LLM_API_KEY is not set.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/agent/service/catalog"
	agentgithub "github.com/nixopus/nixopus/api/internal/features/agent/service/github"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Deploy agent: nonce probe ────────────────────────────────────────────────

// TestAgentLiveProbe_DeployAgent_Nonce injects a nonce app name into a mock Nixopus
// server and asks the deploy agent to list applications. The agent must call
// nixopus_api to retrieve the nonce — it cannot guess it.
func TestAgentLiveProbe_DeployAgent_Nonce(t *testing.T) {
	provider := getLiveProvider(t)

	nonce := fmt.Sprintf("probe-app-%d", time.Now().UnixNano())

	// Route-aware mock: nonce only appears at the correct applications endpoint.
	nixopusSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/deploy/application") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"id": "app-001", "name": nonce, "status": "running"},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": nil})
		}
	}))
	defer nixopusSrv.Close()

	schema := nixopusAPISchemaFull()
	base := llm.NewToolRegistry()
	base.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call any Nixopus API endpoint.",
		Parameters:  schema,
		Handler:     (&AgentService{}).nixopusAPIHandler,
	})

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{
		Model:        getLiveModel(),
		SystemPrompt: "You are a deploy assistant. Use nixopus_api to answer questions. Be concise.\n\n" + catalog.Catalog,
		MaxSteps:     5,
	})

	ctx := ctxWithBase(nixopusSrv.URL)
	result, err := agent.Run(ctx, "What applications do I have? List their names.")
	require.NoError(t, err)

	// The nonce can only appear in the response if the agent actually called the tool
	assert.Contains(t, result.Content, nonce,
		"nonce %q should appear in response — agent must call nixopus_api to retrieve it", nonce)

	calls := rec.CallsFor("nixopus_api")
	require.NotEmpty(t, calls, "agent must have called nixopus_api at least once")

	for _, c := range calls {
		assertSchemaValid(t, c.Args, schema, "nixopus_api")

		var args struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		json.Unmarshal(c.Args, &args)
		t.Logf("nixopus_api call: method=%s path=%s", args.Method, args.Path)

		assert.Equal(t, "GET", args.Method, "listing applications must use GET")
		assert.True(t, strings.HasPrefix(args.Path, "/api/v1/"),
			"path must start with /api/v1/, got %q — LLM is hallucinating a path outside the catalog", args.Path)
		assert.True(t,
			strings.HasPrefix(args.Path, "/api/v1/deploy/application"),
			"deploy agent listing apps must use /api/v1/deploy/application..., got %q", args.Path)
	}
}

// ─── Diagnostic agent: nonce probe ───────────────────────────────────────────

// TestAgentLiveProbe_DiagnosticAgent_Nonce injects a nonce into mock application
// logs and asks the diagnostic agent to fetch them. Route-aware: the nonce is
// only returned at the correct app-logs endpoint.
func TestAgentLiveProbe_DiagnosticAgent_Nonce(t *testing.T) {
	provider := getLiveProvider(t)

	nonce := fmt.Sprintf("CRASH-NONCE-%d", time.Now().UnixNano())
	appID := "app-001"

	// Route-aware mock:
	//   /api/v1/deploy/application/logs/{app_id} → returns logs containing nonce
	//   everything else → empty 200
	nixopusSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/deploy/application/logs/") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{{"level": "ERROR", "message": nonce + ": out of memory"}},
			})
		} else if strings.HasPrefix(r.URL.Path, "/api/v1/deploy/application/deployments") {
			// Return a real-looking deployment list so the LLM knows an app exists
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"id": "dep-abc", "status": "failed", "application_id": appID},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": nil})
		}
	}))
	defer nixopusSrv.Close()

	schema := nixopusDiagnosticSchema()
	base := llm.NewToolRegistry()
	base.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call Nixopus API for diagnostic data.",
		Parameters:  schema,
		Handler:     (&AgentService{}).nixopusAPIHandler,
	})

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{
		Model:        getLiveModel(),
		SystemPrompt: "You are a diagnostic assistant. Use nixopus_api to fetch data. Be concise.\n\n" + catalog.Catalog,
		MaxSteps:     5,
	})

	ctx := ctxWithBase(nixopusSrv.URL)
	result, err := agent.Run(ctx,
		fmt.Sprintf("Fetch the application logs for application %s and tell me about any errors.", appID))
	require.NoError(t, err)

	assert.Contains(t, result.Content, nonce,
		"nonce %q should appear — agent must call /api/v1/deploy/application/logs/%s", nonce, appID)

	calls := rec.CallsFor("nixopus_api")
	require.NotEmpty(t, calls)

	var logsCall *llm.RecordedCall
	for i := range calls {
		var args struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		json.Unmarshal(calls[i].Args, &args)
		t.Logf("nixopus_api call: method=%s path=%s", args.Method, args.Path)
		assertSchemaValid(t, calls[i].Args, schema, "nixopus_api")
		if strings.HasPrefix(args.Path, "/api/v1/deploy/application/logs/") {
			logsCall = &calls[i]
		}
	}

	require.NotNil(t, logsCall,
		"at least one call must be to /api/v1/deploy/application/logs/{id}")
}

// ─── GitHub agent: nonce probe ────────────────────────────────────────────────

// TestAgentLiveProbe_GithubAgent_ListPRs asks the GitHub agent to list PRs.
// A mock GitHub API returns a nonce PR title — the agent must call the correct
// tool with valid owner/repo params to see it.
func TestAgentLiveProbe_GithubAgent_ListPRs(t *testing.T) {
	provider := getLiveProvider(t)

	nonce := fmt.Sprintf("pr-nonce-%d", time.Now().UnixNano())

	// Mock GitHub API
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/probe-org/probe-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"number": 1, "title": nonce, "state": "open"},
		})
	})
	ghSrv := httptest.NewServer(mux)
	defer ghSrv.Close()
	restore := agentgithub.RedirectAPIToTestServer(ghSrv.URL)
	defer restore()

	gc := agentgithub.NewProbeClient("test-token", func(ctx context.Context) string {
		v, _ := ctx.Value(ctxKeyOrgID).(string)
		return v
	})
	tool := agentgithub.ListPullRequestsTool(gc)

	base := llm.NewToolRegistry()
	base.Register(tool)

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{
		Model:        getLiveModel(),
		SystemPrompt: "You are a GitHub assistant. Use github_list_pull_requests to fetch PR data. Be concise.",
		MaxSteps:     5,
	})

	ctx := context.WithValue(context.Background(), ctxKeyOrgID, "")
	result, err := agent.Run(ctx, "List the open pull requests for probe-org/probe-repo.")
	require.NoError(t, err)

	assert.Contains(t, result.Content, nonce,
		"nonce PR title %q must appear in response", nonce)

	calls := rec.CallsFor("github_list_pull_requests")
	require.NotEmpty(t, calls, "agent must have called github_list_pull_requests")

	for _, c := range calls {
		assertSchemaValid(t, c.Args, tool.Parameters, "github_list_pull_requests")

		var args struct {
			Owner string `json:"owner"`
			Repo  string `json:"repo"`
		}
		json.Unmarshal(c.Args, &args)
		assert.NotEmpty(t, args.Owner, "owner must not be empty")
		assert.NotEmpty(t, args.Repo, "repo must not be empty")
		t.Logf("github_list_pull_requests call: owner=%s repo=%s args=%s", args.Owner, args.Repo, c.Args)
	}
}

// ─── HTTP probe agent: nonce probe ────────────────────────────────────────────

// TestAgentLiveProbe_HttpProbe asks the deploy agent to probe a URL.
// The nonce is injected as a response header that the agent should surface.
func TestAgentLiveProbe_HttpProbe(t *testing.T) {
	provider := getLiveProvider(t)

	nonce := fmt.Sprintf("probe-%d", time.Now().UnixNano())
	probeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Probe-Nonce", nonce)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","nonce":"%s"}`, nonce)
	}))
	defer probeSrv.Close()

	httpProbeSchema := json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"},"timeout_seconds":{"type":"integer"}},"required":["url"]}`)

	base := llm.NewToolRegistry()
	base.Register(llm.ToolDefinition{
		Name:        "http_probe",
		Description: "Check if a URL is accessible.",
		Parameters:  httpProbeSchema,
		Handler:     httpProbeHandler,
	})

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{
		Model:        getLiveModel(),
		SystemPrompt: "You are a health-check assistant. Use http_probe to check URL accessibility. Be concise.",
		MaxSteps:     3,
	})

	result, err := agent.Run(context.Background(),
		fmt.Sprintf("Is %s accessible? What status code does it return?", probeSrv.URL))
	require.NoError(t, err)

	calls := rec.CallsFor("http_probe")
	require.NotEmpty(t, calls, "agent must have called http_probe")

	for _, c := range calls {
		assertSchemaValid(t, c.Args, httpProbeSchema, "http_probe")

		var args struct {
			URL string `json:"url"`
		}
		json.Unmarshal(c.Args, &args)
		assert.True(t, strings.HasPrefix(args.URL, "http"), "url must be a valid HTTP URL, got %q", args.URL)
		t.Logf("http_probe call: url=%s", args.URL)
	}

	// Response should mention success (200)
	assert.True(t,
		strings.Contains(result.Content, "200") || strings.Contains(result.Content, "accessible") || strings.Contains(result.Content, "reachable"),
		"response should mention the probe result, got: %s", result.Content)
}

// ─── Machine agent: nonce probe ───────────────────────────────────────────────

// TestAgentLiveProbe_MachineAgent_Stats asks the machine agent for server stats.
// A nonce is embedded in the cpu_model field — the agent must call /api/v1/machines/stats
// to retrieve it. Route-aware mock: only the stats endpoint returns the nonce.
func TestAgentLiveProbe_MachineAgent_Stats(t *testing.T) {
	provider := getLiveProvider(t)

	nonce := fmt.Sprintf("Intel-NONCE-%d", time.Now().UnixNano())

	// Route-aware mock: only /api/v1/machines/stats returns the nonce cpu_model.
	// Any other path gets an empty response, forcing the LLM to call the correct endpoint.
	nixopusSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/machines/stats" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"cpu_model": nonce,
					"cpu":       12.5,
					"memory_gb": 4,
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": nil})
		}
	}))
	defer nixopusSrv.Close()

	schema := nixopusMachineSchema()
	base := llm.NewToolRegistry()
	base.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call Nixopus machine management API.",
		Parameters:  schema,
		Handler:     (&AgentService{}).nixopusAPIHandler,
	})

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{
		Model:        getLiveModel(),
		SystemPrompt: "You are a machine management assistant. Use nixopus_api to get server information. Be concise.\n\n" + catalog.Catalog,
		MaxSteps:     5,
	})

	ctx := ctxWithBase(nixopusSrv.URL)
	result, err := agent.Run(ctx, "What is the CPU model of my server? Use nixopus_api to get machine stats.")
	require.NoError(t, err)

	assert.Contains(t, result.Content, nonce,
		"nonce cpu_model %q must appear in response — agent must call /api/v1/machines/stats", nonce)

	calls := rec.CallsFor("nixopus_api")
	require.NotEmpty(t, calls)

	// At least the first call must be to the stats endpoint
	var statsCall *llm.RecordedCall
	for i := range calls {
		var args struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		json.Unmarshal(calls[i].Args, &args)
		t.Logf("nixopus_api call: %s", calls[i].Args)
		assertSchemaValid(t, calls[i].Args, schema, "nixopus_api")
		if args.Path == "/api/v1/machines/stats" {
			statsCall = &calls[i]
		}
	}

	require.NotNil(t, statsCall,
		"at least one call must be to /api/v1/machines/stats — LLM used wrong endpoint for machine stats")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func getLiveProvider(t *testing.T) llm.Provider {
	t.Helper()
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		t.Skip("LLM_API_KEY not set — skipping live probe test")
	}
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return llm.NewOpenAIProvider(llm.OpenAIConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Headers: map[string]string{
			"HTTP-Referer": "https://nixopus.com",
			"X-Title":      "Nixopus Agent Probe Tests",
		},
	})
}

func getLiveModel() string {
	if m := os.Getenv("LLM_MODEL"); m != "" {
		return m
	}
	return "openai/gpt-4o-mini"
}

// ─── Full catalog live probes ─────────────────────────────────────────────────
//
// TestCatalogLiveProbe_AllOps runs a live LLM probe for every API catalog
// operation defined in allCatalogOps (catalog_allpaths_test.go).
//
// For each operation:
//  1. A route-aware mock Nixopus server injects a nonce ONLY at the correct path.
//  2. The LLM is asked a natural-language question that maps to this operation.
//  3. We assert:
//     - The agent called nixopus_api with the correct method.
//     - The path starts with /api/v1/ and matches the expected prefix.
//     - Schema validation passes.
//     - For READ ops: the nonce appears in the LLM's response.
//
// Run with:
//
//	LLM_API_KEY=<key> go test -tags agent_probe -run TestCatalogLiveProbe_AllOps -v -timeout 600s

// liveProbeEntry maps one catalogOp to the data needed for a live probe.
type liveProbeEntry struct {
	op         catalogOp
	question   string // natural-language question to ask the agent
	nonceField string // JSON field name where nonce is injected in mock response
	skipReason string // non-empty → skip with this reason
}

// buildLiveProbeTable returns the question and mock config for every catalogOp.
func buildLiveProbeTable() []liveProbeEntry {
	return []liveProbeEntry{
		// Applications
		{op: byName("list_applications"), question: "What applications do I have? List their names.", nonceField: "name"},
		{op: byName("list_applications_paged"), question: "List my first page of applications.", nonceField: "name"},
		{op: byName("get_application"), question: "Get the details of application app-uuid-123.", nonceField: "name"},
		{op: byName("list_deployments"), question: "List all deployments for application app-uuid-123.", nonceField: "status"},
		{op: byName("list_deployments_paged"), question: "List the first page of deployments for application app-uuid-123.", nonceField: "status"},
		{op: byName("get_deployment_by_id"), question: "Get the details of deployment dep-uuid-456.", nonceField: "status"},
		{op: byName("get_deployment_logs"), question: "Get the logs for deployment dep-uuid-456.", nonceField: "message"},
		{op: byName("get_deployment_logs_filtered"), question: "Get the error-level logs for deployment dep-uuid-456.", nonceField: "message"},
		{op: byName("get_application_logs"), question: "Get the application logs for app app-uuid-123.", nonceField: "message"},
		{op: byName("create_application"), question: "Deploy the repository owner/repo to Nixopus as my-app on port 3000.", skipReason: "mutating — verified by method+path check"},
		{op: byName("update_application"), question: "Update application app-uuid-123 to run on port 8080.", skipReason: "mutating — verified by method+path check"},
		{op: byName("delete_application"), question: "Delete application app-uuid-123.", skipReason: "destructive — skip live LLM call"},
		{op: byName("redeploy_application"), question: "Rebuild and redeploy application app-uuid-123.", skipReason: "mutating — verified by method+path check"},
		{op: byName("restart_deployment"), question: "Restart deployment dep-uuid-456 without rebuilding.", skipReason: "mutating — verified by method+path check"},
		{op: byName("rollback_deployment"), question: "Roll back application app-uuid-123 to its previous deployment.", skipReason: "mutating — verified by method+path check"},
		{op: byName("cancel_deployment"), question: "Cancel the in-flight deployment dep-uuid-456.", skipReason: "mutating — verified by method+path check"},
		{op: byName("recover_application"), question: "Recover application app-uuid-123.", skipReason: "mutating — verified by method+path check"},
		{op: byName("update_labels"), question: "Update the labels for application app-uuid-123 to prod and v2.", skipReason: "mutating — verified by method+path check"},
		{op: byName("add_domain_to_app"), question: "Add domain myapp.example.com to application app-uuid-123.", skipReason: "mutating — verified by method+path check"},
		{op: byName("remove_domain_from_app"), question: "Remove domain myapp.example.com from application app-uuid-123.", skipReason: "mutating — verified by method+path check"},
		{op: byName("list_compose_services"), question: "List the compose services for application app-uuid-123.", nonceField: "service_name"},
		{op: byName("get_app_servers"), question: "What servers is application app-uuid-123 deployed on?", nonceField: "server_id"},

		// Projects
		{op: byName("create_project"), question: "Create a new project called my-project from repository owner/repo.", skipReason: "mutating — verified by method+path check"},
		{op: byName("get_project_family"), question: "Get the project family for family-id fam-uuid-789.", nonceField: "name"},
		{op: byName("list_family_environments"), question: "List all environments in family fam-uuid-789.", nonceField: "name"},

		// Artifacts
		{op: byName("list_artifacts"), question: "List the deployment artifacts for application app-uuid-123.", nonceField: "artifact_id"},
		{op: byName("download_artifact"), question: "Get the download URL for artifact from deployment dep-uuid-456.", nonceField: "url"},

		// Domains
		{op: byName("list_domains"), question: "List all my domains.", nonceField: "name"},
		{op: byName("list_domains_by_type"), question: "List all my custom domains.", nonceField: "name"},
		{op: byName("generate_subdomain"), question: "Generate a random subdomain for me.", nonceField: "subdomain"},
		{op: byName("dns_check"), question: "Check the DNS status for custom domain dom-uuid-123.", nonceField: "status"},
		{op: byName("create_custom_domain"), question: "Add custom domain custom.example.com.", skipReason: "mutating — verified by method+path check"},

		// GitHub Connectors
		{op: byName("list_github_connectors"), question: "List all my GitHub app connectors.", nonceField: "slug"},
		{op: byName("list_github_repositories"), question: "List all GitHub repositories accessible through my connectors.", nonceField: "full_name"},
		{op: byName("list_github_branches"), question: "List the branches for repository owner/repo.", nonceField: "name"},

		// Containers
		{op: byName("list_containers"), question: "List all my running containers.", nonceField: "name"},
		{op: byName("list_containers_filtered"), question: "List all running containers.", nonceField: "name"},
		{op: byName("get_container"), question: "Get details of container container-id-abc.", nonceField: "image"},
		{op: byName("start_container"), question: "Start container container-id-abc.", skipReason: "mutating — verified by method+path check"},
		{op: byName("stop_container"), question: "Stop container container-id-abc.", skipReason: "mutating — verified by method+path check"},
		{op: byName("restart_container"), question: "Restart container container-id-abc.", skipReason: "mutating — verified by method+path check"},
		{op: byName("remove_container"), question: "Remove container container-id-abc.", skipReason: "destructive — skip live LLM call"},
		{op: byName("list_images"), question: "List all Docker images on the system.", nonceField: "repo_tags"},

		// Machines
		{op: byName("list_machines"), question: "List all my registered servers.", nonceField: "name"},
		{op: byName("ssh_status_all"), question: "List the name of each machine and its SSH connectivity status.", nonceField: "name"},
		{op: byName("ssh_status_one"), question: "Check the SSH status of machine srv-uuid-123.", nonceField: "status"},
		{op: byName("machine_stats"), question: "Get the current CPU model and hardware stats of my server.", nonceField: "cpu_model"},
		{op: byName("machine_status"), question: "Get the lifecycle status of my provisioned machine.", nonceField: "state"},
		{op: byName("machine_metrics"), question: "Get the time-series metrics for my machine from 2024-01-01 to 2024-01-02.", nonceField: "value"},
		{op: byName("machine_metrics_summary"), question: "What metric names are included in the summary for my machine since 2024-01-01?", nonceField: "name"},
		{op: byName("machine_events"), question: "List the lifecycle events for my machine.", nonceField: "event_type"},
		{op: byName("machine_plans"), question: "What machine plans are available?", nonceField: "name"},
		{op: byName("machine_billing"), question: "What is the billing status for my machine?", nonceField: "plan"},
		{op: byName("backup_schedule"), question: "What is the current backup schedule for my machine?", nonceField: "cron"},
		{op: byName("list_backups"), question: "List all my machine backups.", nonceField: "backup_id"},
		{op: byName("trigger_backup"), question: "Trigger an immediate backup of my machine.", skipReason: "mutating — verified by method+path check"},
		{op: byName("restart_machine"), question: "Restart my machine.", skipReason: "mutating — verified by method+path check"},
		{op: byName("machine_exec"), question: "Run the command df -h on my host machine.", skipReason: "mutating — verified by method+path check"},

		// System
		{op: byName("health_check"), question: "Check the overall system health status.", nonceField: "version"},
		{op: byName("check_for_updates"), question: "Check if there are any system updates available.", nonceField: "latest_version"},
		{op: byName("audit_logs"), question: "Show me the recent audit log entries.", nonceField: "action"},
		{op: byName("audit_logs_filtered"), question: "Show me the recent deployment audit log entries.", nonceField: "action"},
		{op: byName("list_feature_flags"), question: "List all feature flags.", nonceField: "name"},
		{op: byName("check_feature_flag"), question: "Use the feature-flags check endpoint to look up the agent_enabled flag and tell me its name.", nonceField: "name"},
		{op: byName("update_feature_flag"), question: "Enable the agent_enabled feature flag.", skipReason: "mutating — verified by method+path check"},

		// MCP
		{op: byName("mcp_catalog"), question: "List all available MCP provider integrations in the catalog.", nonceField: "name"},
		{op: byName("list_mcp_servers"), question: "List all my configured MCP servers.", nonceField: "name"},
		{op: byName("discover_mcp_tools"), question: "Discover what MCP tools are available on my servers.", nonceField: "tool_name"},
		{op: byName("list_enabled_mcp_servers"), question: "List all enabled MCP servers with their credentials.", nonceField: "name"},

		// Extensions
		{op: byName("list_extensions"), question: "List all available Nixopus extensions.", nonceField: "name"},
		{op: byName("list_extension_categories"), question: "List all extension categories.", nonceField: "name"},
		{op: byName("get_extension_by_id"), question: "Get the details of extension ext-uuid-123.", nonceField: "name"},
		{op: byName("get_extension_by_extension_id"), question: "Get the extension with extension ID github-actions.", nonceField: "name"},

		// Notifications
		{op: byName("send_notification_slack"), question: "Send a Slack notification saying 'deploy done'.", skipReason: "mutating — verified by method+path check"},
		{op: byName("get_notification_prefs"), question: "Get my notification preferences.", nonceField: "channel"},
		{op: byName("get_smtp_config"), question: "Get the SMTP configuration for my organisation org-uuid-123.", nonceField: "host"},
		{op: byName("get_webhook_notification"), question: "Get the webhook configuration for deployment events.", nonceField: "url"},

		// Health Checks
		{op: byName("get_health_checks"), question: "Get all health checks configured for application app-uuid-123.", nonceField: "url"},
		{op: byName("health_check_results"), question: "Show the health check results for application app-uuid-123.", nonceField: "status"},
		{op: byName("health_check_stats"), question: "Show the health check statistics for application app-uuid-123.", nonceField: "uptime_percent"},
	}
}

// byName looks up a catalogOp by name. Panics if not found (test authoring error).
func byName(name string) catalogOp {
	for _, op := range allCatalogOps {
		if op.name == name {
			return op
		}
	}
	panic(fmt.Sprintf("byName: operation %q not found in allCatalogOps — check spelling", name))
}

// TestCatalogLiveProbe_AllOps runs one live LLM + route-aware mock probe per
// catalog operation. Each test is fully isolated.
func TestCatalogLiveProbe_AllOps(t *testing.T) {
	provider := getLiveProvider(t)
	schema := nixopusAPISchemaFull()
	entries := buildLiveProbeTable()

	for _, entry := range entries {
		entry := entry
		t.Run(entry.op.name, func(t *testing.T) {
			if entry.skipReason != "" {
				t.Skip(entry.skipReason)
			}

			nonce := fmt.Sprintf("PROBE-%s-%d", entry.op.name, time.Now().UnixNano())

			// Route-aware mock: nonce only served at the canonical path for this operation.
			// All other paths return empty success so the LLM can still make exploratory calls.
			canonicalPath := canonicalPathFor(entry.op)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if pathMatches(r.URL.Path, canonicalPath) {
					json.NewEncoder(w).Encode(buildNonceResponse(entry.nonceField, nonce))
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": nil})
				}
			}))
			defer srv.Close()

			base := llm.NewToolRegistry()
			base.Register(llm.ToolDefinition{
				Name:        "nixopus_api",
				Description: "Call any Nixopus API endpoint using method, path, and optional body.",
				Parameters:  schema,
				Handler:     (&AgentService{}).nixopusAPIHandler,
			})
			rec := &llm.ToolCallRecorder{}
			agent := llm.NewAgent(provider, llm.WrapWithRecorder(base, rec), llm.AgentConfig{
				Model:        getLiveModel(),
				SystemPrompt: "You are a Nixopus assistant. Use nixopus_api to answer questions. Be concise.\n\n" + catalog.Catalog,
				MaxSteps:     6,
			})

			ctx := ctxWithBase(srv.URL)
			result, err := agent.Run(ctx, entry.question)
			require.NoError(t, err)

			calls := rec.CallsFor("nixopus_api")
			require.NotEmpty(t, calls,
				"op %q: agent must call nixopus_api at least once", entry.op.name)

			// Log all calls for visibility
			for _, c := range calls {
				var a struct {
					Method string `json:"method"`
					Path   string `json:"path"`
				}
				json.Unmarshal(c.Args, &a)
				t.Logf("[%s] %s %s", entry.op.name, a.Method, a.Path)
			}

			// All calls must pass schema
			for _, c := range calls {
				assertSchemaValid(t, c.Args, schema, "nixopus_api")
			}

			// At least one call must use correct method and hit the right path area
			var matchedCall *llm.RecordedCall
			for i := range calls {
				var a struct {
					Method string `json:"method"`
					Path   string `json:"path"`
				}
				json.Unmarshal(calls[i].Args, &a)
				if a.Method == entry.op.method && pathMatches(a.Path, canonicalPath) {
					matchedCall = &calls[i]
					break
				}
			}
			require.NotNil(t, matchedCall,
				"op %q: no call matched method=%s path≈%s\nall calls:\n%s",
				entry.op.name, entry.op.method, canonicalPath, formatCalls(calls))

			// All paths must start with /api/v1/ (trim any LLM-added whitespace)
			for _, c := range calls {
				var a struct {
					Path string `json:"path"`
				}
				json.Unmarshal(c.Args, &a)
				trimmed := strings.TrimSpace(a.Path)
				assert.True(t, strings.HasPrefix(trimmed, "/api/v1/"),
					"op %q: all paths must start with /api/v1/, got %q", entry.op.name, a.Path)
			}

			// For READ ops, the nonce must appear in the final response
			if entry.op.method == "GET" {
				assert.Contains(t, result.Content, nonce,
					"op %q: nonce %q must appear in response (LLM must report what it fetched)", entry.op.name, nonce)
			}
		})
	}
}

// canonicalPathFor returns the path prefix used to match mock server routing.
// Path-param placeholders like {app_uuid} are replaced with their example values
// from the catalogOp, so the mock matches the same URL the LLM will call.
func canonicalPathFor(op catalogOp) string {
	// Strip query string — match on path component only
	path := op.path
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	return path
}

// pathMatches returns true if requestPath starts with or equals prefix.
// Handles both exact paths and prefix matching for path-param routes.
func pathMatches(requestPath, prefix string) bool {
	requestPath = strings.TrimSpace(requestPath)
	return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") || strings.HasPrefix(requestPath, prefix+"?")
}

// buildNonceResponse constructs a mock response with the nonce in the
// specified field, nested under the canonical "data" wrapper.
func buildNonceResponse(field, nonce string) map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
		"data": []map[string]interface{}{
			{field: nonce, "id": "item-001"},
		},
	}
}

// formatCalls renders recorded calls as a readable string for failure messages.
func formatCalls(calls []llm.RecordedCall) string {
	var sb strings.Builder
	for _, c := range calls {
		sb.WriteString("  ")
		sb.Write(c.Args)
		sb.WriteByte('\n')
	}
	return sb.String()
}
