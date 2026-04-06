package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── JSON-RPC helpers (local mirror of unexported types) ─────────────────────

type jrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
}

var toolsPayload = json.RawMessage(`{"tools":[{"name":"search_repos","description":"Search GitHub repos"},{"name":"create_issue","description":"Open an issue"}]}`)
var initPayload = json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{}}`)
var callResultPayload = json.RawMessage(`{"content":[{"type":"text","text":"query result here"}]}`)

// ─── HTTP transport mock server ───────────────────────────────────────────────

// newHTTPMockServer creates a Streamable-HTTP MCP server that responds to
// initialize and tools/list over plain JSON-RPC POST.
func newHTTPMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jrpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		var result json.RawMessage
		switch req.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "test-session-123")
			result = initPayload
		case "tools/list":
			result = toolsPayload
		case "tools/call":
			result = callResultPayload
		default:
			http.Error(w, "unknown method", http.StatusNotFound)
			return
		}

		resp := jrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
}

// ─── SSE transport mock server ────────────────────────────────────────────────

// newSSEMockServer creates a legacy-SSE MCP server.
// GET /sse streams events; POST /messages relays JSON-RPC responses back over
// that same stream.
func newSSEMockServer(t *testing.T) *httptest.Server {
	t.Helper()

	outgoing := make(chan string, 4)
	done := make(chan struct{})

	mux := http.NewServeMux()

	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok, "ResponseWriter must implement http.Flusher")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		fmt.Fprintf(w, "event: endpoint\ndata: /messages\n\n") //nolint:errcheck
		flusher.Flush()

		for {
			select {
			case msg, ok := <-outgoing:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg) //nolint:errcheck
				flusher.Flush()
			case <-done:
				return
			case <-r.Context().Done():
				return
			}
		}
	})

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		var req jrpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		var result json.RawMessage
		switch req.Method {
		case "initialize":
			result = initPayload
		case "tools/list":
			result = toolsPayload
		default:
			http.Error(w, "unknown method", http.StatusNotFound)
			return
		}

		resp := jrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
		data, _ := json.Marshal(resp)
		outgoing <- string(data)
		w.WriteHeader(http.StatusAccepted)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(func() {
		close(done)
		srv.Close()
	})
	return srv
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestDiscoverServerTools_HTTP(t *testing.T) {
	srv := newHTTPMockServer(t)
	defer srv.Close()

	provider := &mcp.MCPProvider{
		ID:        "github",
		Name:      "GitHub",
		URL:       srv.URL + "/",
		Transport: "http",
		Fields: []mcp.ProviderField{
			{Key: "token", HeaderName: "Authorization", HeaderPrefix: "Bearer", Required: true},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	tools, err := service.DiscoverServerTools(ctx, provider, "", map[string]string{
		"token": "ghp_test_token",
	})

	require.NoError(t, err)
	require.Len(t, tools, 2)
	assert.Equal(t, "search_repos", tools[0].Name)
	assert.Equal(t, "create_issue", tools[1].Name)
}

func TestDiscoverServerTools_HTTP_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	provider := &mcp.MCPProvider{
		ID:        "github",
		Name:      "GitHub",
		URL:       srv.URL + "/",
		Transport: "http",
		Fields:    []mcp.ProviderField{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	_, err := service.DiscoverServerTools(ctx, provider, "", map[string]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initialize")
}

func TestDiscoverServerTools_SSE(t *testing.T) {
	srv := newSSEMockServer(t)

	provider := &mcp.MCPProvider{
		ID:        "supabase",
		Name:      "Supabase",
		URL:       srv.URL + "/sse",
		Transport: "sse",
		Fields: []mcp.ProviderField{
			{Key: "access_token", HeaderName: "Authorization", HeaderPrefix: "Bearer", Required: true},
			{Key: "project_ref", IsQueryParam: true},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	tools, err := service.DiscoverServerTools(ctx, provider, "", map[string]string{
		"access_token": "sbp_test_token",
		"project_ref":  "abcdef",
	})

	require.NoError(t, err)
	require.Len(t, tools, 2)
	assert.Equal(t, "search_repos", tools[0].Name)
}

func TestDiscoverServerTools_CustomHTTP(t *testing.T) {
	srv := newHTTPMockServer(t)
	defer srv.Close()

	provider := &mcp.MCPProvider{
		ID:        "custom",
		Name:      "Custom",
		URL:       "",
		Transport: "http",
		Fields:    []mcp.ProviderField{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	tools, err := service.DiscoverServerTools(ctx, provider, srv.URL+"/", map[string]string{})

	require.NoError(t, err)
	require.Len(t, tools, 2)
}

func TestCallToolOnServer_HTTP(t *testing.T) {
	srv := newHTTPMockServer(t)
	defer srv.Close()

	provider := &mcp.MCPProvider{
		ID:        "supabase",
		Name:      "Supabase",
		URL:       srv.URL + "/",
		Transport: "http",
		Fields: []mcp.ProviderField{
			{Key: "access_token", HeaderName: "Authorization", HeaderPrefix: "Bearer", Required: true},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	result, err := service.CallToolOnServer(ctx, provider, "", map[string]string{
		"access_token": "sbp_test_token",
	}, service.CallToolParams{
		Name:      "list_tables",
		Arguments: map[string]any{"project_ref": "abc"},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "query result here", result.Content[0].Text)
}

func TestCallToolOnServer_HTTP_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	provider := &mcp.MCPProvider{
		ID:        "supabase",
		Name:      "Supabase",
		URL:       srv.URL + "/",
		Transport: "http",
		Fields:    []mcp.ProviderField{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	_, err := service.CallToolOnServer(ctx, provider, "", map[string]string{}, service.CallToolParams{
		Name: "list_tables",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initialize")
}

func TestCallToolOnServer_SessionHeader(t *testing.T) {
	var receivedSessionID string
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req jrpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		var result json.RawMessage
		switch req.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "sess-abc-123")
			result = initPayload
		case "tools/call":
			receivedSessionID = r.Header.Get("Mcp-Session-Id")
			result = callResultPayload
		default:
			http.Error(w, "unknown method", http.StatusNotFound)
			return
		}

		resp := jrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	provider := &mcp.MCPProvider{
		ID:        "supabase",
		Name:      "Supabase",
		URL:       srv.URL + "/",
		Transport: "http",
		Fields:    []mcp.ProviderField{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tenSeconds)
	defer cancel()

	result, err := service.CallToolOnServer(ctx, provider, "", map[string]string{}, service.CallToolParams{
		Name: "execute_sql",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, callCount)
	assert.Equal(t, "sess-abc-123", receivedSessionID,
		"tools/call must include the Mcp-Session-Id from initialize")
}
