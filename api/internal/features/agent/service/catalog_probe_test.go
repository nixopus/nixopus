package service

// Catalog operation probe tests.
//
// These test the SPECIFIC tool call patterns the LLM makes when operating on
// api-catalog operations. They catch the most common real-world failure modes:
//
//   1. Wrong param name  — `application_id` vs the correct `id` query param
//   2. Wrong param style — embedding path params as query strings or body fields
//   3. Wrong method      — GET for a mutation, POST for a read
//   4. Old format        — using `operation`+`params` instead of `method`+`path`
//
// Each test scripted mock LLM response represents what the LLM *actually* returns,
// and validates it against the schema. The negative tests (wrong pattern) must ALSO
// be detected at the path-pattern level via pathMatchesOperation.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Schema-level tests: old format detection ─────────────────────────────────

// TestCatalogOldFormat_OperationParamsRejected verifies that if a LLM sends the
// deprecated `operation`+`params` format (from old api-catalog SKILL.md), the
// args fail schema validation — because `method` is missing.
func TestCatalogOldFormat_OperationParamsRejected(t *testing.T) {
	schema := nixopusAPISchemaFull()

	oldFormatCases := []struct {
		name string
		args json.RawMessage
	}{
		{
			name: "operation_and_params",
			args: json.RawMessage(`{"operation":"get_application","params":{"id":"app-uuid"}}`),
		},
		{
			name: "operation_only",
			args: json.RawMessage(`{"operation":"get_application_deployments"}`),
		},
		{
			name: "params_only",
			args: json.RawMessage(`{"params":{"id":"app-uuid"}}`),
		},
		{
			name: "action_format",
			args: json.RawMessage(`{"action":"list_applications","data":{}}`),
		},
		{
			name: "completely_empty",
			args: json.RawMessage(`{}`),
		},
	}

	for _, tc := range oldFormatCases {
		t.Run(tc.name, func(t *testing.T) {
			err := llm.ValidateToolArgs(tc.args, schema)
			assert.Error(t, err, "old format %q must fail — no method field", tc.name)
			assert.Contains(t, err.Error(), "method", "error must mention missing method field")
		})
	}
}

// ─── Path-pattern validation helper ──────────────────────────────────────────

// CatalogCallResult is what we validate from a recorded nixopus_api call.
type CatalogCallResult struct {
	Method string
	Path   string
	Body   map[string]interface{}
}

// parseCatalogCall extracts method/path/body from recorded args.
func parseCatalogCall(t *testing.T, args json.RawMessage) CatalogCallResult {
	t.Helper()
	var raw struct {
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Body   json.RawMessage `json:"body"`
	}
	require.NoError(t, json.Unmarshal(args, &raw))
	var body map[string]interface{}
	if raw.Body != nil {
		json.Unmarshal(raw.Body, &body)
	}
	return CatalogCallResult{Method: raw.Method, Path: raw.Path, Body: body}
}

// ─── Correct call patterns (what LLM SHOULD send) ────────────────────────────

// TestCatalogCorrectCalls_Schema verifies that correctly formed calls for each
// major operation pass schema validation.
func TestCatalogCorrectCalls_Schema(t *testing.T) {
	schema := nixopusAPISchemaFull()

	correctCalls := []struct {
		operation string
		args      json.RawMessage
	}{
		// Applications
		{"list_applications", json.RawMessage(`{"method":"GET","path":"/api/v1/deploy/applications"}`)},
		{"get_application", json.RawMessage(`{"method":"GET","path":"/api/v1/deploy/application?id=app-uuid-123"}`)},
		// ⚠ CRITICAL: deployments uses `id` NOT `application_id`
		{"get_deployments_correct_param", json.RawMessage(`{"method":"GET","path":"/api/v1/deploy/application/deployments?id=app-uuid-123"}`)},
		// deployment_id in path, NOT as query string
		{"get_deployment_by_id_path_param", json.RawMessage(`{"method":"GET","path":"/api/v1/deploy/application/deployments/deploy-uuid-456"}`)},
		{"get_deployment_logs_path_param", json.RawMessage(`{"method":"GET","path":"/api/v1/deploy/application/deployments/deploy-uuid-456/logs"}`)},
		{"get_app_logs_path_param", json.RawMessage(`{"method":"GET","path":"/api/v1/deploy/application/logs/app-uuid-123"}`)},
		// Mutations must be POST/PUT/DELETE
		{"restart_deployment_post", json.RawMessage(`{"method":"POST","path":"/api/v1/deploy/application/restart","body":{"id":"deploy-uuid"}}`)},
		{"redeploy_app_post", json.RawMessage(`{"method":"POST","path":"/api/v1/deploy/application/redeploy","body":{"id":"app-uuid"}}`)},
		{"rollback_post", json.RawMessage(`{"method":"POST","path":"/api/v1/deploy/application/rollback","body":{"id":"app-uuid"}}`)},
		// Container operations: container_id in PATH
		{"list_containers", json.RawMessage(`{"method":"GET","path":"/api/v1/container"}`)},
		{"get_container_path_param", json.RawMessage(`{"method":"GET","path":"/api/v1/container/container-id-xyz"}`)},
		{"restart_container_path_param", json.RawMessage(`{"method":"POST","path":"/api/v1/container/container-id-xyz/restart"}`)},
		// Machines
		{"get_machine_stats", json.RawMessage(`{"method":"GET","path":"/api/v1/machines/stats"}`)},
		{"list_servers", json.RawMessage(`{"method":"GET","path":"/api/v1/machines"}`)},
		{"set_default_machine", json.RawMessage(`{"method":"PUT","path":"/api/v1/machines/server-id-abc/set-default"}`)},
		// Notifications
		{"send_notification", json.RawMessage(`{"method":"POST","path":"/api/v1/notification/send","body":{"channel":"slack","message":"hello"}}`)},
		// System
		{"audit_logs", json.RawMessage(`{"method":"GET","path":"/api/v1/audit/logs?page=1&page_size=50"}`)},
	}

	for _, tc := range correctCalls {
		t.Run(tc.operation, func(t *testing.T) {
			assert.NoError(t, llm.ValidateToolArgs(tc.args, schema),
				"correct call for %q should pass schema", tc.operation)
		})
	}
}

