package realtime

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
)

func TestHandleTerminal_invalidData(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleTerminal(conn, types.Payload{Data: "x"})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestHandleTerminal_missingTerminalId(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleTerminal(conn, types.Payload{Data: map[string]interface{}{"value": "x"}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestHandleTerminal_invalidInput(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleTerminal(conn, types.Payload{Data: map[string]interface{}{"terminalId": "t1", "value": 1}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestCreateTerminal_noOrg(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleTerminal(conn, types.Payload{Data: map[string]interface{}{"terminalId": "t1", "value": "hi"}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestCreateTerminal_invalidOrgUUID(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.orgIDs.Store(conn, "not-a-uuid")
	s.handleTerminal(conn, types.Payload{Data: map[string]interface{}{"terminalId": "t1", "value": "hi"}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestCreateTerminal_validOrg_terminaFailsWithoutSSH(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.orgIDs.Store(conn, uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8").String())
	s.handleTerminal(conn, types.Payload{Data: map[string]interface{}{"terminalId": "t1", "value": "x"}})
	_ = client.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestHandleTerminalResize_errors(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleTerminalResize(conn, types.Payload{Data: "nope"})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func TestHandleTerminalResize_notFound(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.handleTerminalResize(conn, types.Payload{Data: map[string]interface{}{"terminalId": "n", "rows": float64(1), "cols": float64(1)}})
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, b, err := client.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAction(b, "error") {
		t.Fatalf("expected error: %s", b)
	}
}

func containsAction(b []byte, want string) bool {
	return strings.Contains(string(b), `"action":"`+want+`"`)
}
