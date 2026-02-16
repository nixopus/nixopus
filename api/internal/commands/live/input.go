package live

import (
	"bufio"
	"context"
	"os"
)

// InputReader reads lines from stdin in a background goroutine
// and sends them to a channel. The caller consumes via Chan().
// When stdout is not a TTY, Start is a no-op (non-interactive mode).
type InputReader struct {
	ch    chan string
	isTTY bool
}

// NewInputReader creates an input reader.
// Pass isTTY=false to disable input (e.g. when piped).
func NewInputReader(isTTY bool) *InputReader {
	return &InputReader{
		ch:    make(chan string, 1),
		isTTY: isTTY,
	}
}

// Start begins reading stdin in a goroutine. Returns immediately.
// Stops when ctx is cancelled or stdin closes.
// When not a TTY, this is a no-op.
func (r *InputReader) Start(ctx context.Context) {
	if !r.isTTY {
		return
	}
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case r.ch <- scanner.Text():
			}
		}
	}()
}

// Chan returns the channel that receives user input lines.
func (r *InputReader) Chan() <-chan string {
	return r.ch
}
