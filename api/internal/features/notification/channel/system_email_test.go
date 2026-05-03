package channel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constructor + Type
// ---------------------------------------------------------------------------

func TestNewSystemEmailChannel_ReturnsNonNil(t *testing.T) {
	ch := NewSystemEmailChannel("key", "from@example.com")
	require.NotNil(t, ch)
	assert.NotNil(t, ch.client)
}

func TestSystemEmailChannel_Type(t *testing.T) {
	ch := NewSystemEmailChannel("key", "from@example.com")
	assert.Equal(t, "system_email", ch.Type())
}

// ---------------------------------------------------------------------------
// Send — early validation errors
// ---------------------------------------------------------------------------

func TestSystemEmailChannel_Send_NoAPIKey(t *testing.T) {
	ch := NewSystemEmailChannel("", "from@example.com")

	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Metadata: map[string]string{"resend_template_id": "tmpl_123"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RESEND_API_KEY not configured")
}

func TestSystemEmailChannel_Send_NoTemplateID_Key(t *testing.T) {
	ch := NewSystemEmailChannel("api-key", "from@example.com")

	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Metadata: map[string]string{}, // resend_template_id missing
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend_template_id not set")
}

func TestSystemEmailChannel_Send_EmptyTemplateID_Value(t *testing.T) {
	ch := NewSystemEmailChannel("api-key", "from@example.com")

	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Metadata: map[string]string{"resend_template_id": ""},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend_template_id not set")
}

func TestSystemEmailChannel_Send_EmptyRecipient(t *testing.T) {
	ch := NewSystemEmailChannel("api-key", "from@example.com")

	err := ch.Send(context.Background(), Message{
		To:       "",
		Metadata: map[string]string{"resend_template_id": "tmpl_123"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recipient address is required")
}

// ---------------------------------------------------------------------------
// Send — HTTP errors
// ---------------------------------------------------------------------------

func TestSystemEmailChannel_Send_HTTPError(t *testing.T) {
	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closed.Close()

	ch := NewSystemEmailChannel("api-key", "from@example.com")
	// Point the client to a closed server by replacing the default client.
	ch.client = closed.Client()

	// Use a custom transport that always fails.
	ch.client.Transport = &failTransport{}

	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Metadata: map[string]string{"resend_template_id": "tmpl_123"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend API request failed")
}

func TestSystemEmailChannel_Send_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	ch := NewSystemEmailChannel("bad-key", "from@example.com")
	ch.client = srv.Client()
	// Redirect the Resend API calls to our test server.
	ch.client.Transport = &rewriteTransport{target: srv.URL, inner: srv.Client().Transport}

	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Metadata: map[string]string{"resend_template_id": "tmpl_123"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend API error (HTTP 401)")
}

// ---------------------------------------------------------------------------
// Send — success paths
// ---------------------------------------------------------------------------

func TestSystemEmailChannel_Send_Success_NoTemplateData(t *testing.T) {
	var captured map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		assert.Equal(t, "Bearer api-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ch := NewSystemEmailChannel("api-key", "from@example.com")
	ch.client = srv.Client()
	ch.client.Transport = &rewriteTransport{target: srv.URL, inner: srv.Client().Transport}

	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Metadata: map[string]string{"resend_template_id": "tmpl_123"},
		// no TemplateData → variables NOT included in payload
	})
	require.NoError(t, err)
	tmpl := captured["template"].(map[string]interface{})
	assert.Equal(t, "tmpl_123", tmpl["id"])
	_, hasVars := tmpl["variables"]
	assert.False(t, hasVars)
}

func TestSystemEmailChannel_Send_Success_WithTemplateData(t *testing.T) {
	var captured map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ch := NewSystemEmailChannel("api-key", "from@example.com")
	ch.client = srv.Client()
	ch.client.Transport = &rewriteTransport{target: srv.URL, inner: srv.Client().Transport}

	err := ch.Send(context.Background(), Message{
		To:           "user@example.com",
		Metadata:     map[string]string{"resend_template_id": "tmpl_456"},
		TemplateData: map[string]interface{}{"Name": "Alice", "Action": "login"},
	})
	require.NoError(t, err)
	tmpl := captured["template"].(map[string]interface{})
	assert.Equal(t, "tmpl_456", tmpl["id"])
	assert.NotNil(t, tmpl["variables"])
}

// ---------------------------------------------------------------------------
// Test transport helpers
// ---------------------------------------------------------------------------

// failTransport always returns an error, simulating a network failure.
type failTransport struct{}

func (f *failTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, assert.AnError
}

// rewriteTransport rewrites requests destined for "https://api.resend.com" to
// the given target (our httptest server URL) so we can intercept them.
type rewriteTransport struct {
	target string
	inner  http.RoundTripper
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Host = req.URL.Host // preserve path/query
	// Replace the scheme+host with our test server.
	import_req, _ := http.NewRequestWithContext(req.Context(), req.Method, rt.target+req.URL.Path, req.Body)
	if import_req != nil {
		import_req.Header = req.Header
		return rt.inner.RoundTrip(import_req)
	}
	return rt.inner.RoundTrip(req)
}
