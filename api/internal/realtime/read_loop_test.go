package realtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nixopus/nixopus/api/internal/types"
)

func TestReadLoop_ping(t *testing.T) {
	s := newTestSocketServer()
	client, server, done := dialWebSocketPair(t)
	defer done()
	go s.readLoop(server)
	_ = client.WriteMessage(websocket.TextMessage, []byte(`{"action":"ping"}`))
	client.Close()
}

func TestReadLoop_invalidJSON(t *testing.T) {
	s := newTestSocketServer()
	client, server, done := dialWebSocketPair(t)
	defer done()
	go s.readLoop(server)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_ = client.WriteMessage(websocket.TextMessage, []byte(`not json`))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("got %q", p.Action)
	}
	_ = client.Close()
}

func TestReadLoop_unknownAction(t *testing.T) {
	s := newTestSocketServer()
	client, server, done := dialWebSocketPair(t)
	defer done()
	go s.readLoop(server)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_ = client.WriteMessage(websocket.TextMessage, []byte(`{"action":"nope"}`))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("got %q", p.Action)
	}
}

func TestReadLoop_subscribe(t *testing.T) {
	s := newTestSocketServer()
	client, server, done := dialWebSocketPair(t)
	defer done()
	go s.readLoop(server)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, _ := json.Marshal(types.Payload{
		Action: types.SUBSCRIBE,
		Topic:  string(MonitorApplicationDeployment),
		Data:   map[string]interface{}{"resource_id": "sub-1"},
	})
	_ = client.WriteMessage(websocket.TextMessage, msg)
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "subscribed" {
		t.Fatalf("action %q", p.Action)
	}
}

func TestReadLoop_stopDashboardMonitor(t *testing.T) {
	s := newTestSocketServer()
	client, server, done := dialWebSocketPair(t)
	defer done()
	go s.readLoop(server)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, _ := json.Marshal(types.Payload{Action: types.STOP_DASHBOARD_MONITOR})
	_ = client.WriteMessage(websocket.TextMessage, msg)
}

func TestReadLoop_monitorApplication_noop(t *testing.T) {
	s := newTestSocketServer()
	client, server, done := dialWebSocketPair(t)
	defer done()
	go s.readLoop(server)
	msg, _ := json.Marshal(types.Payload{Action: types.MONITOR_APPLICATION})
	_ = client.WriteMessage(websocket.TextMessage, msg)
	_ = client.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err := client.ReadMessage()
	if err == nil {
		t.Fatal("expected no reply from unimplemented monitor action")
	}
	_ = client.Close()
}
