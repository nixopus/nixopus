package dockerfile_generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

type GenerateResponse struct {
	Dockerfile   string
	Port         int
	Workdir      string
	Dockerignore string // optional, may be empty
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type pipelineRunRequest struct {
	Source        string `json:"source"`
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
	if c.baseURL == "" {
		return nil, fmt.Errorf("agent endpoint not configured")
	}

	reqBody := pipelineRunRequest{Source: source}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/pipeline/run", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
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

	var result pipelineRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

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

	out := &GenerateResponse{
		Dockerfile:   result.Dockerfile,
		Dockerignore: result.Dockerignore,
		Port:         port,
		Workdir:      parseWorkdirFromDockerfile(result.Dockerfile),
	}

	return out, nil
}
