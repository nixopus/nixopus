package memory

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { db.Close() })
	return db
}

func createTestTables(t *testing.T, db *bun.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS agent_threads (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			metadata TEXT DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_threads_user ON agent_threads(user_id)`,
		`CREATE TABLE IF NOT EXISTS agent_messages (
			id TEXT PRIMARY KEY,
			thread_id TEXT NOT NULL REFERENCES agent_threads(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			tool_calls TEXT DEFAULT '[]',
			tool_call_id TEXT DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			seq INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_messages_thread_seq ON agent_messages(thread_id, seq)`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(context.Background(), s); err != nil {
			t.Fatalf("create test tables: %v", err)
		}
	}
}

func setupStore(t *testing.T) *PostgresStore {
	t.Helper()
	db := setupTestDB(t)
	createTestTables(t, db)
	return NewPostgresStore(db)
}

func TestCreateThread(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Test Thread", Metadata: JSONMap{"user": "123"}}
	err := store.CreateThread(ctx, thread)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if thread.ID == "" {
		t.Error("expected ID to be generated")
	}
	if thread.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if thread.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", thread.UserID)
	}
}

func TestCreateThread_WithID(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{ID: "custom-id", UserID: "user-1", Title: "Custom"}
	err := store.CreateThread(ctx, thread)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if thread.ID != "custom-id" {
		t.Errorf("expected custom-id, got %s", thread.ID)
	}
}

func TestGetThread(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Get Me"}
	store.CreateThread(ctx, thread)

	got, err := store.GetThread(ctx, thread.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Get Me" {
		t.Errorf("expected 'Get Me', got %q", got.Title)
	}
}

func TestGetThreadForUser(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "My Thread"}
	store.CreateThread(ctx, thread)

	got, err := store.GetThreadForUser(ctx, thread.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "My Thread" {
		t.Errorf("expected 'My Thread', got %q", got.Title)
	}

	_, err = store.GetThreadForUser(ctx, thread.ID, "other-user")
	if err == nil {
		t.Fatal("expected error for wrong user")
	}
}

func TestGetThread_NotFound(t *testing.T) {
	store := setupStore(t)
	_, err := store.GetThread(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing thread")
	}
}

func TestDeleteThread(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Delete Me"}
	store.CreateThread(ctx, thread)

	msgs := []StoredMessage{
		{ID: "m1", ThreadID: thread.ID, Role: llm.RoleUser, Content: "hello", Seq: 1, CreatedAt: time.Now()},
	}
	store.AppendMessages(ctx, thread.ID, msgs)

	// Wrong user should not delete
	err := store.DeleteThread(ctx, thread.ID, "other-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = store.GetThread(ctx, thread.ID)
	if err != nil {
		t.Fatal("thread should still exist after wrong-user delete")
	}

	// Correct user deletes
	err = store.DeleteThread(ctx, thread.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.GetThread(ctx, thread.ID)
	if err == nil {
		t.Error("expected thread to be deleted")
	}

	messages, _ := store.GetMessages(ctx, thread.ID, 0)
	if len(messages) != 0 {
		t.Errorf("expected messages to be deleted, got %d", len(messages))
	}
}

func TestListThreads(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		store.CreateThread(ctx, &Thread{UserID: "user-1", Title: "Thread"})
		time.Sleep(time.Millisecond)
	}
	store.CreateThread(ctx, &Thread{UserID: "user-2", Title: "Other User Thread"})

	threads, err := store.ListThreads(ctx, "user-1", 3, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(threads) != 3 {
		t.Errorf("expected 3 threads, got %d", len(threads))
	}

	threads, err = store.ListThreads(ctx, "user-1", 10, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(threads) != 2 {
		t.Errorf("expected 2 threads (offset 3), got %d", len(threads))
	}

	// user-2 should only see their own thread
	threads, err = store.ListThreads(ctx, "user-2", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(threads) != 1 {
		t.Errorf("expected 1 thread for user-2, got %d", len(threads))
	}
}

func TestAppendMessages(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Messages"}
	store.CreateThread(ctx, thread)

	msgs := MessagesFromLLM(thread.ID, []llm.Message{
		{Role: llm.RoleUser, Content: "hello"},
		{Role: llm.RoleAssistant, Content: "hi there"},
	}, 1)

	err := store.AppendMessages(ctx, thread.ID, msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, err := store.GetMessages(ctx, thread.ID, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(stored))
	}
	if stored[0].Content != "hello" {
		t.Errorf("expected 'hello', got %q", stored[0].Content)
	}
	if stored[1].Content != "hi there" {
		t.Errorf("expected 'hi there', got %q", stored[1].Content)
	}
}

func TestAppendMessages_Empty(t *testing.T) {
	store := setupStore(t)
	err := store.AppendMessages(context.Background(), "any", nil)
	if err != nil {
		t.Fatalf("unexpected error for empty append: %v", err)
	}
}

func TestAppendMessages_WithToolCalls(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Tool Thread"}
	store.CreateThread(ctx, thread)

	msgs := []StoredMessage{
		{
			ID:       "m1",
			ThreadID: thread.ID,
			Role:     llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{{
				ID:       "call_1",
				Type:     "function",
				Function: llm.FunctionCall{Name: "search", Arguments: `{"q":"test"}`},
			}},
			Seq:       1,
			CreatedAt: time.Now(),
		},
		{
			ID:         "m2",
			ThreadID:   thread.ID,
			Role:       llm.RoleTool,
			Content:    `{"results":[]}`,
			ToolCallID: "call_1",
			Seq:        2,
			CreatedAt:  time.Now(),
		},
	}

	err := store.AppendMessages(ctx, thread.ID, msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.GetMessages(ctx, thread.ID, 0)
	if len(stored) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(stored))
	}

	if len(stored[0].ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(stored[0].ToolCalls))
	}
	if stored[0].ToolCalls[0].Function.Name != "search" {
		t.Errorf("expected 'search', got %q", stored[0].ToolCalls[0].Function.Name)
	}
	if stored[1].ToolCallID != "call_1" {
		t.Errorf("expected tool_call_id 'call_1', got %q", stored[1].ToolCallID)
	}
}

func TestGetMessages_WithLimit(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Limit Test"}
	store.CreateThread(ctx, thread)

	msgs := MessagesFromLLM(thread.ID, []llm.Message{
		{Role: llm.RoleUser, Content: "1"},
		{Role: llm.RoleAssistant, Content: "2"},
		{Role: llm.RoleUser, Content: "3"},
		{Role: llm.RoleAssistant, Content: "4"},
		{Role: llm.RoleUser, Content: "5"},
	}, 1)
	store.AppendMessages(ctx, thread.ID, msgs)

	stored, err := store.GetMessages(ctx, thread.ID, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) != 3 {
		t.Errorf("expected 3 messages with limit, got %d", len(stored))
	}
	if stored[0].Content != "1" {
		t.Errorf("expected first message '1', got %q", stored[0].Content)
	}
}

func TestGetMessages_Ordering(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Order Test"}
	store.CreateThread(ctx, thread)

	msgs := MessagesFromLLM(thread.ID, []llm.Message{
		{Role: llm.RoleUser, Content: "first"},
		{Role: llm.RoleAssistant, Content: "second"},
		{Role: llm.RoleUser, Content: "third"},
	}, 1)
	store.AppendMessages(ctx, thread.ID, msgs)

	stored, _ := store.GetMessages(ctx, thread.ID, 0)
	for i, m := range stored {
		if m.Seq != i+1 {
			t.Errorf("message %d: expected seq %d, got %d", i, i+1, m.Seq)
		}
	}
}

func TestGetMessageCount(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	thread := &Thread{UserID: "user-1", Title: "Count"}
	store.CreateThread(ctx, thread)

	count, err := store.GetMessageCount(ctx, thread.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	msgs := MessagesFromLLM(thread.ID, []llm.Message{
		{Role: llm.RoleUser, Content: "a"},
		{Role: llm.RoleAssistant, Content: "b"},
	}, 1)
	store.AppendMessages(ctx, thread.ID, msgs)

	count, _ = store.GetMessageCount(ctx, thread.ID)
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestStoredMessage_ToLLMMessage(t *testing.T) {
	msg := StoredMessage{
		Role:    llm.RoleAssistant,
		Content: "hello",
		ToolCalls: []llm.ToolCall{{
			ID:       "c1",
			Type:     "function",
			Function: llm.FunctionCall{Name: "test", Arguments: "{}"},
		}},
		ToolCallID: "c0",
	}

	lm := msg.ToLLMMessage()
	if lm.Role != llm.RoleAssistant {
		t.Errorf("expected assistant, got %s", lm.Role)
	}
	if lm.Content != "hello" {
		t.Errorf("expected 'hello', got %q", lm.Content)
	}
	if len(lm.ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(lm.ToolCalls))
	}
	if lm.ToolCallID != "c0" {
		t.Errorf("expected 'c0', got %q", lm.ToolCallID)
	}
}

func TestMessagesFromLLM(t *testing.T) {
	msgs := MessagesFromLLM("thread-1", []llm.Message{
		{Role: llm.RoleUser, Content: "hi"},
		{Role: llm.RoleAssistant, Content: "hello"},
	}, 5)

	if len(msgs) != 2 {
		t.Fatalf("expected 2, got %d", len(msgs))
	}
	if msgs[0].ThreadID != "thread-1" {
		t.Errorf("expected thread-1, got %s", msgs[0].ThreadID)
	}
	if msgs[0].Seq != 5 {
		t.Errorf("expected seq 5, got %d", msgs[0].Seq)
	}
	if msgs[1].Seq != 6 {
		t.Errorf("expected seq 6, got %d", msgs[1].Seq)
	}
	if msgs[0].ID == "" || msgs[1].ID == "" {
		t.Error("expected IDs to be generated")
	}
	if msgs[0].ID == msgs[1].ID {
		t.Error("expected unique IDs")
	}
}

func TestJSONMap_ScanNil(t *testing.T) {
	var m JSONMap
	err := m.Scan(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Error("expected initialized map")
	}
}

func TestJSONMap_ScanBytes(t *testing.T) {
	var m JSONMap
	err := m.Scan([]byte(`{"key":"val"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["key"] != "val" {
		t.Errorf("expected 'val', got %q", m["key"])
	}
}

func TestJSONMap_ScanString(t *testing.T) {
	var m JSONMap
	err := m.Scan(`{"a":"b"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["a"] != "b" {
		t.Errorf("expected 'b', got %q", m["a"])
	}
}

func TestJSONMap_ScanUnsupported(t *testing.T) {
	var m JSONMap
	err := m.Scan(12345)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestJSONMap_ValueNil(t *testing.T) {
	var m JSONMap
	v, err := m.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "{}" {
		t.Errorf("expected '{}', got %v", v)
	}
}

func TestJSONToolCalls_ScanNil(t *testing.T) {
	var tc JSONToolCalls
	err := tc.Scan(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc != nil {
		t.Error("expected nil for nil scan")
	}
}

func TestJSONToolCalls_ScanBytes(t *testing.T) {
	var tc JSONToolCalls
	err := tc.Scan([]byte(`[{"id":"c1","type":"function","function":{"name":"test","arguments":"{}"}}]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tc) != 1 || tc[0].ID != "c1" {
		t.Errorf("unexpected result: %+v", tc)
	}
}

func TestJSONToolCalls_ScanString(t *testing.T) {
	var tc JSONToolCalls
	err := tc.Scan(`[{"id":"c2","type":"function","function":{"name":"x","arguments":"{}"}}]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tc) != 1 || tc[0].ID != "c2" {
		t.Errorf("unexpected result: %+v", tc)
	}
}

func TestJSONToolCalls_ScanUnsupported(t *testing.T) {
	var tc JSONToolCalls
	err := tc.Scan(99999)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestJSONToolCalls_ValueNil(t *testing.T) {
	var tc JSONToolCalls
	v, err := tc.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "[]" {
		t.Errorf("expected '[]', got %v", v)
	}
}

func closedDBStore(t *testing.T) *PostgresStore {
	t.Helper()
	db := setupTestDB(t)
	createTestTables(t, db)
	store := NewPostgresStore(db)
	db.Close()
	return store
}

func TestCreateThread_DBError(t *testing.T) {
	store := closedDBStore(t)
	err := store.CreateThread(context.Background(), &Thread{UserID: "u1", Title: "fail"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetThread_DBError(t *testing.T) {
	store := closedDBStore(t)
	_, err := store.GetThread(context.Background(), "id")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteThread_DBError(t *testing.T) {
	store := closedDBStore(t)
	err := store.DeleteThread(context.Background(), "id", "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListThreads_DBError(t *testing.T) {
	store := closedDBStore(t)
	_, err := store.ListThreads(context.Background(), "user-1", 10, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppendMessages_DBError(t *testing.T) {
	store := closedDBStore(t)
	msgs := []StoredMessage{{ID: "m1", ThreadID: "t1", Role: llm.RoleUser, Content: "hi", Seq: 1, CreatedAt: time.Now()}}
	err := store.AppendMessages(context.Background(), "t1", msgs)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMessages_DBError(t *testing.T) {
	store := closedDBStore(t)
	_, err := store.GetMessages(context.Background(), "t1", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMessageCount_DBError(t *testing.T) {
	store := closedDBStore(t)
	_, err := store.GetMessageCount(context.Background(), "t1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteThread_SecondDeleteError(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	store := NewPostgresStore(db)

	thread := &Thread{UserID: "user-1", Title: "test"}
	store.CreateThread(context.Background(), thread)
	msgs := []StoredMessage{{ID: "m1", ThreadID: thread.ID, Role: llm.RoleUser, Content: "hi", Seq: 1, CreatedAt: time.Now()}}
	store.AppendMessages(context.Background(), thread.ID, msgs)

	// Drop threads table to make second delete fail
	db.ExecContext(context.Background(), "DROP TABLE agent_threads")

	err := store.DeleteThread(context.Background(), thread.ID, "user-1")
	if err == nil {
		t.Fatal("expected error from missing threads table on second delete")
	}
}

func TestAppendMessages_UpdateTimestampError(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	store := NewPostgresStore(db)

	thread := &Thread{UserID: "user-1", Title: "test"}
	store.CreateThread(context.Background(), thread)

	// Drop the threads table to force update error
	db.ExecContext(context.Background(), "DROP TABLE agent_threads")

	msgs := []StoredMessage{{ID: "m1", ThreadID: thread.ID, Role: llm.RoleUser, Content: "hi", Seq: 1, CreatedAt: time.Now()}}
	err := store.AppendMessages(context.Background(), thread.ID, msgs)
	if err == nil {
		t.Fatal("expected error from missing threads table")
	}
}
