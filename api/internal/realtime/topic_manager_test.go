package realtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/internal/types"
)

func TestSubscribeToTopic_withResourceID(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.SubscribeToTopic(MonitorApplicationDeployment, "res-1", conn)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
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
	if p.Topic != "monitor_application_deployment:res-1" {
		t.Fatalf("topic %q", p.Topic)
	}
}

func TestSubscribeToTopic_emptyResourceID(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.SubscribeToTopic(MonitorHealthCheck, "", conn)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Topic != "monitor_health_check" {
		t.Fatalf("topic %q", p.Topic)
	}
}

func TestUnsubscribeFromTopic(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.SubscribeToTopic(MonitorApplicationDeployment, "r1", conn)
	_, _, _ = client.ReadMessage()
	s.UnsubscribeFromTopic(MonitorApplicationDeployment, "r1", conn)
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "unsubscribed" {
		t.Fatalf("action %q", p.Action)
	}
}

func TestUnsubscribeFromTopic_notSubscribed(t *testing.T) {
	s := newTestSocketServer()
	_, conn, done := dialWebSocketPair(t)
	defer done()
	s.UnsubscribeFromTopic(MonitorApplicationDeployment, "nope", conn)
}

func TestBroadcastToTopic(t *testing.T) {
	s := newTestSocketServer()
	c1, s1, d1 := dialWebSocketPair(t)
	defer d1()
	s.SubscribeToTopic(MonitorApplicationDeployment, "b1", s1)
	_, _, _ = c1.ReadMessage()
	s.BroadcastToTopic(MonitorApplicationDeployment, "b1", map[string]string{"k": "v"})
	_ = c1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := c1.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "message" {
		t.Fatalf("action %q", p.Action)
	}
}

func TestHandleSubscribe_valid(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleSubscribe(conn, types.Payload{
		Action: types.SUBSCRIBE,
		Topic:  string(MonitorApplicationDeployment),
		Data:   map[string]interface{}{"resource_id": "r-valid"},
	})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
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

func TestHandleUnsubscribe_valid(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleSubscribe(conn, types.Payload{
		Action: types.SUBSCRIBE,
		Topic:  string(MonitorApplicationDeployment),
		Data:   map[string]interface{}{"resource_id": "r-u"},
	})
	_, _, _ = client.ReadMessage()
	s.handleUnsubscribe(conn, types.Payload{
		Action: types.UNSUBSCRIBE,
		Topic:  string(MonitorApplicationDeployment),
		Data:   map[string]interface{}{"resource_id": "r-u"},
	})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "unsubscribed" {
		t.Fatalf("action %q", p.Action)
	}
}

func TestHandleSubscribe_errors(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleSubscribe(conn, types.Payload{Action: "subscribe", Topic: string(MonitorApplicationDeployment), Data: map[string]interface{}{}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("expected error, got %q", p.Action)
	}
}

func TestHandleSubscribe_invalid(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleSubscribe(conn, types.Payload{Action: "subscribe"})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("expected error, got %q", p.Action)
	}
}

func TestHandleUnsubscribe_errors(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleUnsubscribe(conn, types.Payload{Action: "unsubscribe", Topic: string(MonitorApplicationDeployment), Data: map[string]interface{}{}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("expected error, got %q", p.Action)
	}
}

func TestBroadcastToTopic_unsubscribeOnWriteError(t *testing.T) {
	s := newTestSocketServer()
	c1, s1, d1 := dialWebSocketPair(t)
	c2, s2, d2 := dialWebSocketPair(t)
	defer d1()
	defer d2()
	s.SubscribeToTopic(MonitorApplicationDeployment, "w1", s1)
	_, _, _ = c1.ReadMessage()
	s.SubscribeToTopic(MonitorApplicationDeployment, "w1", s2)
	_, _, _ = c2.ReadMessage()
	_ = s1.Close()
	s.BroadcastToTopic(MonitorApplicationDeployment, "w1", map[string]int{"n": 1})
	time.Sleep(50 * time.Millisecond)
	c2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, msg, err := c2.ReadMessage()
	if err != nil {
		t.Fatalf("c2 read: %v", err)
	}
	var p types.Payload
	if err := json.Unmarshal(msg, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "message" {
		t.Fatalf("action %q", p.Action)
	}
}

func TestHandleUnsubscribe_invalid(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleUnsubscribe(conn, types.Payload{Action: "unsubscribe"})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var p types.Payload
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Action != "error" {
		t.Fatalf("expected error, got %q", p.Action)
	}
}