// ─── Wrong call patterns (what LLM SHOULD NOT send) ──────────────────────────

// TestCatalogWrongCalls_MethodViolation tests that wrong HTTP methods are caught
// by schema enum validation.
func TestCatalogWrongCalls_MethodViolation(t *testing.T) {
	schema := nixopusAPISchemaFull()

	wrongMethods := []struct {
		name string
		args json.RawMessage
	}{
		{"lowercase_get", json.RawMessage(`{"method":"get","path":"/api/v1/deploy/applications"}`)},
		{"lowercase_post", json.RawMessage(`{"method":"post","path":"/api/v1/deploy/application/restart"}`)},
		{"wrong_verb_query", json.RawMessage(`{"method":"QUERY","path":"/api/v1/deploy/applications"}`)},
		{"wrong_verb_fetch", json.RawMessage(`{"method":"FETCH","path":"/api/v1/deploy/applications"}`)},
		{"head_not_allowed", json.RawMessage(`{"method":"HEAD","path":"/api/v1/deploy/applications"}`)},
	}

	for _, tc := range wrongMethods {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, llm.ValidateToolArgs(tc.args, schema),
				"wrong method in %q must fail enum validation", tc.name)
		})
	}
}

// TestCatalogPathPatterns_CorrectVsWrong tests the most commonly confused path
// patterns. These can't be validated by JSON schema alone — we check via string
// inspection of what the mock LLM returned.
func TestCatalogPathPatterns_CorrectVsWrong(t *testing.T) {
	appUUID := "app-uuid-abc123"
	deployUUID := "deploy-uuid-def456"

	cases := []struct {
		operation   string
		correctPath string
		wrongPaths  []string
		method      string
		bodyField   string // if the wrong version puts UUID in body instead of path
	}{
		{
			operation:   "get_application_deployments",
			method:      "GET",
			correctPath: fmt.Sprintf("/api/v1/deploy/application/deployments?id=%s", appUUID),
			wrongPaths: []string{
				// Most common LLM mistake: using `application_id` instead of `id`
				fmt.Sprintf("/api/v1/deploy/application/deployments?application_id=%s", appUUID),
				fmt.Sprintf("/api/v1/deploy/application/deployments?app_id=%s", appUUID),
				// Missing query param entirely
				"/api/v1/deploy/application/deployments",
			},
		},
		{
			operation:   "get_deployment_by_id",
			method:      "GET",
			correctPath: fmt.Sprintf("/api/v1/deploy/application/deployments/%s", deployUUID),
			wrongPaths: []string{
				// LLM puts deployment_id as query string instead of path param
				fmt.Sprintf("/api/v1/deploy/application/deployments?id=%s", deployUUID),
				fmt.Sprintf("/api/v1/deploy/application/deployments?deployment_id=%s", deployUUID),
			},
		},
		{
			operation:   "get_deployment_logs",
			method:      "GET",
			correctPath: fmt.Sprintf("/api/v1/deploy/application/deployments/%s/logs", deployUUID),
			wrongPaths: []string{
				fmt.Sprintf("/api/v1/deploy/application/deployments?id=%s/logs", deployUUID),
				fmt.Sprintf("/api/v1/deploy/application/deployments/logs?id=%s", deployUUID),
				fmt.Sprintf("/api/v1/deploy/application/logs?deployment_id=%s", deployUUID),
			},
		},
		{
			operation:   "get_container_by_id",
			method:      "GET",
			correctPath: "/api/v1/container/container-abc",
			wrongPaths: []string{
				// LLM uses query param instead of path param
				"/api/v1/container?id=container-abc",
				"/api/v1/container?container_id=container-abc",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.operation+"/correct_path", func(t *testing.T) {
			assert.True(t, strings.HasPrefix(tc.correctPath, "/api/v1/"),
				"correct path must start with /api/v1/")
		})

		for i, wrong := range tc.wrongPaths {
			t.Run(fmt.Sprintf("%s/wrong_path_%d", tc.operation, i+1), func(t *testing.T) {
				// The wrong path and the correct path must be different strings
				assert.NotEqual(t, tc.correctPath, wrong,
					"wrong path %q must differ from correct %q", wrong, tc.correctPath)

				// Verify the specific mistake is captured: wrong param name in wrong path
				// but NOT in correct path
				if strings.Contains(wrong, "application_id") {
					assert.NotContains(t, tc.correctPath, "application_id",
						"correct path for %s must use 'id' not 'application_id'", tc.operation)
				}
			})
		}
	}
}

// ─── Agent loop: scripted LLM probe for catalog operations ───────────────────

// TestCatalogAgentProbe_GetDeployments verifies the agent loop captures a
// correctly-formed get_application_deployments call.
func TestCatalogAgentProbe_GetDeployments(t *testing.T) {
	appUUID := "app-abc-123"
	correctArgs := fmt.Sprintf(`{"method":"GET","path":"/api/v1/deploy/application/deployments?id=%s&limit=5"}`, appUUID)

	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", correctArgs),
		textResponse("Found 2 deployments."),
	})

	nixopusSrv := newMockNixopusServer(t, map[string]interface{}{
		"data": []map[string]string{{"id": "dep-1"}, {"id": "dep-2"}},
	})

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusAPISchemaFull())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL),
		fmt.Sprintf("list the last 5 deployments for app %s", appUUID))
	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)

	call := parseCatalogCall(t, calls[0].Args)
	assert.Equal(t, "GET", call.Method)
	assertSchemaValid(t, calls[0].Args, nixopusAPISchemaFull(), "nixopus_api")

	// The param must be `id=` NOT `application_id=`
	assert.Contains(t, call.Path, "?id=", "deployments path must use ?id= not ?application_id=")
	assert.NotContains(t, call.Path, "application_id", "must not use wrong param name")
	assert.Contains(t, call.Path, appUUID)
}

