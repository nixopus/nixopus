package channel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlackChannel_ReturnsNonNil(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewSlackChannel(db, context.Background())
	require.NotNil(t, ch)
}

func TestSlackChannel_Type(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewSlackChannel(db, context.Background())
	assert.Equal(t, "slack", ch.Type())
}

func TestSlackChannel_Send_MissingOrgID(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewSlackChannel(db, context.Background())

	err := ch.Send(context.Background(), Message{Body: "hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization_id required")
}

func TestSlackChannel_Send_WebhookLookupError(t *testing.T) {
	db := newChannelTestDB(t)
	// No webhook row inserted → getWebhookURL returns an error.
	ch := NewSlackChannel(db, context.Background())

	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": uuid.New().String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active slack webhook")
}

func TestSlackChannel_Send_HTTPError(t *testing.T) {
	// Insert a webhook that points to a closed server.
	db := newChannelTestDB(t)
	orgID := uuid.New()
	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closed.Close()
	insertActiveWebhook(t, db, "slack", closed.URL, orgID)

	ch := NewSlackChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send slack notification")
}

func TestSlackChannel_Send_Non2xx(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()
	insertActiveWebhook(t, db, "slack", srv.URL, orgID)

	ch := NewSlackChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slack webhook returned status 400")
}

func TestSlackChannel_Send_Success_StatusOK(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 512)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	insertActiveWebhook(t, db, "slack", srv.URL, orgID)

	ch := NewSlackChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "deploy complete",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
	assert.Contains(t, receivedBody, "deploy complete")
}

func TestSlackChannel_Send_Success_StatusNoContent(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	insertActiveWebhook(t, db, "slack", srv.URL, orgID)

	ch := NewSlackChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
}
