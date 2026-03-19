package queue

import (
	"context"
	"log"
	"strings"
	"sync"
)

const replyChannelPrefix = "machine:reply:"

type ReplyMultiplexer struct {
	waiters sync.Map
}

func NewReplyMultiplexer() *ReplyMultiplexer {
	return &ReplyMultiplexer{}
}

func (m *ReplyMultiplexer) Start(ctx context.Context) {
	if redisClient == nil {
		log.Println("[reply-mux] Redis client not initialized, skipping PSUBSCRIBE")
		return
	}

	go func() {
		pubsub := redisClient.PSubscribe(ctx, replyChannelPrefix+"*")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				requestID := ExtractRequestIDFromChannel(msg.Channel)
				if requestID != "" {
					m.Dispatch(requestID, msg.Payload)
				}
			}
		}
	}()
}

func (m *ReplyMultiplexer) RegisterWaiter(requestID string) <-chan string {
	ch := make(chan string, 1)
	m.waiters.Store(requestID, ch)
	return ch
}

func (m *ReplyMultiplexer) RemoveWaiter(requestID string) {
	m.waiters.Delete(requestID)
}

func (m *ReplyMultiplexer) Dispatch(requestID string, data string) {
	val, ok := m.waiters.Load(requestID)
	if !ok {
		return
	}
	ch := val.(chan string)
	select {
	case ch <- data:
	default:
	}
}

func ExtractRequestIDFromChannel(channel string) string {
	if !strings.HasPrefix(channel, replyChannelPrefix) {
		return ""
	}
	return strings.TrimPrefix(channel, replyChannelPrefix)
}
