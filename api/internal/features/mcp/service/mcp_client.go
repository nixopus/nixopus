package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
)

// MCPTool mirrors the MCP tools/list schema.
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
}

// ServerToolSet is the per-server result returned by DiscoverAllTools.
type ServerToolSet struct {
	ServerID   string    `json:"server_id"`
	ServerName string    `json:"server_name"`
	ProviderID string    `json:"provider_id"`
	Tools      []MCPTool `json:"tools"`
	Error      string    `json:"error,omitempty"`
}

// ─── JSON-RPC types ───────────────────────────────────────────────────────────

type mcpRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type mcpInitParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      map[string]any `json:"clientInfo"`
}

// ─── Header / URL helpers ─────────────────────────────────────────────────────

func buildServerHeaders(provider *mcp.MCPProvider, creds map[string]string) map[string]string {
	h := map[string]string{"Content-Type": "application/json"}
	for _, f := range provider.Fields {
		if f.HeaderName == "" {
			continue
		}
		v := creds[f.Key]
		if v == "" {
			continue
		}
		if f.HeaderPrefix != "" {
			v = f.HeaderPrefix + " " + v
		}
		h[f.HeaderName] = v
	}
	return h
}

func buildServerURL(provider *mcp.MCPProvider, customURL string, creds map[string]string) (string, error) {
	raw := provider.URL
	if provider.ID == "custom" {
		raw = customURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for _, f := range provider.Fields {
		if f.IsQueryParam {
			if v := creds[f.Key]; v != "" {
				q.Set(f.Key, v)
			}
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ─── HTTP (Streamable HTTP) transport ────────────────────────────────────────

func postRPC(ctx context.Context, client *http.Client, serverURL string, headers map[string]string, req mcpRequest) (*mcpResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "text/event-stream") {
		return firstSSEMessage(resp.Body)
	}

	var rpc mcpResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		return nil, err
	}
	return &rpc, nil
}

func discoverHTTP(ctx context.Context, serverURL string, headers map[string]string) ([]MCPTool, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	initResp, err := postRPC(ctx, client, serverURL, headers, mcpRequest{
		JSONRPC: "2.0", ID: 1, Method: "initialize",
		Params: mcpInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]any{},
			ClientInfo:      map[string]any{"name": "nixopus", "version": "1.0"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}
	if initResp.Error != nil {
		return nil, fmt.Errorf("initialize: %s", initResp.Error.Message)
	}

	listResp, err := postRPC(ctx, client, serverURL, headers, mcpRequest{
		JSONRPC: "2.0", ID: 2, Method: "tools/list",
		Params: map[string]any{},
	})
	if err != nil {
		return nil, fmt.Errorf("tools/list: %w", err)
	}
	if listResp.Error != nil {
		return nil, fmt.Errorf("tools/list: %s", listResp.Error.Message)
	}

	return parseToolsResult(listResp.Result)
}

// ─── SSE (legacy) transport ───────────────────────────────────────────────────

type sseEvent struct{ name, data string }

func streamSSEEvents(ctx context.Context, r io.Reader) <-chan sseEvent {
	ch := make(chan sseEvent, 20)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(r)
		var ev sseEvent
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line := scanner.Text()
			if line == "" {
				if ev.name != "" || ev.data != "" {
					select {
					case ch <- ev:
					case <-ctx.Done():
						return
					}
					ev = sseEvent{}
				}
				continue
			}
			if after, ok := strings.CutPrefix(line, "event:"); ok {
				ev.name = strings.TrimSpace(after)
			} else if after, ok := strings.CutPrefix(line, "data:"); ok {
				ev.data = strings.TrimSpace(after)
			}
		}
	}()
	return ch
}

func firstSSEMessage(r io.Reader) (*mcpResponse, error) {
	scanner := bufio.NewScanner(r)
	var data string
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "data:"); ok {
			data = strings.TrimSpace(after)
		}
		if line == "" && data != "" {
			var rpc mcpResponse
			if err := json.Unmarshal([]byte(data), &rpc); err != nil {
				return nil, err
			}
			return &rpc, nil
		}
	}
	return nil, fmt.Errorf("no SSE message data received")
}

