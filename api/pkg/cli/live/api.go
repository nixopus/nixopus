package live

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/syncproto"
	"github.com/raghavyuva/nixopus-api/pkg/cli/cliconfig"
)

var workdirRe = regexp.MustCompile(`(?m)^WORKDIR\s+(\S+)`)

func truncMsg(s string, max int) string {
	if max <= 0 {
		max = 60
	}
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func parseWorkdirFromDockerfile(dockerfile string) string {
	matches := workdirRe.FindAllStringSubmatch(dockerfile, -1)
	if len(matches) == 0 {
		return "/app"
	}
	last := matches[len(matches)-1]
	if len(last) >= 2 {
		return last[1]
	}
	return "/app"
}

// DeploymentWorkflowResult is the result of a successful deployment workflow run.
type DeploymentWorkflowResult struct {
	Dockerfile   string
	Dockerignore string
	Port         int
	Workdir      string
}

// ProgressFunc is called for each progress event during workflow execution.
type ProgressFunc func(stepID, message string)

// ReasoningChunkFunc is called for each streaming reasoning chunk from the agent.
// When provided, chunks are streamed incrementally; the final deployment-progress still provides the full message.
type ReasoningChunkFunc func(step, chunk string)

// OnBuildLog is called for each build log line from the listen-for-build step in the workflow stream.
type OnBuildLog func(step, log string)

// ApprovalContext holds the Dockerfile proposal and metadata shown to the user before approval.
type ApprovalContext struct {
	Dockerfile       string   `json:"dockerfile"`
	Summary          string   `json:"summary"`
	ValidationScore  int      `json:"validationScore"`
	Suggestions      []string `json:"suggestions"`
	DependenciesJSON string   `json:"-"` // Raw JSON for display if needed
}

// OnRequestApproval is called when the workflow suspends at request-approval.
// ApprovalContext contains the proposed Dockerfile and summary for the user to review.
// Return true to approve, false to reject.
// If nil, the client auto-approves (approved: true) without prompting.
type OnRequestApproval func(ctx context.Context, approval *ApprovalContext) (approved bool, err error)

// DeploymentWorkflowClient calls the Mastra deployment workflow via HTTP.
type DeploymentWorkflowClient struct {
	endpoint    string
	accessToken string
	orgID       string
	httpClient  *http.Client
}

// NewDeploymentWorkflowClient creates a client for the deployment workflow.
func NewDeploymentWorkflowClient(accessToken, orgID string) *DeploymentWorkflowClient {
	timeout, err := cliconfig.GetWorkflowTimeout()
	if err != nil {
		panic("cliconfig: " + err.Error())
	}
	return &DeploymentWorkflowClient{
		endpoint:    cliconfig.GetAgentEndpoint(),
		accessToken: accessToken,
		orgID:       orgID,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// createRun creates a workflow run and returns the runId. Mastra requires runId to stream.
func (c *DeploymentWorkflowClient) createRun(ctx context.Context, applicationID string) (string, error) {
	reqBody := map[string]interface{}{
		"resourceId": applicationID,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal create-run request: %w", err)
	}
	wfID := cliconfig.GetWorkflowID()
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/workflows/"+wfID+"/create-run", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create run request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	if c.orgID != "" {
		req.Header.Set("X-Organization-Id", c.orgID)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create-run request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read create-run response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create-run returned %d: %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		RunID string `json:"runId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse create-run response: %w", err)
	}
	if result.RunID == "" {
		return "", fmt.Errorf("create-run response missing runId: %s", string(respBody))
	}
	return result.RunID, nil
}

// Run executes the deployment workflow with auto-resume on suspend.
// Returns the Dockerfile and related metadata on success.
// When the workflow suspends at request-approval, if onRequestApproval is non-nil
// it is called to obtain user approval before resuming; otherwise auto-approves.
// If onReasoningChunk is non-nil, LLM reasoning is streamed chunk-by-chunk for a rich agent UI.
// If onBuildLog is non-nil, build log lines from listen-for-build are streamed to the caller.
func (c *DeploymentWorkflowClient) Run(ctx context.Context, applicationID, source, mode string, onProgress ProgressFunc, onReasoningChunk ReasoningChunkFunc, onBuildLog OnBuildLog, onRequestApproval OnRequestApproval) (*DeploymentWorkflowResult, error) {
	if mode == "" {
		mode = "development"
	}

	runID, err := c.createRun(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("create workflow run: %w", err)
	}

	inputData := map[string]interface{}{
		"applicationId":  applicationID,
		"source":         source,
		"mode":           mode,
		"organizationId": c.orgID,
	}

	reqBody := map[string]interface{}{
		"inputData":      inputData,
		"closeOnSuspend": true, // Stream closes on suspend so we can prompt for approval and resume
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	streamURL := c.endpoint + "/api/workflows/" + cliconfig.GetWorkflowID() + "/stream?runId=" + url.QueryEscape(runID)
	req, err := http.NewRequestWithContext(ctx, "POST", streamURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	if c.orgID != "" {
		req.Header.Set("X-Organization-Id", c.orgID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workflow request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("workflow returned %d: %s", resp.StatusCode, string(errBody))
	}

	if cliconfig.IsDebugStream() {
		log.Printf("[workflow-stream] DEBUG: connected to stream, Content-Type=%s", resp.Header.Get("Content-Type"))
	}

	// Parse stream (SSE or NDJSON)
	runID, result, err := c.readWorkflowStream(ctx, resp.Body, onProgress, onReasoningChunk, onBuildLog)
	if err != nil {
		return nil, err
	}

	effectiveStatus := func(r *workflowResultPayload) string {
		if r == nil {
			return ""
		}
		if r.Status != "" {
			return r.Status
		}
		return r.WorkflowStatus
	}

	// Resume loop: if suspended, resume (with human approval for request-approval when handler provided)
	for result != nil && effectiveStatus(result) == "suspended" {
		step := result.SuspendedStep
		if step == "" && len(result.Suspended) > 0 {
			step = result.Suspended[0]
		}
		if step == "" {
			return nil, fmt.Errorf("workflow suspended but no step information")
		}

		// listen-for-build (or listenForBuild) suspends waiting for build logs, but the build only starts after we send trigger_build.
		// We already have the Dockerfile from success-output; treat as success and return so we can send trigger_build.
		if isListenForBuildStep(step) && len(result.Result) > 0 && bytes.Contains(result.Result, []byte(`"dockerfile"`)) {
			log.Printf("[live] workflow suspended at %q with Dockerfile, treating as success (sending trigger_build)", step)
			return c.parseResult(result)
		}

		var resumeData map[string]interface{}
		approvalCtx := result.ApprovalContext
		if approvalCtx == nil {
			approvalCtx = &ApprovalContext{}
		}
		if step == "request-approval" && onRequestApproval != nil {
			approved, err := onRequestApproval(ctx, approvalCtx)
			if err != nil {
				return nil, fmt.Errorf("approval: %w", err)
			}
			resumeData = map[string]interface{}{"approved": approved}
		} else {
			resumeData = c.getResumeDataForStep(step)
		}

		result, err = c.resume(ctx, runID, step, resumeData, onProgress, onReasoningChunk, onBuildLog)
		if err != nil {
			return nil, err
		}

		if effectiveStatus(result) == "failed" {
			return nil, fmt.Errorf("workflow failed: %s", result.Error)
		}

		if effectiveStatus(result) == "success" {
			return c.parseResult(result)
		}
	}

	if result != nil && effectiveStatus(result) == "success" {
		return c.parseResult(result)
	}

	// Build descriptive error for debugging
	var errMsg string
	if result != nil {
		errMsg = fmt.Sprintf("workflow did not complete: status=%q error=%q runId=%s", effectiveStatus(result), result.Error, result.RunID)
	} else {
		errMsg = "workflow did not complete: stream ended without workflow-finish (no result)"
	}
	return nil, fmt.Errorf("%s", errMsg)
}

type workflowStreamEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	RunID   string          `json:"runId"`
}

// tryDeploymentReasoningChunk parses deployment-reasoning-chunk format.
// Returns true if handled.
func tryDeploymentReasoningChunk(payload []byte, onReasoningChunk ReasoningChunkFunc) bool {
	if onReasoningChunk == nil || len(payload) == 0 {
		return false
	}
	var p struct {
		Type  string `json:"type"`
		Step  string `json:"step"`
		Chunk string `json:"chunk"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return false
	}
	if p.Type != "deployment-reasoning-chunk" {
		return false
	}
	if cliconfig.IsDebugStream() {
		preview := p.Chunk
		if len(preview) > 30 {
			preview = preview[:30] + "..."
		}
		log.Printf("[workflow-stream] reasoning-chunk step=%q chunk=%q", p.Step, preview)
	}
	onReasoningChunk(p.Step, p.Chunk)
	return true
}

// tryDeploymentProgress parses deployment-progress format (from streamProgress in agent).
// Returns true if handled. Handles both top-level and nested (inside payload.output).
func tryDeploymentProgress(payload []byte, onProgress ProgressFunc) bool {
	if onProgress == nil || len(payload) == 0 {
		return false
	}
	var p struct {
		Type    string `json:"type"`
		Step    string `json:"step"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return false
	}
	if p.Step == "" && p.Message == "" {
		return false
	}
	// Accept if type is deployment-progress, or if we have step+message (shape matches)
	if p.Type != "" && p.Type != "deployment-progress" {
		return false
	}
	if p.Message == "" {
		p.Message = p.Step
	}
	if cliconfig.IsDebugStream() {
		log.Printf("[workflow-stream] tryDeploymentProgress HANDLED step=%q msg=%q", p.Step, truncMsg(p.Message, 50))
	}
	onProgress(p.Step, p.Message)
	return true
}

// tryBuildLog parses build-log format (from listen-for-build step in the Mastra workflow).
// Returns true if handled.
func tryBuildLog(payload []byte, onBuildLog OnBuildLog) bool {
	if onBuildLog == nil || len(payload) == 0 {
		return false
	}
	var p struct {
		Type string `json:"type"`
		Step string `json:"step"`
		Log  string `json:"log"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return false
	}
	if p.Type != "build-log" {
		return false
	}
	onBuildLog(p.Step, p.Log)
	return true
}

type workflowStepPayload struct {
	StepID   string      `json:"stepId"`
	ID       string      `json:"id"`
	StepName string      `json:"stepName"`
	Message  string      `json:"message"`
	Output   interface{} `json:"output"`
}

// workflowStepSuspendedPayload is the payload for workflow-step-suspended events.
// Mastra sends this when a step calls suspend().
type workflowStepSuspendedPayload struct {
	StepName       string          `json:"stepName"`
	ID             string          `json:"id"`
	StepCallID     string          `json:"stepCallId"`
	Status         string          `json:"status"`
	SuspendPayload json.RawMessage `json:"suspendPayload"`
}

// parseApprovalContext extracts ApprovalContext from request-approval suspendPayload.
func parseApprovalContext(raw json.RawMessage) *ApprovalContext {
	if len(raw) == 0 {
		return nil
	}
	var p struct {
		Dockerfile      string   `json:"dockerfile"`
		Summary         string   `json:"summary"`
		ValidationScore int      `json:"validationScore"`
		Suggestions     []string `json:"suggestions"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil
	}
	return &ApprovalContext{
		Dockerfile:      p.Dockerfile,
		Summary:         p.Summary,
		ValidationScore: p.ValidationScore,
		Suggestions:     p.Suggestions,
	}
}

// workflowFinishPayload is the payload for workflow-finish events.
type workflowFinishPayload struct {
	WorkflowStatus string          `json:"workflowStatus"`
	Metadata       json.RawMessage `json:"metadata"`
	Output         json.RawMessage `json:"output"`
}

// workflowResultPayload is the internal result we pass to the resume loop.
// Built from workflow-step-suspended or workflow-finish.
type workflowResultPayload struct {
	Status          string
	WorkflowStatus  string
	SuspendedStep   string
	Suspended       []string
	Error           string
	Result          json.RawMessage
	RunID           string
	ApprovalContext *ApprovalContext // Set when suspended at request-approval
}

// firstByteReader logs when the first bytes are read from the stream (helps debug buffering).
type firstByteReader struct {
	r      io.Reader
	logged bool
}

func (r *firstByteReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 && !r.logged {
		r.logged = true
		if cliconfig.IsDebugStream() {
			log.Printf("[workflow-stream] first bytes received (%d bytes)", n)
		}
	}
	return n, err
}

// splitRecordSeparator splits on ASCII Record Separator (0x1e).
// Mastra HTTP API uses 0x1e as the event delimiter, not newlines.
func splitRecordSeparator(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\x1e'); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (c *DeploymentWorkflowClient) readWorkflowStream(ctx context.Context, body io.Reader, onProgress ProgressFunc, onReasoningChunk ReasoningChunkFunc, onBuildLog OnBuildLog) (string, *workflowResultPayload, error) {
	body = &firstByteReader{r: body}
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)
	scanner.Split(splitRecordSeparator)

	var runID string
	var lastResult *workflowResultPayload
	var collectedResult json.RawMessage
	var eventCount int
	var lastEventType string

	if cliconfig.IsDebugStream() {
		log.Printf("[workflow-stream] started reading stream")
	}

	// Log periodically while waiting for first line (helps debug: no data vs buffering)
	firstLineCh := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-firstLineCh:
				return
			case <-ticker.C:
				if cliconfig.IsDebugStream() {
					log.Printf("[workflow-stream] still waiting for stream data (no complete line received yet)...")
				}
			}
		}
	}()

	for scanner.Scan() {
		select {
		case firstLineCh <- struct{}{}:
		default:
		}

		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		default:
		}

		rawLine := strings.TrimSpace(scanner.Text())
		if rawLine == "" || rawLine == "[DONE]" {
			continue
		}
		if strings.HasPrefix(rawLine, "data:") {
			rawLine = strings.TrimSpace(strings.TrimPrefix(rawLine, "data:"))
		}
		if rawLine == "" {
			continue
		}
		if cliconfig.IsDebugStream() {
			preview := rawLine
			if len(preview) > 300 {
				preview = preview[:300] + "..."
			}
			log.Printf("[workflow-stream] RAW: %s", preview)
		}

		var event workflowStreamEvent
		if err := json.Unmarshal([]byte(rawLine), &event); err != nil {
			if cliconfig.IsDebugStream() {
				preview := rawLine
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				log.Printf("[workflow-stream] JSON parse error: %v | line=%s", err, preview)
			}
			continue
		}

		eventCount++
		lastEventType = event.Type

		if cliconfig.IsDebugStream() {
			payloadPreview := string(event.Payload)
			if len(payloadPreview) > 500 {
				payloadPreview = payloadPreview[:500] + "..."
			}
			log.Printf("[workflow-stream] #%d type=%q runId=%q payload=%s", eventCount, event.Type, event.RunID, payloadPreview)
		}

		if event.RunID != "" {
			runID = event.RunID
		}

		payloadBytes := event.Payload
		// workflow-step-output nests deployment-progress in payload.output; unwrap only for that event
		deployProgressBytes := payloadBytes
		if event.Type == "workflow-step-output" {
			var wrapped struct {
				Output json.RawMessage `json:"output"`
			}
			if err := json.Unmarshal(payloadBytes, &wrapped); err == nil && len(wrapped.Output) > 0 {
				deployProgressBytes = wrapped.Output
			}
		}

		// deployment-reasoning-chunk, deployment-progress, and build-log can be top-level or nested inside workflow-step-output
		if tryDeploymentReasoningChunk(deployProgressBytes, onReasoningChunk) {
			continue
		}
		if tryDeploymentProgress(deployProgressBytes, onProgress) {
			continue
		}
		if tryBuildLog(deployProgressBytes, onBuildLog) {
			continue
		}
		if tryBuildLog(payloadBytes, onBuildLog) {
			continue
		}
		if event.Type == "deployment-progress" || event.Type == "deployment-reasoning-chunk" || event.Type == "build-log" {
			continue
		}

		switch event.Type {
		case "workflow-step-suspended":
			var p workflowStepSuspendedPayload
			if err := json.Unmarshal(payloadBytes, &p); err == nil && p.Status == "suspended" {
				stepID := p.StepName
				if stepID == "" {
					stepID = p.ID
				}
				if stepID != "" {
					lastResult = &workflowResultPayload{
						Status:          "suspended",
						SuspendedStep:   stepID,
						Suspended:       []string{stepID},
						RunID:           runID,
						ApprovalContext: parseApprovalContext(p.SuspendPayload),
					}
					if len(collectedResult) > 0 {
						lastResult.Result = collectedResult
					}
					return runID, lastResult, nil
				}
			}
		case "workflow-step-start", "workflow-step-output", "workflow-step-progress":
			var payload workflowStepPayload
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				stepID := payload.StepID
				if stepID == "" {
					stepID = payload.ID
				}
				if stepID == "" {
					stepID = payload.StepName
				}
				// listen-for-build keeps stream open waiting for build logs, but build only starts after we send trigger_build.
				// We already have the Dockerfile from success-output; return now so we can send trigger_build.
				if isListenForBuildStep(stepID) && len(collectedResult) > 0 && bytes.Contains(collectedResult, []byte(`"dockerfile"`)) {
					log.Printf("[live] reached %q with Dockerfile in collectedResult, treating as success (sending trigger_build)", stepID)
					return runID, &workflowResultPayload{
						Status:         "success",
						WorkflowStatus: "success",
						RunID:          runID,
						Result:         collectedResult,
					}, nil
				}
				if onProgress != nil {
					msg := payload.Message
					if msg == "" {
						msg = payload.StepName
					}
					if msg == "" {
						msg = payload.StepID
					}
					if msg == "" {
						msg = payload.ID
					}
					if msg != "" {
						onProgress(stepID, msg)
					}
				}
			}
		case "workflow-step-result":
			// Collect output from steps (e.g. propose-dockerfile, success-output).
			// Prefer outputs containing dockerfile (success-output) over mapping step outputs.
			var payload struct {
				StepID string          `json:"stepId"`
				Output json.RawMessage `json:"output"`
			}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil && len(payload.Output) > 0 {
				if bytes.Contains(payload.Output, []byte(`"dockerfile"`)) || len(collectedResult) == 0 {
					collectedResult = payload.Output
				}
			}
		case "workflow-finish", "finish":
			var p workflowFinishPayload
			if err := json.Unmarshal(payloadBytes, &p); err == nil {
				status := p.WorkflowStatus
				// Prefer collectedResult (from success-output etc.) over workflow-finish output.
				// workflow-finish output is usually just usage metadata; the Dockerfile comes from workflow-step-result.
				resultData := collectedResult
				if len(resultData) == 0 {
					resultData = p.Output
				} else if !bytes.Contains(resultData, []byte(`"dockerfile"`)) && len(p.Output) > 0 {
					// collectedResult has no dockerfile; fall back to workflow output
					resultData = p.Output
				}
				lastResult = &workflowResultPayload{
					Status:         status,
					WorkflowStatus: status,
					RunID:          runID,
					Result:         resultData,
				}
				if status == "success" {
					return runID, lastResult, nil
				}
				if status == "suspended" {
					// workflow-step-suspended normally fires first with step info; if we reach here without it, we don't have it
					return runID, lastResult, nil
				}
			}
		case "workflow-suspend":
			// Legacy/different format; parse like workflowResultPayload for compatibility
			var p struct {
				Status        string   `json:"status"`
				SuspendedStep string   `json:"suspendedStep"`
				Suspended     []string `json:"suspended"`
				RunID         string   `json:"runId"`
			}
			if err := json.Unmarshal(payloadBytes, &p); err == nil && p.Status == "suspended" {
				stepID := p.SuspendedStep
				if stepID == "" && len(p.Suspended) > 0 {
					stepID = p.Suspended[0]
				}
				lastResult = &workflowResultPayload{
					Status:        "suspended",
					SuspendedStep: stepID,
					Suspended:     p.Suspended,
					RunID:         runID,
				}
				if len(collectedResult) > 0 {
					lastResult.Result = collectedResult
				}
				return runID, lastResult, nil
			}
		default:
			// workflow-start, etc. - no-op
		}
	}

	if err := scanner.Err(); err != nil {
		if cliconfig.IsDebugStream() {
			log.Printf("[workflow-stream] scanner error: %v", err)
		}
		return "", nil, err
	}

	if cliconfig.IsDebugStream() {
		lastStatus, lastStep := "", ""
		if lastResult != nil {
			lastStatus = lastResult.Status
			if lastStatus == "" {
				lastStatus = lastResult.WorkflowStatus
			}
			lastStep = lastResult.SuspendedStep
		}
		log.Printf("[workflow-stream] stream ended eventCount=%d lastEvent=%q runID=%q status=%q suspendedStep=%q",
			eventCount, lastEventType, runID, lastStatus, lastStep)
	}

	if lastResult != nil && len(collectedResult) > 0 && len(lastResult.Result) == 0 {
		lastResult.Result = collectedResult
	}

	// Stream ended without workflow-finish or workflow-step-suspended (e.g. agent closes after success-output).
	// If we have a Dockerfile from workflow-step-result, treat as success so we can send trigger_build.
	if lastResult == nil && len(collectedResult) > 0 && bytes.Contains(collectedResult, []byte(`"dockerfile"`)) {
		log.Printf("[live] stream ended with Dockerfile in collectedResult (no workflow-finish), treating as success (sending trigger_build)")
		lastResult = &workflowResultPayload{
			Status:         "success",
			WorkflowStatus: "success",
			RunID:          runID,
			Result:         collectedResult,
		}
	}

	return runID, lastResult, nil
}

// isListenForBuildStep returns true if the step is the listen-for-build step (name may vary by agent).
func isListenForBuildStep(step string) bool {
	s := strings.ToLower(strings.ReplaceAll(step, "_", "-"))
	return strings.Contains(s, "listen") && strings.Contains(s, "build")
}

func (c *DeploymentWorkflowClient) getResumeDataForStep(step string) map[string]interface{} {
	switch step {
	case "detect-project-structure":
		return map[string]interface{}{"selectedPath": "."}
	case "detect-env-issues", "detect-dependency-audit", "detect-security-scan":
		return map[string]interface{}{"acknowledged": true, "proceedAnyway": true}
	case "request-approval":
		return map[string]interface{}{"approved": true}
	default:
		return map[string]interface{}{"acknowledged": true, "approved": true}
	}
}

func (c *DeploymentWorkflowClient) resume(ctx context.Context, runID, step string, resumeData map[string]interface{}, onProgress ProgressFunc, onReasoningChunk ReasoningChunkFunc, onBuildLog OnBuildLog) (*workflowResultPayload, error) {
	// Mastra resume-stream expects step as array (path to suspended step)
	reqBody := map[string]interface{}{
		"step":       []string{step},
		"resumeData": resumeData,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal resume request: %w", err)
	}

	resumeStreamURL := c.endpoint + "/api/workflows/" + cliconfig.GetWorkflowID() + "/resume-stream?runId=" + url.QueryEscape(runID)
	req, err := http.NewRequestWithContext(ctx, "POST", resumeStreamURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create resume request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	if c.orgID != "" {
		req.Header.Set("X-Organization-Id", c.orgID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("resume request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("resume-stream returned %d: %s", resp.StatusCode, string(respBody))
	}

	// resume-stream returns a stream; parse it like the initial stream
	_, result, err := c.readWorkflowStream(ctx, resp.Body, onProgress, onReasoningChunk, onBuildLog)
	if err != nil {
		return nil, fmt.Errorf("resume stream: %w", err)
	}
	return result, nil
}

func (c *DeploymentWorkflowClient) parseResult(r *workflowResultPayload) (*DeploymentWorkflowResult, error) {
	if len(r.Result) == 0 {
		return nil, fmt.Errorf("workflow result has no output")
	}

	// Try top-level structure first
	var output struct {
		Dockerfile string `json:"dockerfile"`
		Status     string `json:"status"`
		BuildLog   string `json:"buildLog"`
		Validation *struct {
			DockerignoreSuggestion string `json:"dockerignoreSuggestion"`
			ExposedPort            int    `json:"exposedPort"`
		} `json:"validation"`
	}

	if err := json.Unmarshal(r.Result, &output); err != nil {
		return nil, fmt.Errorf("parse workflow output: %w", err)
	}

	// Result might be nested (e.g. { result: { dockerfile: "..." } })
	if output.Dockerfile == "" {
		var nested struct {
			Result *struct {
				Dockerfile string `json:"dockerfile"`
				Validation *struct {
					DockerignoreSuggestion string `json:"dockerignoreSuggestion"`
					ExposedPort            int    `json:"exposedPort"`
				} `json:"validation"`
			} `json:"result"`
		}
		if err := json.Unmarshal(r.Result, &nested); err == nil && nested.Result != nil && nested.Result.Dockerfile != "" {
			output.Dockerfile = nested.Result.Dockerfile
			output.Validation = nested.Result.Validation
		}
	}

	if output.Dockerfile == "" {
		return nil, fmt.Errorf("workflow did not produce a Dockerfile")
	}

	port := 3000
	if output.Validation != nil && output.Validation.ExposedPort > 0 {
		port = output.Validation.ExposedPort
	}

	dockerignore := ""
	if output.Validation != nil && output.Validation.DockerignoreSuggestion != "" {
		dockerignore = output.Validation.DockerignoreSuggestion
	}

	workdir := parseWorkdirFromDockerfile(output.Dockerfile)

	return &DeploymentWorkflowResult{
		Dockerfile:   output.Dockerfile,
		Dockerignore: dockerignore,
		Port:         port,
		Workdir:      workdir,
	}, nil
}

// ToTriggerBuildPayload converts the result to the WebSocket payload format.
func (r *DeploymentWorkflowResult) ToTriggerBuildPayload() syncproto.TriggerBuildPayload {
	return syncproto.TriggerBuildPayload{
		Dockerfile:   r.Dockerfile,
		Dockerignore: r.Dockerignore,
		Port:         r.Port,
		Workdir:      r.Workdir,
	}
}
