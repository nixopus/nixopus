package dockerfile_generator

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type GenerateResponse struct {
	Dockerfile   string
	Port         int
	Workdir      string
	Dockerignore string // optional, may be empty
}

// ProgressFunc is called for each progress event during SSE streaming.
// stageId identifies the pipeline phase (e.g. "resolve-repo", "dockerfile-generate").
// message is a human-readable description of what's happening.
type ProgressFunc func(stageId, message string)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

type pipelineRunRequest struct {
	Source        string `json:"source"`
	Mode          string `json:"mode,omitempty"`
	ModelOverride string `json:"modelOverride,omitempty"`
}

type pipelineRunResponse struct {
	Status       string          `json:"status"`
	Dockerfile   string          `json:"dockerfile"`
	Dockerignore string          `json:"dockerignore"`
	Stages       []pipelineStage `json:"stages"`
}

type pipelineStage struct {
	StageId string          `json:"stageId"`
	Status  string          `json:"status"`
	Output  json.RawMessage `json:"output"`
	Error   string          `json:"error"`
}

type dockerfileStageOutput struct {
	Dockerfile *struct {
		Dockerfile  string `json:"dockerfile"`
		ExposedPort int    `json:"exposed_port"`
	} `json:"dockerfile"`
}

var workdirRe = regexp.MustCompile(`(?m)^WORKDIR\s+(\S+)`)

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

type AuthHeaders struct {
	Token  string
	Cookie string
}

func (c *Client) Generate(ctx context.Context, source, organizationID string, auth AuthHeaders) (*GenerateResponse, error) {
	return c.GenerateWithMode(ctx, source, organizationID, "", auth)
}

// GenerateWithMode generates a Dockerfile. Pass mode "development" for live reload.
// When onProgress is non-nil, uses SSE streaming to receive real-time progress events.
// When onProgress is nil, falls back to a standard JSON request.
func (c *Client) GenerateWithMode(ctx context.Context, source, organizationID, mode string, auth AuthHeaders) (*GenerateResponse, error) {
	return c.GenerateWithProgress(ctx, source, organizationID, mode, auth, nil)
}

// GenerateWithProgress generates a Dockerfile with real-time progress streaming.
// When onProgress is non-nil, requests SSE from the pipeline endpoint and calls
// onProgress for each progress event before returning the final result.
func (c *Client) GenerateWithProgress(ctx context.Context, source, organizationID, mode string, auth AuthHeaders, onProgress ProgressFunc) (*GenerateResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("agent endpoint not configured")
	}

	reqBody := pipelineRunRequest{Source: source, Mode: mode}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/pipeline/run", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if onProgress != nil {
		req.Header.Set("Accept", "text/event-stream")
	}
	if auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	}
	if auth.Cookie != "" {
		req.Header.Set("Cookie", auth.Cookie)
	}
	if organizationID != "" {
		req.Header.Set("X-Organization-Id", organizationID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent pipeline returned status %d", resp.StatusCode)
	}

	// SSE streaming path
	if onProgress != nil && strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return c.handleSSEResponse(resp, onProgress)
	}

	// Standard JSON path (fallback)
	return c.handleJSONResponse(resp)
}

// handleSSEResponse parses SSE events from the pipeline endpoint.
// Events: "progress" (stageId + message), "result" (final response), "error".
func (c *Client) handleSSEResponse(resp *http.Response, onProgress ProgressFunc) (*GenerateResponse, error) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)

	var currentEvent string
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if currentEvent != "" && len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				result, err := c.processSSEEvent(currentEvent, data, onProgress)
				if err != nil {
					return nil, err
				}
				if result != nil {
					return result, nil
				}
			}
			currentEvent = ""
			dataLines = nil
			continue
		}

		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	// Process any remaining event
	if currentEvent != "" && len(dataLines) > 0 {
		data := strings.Join(dataLines, "\n")
		result, err := c.processSSEEvent(currentEvent, data, onProgress)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("SSE stream ended without a result event")
}

// processSSEEvent handles a single SSE event.
// Returns a non-nil GenerateResponse when the "result" event is received.
func (c *Client) processSSEEvent(event, data string, onProgress ProgressFunc) (*GenerateResponse, error) {
	switch event {
	case "progress":
		var p struct {
			StageId string `json:"stageId"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(data), &p); err == nil && onProgress != nil {
			onProgress(p.StageId, p.Message)
		}
		return nil, nil

	case "result":
		var result pipelineRunResponse
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return nil, fmt.Errorf("decode SSE result: %w", err)
		}
		return c.parseResult(&result)

	case "error":
		var errData struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &errData); err == nil {
			return nil, fmt.Errorf("pipeline error: %s", errData.Error)
		}
		return nil, fmt.Errorf("pipeline error: %s", data)
	}
	return nil, nil
}

// handleJSONResponse parses a standard (non-streaming) JSON response.
func (c *Client) handleJSONResponse(resp *http.Response) (*GenerateResponse, error) {
	var result pipelineRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return c.parseResult(&result)
}

// parseResult extracts the GenerateResponse from a pipeline response.
func (c *Client) parseResult(result *pipelineRunResponse) (*GenerateResponse, error) {
	if result.Status == "failed" {
		for _, s := range result.Stages {
			if s.Error != "" {
				return nil, fmt.Errorf("agent pipeline failed: %s", s.Error)
			}
		}
		return nil, fmt.Errorf("agent pipeline failed with status %s", result.Status)
	}

	if result.Dockerfile == "" {
		return nil, fmt.Errorf("agent returned empty dockerfile")
	}

	port := 3000
	for _, s := range result.Stages {
		if s.StageId == "dockerfile-generate" && s.Output != nil {
			var out dockerfileStageOutput
			if err := json.Unmarshal(s.Output, &out); err == nil && out.Dockerfile != nil && out.Dockerfile.ExposedPort > 0 {
				port = out.Dockerfile.ExposedPort
				break
			}
		}
	}

	return &GenerateResponse{
		Dockerfile:   result.Dockerfile,
		Dockerignore: result.Dockerignore,
		Port:         port,
		Workdir:      parseWorkdirFromDockerfile(result.Dockerfile),
	}, nil
}