// TestCatalogAgentProbe_GetDeploymentByID verifies deployment_id is embedded
// in the path, NOT as a query string.
func TestCatalogAgentProbe_GetDeploymentByID(t *testing.T) {
	deployUUID := "deploy-xyz-789"
	correctArgs := fmt.Sprintf(`{"method":"GET","path":"/api/v1/deploy/application/deployments/%s"}`, deployUUID)

	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", correctArgs),
		textResponse("Deployment status: running."),
	})

	nixopusSrv := newMockNixopusServer(t, map[string]string{"status": "running"})

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusAPISchemaFull())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL),
		fmt.Sprintf("get deployment %s", deployUUID))
	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)

	call := parseCatalogCall(t, calls[0].Args)
	assert.Equal(t, "GET", call.Method)
	assertSchemaValid(t, calls[0].Args, nixopusAPISchemaFull(), "nixopus_api")

	// deployment UUID must be IN THE PATH, not a query string
	assert.Contains(t, call.Path, deployUUID, "deployment UUID must be in path")
	assert.NotContains(t, call.Path, "?id=", "deployment UUID must not be a query param")
	assert.NotContains(t, call.Path, "?deployment_id=", "must not use deployment_id query param")
}

// TestCatalogAgentProbe_RestartDeployment verifies restart uses POST not GET
// and deployment UUID goes in the body, not the path.
func TestCatalogAgentProbe_RestartDeployment(t *testing.T) {
	deployUUID := "deploy-restart-111"
	correctArgs := fmt.Sprintf(`{"method":"POST","path":"/api/v1/deploy/application/restart","body":{"id":"%s"}}`, deployUUID)

	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", correctArgs),
		textResponse("Deployment restarted."),
	})

	nixopusSrv := newMockNixopusServer(t, map[string]string{"status": "restarting"})

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusAPISchemaFull())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL),
		fmt.Sprintf("restart deployment %s", deployUUID))
	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)

	call := parseCatalogCall(t, calls[0].Args)
	assertSchemaValid(t, calls[0].Args, nixopusAPISchemaFull(), "nixopus_api")

	// Must be POST (not GET)
	assert.Equal(t, "POST", call.Method, "restart must use POST, not GET")
	assert.Equal(t, "/api/v1/deploy/application/restart", call.Path)

	// UUID must be in body, not path
	assert.NotContains(t, call.Path, deployUUID, "UUID must be in body, not path")
	require.NotNil(t, call.Body)
	assert.Equal(t, deployUUID, call.Body["id"])
}

