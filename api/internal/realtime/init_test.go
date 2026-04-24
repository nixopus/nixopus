package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/nixopus/nixopus/api/internal/types"
)

func TestNewSocketServer(t *testing.T) {
	old := startListeningForSocketServer
	startListeningForSocketServer = func(*PostgresListener, context.Context, *SocketServer) error { return nil }
	t.Cleanup(func() { startListeningForSocketServer = old })
	s, err := NewSocketServer(nil, nil, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("nil server")
	}
	s.Shutdown()
}

func TestNewSocketServer_listenError(t *testing.T) {
	oldL := startListeningForSocketServer
	oldP := pgxConnect
	startListeningForSocketServer = startListeningForSocketServerImpl
	pgxConnect = func(context.Context, string) (*pgx.Conn, error) { return nil, errors.New("unreachable db") }
	t.Cleanup(func() {
		startListeningForSocketServer = oldL
		pgxConnect = oldP
	})
	_, err := NewSocketServer(nil, nil, context.Background())
	if err == nil {
		t.Fatal("expected error from StartListeningAndNotify")
	}
}

func TestGetConnWriteMu_SameInstance(t *testing.T) {
	s := newTestSocketServer()
	_, conn, done := dialWebSocketPair(t)
	defer done()
	a := s.getConnWriteMu(conn)
	b := s.getConnWriteMu(conn)
	if a != b {
		t.Fatal("expected same mutex for same connection")
	}
}

func TestWriteJSONAndSendError(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	if err := s.writeJSON(conn, map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(b)) != `{"a":1}` {
		t.Fatalf("got %q", b)
	}
	s.sendError(conn, "oops")
	_, b, err = client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("action: %q", p.Action)
	}
	if p.Data != "oops" {
		t.Fatalf("data: %v", p.Data)
	}
}

func TestHandlePing(t *testing.T) {
	s := newTestSocketServer()
	_, conn, done := dialWebSocketPair(t)
	defer done()
	s.conns.Store(conn, newUser(t).String())
	s.handlePing(conn)
}

func TestSetLiveDevHandler(t *testing.T) {
	s := newTestSocketServer()
	calls := 0
	s.SetLiveDevHandler(func(channel, payload string) {
		calls++
		if channel != "live_dev_logs" || payload != "p" {
			t.Errorf("bad args %q %q", channel, payload)
		}
	})
	s.liveDevHandlerMu.RLock()
	h := s.liveDevHandler
	s.liveDevHandlerMu.RUnlock()
	if h == nil {
		t.Fatal("expected handler")
	}
	h("live_dev_logs", "p")
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestHandleDisconnect_clearsTopics(t *testing.T) {
	s := newTestSocketServer()
	_, conn, done := dialWebSocketPair(t)
	s.SubscribeToTopic(MonitorApplicationDeployment, "app-1", conn)
	done()
	s.handleDisconnect(conn)
	s.topicsMu.RLock()
	_, has := s.topics[string(MonitorApplicationDeployment)+":app-1"]
	s.topicsMu.RUnlock()
	if has {
		t.Fatal("topic should be removed")
	}
	_, in := s.conns.Load(conn)
	if in {
		t.Fatal("conn should be removed")
	}
}

func TestHandleHTTP_TokenRequired(t *testing.T) {
	s := newTestSocketServer()
	ts := httptest.NewServer(http.HandlerFunc(s.HandleHTTP))
	t.Cleanup(ts.Close)
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: %d", res.StatusCode)
	}
}

func TestHandleHTTP_OrganizationQuerySetsHeader(t *testing.T) {
	s := newTestSocketServer()
	var orgHeader string
	s.verifyTokenOverride = func(_ string, r *http.Request) (*types.User, string, error) {
		if r != nil {
			orgHeader = r.Header.Get("X-Organization-Id")
		}
		return &types.User{ID: newUser(t)}, "from-session", nil
	}
	ts := httptest.NewServer(http.HandlerFunc(s.HandleHTTP))
	t.Cleanup(ts.Close)
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok&organization-id=org-from-query"
	client, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if orgHeader != "org-from-query" {
		t.Fatalf("org header: %q", orgHeader)
	}
}

func TestHandleHTTP_UpgradeWithOverride(t *testing.T) {
	s := newTestSocketServer()
	s.verifyTokenOverride = func(token string, r *http.Request) (*types.User, string, error) {
		if token != "ok" {
			return nil, "", errors.New("bad")
		}
		return &types.User{ID: newUser(t)}, "org-1", nil
	}
	ts := httptest.NewServer(http.HandlerFunc(s.HandleHTTP))
	t.Cleanup(ts.Close)
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=ok"
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.WriteMessage(websocket.TextMessage, []byte(`{"action":"ping"}`))
}

func TestHandleHTTP_InvalidToken(t *testing.T) {
	s := newTestSocketServer()
	s.verifyTokenOverride = func(string, *http.Request) (*types.User, string, error) {
		return nil, "", errors.New("nope")
	}
	ts := httptest.NewServer(http.HandlerFunc(s.HandleHTTP))
	t.Cleanup(ts.Close)
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=x"
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := c.ReadMessage()
	if err != nil {
		return
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err == nil && p.Action != "error" {
		t.Fatalf("expected error action, got %q", p.Action)
	}
}

func TestHandleHTTP_BearerStripsPrefix(t *testing.T) {
	s := newTestSocketServer()
	s.verifyTokenOverride = func(token string, _ *http.Request) (*types.User, string, error) {
		if token != "raw" {
			return nil, "", fmt.Errorf("got %q want raw", token)
		}
		return &types.User{ID: newUser(t)}, "", nil
	}
	ts := httptest.NewServer(http.HandlerFunc(s.HandleHTTP))
	t.Cleanup(ts.Close)
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=Bearer%20raw"
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}

func TestShutdown_closesConns(t *testing.T) {
	s := newTestSocketServer()
	_, server, done := dialWebSocketPair(t)
	t.Cleanup(done)
	s.conns.Store(server, "u")
	s.Shutdown()
}
