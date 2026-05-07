package service

import (
	"context"
	"sync"
)

// activeStreams tracks in-flight streaming requests by thread ID
// so they can be explicitly cancelled via API.
var (
	activeStreams   = make(map[string]context.CancelFunc)
	activeStreamsMu sync.Mutex
)

func RegisterStream(threadID string, cancel context.CancelFunc) {
	activeStreamsMu.Lock()
	activeStreams[threadID] = cancel
	activeStreamsMu.Unlock()
}

func UnregisterStream(threadID string) {
	activeStreamsMu.Lock()
	delete(activeStreams, threadID)
	activeStreamsMu.Unlock()
}

func CancelStream(threadID string) bool {
	activeStreamsMu.Lock()
	cancel, ok := activeStreams[threadID]
	if ok {
		delete(activeStreams, threadID)
	}
	activeStreamsMu.Unlock()

	if ok {
		cancel()
	}
	return ok
}
