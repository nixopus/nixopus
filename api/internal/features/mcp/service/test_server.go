package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
)

type TestResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func (s *MCPService) TestServer(req *validation.TestServerRequest) *TestResult {
	s.logger.Log(logger.Info, "Testing MCP server connection", req.ProviderID)

	provider := mcp.GetProvider(req.ProviderID)
	if provider == nil {
		return &TestResult{OK: false, Error: "unknown provider"}
	}

	rawURL := provider.URL
	if req.ProviderID == "custom" {
		rawURL = req.CustomURL
	}

	if len(provider.Fields) > 0 {
		u, err := url.Parse(rawURL)
		if err != nil {
			return &TestResult{OK: false, Error: "invalid URL"}
		}
		q := u.Query()
		for _, field := range provider.Fields {
			if field.IsQueryParam {
				if v, ok := req.Credentials[field.Key]; ok && v != "" {
					q.Set(field.Key, v)
				}
			}
		}
		u.RawQuery = q.Encode()
		rawURL = u.String()
	}

	headers := map[string]string{}
	for _, field := range provider.Fields {
		if field.HeaderName == "" {
			continue
		}
		v, ok := req.Credentials[field.Key]
		if !ok || v == "" {
			continue
		}
		headerValue := v
		if field.HeaderPrefix != "" {
			headerValue = field.HeaderPrefix + " " + v
		}
		headers[field.HeaderName] = headerValue
	}

	client := &http.Client{Timeout: 10 * time.Second}

	var resp *http.Response
	var err error

	if provider.Transport == "sse" {
		resp, err = doTestGET(client, rawURL, headers)
	} else {
		resp, err = doTestRPC(client, rawURL, headers)
	}

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return &TestResult{OK: false, Error: "connection refused"}
		}
		return &TestResult{OK: false, Error: "connection timed out"}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return &TestResult{OK: false, Error: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(body))}
	}

	return &TestResult{OK: true}
}

func doTestGET(client *http.Client, rawURL string, headers map[string]string) (*http.Response, error) {
	httpReq, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	return client.Do(httpReq)
}

func doTestRPC(client *http.Client, rawURL string, headers map[string]string) (*http.Response, error) {
	rpc := mcpRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: mcpInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]any{},
			ClientInfo:      map[string]any{"name": "nixopus", "version": "1.0"},
		},
	}
	body, err := json.Marshal(rpc)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	return client.Do(httpReq)
}
