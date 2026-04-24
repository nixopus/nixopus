package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nixopus/nixopus/api/internal/features/dashboard"
	deployrt "github.com/nixopus/nixopus/api/internal/features/deploy/realtime"
	"github.com/nixopus/nixopus/api/internal/features/terminal"
)

func newTestSocketServer() *SocketServer {
	return &SocketServer{
		conns:               &sync.Map{},
		orgIDs:              &sync.Map{},
		shutdown:            make(chan struct{}),
		ctx:                 context.Background(),
		topics:              make(map[string]map[*websocket.Conn]bool),
		terminals:           make(map[*websocket.Conn]map[string]*terminal.Terminal),
		dashboardMonitors:   make(map[*websocket.Conn]*dashboard.DashboardMonitor),
		applicationMonitors: make(map[*websocket.Conn]*deployrt.ApplicationMonitor),
	}
}

// dialWebSocketPair opens one WebSocket from client to a test server; returns
// the client end, the upgraded server *websocket.Conn, and a cleanup func.
func dialWebSocketPair(t *testing.T) (client, server *websocket.Conn, cleanup func()) {
	t.Helper()
	srvCh := make(chan *websocket.Conn, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade: %v", err)
			return
		}
		srvCh <- c
	}))

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	srv := <-srvCh
	return client, srv, func() {
		_ = client.Close()
		_ = srv.Close()
		ts.Close()
	}
}

func newUser(t *testing.T) uuid.UUID {
	t.Helper()
	return uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
}
