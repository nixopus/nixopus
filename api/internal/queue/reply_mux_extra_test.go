package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReplyMultiplexer_Dispatch_dropsWhenBufferFull(t *testing.T) {
	m := NewReplyMultiplexer()
	ch := m.RegisterWaiter("rid")
	m.Dispatch("rid", "first")
	m.Dispatch("rid", "second")
	select {
	case v := <-ch:
		assert.Equal(t, "first", v)
	case <-time.After(time.Second):
		t.Fatal("expected first message")
	}
	m.RemoveWaiter("rid")
}

func TestReplyMultiplexerWithPrefix_Start_skipsWithoutRedis(t *testing.T) {
	prev := redisClient
	redisClient = nil
	t.Cleanup(func() { redisClient = prev })

	m := NewReplyMultiplexerWithPrefix("custom:")
	pctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	m.Start(pctx)
}
