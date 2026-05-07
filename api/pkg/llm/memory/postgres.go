package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PostgresStore struct {
	db *bun.DB
}

func NewPostgresStore(db *bun.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateThread(ctx context.Context, thread *Thread) error {
	if thread.ID == "" {
		thread.ID = uuid.New().String()
	}
	now := time.Now()
	thread.CreatedAt = now
	thread.UpdatedAt = now

	_, err := s.db.NewInsert().
		Model(thread).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: insert thread: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetThread(ctx context.Context, id string) (*Thread, error) {
	thread := &Thread{}
	err := s.db.NewSelect().
		Model(thread).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("memory: get thread %s: %w", id, err)
	}
	return thread, nil
}

func (s *PostgresStore) GetThreadForUser(ctx context.Context, id, userID string) (*Thread, error) {
	thread := &Thread{}
	err := s.db.NewSelect().
		Model(thread).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("memory: get thread %s for user %s: %w", id, userID, err)
	}
	return thread, nil
}

func (s *PostgresStore) UpdateThread(ctx context.Context, id, userID, title string) error {
	_, err := s.db.NewUpdate().
		Model((*Thread)(nil)).
		Set("title = ?", title).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: update thread %s: %w", id, err)
	}
	return nil
}

func (s *PostgresStore) DeleteThread(ctx context.Context, id, userID string) error {
	_, err := s.db.NewDelete().
		Model((*StoredMessage)(nil)).
		Where("thread_id = ?", id).
		Where("thread_id IN (SELECT id FROM agent_threads WHERE id = ? AND user_id = ?)", id, userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: delete messages for thread %s: %w", id, err)
	}

	_, err = s.db.NewDelete().
		Model((*Thread)(nil)).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: delete thread %s: %w", id, err)
	}
	return nil
}

func (s *PostgresStore) ListThreads(ctx context.Context, userID string, limit, offset int) ([]Thread, error) {
	var threads []Thread
	err := s.db.NewSelect().
		Model(&threads).
		Where("user_id = ?", userID).
		OrderExpr("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("memory: list threads: %w", err)
	}
	return threads, nil
}

func (s *PostgresStore) AppendMessages(ctx context.Context, threadID string, messages []StoredMessage) error {
	if len(messages) == 0 {
		return nil
	}

	_, err := s.db.NewInsert().
		Model(&messages).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: append messages to thread %s: %w", threadID, err)
	}

	_, err = s.db.NewUpdate().
		Model((*Thread)(nil)).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", threadID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: update thread timestamp %s: %w", threadID, err)
	}

	return nil
}

func (s *PostgresStore) GetMessages(ctx context.Context, threadID string, limit int) ([]StoredMessage, error) {
	var messages []StoredMessage
	q := s.db.NewSelect().
		Model(&messages).
		Where("thread_id = ?", threadID).
		OrderExpr("seq ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("memory: get messages for thread %s: %w", threadID, err)
	}
	return messages, nil
}

func (s *PostgresStore) GetMessageCount(ctx context.Context, threadID string) (int, error) {
	count, err := s.db.NewSelect().
		Model((*StoredMessage)(nil)).
		Where("thread_id = ?", threadID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("memory: count messages for thread %s: %w", threadID, err)
	}
	return count, nil
}
