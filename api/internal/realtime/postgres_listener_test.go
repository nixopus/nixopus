package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nixopus/nixopus/api/internal/types"
)

func TestListenToApplicationChanges_connectError(t *testing.T) {
	old := pgxConnect
	pgxConnect = func(context.Context, string) (*pgx.Conn, error) { return nil, errors.New("no db") }
	t.Cleanup(func() { pgxConnect = old })
	l := NewPostgresListener()
	_, err := l.ListenToApplicationChanges(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStartListeningAndNotify_propagatesError(t *testing.T) {
	old := pgxConnect
	pgxConnect = func(context.Context, string) (*pgx.Conn, error) { return nil, errors.New("no db") }
	t.Cleanup(func() { pgxConnect = old })
	s := newTestSocketServer()
	pl := NewPostgresListener()
	err := StartListeningAndNotify(pl, context.Background(), s)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetConnString(t *testing.T) {
	c := types.Config{
		Database: types.DatabaseConfig{
			Host:     "db.example",
			Port:     "5432",
			Username: "u",
			Password: "p",
			Name:     "app",
			SSLMode:  "disable",
		},
	}
	s := getConnString(c)
	if s != "host=db.example port=5432 user=u password=p dbname=app sslmode=disable" {
		t.Fatalf("unexpected: %q", s)
	}
}

func TestNewPostgresListener(t *testing.T) {
	l := NewPostgresListener()
	if l == nil {
		t.Fatal("nil listener")
	}
	if l.notificationChan == nil {
		t.Fatal("expected channel")
	}
}

func TestHandleNotifications_applicationChanges(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.SubscribeToTopic(MonitorApplicationDeployment, "app-99", conn)
	_, _, _ = client.ReadMessage()
	payload := map[string]interface{}{
		"table":          "applications",
		"action":         "update",
		"application_id": "app-99",
		"data":           map[string]interface{}{},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *PostgresNotification, 2)
	go s.handleNotifications(ch)
	ch <- &PostgresNotification{Channel: "application_changes", Payload: string(b)}
	close(ch)
	time.Sleep(30 * time.Millisecond)
	_ = client.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var out types.Payload
	if err := json.Unmarshal(msg, &out); err != nil {
		t.Fatal(err)
	}
	if out.Action != "message" {
		t.Fatalf("action %q", out.Action)
	}
}

func TestHandleNotifications_applicationLogs_deploymentTopic(t *testing.T) {
	s := newTestSocketServer()
	client, conn, done := dialWebSocketPair(t)
	defer done()
	s.SubscribeToTopic(MonitorApplicationDeployment, "dep-1", conn)
	_, _, _ = client.ReadMessage()
	payload := map[string]interface{}{
		"table":          "application_logs",
		"action":         "insert",
		"application_id": "app-x",
		"data": map[string]interface{}{
			"application_deployment_id": "dep-1",
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *PostgresNotification, 1)
	go s.handleNotifications(ch)
	ch <- &PostgresNotification{Channel: "application_changes", Payload: string(b)}
	close(ch)
	time.Sleep(30 * time.Millisecond)
	_ = client.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var out types.Payload
	if err := json.Unmarshal(msg, &out); err != nil {
		t.Fatal(err)
	}
	if out.Action != "message" {
		t.Fatalf("action %q", out.Action)
	}
}

func TestHandleNotifications_invalidJSON_skips(t *testing.T) {
	s := newTestSocketServer()
	_, _, done := dialWebSocketPair(t)
	t.Cleanup(done)
	ch := make(chan *PostgresNotification, 1)
	go s.handleNotifications(ch)
	ch <- &PostgresNotification{Channel: "application_changes", Payload: "not-json"}
	close(ch)
}

func TestHandleNotifications_liveDev(t *testing.T) {
	s := newTestSocketServer()
	var gotCh, gotPay string
	s.SetLiveDevHandler(func(ch, pay string) {
		gotCh, gotPay = ch, pay
	})
	ch := make(chan *PostgresNotification, 1)
	go s.handleNotifications(ch)
	ch <- &PostgresNotification{Channel: "live_dev_logs", Payload: "hello"}
	close(ch)
	time.Sleep(20 * time.Millisecond)
	if gotCh != "live_dev_logs" || gotPay != "hello" {
		t.Fatalf("got %q %q", gotCh, gotPay)
	}
}

func TestHandleNotifications_liveDev_nilHandler(t *testing.T) {
	s := newTestSocketServer()
	ch := make(chan *PostgresNotification, 1)
	go s.handleNotifications(ch)
	ch <- &PostgresNotification{Channel: "live_dev_status", Payload: "x"}
	close(ch)
}