func discoverSSE(ctx context.Context, sseURL string, headers map[string]string) ([]MCPTool, error) {
	noTimeoutClient := &http.Client{}
	sseReq, err := http.NewRequestWithContext(ctx, http.MethodGet, sseURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		sseReq.Header.Set(k, v)
	}
	sseReq.Header.Set("Accept", "text/event-stream")

	sseResp, err := noTimeoutClient.Do(sseReq)
	if err != nil {
		return nil, err
	}
	defer sseResp.Body.Close()

	if sseResp.StatusCode >= 400 {
		return nil, fmt.Errorf("SSE endpoint returned %d", sseResp.StatusCode)
	}

	events := streamSSEEvents(ctx, sseResp.Body)

	// Wait for the endpoint event
	msgEndpoint := ""
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
waitEndpoint:
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return nil, fmt.Errorf("SSE stream closed before endpoint event")
			}
			if ev.name == "endpoint" {
				msgEndpoint = ev.data
				break waitEndpoint
			}
		case <-deadline.C:
			return nil, fmt.Errorf("timeout waiting for endpoint event")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Resolve relative endpoint URL
	if !strings.HasPrefix(msgEndpoint, "http") {
		base, _ := url.Parse(sseURL)
		rel, _ := url.Parse(msgEndpoint)
		msgEndpoint = base.ResolveReference(rel).String()
	}

	postHeaders := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		postHeaders[k] = v
	}
	postHeaders["Content-Type"] = "application/json"

	postAndWait := func(req mcpRequest) (*mcpResponse, error) {
		body, _ := json.Marshal(req)
		postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, msgEndpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		for k, v := range postHeaders {
			postReq.Header.Set(k, v)
		}
		pc := &http.Client{Timeout: 10 * time.Second}
		postResp, err := pc.Do(postReq)
		if err != nil {
			return nil, err
		}
		defer postResp.Body.Close()
		if postResp.StatusCode >= 400 {
			return nil, fmt.Errorf("message endpoint returned %d", postResp.StatusCode)
		}

		// Read the matching response from the shared SSE stream
		rTimeout := time.NewTimer(10 * time.Second)
		defer rTimeout.Stop()
		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return nil, fmt.Errorf("SSE stream closed waiting for response")
				}
				if ev.name == "message" && ev.data != "" {
					var rpc mcpResponse
					if err := json.Unmarshal([]byte(ev.data), &rpc); err != nil {
						continue
					}
					if rpc.ID == req.ID {
						return &rpc, nil
					}
				}
			case <-rTimeout.C:
				return nil, fmt.Errorf("timeout waiting for response to %s", req.Method)
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	initResp, err := postAndWait(mcpRequest{
		JSONRPC: "2.0", ID: 1, Method: "initialize",
		Params: mcpInitParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]any{},
			ClientInfo:      map[string]any{"name": "nixopus", "version": "1.0"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}
	if initResp.Error != nil {
		return nil, fmt.Errorf("initialize: %s", initResp.Error.Message)
	}

	listResp, err := postAndWait(mcpRequest{
		JSONRPC: "2.0", ID: 2, Method: "tools/list",
		Params: map[string]any{},
	})
	if err != nil {
		return nil, fmt.Errorf("tools/list: %w", err)
	}
	if listResp.Error != nil {
		return nil, fmt.Errorf("tools/list: %s", listResp.Error.Message)
	}

	return parseToolsResult(listResp.Result)
}

// ─── Shared ───────────────────────────────────────────────────────────────────

func parseToolsResult(raw json.RawMessage) ([]MCPTool, error) {
	var result struct {
		Tools []MCPTool `json:"tools"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

// DiscoverServerTools calls tools/list on a single MCP server using the
// appropriate transport (HTTP or SSE) and returns the tool list.
func DiscoverServerTools(ctx context.Context, provider *mcp.MCPProvider, customURL string, creds map[string]string) ([]MCPTool, error) {
	serverURL, err := buildServerURL(provider, customURL, creds)
	if err != nil {
		return nil, fmt.Errorf("bad server URL: %w", err)
	}
	headers := buildServerHeaders(provider, creds)

	if provider.Transport == "sse" {
		return discoverSSE(ctx, serverURL, headers)
	}
	return discoverHTTP(ctx, serverURL, headers)
}
