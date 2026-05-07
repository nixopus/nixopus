package llm

import "context"

type Provider interface {
	Complete(ctx context.Context, params CompletionParams) (*Response, error)
	Stream(ctx context.Context, params CompletionParams) (*StreamIterator, error)
}

type StreamIterator struct {
	ch     <-chan StreamEvent
	cancel context.CancelFunc
}

type StreamEventType int

const (
	EventChunk StreamEventType = iota
	EventDone
	EventError
)

type StreamEvent struct {
	Type  StreamEventType
	Chunk *StreamChunk
	Err   error
}

func newStreamIterator(ch <-chan StreamEvent, cancel context.CancelFunc) *StreamIterator {
	return &StreamIterator{ch: ch, cancel: cancel}
}

func (s *StreamIterator) Next() (StreamEvent, bool) {
	event, ok := <-s.ch
	return event, ok
}

func (s *StreamIterator) Close() {
	s.cancel()
	for range s.ch {
	}
}
