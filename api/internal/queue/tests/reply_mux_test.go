package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/queue"
	"github.com/stretchr/testify/assert"
)

func TestReplyMux_RegisterAndRemoveWaiter(t *testing.T) {
	mux := queue.NewReplyMultiplexer()

	ch := mux.RegisterWaiter("req-1")
	assert.NotNil(t, ch, "RegisterWaiter should return a non-nil channel")

	mux.RemoveWaiter("req-1")

	ch2 := mux.RegisterWaiter("req-1")
	assert.NotNil(t, ch2, "should be able to re-register after removal")
	mux.RemoveWaiter("req-1")
}

func TestReplyMux_DispatchToCorrectWaiter(t *testing.T) {
	mux := queue.NewReplyMultiplexer()

	ch1 := mux.RegisterWaiter("req-1")
	ch2 := mux.RegisterWaiter("req-2")

	mux.Dispatch("req-1", `{"success":true}`)

	select {
	case msg := <-ch1:
		assert.Equal(t, `{"success":true}`, msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch1 should have received the message")
	}

	select {
	case <-ch2:
		t.Fatal("ch2 should NOT have received any message")
	case <-time.After(50 * time.Millisecond):
	}

	mux.RemoveWaiter("req-1")
	mux.RemoveWaiter("req-2")
}

func TestReplyMux_DispatchUnknownRequestIgnored(t *testing.T) {
	mux := queue.NewReplyMultiplexer()
	assert.NotPanics(t, func() {
		mux.Dispatch("nonexistent", `{"data":"ignored"}`)
	})
}

func TestReplyMux_ConcurrentWaiters(t *testing.T) {
	mux := queue.NewReplyMultiplexer()
	const n = 20

	ids := make([]string, n)
	channels := make([]<-chan string, n)
	for i := 0; i < n; i++ {
		ids[i] = "req-" + time.Now().Format("150405") + "-" + string(rune('A'+i))
		channels[i] = mux.RegisterWaiter(ids[i])
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mux.Dispatch(ids[idx], `{"idx":`+ids[idx]+`}`)
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		select {
		case msg := <-channels[i]:
			assert.NotEmpty(t, msg, "waiter %d received empty message", i)
		case <-time.After(200 * time.Millisecond):
			t.Errorf("waiter %d timed out", i)
		}
	}

	for i := 0; i < n; i++ {
		mux.RemoveWaiter(ids[i])
	}
}

func TestReplyMux_ChannelBuffered(t *testing.T) {
	mux := queue.NewReplyMultiplexer()
	ch := mux.RegisterWaiter("req-buf")

	mux.Dispatch("req-buf", `{"buffered":true}`)

	msg := <-ch
	assert.Equal(t, `{"buffered":true}`, msg)

	mux.RemoveWaiter("req-buf")
}

func TestReplyMux_ExtractRequestID(t *testing.T) {
	tests := []struct {
		channel string
		want    string
	}{
		{"machine:reply:abc-123", "abc-123"},
		{"machine:reply:req-with-dashes", "req-with-dashes"},
		{"machine:reply:", ""},
		{"other:channel", ""},
		{"machine:reply:uuid-1234-5678", "uuid-1234-5678"},
	}

	for _, tt := range tests {
		got := queue.ExtractRequestIDFromChannel(tt.channel)
		assert.Equal(t, tt.want, got, "ExtractRequestIDFromChannel(%q)", tt.channel)
	}
}
