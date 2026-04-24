package realtime

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nixopus/nixopus/api/internal/types"
)

func TestHandleStopDashboardMonitor_noMonitor(t *testing.T) {
	s := newTestSocketServer()
	_, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleStopDashboardMonitor(conn)
}

func TestSendResponse_marshalable(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.sendResponse(conn, types.Payload{Action: "ok", Data: map[string]string{"a": "b"}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	mt, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if mt != websocket.TextMessage {
		t.Fatalf("message type %d", mt)
	}
	if string(b) == "" {
		t.Fatal("empty body")
	}
}

func TestSendResponse_marshalError(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.sendResponse(conn, types.Payload{Data: map[string]interface{}{"c": make(chan int)}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestHandleDashboardMonitor_nilDeployController(t *testing.T) {
	s := newTestSocketServer()
	s.deployController = nil
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleDashboardMonitor(conn, types.Payload{
		Action: types.DASHBOARD_MONITOR,
		Data: map[string]interface{}{
			"organization_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
	})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}