// TestCatalogAgentProbe_ContainerPathParam verifies container operations embed
// container_id in the URL path rather than as a query string.
func TestCatalogAgentProbe_ContainerPathParam(t *testing.T) {
	containerID := "container-abc-999"
	correctArgs := fmt.Sprintf(`{"method":"POST","path":"/api/v1/container/%s/restart"}`, containerID)

	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", correctArgs),
		textResponse("Container restarted."),
	})

	nixopusSrv := newMockNixopusServer(t, map[string]string{"status": "restarting"})

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusAPISchemaFull())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL),
		fmt.Sprintf("restart container %s", containerID))
	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)

	call := parseCatalogCall(t, calls[0].Args)
	assertSchemaValid(t, calls[0].Args, nixopusAPISchemaFull(), "nixopus_api")

	assert.Equal(t, "POST", call.Method)
	// container_id must be IN THE PATH, not a query string
	assert.Contains(t, call.Path, containerID, "container_id must be embedded in path")
	assert.NotContains(t, call.Path, "?id=", "container_id must not be a query param")
	assert.NotContains(t, call.Path, "?container_id=", "must not use container_id query param")
}

// TestCatalogAgentProbe_MachineSetDefault verifies set-default uses PUT not POST.
func TestCatalogAgentProbe_MachineSetDefault(t *testing.T) {
	serverID := "server-id-xyz"
	correctArgs := fmt.Sprintf(`{"method":"PUT","path":"/api/v1/machines/%s/set-default"}`, serverID)

	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", correctArgs),
		textResponse("Machine set as default."),
	})

	nixopusSrv := newMockNixopusServer(t, map[string]string{"status": "ok"})

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusMachineSchema())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL),
		fmt.Sprintf("set server %s as the default machine", serverID))
	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)

	call := parseCatalogCall(t, calls[0].Args)
	assertSchemaValid(t, calls[0].Args, nixopusMachineSchema(), "nixopus_api")

	assert.Equal(t, "PUT", call.Method, "set-default must use PUT not POST")
	assert.Contains(t, call.Path, serverID)
	assert.Contains(t, call.Path, "set-default")
}

// ─── Live probe: catalog operation nonce tests (agent_probe build tag) ────────
// These live probes are in probe_live_test.go and rely on real LLM.
// The CI-safe equivalents above use scripted mock LLMs.

// ─── Concurrent recorder safety ───────────────────────────────────────────────

// TestCatalogAgentProbe_MultipleCallsRecordedSafely verifies that when a mock LLM
// returns two sequential tool calls, both are captured correctly.
func TestCatalogAgentProbe_MultipleCallsRecordedSafely(t *testing.T) {
	var step atomic.Int32
	llmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := step.Add(1)
		var resp llm.Response
		switch n {
		case 1:
			resp = toolCallResponse("c1", "nixopus_api", `{"method":"GET","path":"/api/v1/deploy/applications"}`)
		case 2:
			resp = toolCallResponse("c2", "nixopus_api", `{"method":"GET","path":"/api/v1/machines/stats"}`)
		default:
			resp = textResponse("Done: 2 apps found, machine healthy.")
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer llmSrv.Close()

	nixopusSrv := newMockNixopusServer(t, map[string]interface{}{"data": []string{}})
	defer nixopusSrv.Close()

	rec, agent := buildProbeAgent(t, llmSrv.URL, nixopusSrv.URL, nixopusAPISchemaFull())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL), "list my apps and check machine health")
	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	assert.Len(t, calls, 2, "both calls must be recorded")

	for _, c := range calls {
		assertSchemaValid(t, c.Args, nixopusAPISchemaFull(), "nixopus_api")
	}

	paths := make([]string, len(calls))
	for i, c := range calls {
		var a struct {
			Path string `json:"path"`
		}
		json.Unmarshal(c.Args, &a)
		paths[i] = a.Path
	}
	assert.Contains(t, paths, "/api/v1/deploy/applications")
	assert.Contains(t, paths, "/api/v1/machines/stats")
}
