package memory

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/uptrace/bun"
)

type JSONMap map[string]string

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *JSONMap) Scan(src interface{}) error {
	if src == nil {
		*m = make(map[string]string)
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("memory: unsupported type for JSONMap: %T", src)
	}
	return json.Unmarshal(data, m)
}

type JSONToolCalls []llm.ToolCall

func (tc JSONToolCalls) Value() (driver.Value, error) {
	if tc == nil {
		return "[]", nil
	}
	b, err := json.Marshal(tc)
	return string(b), err
}

func (tc *JSONToolCalls) Scan(src interface{}) error {
	if src == nil {
		*tc = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("memory: unsupported type for JSONToolCalls: %T", src)
	}
	return json.Unmarshal(data, tc)
}

type Thread struct {
	bun.BaseModel `bun:"table:agent_threads"`

	ID        string    `json:"id" bun:"id,pk"`
	UserID    string    `json:"user_id" bun:"user_id,notnull"`
	Title     string    `json:"title" bun:"title"`
	Metadata  JSONMap   `json:"metadata" bun:"metadata"`
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`
}

type StoredMessage struct {
	bun.BaseModel `bun:"table:agent_messages"`

	ID         string        `json:"id" bun:"id,pk"`
	ThreadID   string        `json:"thread_id" bun:"thread_id"`
	Role       llm.Role      `json:"role" bun:"role"`
	Content    string        `json:"content" bun:"content"`
	ToolCalls  JSONToolCalls `json:"tool_calls,omitempty" bun:"tool_calls"`
	ToolCallID string        `json:"tool_call_id,omitempty" bun:"tool_call_id"`
	CreatedAt  time.Time     `json:"created_at" bun:"created_at"`
	Seq        int           `json:"seq" bun:"seq"`
}

func (m *StoredMessage) ToLLMMessage() llm.Message {
	return llm.Message{
		Role:       m.Role,
		Content:    m.Content,
		ToolCalls:  m.ToolCalls,
		ToolCallID: m.ToolCallID,
	}
}

func MessagesFromLLM(threadID string, messages []llm.Message, startSeq int) []StoredMessage {
	stored := make([]StoredMessage, 0, len(messages))
	for i, msg := range messages {
		stored = append(stored, StoredMessage{
			ID:         uuid.New().String(),
			ThreadID:   threadID,
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  msg.ToolCalls,
			ToolCallID: msg.ToolCallID,
			CreatedAt:  time.Now(),
			Seq:        startSeq + i,
		})
	}
	return stored
}

type Store interface {
	CreateThread(ctx context.Context, thread *Thread) error
	GetThread(ctx context.Context, id string) (*Thread, error)
	GetThreadForUser(ctx context.Context, id, userID string) (*Thread, error)
	UpdateThread(ctx context.Context, id, userID, title string) error
	DeleteThread(ctx context.Context, id, userID string) error
	ListThreads(ctx context.Context, userID string, limit, offset int) ([]Thread, error)

	AppendMessages(ctx context.Context, threadID string, messages []StoredMessage) error
	GetMessages(ctx context.Context, threadID string, limit int) ([]StoredMessage, error)
	GetMessageCount(ctx context.Context, threadID string) (int, error)
}
