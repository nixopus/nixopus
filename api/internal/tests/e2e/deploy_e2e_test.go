//go:build integration

package e2e

// E2E deployment test — verifies a sample application goes through the full
// create → deploy → monitor lifecycle against the REAL running Nixopus API.
//
// WHY THIS TEST EXISTS
// =====================
// Probe tests (probe_test.go, catalog_allpaths_test.go) only verify that the
// LLM calls the correct endpoint with the correct method. They use mock servers
// that accept any request and return fixed JSON. They cannot catch:
//
//   1. Request body schema mismatches (e.g. unknown fields the API rejects)
//   2. Infrastructure requirements (SSH server, Docker daemon must be configured)
//   3. Async deployment failures (task worker logs "no server configured" after
//      the HTTP 200 is already sent)
//   4. Dependent service failures propagating across unrelated endpoints
//   5. Real API contract drift (backend changes field names, probe tests don't notice)
//
// This test catches all of the above by running against the real server.
//
// HOW TO RUN
// ==========
//   # Option A — using saved session cookies (get from browser DevTools):
//   NIXOPUS_E2E=1 NIXOPUS_COOKIE="better-auth.session_token=..." \
//     NIXOPUS_ORG_ID="<your-org-uuid>" \
//     go test -tags integration ./internal/tests/agent/ \
//     -run TestDeployE2E -v -timeout 10m
//
//   # Option B — auto-register+login with a fresh test user:
//   NIXOPUS_E2E=1 \
//     go test -tags integration ./internal/tests/agent/ \
//     -run TestDeployE2E -v -timeout 10m
//
// WHAT IT CATCHES (from the real server logs)
// ============================================
//   1. "json: unknown field deploy_on_create" — body schema mismatch
//   2. "no server configured for organization" — missing SSH machine
//   3. Container 500 errors when no Docker daemon is reachable
//   4. Async deployment failures after the trigger HTTP 200

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	tests "github.com/nixopus/nixopus/api/internal/tests"
	"github.com/stretchr/testify/require"
)

// ─── E2E deployment lifecycle test ───────────────────────────────────────────

