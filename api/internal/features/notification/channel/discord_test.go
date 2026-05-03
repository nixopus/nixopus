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

func TestNewDiscordChannel_ReturnsNonNil(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewDiscordChannel(db, context.Background())
	require.NotNil(t, ch)
}

func TestDiscordChannel_Type(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewDiscordChannel(db, context.Background())
	assert.Equal(t, "discord", ch.Type())
}

func TestDiscordChannel_Send_MissingOrgID(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewDiscordChannel(db, context.Background())

	err := ch.Send(context.Background(), Message{Body: "hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization_id required")
}

func TestDiscordChannel_Send_WebhookLookupError(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewDiscordChannel(db, context.Background())

	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": uuid.New().String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active discord webhook")
}

func TestDiscordChannel_Send_HTTPError(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closed.Close()
	insertActiveWebhook(t, db, "discord", closed.URL, orgID)

	ch := NewDiscordChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send discord notification")
}

func TestDiscordChannel_Send_Non2xx(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer srv.Close()
	insertActiveWebhook(t, db, "discord", srv.URL, orgID)

	ch := NewDiscordChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "discord webhook returned status 500")
}

func TestDiscordChannel_Send_Success_StatusOK(t *testing.T) {
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
	insertActiveWebhook(t, db, "discord", srv.URL, orgID)

	ch := NewDiscordChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "server alert",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
	assert.Contains(t, receivedBody, "server alert")
}

func TestDiscordChannel_Send_Success_StatusNoContent(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	insertActiveWebhook(t, db, "discord", srv.URL, orgID)

	ch := NewDiscordChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		Body:     "hi",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
}