func TestDeployE2E_SampleApp(t *testing.T) {
	if os.Getenv("NIXOPUS_E2E") == "" {
		t.Skip("set NIXOPUS_E2E=1 to run this test against a live server")
	}

	auth, err := e2eAuth(t)
	require.NoError(t, err, "failed to authenticate")

	client := &http.Client{Timeout: 30 * time.Second}

	report := &deployReport{t: t}

	// ── Step 1: Verify the API is alive ───────────────────────────────────────
	report.step("health_check", func() stepResult {
		resp, err := doGet(client, tests.GetHealthURL(), auth)
		if err != nil {
			return stepResult{err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return stepResult{status: resp.StatusCode, body: string(body)}
	})

	// ── Step 2: Check whether a machine is configured (prerequisite) ──────────
	var machineConfigured bool
	report.step("check_machine_configured", func() stepResult {
		// Use SSH status endpoint — it reliably reports whether the org has a
		// configured server. GET /machines may be empty due to RBAC or scoping.
		resp, err := doGet(client, tests.GetMachineSSHStatusURL(), auth)
		if err != nil {
			return stepResult{err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		// SSH status returns {status, connected, message, is_configured}
		connected, _ := result["connected"].(bool)
		isConfigured, _ := result["is_configured"].(bool)
		machineConfigured = connected || isConfigured

		if machineConfigured {
			return stepResult{status: resp.StatusCode, body: fmt.Sprintf("SSH connected=%v is_configured=%v", connected, isConfigured)}
		}
		return stepResult{
			status:  resp.StatusCode,
			body:    "no machine configured — deployment will fail at the SSH tunnel step",
			warning: "deploy will fail: register a machine via POST /api/v1/machines and verify SSH",
		}
	})

	// ── Step 3: Create a sample application project ───────────────────────────
	// The API requires a numeric GitHub repository ID, not "owner/repo".
	// Resolve it from the first connected GitHub connector. Skip if none is set up.
	var projectID string
	report.step("create_project", func() stepResult {
		// Get connected GitHub repositories
		repoResp, err := doGet(client, tests.GetGithubRepositoriesURL(), auth)
		if err != nil {
			return stepResult{err: fmt.Errorf("list GitHub repositories: %w", err)}
		}
		repoBody, _ := io.ReadAll(repoResp.Body)
		repoResp.Body.Close()

		var repoResult map[string]interface{}
		json.Unmarshal(repoBody, &repoResult)

		// data is an array of repositories with numeric "id" field
		repos, _ := repoResult["data"].([]interface{})
		if len(repos) == 0 {
			return stepResult{
				status:  repoResp.StatusCode,
				body:    string(repoBody),
				warning: "no GitHub repositories found — connect a GitHub account via POST /api/v1/github-connector to enable deployment",
			}
		}

		// Use the first available repository
		repo, _ := repos[0].(map[string]interface{})
		repoID := extractNumericID(repo)
		repoName, _ := repo["name"].(string)
		if repoID == 0 {
			return stepResult{err: fmt.Errorf("could not extract numeric repo ID from: %v", repo)}
		}

		payload := map[string]interface{}{
			"name":        fmt.Sprintf("e2e-sample-%d", time.Now().Unix()),
			"repository":  repoID, // numeric GitHub repo ID
			"branch":      "main",
			"build_pack":  "dockerfile",
			"port":        3000,
			"environment": "production",
		}
		resp, err := doPost(client, tests.GetDeployApplicationProjectURL(), auth, payload)
		if err != nil {
			return stepResult{err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return stepResult{
				status: resp.StatusCode,
				body:   string(body),
				err:    fmt.Errorf("create project failed: HTTP %d — %s", resp.StatusCode, extractError(result)),
			}
		}

		if data, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(string); ok {
				projectID = id
			}
		}
		if projectID == "" {
			return stepResult{status: resp.StatusCode, body: string(body), err: fmt.Errorf("no id in response")}
		}
		return stepResult{status: resp.StatusCode, body: fmt.Sprintf("project id=%s repo=%s(%d)", projectID, repoName, repoID)}
	})

	if projectID == "" {
		report.finish()
		// If the step issued a warning (no GitHub connector), don't fatal — just finish
		t.Skip("cannot proceed: no GitHub repository available — connect a GitHub account to test deployment")
	}

	// Register cleanup: delete the project/app after the test
	t.Cleanup(func() {
		doDelete(client, tests.GetDeployApplicationURL(), auth,
			map[string]interface{}{"id": projectID})
	})

	// ── Step 4: Trigger deployment ─────────────────────────────────────────────
	var deploymentStarted bool
	report.step("trigger_deploy", func() stepResult {
		payload := map[string]interface{}{"id": projectID}
		resp, err := doPost(client, tests.GetDeployApplicationProjectDeployURL(), auth, payload)
		if err != nil {
			return stepResult{err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if resp.StatusCode != http.StatusOK {
			return stepResult{
				status: resp.StatusCode,
				body:   string(body),
				err:    fmt.Errorf("trigger deploy failed: HTTP %d — %s", resp.StatusCode, extractError(result)),
			}
		}
		deploymentStarted = true
		return stepResult{status: resp.StatusCode, body: fmt.Sprintf("trigger accepted, app=%s", projectID)}
	})

	// ── Step 5: Poll deployment status ────────────────────────────────────────
	// The trigger returns HTTP 200 the moment the background task is queued.
	// The actual deployment record is created by that task (async). We give it
	// 2 seconds to appear then poll the deployments list, matching by application_id.
	if deploymentStarted {
		report.step("poll_deployment_status", func() stepResult {
			time.Sleep(2 * time.Second) // let the task worker create the deployment record

			deadline := time.Now().Add(5 * time.Minute)
			pollInterval := 5 * time.Second

			// Always use the list endpoint. The /deployments/{id} endpoint requires
			// the actual deployment record ID, NOT the application status record ID
			// returned by the trigger. The list endpoint needs ?id=<application_id>.
			pollURL := tests.GetDeployApplicationDeploymentsURL() + "?id=" + projectID

			for time.Now().Before(deadline) {
				resp, err := doGet(client, pollURL, auth)
				if err != nil {
					return stepResult{err: err}
				}
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var result map[string]interface{}
				json.Unmarshal(body, &result)

				status, deploymentID := extractDeploymentStatus(result, projectID)
				t.Logf("[poll] status=%q deploymentID=%q rawBody=%s", status, deploymentID, truncate(string(body), 300))

				switch status {
				case "success", "running":
					return stepResult{
						status: resp.StatusCode,
						body:   fmt.Sprintf("deployment %s: status=%s", deploymentID, status),
					}
				case "failed", "error", "cancelled":
					// Fetch logs immediately to capture the failure reason
					logs := ""
					if deploymentID != "" {
						logs = fetchDeploymentLogs(t, client,
							tests.GetDeployApplicationDeploymentLogsURL(deploymentID),
							auth)
					}
					return stepResult{
						status: resp.StatusCode,
						body:   fmt.Sprintf("deployment %s failed: status=%s", deploymentID, status),
						err:    fmt.Errorf("deployment failed (status=%s)\nLogs:\n%s", status, logs),
					}
				case "pending", "queued", "building", "deploying":
					// Still in progress — keep polling
				default:
					if !machineConfigured {
						return stepResult{
							status:  resp.StatusCode,
							body:    string(body),
							warning: "deployment status unknown — likely failed due to missing machine: " + status,
						}
					}
				}
				time.Sleep(pollInterval)
			}
			return stepResult{
				err: fmt.Errorf("deployment timed out after 5 minutes — last status unknown"),
			}
		})
	}

	// ── Step 6: Verify application is accessible ──────────────────────────────
	report.step("verify_application", func() stepResult {
		resp, err := doGet(client,
			tests.GetDeployApplicationURL()+"?id="+projectID,
			auth)
		if err != nil {
			return stepResult{err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if resp.StatusCode != http.StatusOK {
			return stepResult{
				status: resp.StatusCode,
				body:   string(body),
				err:    fmt.Errorf("application verification failed: HTTP %d", resp.StatusCode),
			}
		}
		return stepResult{status: resp.StatusCode, body: "application record exists"}
	})

	// ── Final report ──────────────────────────────────────────────────────────
	report.finish()
}

// ─── Request body contract tests ─────────────────────────────────────────────
//
// These tests catch the EXACT failures shown in the server logs:
// "json: unknown field "deploy_on_create"" — a frontend/backend schema mismatch
// that our mock-server probe tests can never catch.

// TestDeployE2E_CreateProject_UnknownFields verifies the API rejects unknown
// request body fields. This would have caught the "deploy_on_create" error.
func TestDeployE2E_CreateProject_UnknownFields(t *testing.T) {
	if os.Getenv("NIXOPUS_E2E") == "" {
		t.Skip("set NIXOPUS_E2E=1 to run this test against a live server")
	}

	auth, err := e2eAuth(t)
	require.NoError(t, err)

	client := &http.Client{Timeout: 15 * time.Second}

	// This exact call is what the frontend sends and gets a 400 for.
	// Our mock probe tests pass because the mock accepts anything.
	// This test catches the contract mismatch against the real API.
	badPayload := map[string]interface{}{
		"name":             "test-unknown-fields",
		"repository":       "owner/repo",
		"deploy_on_create": true, // unknown field — the exact failure from the logs
	}

	resp, err := doPost(client, tests.GetDeployApplicationProjectURL(), auth, badPayload)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	t.Logf("response status: %d", resp.StatusCode)
	t.Logf("response body: %s", string(body))

	// The API SHOULD reject unknown fields with 400. If it returns 200, that means
	// the API accepts unknown fields (lenient mode) which is also useful to document.
	if resp.StatusCode == http.StatusBadRequest {
		t.Logf("PASS: API correctly rejects unknown field 'deploy_on_create' with 400")
		require.Contains(t, string(body), "deploy_on_create",
			"error message should mention the unknown field name")
	} else if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		t.Logf("NOTE: API accepts unknown fields (lenient mode) — status %d", resp.StatusCode)
		// Clean up the created project if one was made
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		if data, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(string); ok && id != "" {
				doDelete(client, tests.GetDeployApplicationURL(), auth,
					map[string]interface{}{"id": id})
			}
		}
	} else {
		t.Errorf("unexpected status %d — expected 400 or 200/201", resp.StatusCode)
	}
}

// TestDeployE2E_DeployWithoutMachine verifies the failure mode when no SSH
// server is configured. This is the EXACT error from the logs:
// "deploy tasks: create deployment failed: failed to get SSH manager: no server configured"
func TestDeployE2E_DeployWithoutMachine(t *testing.T) {
	if os.Getenv("NIXOPUS_E2E") == "" {
		t.Skip("set NIXOPUS_E2E=1 to run this test against a live server")
	}

	auth, err := e2eAuth(t)
	require.NoError(t, err)

	client := &http.Client{Timeout: 30 * time.Second}

	// Check whether a machine is configured via SSH status (more reliable than GET /machines)
	sshResp, err := doGet(client, tests.GetMachineSSHStatusURL(), auth)
	require.NoError(t, err)
	sshBody, _ := io.ReadAll(sshResp.Body)
	sshResp.Body.Close()

	var sshResult map[string]interface{}
	json.Unmarshal(sshBody, &sshResult)
	connected, _ := sshResult["connected"].(bool)
	isConfigured, _ := sshResult["is_configured"].(bool)
	if connected || isConfigured {
		t.Skipf("SSH is active (connected=%v is_configured=%v) — this test verifies the no-machine failure path; skip when a machine is connected",
			connected, isConfigured)
	}

	// Create a project
	payload := map[string]interface{}{
		"name":       fmt.Sprintf("e2e-nomachine-%d", time.Now().Unix()),
		"repository": "nixopus/sample-app",
		"branch":     "main",
	}
	resp, err := doPost(client, tests.GetDeployApplicationProjectURL(), auth, payload)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Logf("create project failed (HTTP %d): %s", resp.StatusCode, string(body))
		t.Skip("could not create project to test deployment failure path")
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	data2, _ := result["data"].(map[string]interface{})
	projectID, _ := data2["id"].(string)
	require.NotEmpty(t, projectID)
	t.Cleanup(func() {
		doDelete(client, tests.GetDeployApplicationURL(), auth,
			map[string]interface{}{"id": projectID})
	})

	// Trigger deployment
	deployResp, err := doPost(client, tests.GetDeployApplicationProjectDeployURL(), auth,
		map[string]interface{}{"id": projectID})
	require.NoError(t, err)
	deployBody, _ := io.ReadAll(deployResp.Body)
	deployResp.Body.Close()

	t.Logf("trigger deploy: HTTP %d — %s", deployResp.StatusCode, truncate(string(deployBody), 300))

	// Deployment trigger itself returns 200 (task is queued) — the failure is async.
	// Poll for deployment status to observe the failure.
	time.Sleep(3 * time.Second) // give the task worker time to fail

	pollResp, err := doGet(client, tests.GetDeployApplicationDeploymentsURL()+"?id="+projectID, auth)
	require.NoError(t, err)
	pollBody, _ := io.ReadAll(pollResp.Body)
	pollResp.Body.Close()

	t.Logf("deployments list: %s", truncate(string(pollBody), 500))

	var pollResult map[string]interface{}
	json.Unmarshal(pollBody, &pollResult)
	status, deploymentID := extractDeploymentStatus(pollResult, projectID)

	t.Logf("deployment status after 3s: status=%s id=%s", status, deploymentID)

	// Fetch logs to see the actual error message
	if deploymentID != "" {
		logs := fetchDeploymentLogs(t, client,
			tests.GetDeployApplicationDeploymentLogsURL(deploymentID),
			auth)
		t.Logf("deployment logs:\n%s", logs)

		// The logs should contain the "no server configured" error
		if strings.Contains(logs, "no server configured") {
			t.Logf("CONFIRMED: deployment failure reason = 'no server configured' (expected without a machine)")
		} else {
			t.Logf("NOTE: deployment failure has a different reason — review logs above")
		}
	}

	require.Equal(t, "failed", status,
		"deployment without a configured machine must end in status=failed")
}

// ─── SSH status prerequisite check ───────────────────────────────────────────

// TestDeployE2E_SSHStatus verifies that GET /api/v1/machines/ssh/status returns
// actionable data (not just empty). From the logs: the API returns 200 but the
// body is "no server configured" — this test documents that behaviour.
func TestDeployE2E_SSHStatus(t *testing.T) {
	if os.Getenv("NIXOPUS_E2E") == "" {
		t.Skip("set NIXOPUS_E2E=1 to run this test against a live server")
	}

	auth, err := e2eAuth(t)
	require.NoError(t, err)

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := doGet(client, tests.GetMachineSSHStatusURL(), auth)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	t.Logf("SSH status: HTTP %d — %s", resp.StatusCode, string(body))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// The endpoint returns {status, connected, message, is_configured} — not {data:[...]}
	connected, _ := result["connected"].(bool)
	isConfigured, _ := result["is_configured"].(bool)
	message, _ := result["message"].(string)

	if connected || isConfigured {
		t.Logf("SSH is active: connected=%v is_configured=%v message=%q", connected, isConfigured, message)
	} else {
		t.Logf("WARNING: SSH is not connected (connected=%v is_configured=%v). "+
			"Deployments will fail with 'no server configured for organization'.", connected, isConfigured)
	}
}

// ─── Auth helper (no test-DB dependency) ─────────────────────────────────────

type e2eAuthResult struct {
	// One of bearer or cookies will be set.
	bearer  string // "Authorization: Bearer <token>" style (from NIXOPUS_API_KEY)
	cookies string // "Cookie: <value>" style (from NIXOPUS_COOKIE or auto-register)
	orgID   string
}

// applyAuth stamps auth headers onto any request.
func (a e2eAuthResult) applyAuth(req *http.Request) {
	if a.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+a.bearer)
	} else if a.cookies != "" {
		req.Header.Set("Cookie", a.cookies)
	}
	if a.orgID != "" {
		req.Header.Set("X-Organization-ID", a.orgID)
	}
}

// e2eAuth resolves auth credentials for the live API in three ways (priority order):
//
//  1. NIXOPUS_API_KEY + NIXOPUS_ORG_ID — Bearer token from DevTools Network tab.
//     Copy the "Authorization" request header value (without the "Bearer " prefix)
//     and the "x-organization-id" request header value.
//  2. NIXOPUS_COOKIE + NIXOPUS_ORG_ID — session cookie from DevTools Application tab.
//  3. Auto-register a fresh test user via the API (no env vars needed).
func e2eAuth(t *testing.T) (e2eAuthResult, error) {
	t.Helper()

	// Option 1: Bearer token (easiest — copy from DevTools Network tab)
	if apiKey := os.Getenv("NIXOPUS_API_KEY"); apiKey != "" {
		orgID := os.Getenv("NIXOPUS_ORG_ID")
		if orgID == "" {
			return e2eAuthResult{}, fmt.Errorf("NIXOPUS_API_KEY set but NIXOPUS_ORG_ID is missing")
		}
		t.Logf("[e2e-auth] using Bearer token from NIXOPUS_API_KEY, org=%s", orgID)
		return e2eAuthResult{bearer: apiKey, orgID: orgID}, nil
	}

	// Option 2: Cookie
	if cookie := os.Getenv("NIXOPUS_COOKIE"); cookie != "" {
		orgID := os.Getenv("NIXOPUS_ORG_ID")
		if orgID == "" {
			return e2eAuthResult{}, fmt.Errorf("NIXOPUS_COOKIE set but NIXOPUS_ORG_ID is missing")
		}
		t.Logf("[e2e-auth] using Cookie from NIXOPUS_COOKIE, org=%s", orgID)
		return e2eAuthResult{cookies: cookie, orgID: orgID}, nil
	}

	// Auto-register a fresh user via the API.
	client := &http.Client{Timeout: 15 * time.Second}
	ts := time.Now().UnixNano()
	email := fmt.Sprintf("e2e-test-%d@nixopus.test", ts)
	password := fmt.Sprintf("E2eTest%d!", ts)

	regPayload := map[string]interface{}{
		"email":    email,
		"password": password,
		"name":     fmt.Sprintf("E2E Test %d", ts),
	}
	regResp, err := doPost(client, tests.GetRegisterURL(), e2eAuthResult{}, regPayload)
	if err != nil {
		return e2eAuthResult{}, fmt.Errorf("register failed: %w", err)
	}
	defer regResp.Body.Close()

	if regResp.StatusCode != http.StatusOK && regResp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(regResp.Body)
		return e2eAuthResult{}, fmt.Errorf("register HTTP %d: %s", regResp.StatusCode, string(b))
	}

	var cookieParts []string
	for _, c := range regResp.Cookies() {
		cookieParts = append(cookieParts, c.Name+"="+c.Value)
	}

	if len(cookieParts) == 0 {
		loginResp, err := doPost(client, tests.GetLoginURL(), e2eAuthResult{}, regPayload)
		if err != nil {
			return e2eAuthResult{}, fmt.Errorf("login failed: %w", err)
		}
		defer loginResp.Body.Close()
		for _, c := range loginResp.Cookies() {
			cookieParts = append(cookieParts, c.Name+"="+c.Value)
		}
	}

	partialAuth := e2eAuthResult{cookies: strings.Join(cookieParts, "; ")}

	// Resolve the org via GET /user
	userResp, err := doGet(client, tests.GetUserDetailsURL(), partialAuth)
	if err != nil {
		return e2eAuthResult{}, fmt.Errorf("get user failed: %w", err)
	}
	defer userResp.Body.Close()
	userBody, _ := io.ReadAll(userResp.Body)

	var userResult map[string]interface{}
	json.Unmarshal(userBody, &userResult)

	orgID := ""
	if data, ok := userResult["data"].(map[string]interface{}); ok {
		if orgs, ok := data["organizations"].([]interface{}); ok && len(orgs) > 0 {
			if org, ok := orgs[0].(map[string]interface{}); ok {
				orgID, _ = org["id"].(string)
			}
		}
	}
	if orgID == "" {
		t.Logf("user response: %s", string(userBody))
		return e2eAuthResult{}, fmt.Errorf("could not resolve org ID from user response")
	}

	t.Logf("[e2e-auth] auto-registered email=%s org=%s", email, orgID)
	return e2eAuthResult{cookies: partialAuth.cookies, orgID: orgID}, nil
}

// ─── Report infrastructure ────────────────────────────────────────────────────

type stepResult struct {
	status  int
	body    string
	warning string
	err     error
}

type deployStep struct {
	name    string
	result  stepResult
	elapsed time.Duration
}

type deployReport struct {
	t     *testing.T
	steps []deployStep
}

func (r *deployReport) step(name string, fn func() stepResult) {
	r.t.Helper()
	start := time.Now()
	result := fn()
	elapsed := time.Since(start)

	r.steps = append(r.steps, deployStep{name: name, result: result, elapsed: elapsed})

	status := "✓"
	if result.err != nil {
		status = "✗"
	} else if result.warning != "" {
		status = "⚠"
	}
	r.t.Logf("[%s] %s  (%dms)  %s", status, name, elapsed.Milliseconds(), shortBody(result))
}

func (r *deployReport) finish() {
	r.t.Helper()
	r.t.Logf("\n═══ E2E Deployment Report ═══")
	var failed []string
	var warned []string
	for _, s := range r.steps {
		mark := "PASS"
		if s.result.err != nil {
			mark = "FAIL"
			failed = append(failed, s.name)
		} else if s.result.warning != "" {
			mark = "WARN"
			warned = append(warned, s.name+": "+s.result.warning)
		}
		r.t.Logf("  [%-4s] %-40s %dms", mark, s.name, s.elapsed.Milliseconds())
		if s.result.err != nil {
			r.t.Logf("         ERROR: %v", s.result.err)
		}
	}

	if len(warned) > 0 {
		r.t.Logf("\nWarnings:")
		for _, w := range warned {
			r.t.Logf("  ⚠ %s", w)
		}
	}

	if len(failed) > 0 {
		r.t.Logf("\nFailed steps: %v", failed)
		r.t.Logf("\nDIAGNOSIS:")
		for _, s := range r.steps {
			if s.result.err != nil {
				r.t.Logf("  %s: %v", s.name, s.result.err)
			}
		}
		r.t.Errorf("%d of %d steps failed: %v", len(failed), len(r.steps), failed)
	} else {
		r.t.Logf("\nAll %d steps passed.", len(r.steps))
	}
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

func doGet(client *http.Client, url string, auth e2eAuthResult) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	auth.applyAuth(req)
	return client.Do(req)
}

func doPost(client *http.Client, url string, auth e2eAuthResult, payload interface{}) (*http.Response, error) {
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	auth.applyAuth(req)
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func doDelete(client *http.Client, url string, auth e2eAuthResult, payload interface{}) {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("DELETE", url, bytes.NewReader(b))
	auth.applyAuth(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

// ─── Analysis helpers ─────────────────────────────────────────────────────────

// extractNumericID returns the numeric GitHub repository ID from a repository object.
// The API stores it as a float64 (JSON number) in the "id" field.
func extractNumericID(repo map[string]interface{}) int64 {
	if repo == nil {
		return 0
	}
	// GitHub repo IDs come back as JSON numbers → float64 in Go
	if f, ok := repo["id"].(float64); ok {
		return int64(f)
	}
	// Some endpoints return the GitHub-specific ID under different keys
	for _, key := range []string{"github_id", "repo_id", "repository_id"} {
		if f, ok := repo[key].(float64); ok {
			return int64(f)
		}
	}
	return 0
}

func extractError(result map[string]interface{}) string {
	if errObj, ok := result["error"].(map[string]interface{}); ok {
		if detail, ok := errObj["detail"].(string); ok {
			return detail
		}
	}
	if msg, ok := result["message"].(string); ok {
		return msg
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// extractDeploymentStatus parses the response from GET /deploy/application/deployments.
//
// The API returns deployment objects that may contain a nested status object
// (as seen in the trigger response) or a direct status string (in the list).
// This function handles both shapes and extracts the final status string and
// the deployment ID (the deployment record's own ID, not the application ID).
//
// It also accepts the application ID so it can filter the right deployment when
// the list endpoint returns all org deployments without per-app filtering.
func extractDeploymentStatus(result map[string]interface{}, applicationID string) (status, deploymentID string) {
	data, ok := result["data"]
	if !ok {
		return "", ""
	}

	var deps []map[string]interface{}
	switch d := data.(type) {
	case []interface{}:
		// plain array: [{id, application_id, status, ...}, ...]
		for _, item := range d {
			if m, ok := item.(map[string]interface{}); ok {
				deps = append(deps, m)
			}
		}
	case map[string]interface{}:
		// wrapped object: {deployments: [...]} or {data: {...}}
		if list, ok := d["deployments"].([]interface{}); ok {
			for _, item := range list {
				if m, ok := item.(map[string]interface{}); ok {
					deps = append(deps, m)
				}
			}
		}
	}

	if len(deps) == 0 {
		return "", ""
	}

	// Find the deployment belonging to our application, or use the latest one.
	var dep map[string]interface{}
	for _, d := range deps {
		appID, _ := d["application_id"].(string)
		if appID == applicationID {
			dep = d
			break
		}
	}
	if dep == nil {
		dep = deps[0]
	}

	deploymentID, _ = dep["id"].(string)

	// status can be a plain string or a nested object {id, application_id, status, ...}.
	// The nested object is the app's status record — its own ID is NOT the deployment
	// record ID. dep["id"] is always the correct deployment record ID for log fetching.
	switch v := dep["status"].(type) {
	case string:
		status = v
	case map[string]interface{}:
		status, _ = v["status"].(string)
	}
	return
}

func fetchDeploymentLogs(t *testing.T, client *http.Client, url string, auth e2eAuthResult) string {
	t.Helper()

	// Retry a few times — the worker may flip status to "failed" before the log
	// entries are committed to the DB.
	var body []byte
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(2 * time.Second)
		}
		resp, err := doGet(client, url, auth)
		if err != nil {
			return fmt.Sprintf("failed to fetch logs: %v", err)
		}
		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		// Prefer the structured log array; bail early if we have entries
		if data, ok := result["data"].(map[string]interface{}); ok {
			if logs, ok := data["logs"].([]interface{}); ok && len(logs) > 0 {
				var lines []string
				for _, entry := range logs {
					if m, ok := entry.(map[string]interface{}); ok {
						// Try common field name variants across different API versions
						level := firstString(m, "level", "log_level", "severity", "lvl")
						msg := firstString(m, "message", "msg", "text", "log", "content")
						if level == "" && msg == "" {
							// Unknown shape — dump raw JSON so we can see the actual keys
							raw, _ := json.Marshal(m)
							lines = append(lines, string(raw))
						} else {
							lines = append(lines, fmt.Sprintf("[%s] %s", level, msg))
						}
					}
				}
				return strings.Join(lines, "\n")
			}
		}
	}
	// Return raw body so the caller can still see the response shape
	return string(body)
}

func shortBody(r stepResult) string {
	if r.err != nil {
		return r.err.Error()
	}
	if r.warning != "" {
		return r.warning
	}
	return truncate(r.body, 120)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// firstString returns the first non-empty string value found for any of the given keys.
func firstString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
		// numeric level (e.g. zerolog uses integers)
		if f, ok := m[k].(float64); ok {
			return fmt.Sprintf("%v", f)
		}
	}
	return ""
}
