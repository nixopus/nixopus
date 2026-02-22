package live

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// Terminal is a thread-safe stdout wrapper with TTY detection.
// All output in the agent UI goes through here to prevent garbled concurrent writes.
type Terminal struct {
	mu    sync.Mutex
	out   io.Writer
	isTTY bool
	width int
}

// NewTerminal creates a terminal writing to stdout.
func NewTerminal() *Terminal {
	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	w := 80
	if tty {
		if tw, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && tw > 0 {
			w = tw
		}
	}
	return &Terminal{
		out:   os.Stdout,
		isTTY: tty,
		width: w,
	}
}

// IsTTY returns true when stdout is an interactive terminal.
func (t *Terminal) IsTTY() bool { return t.isTTY }

// Width returns the terminal column count.
func (t *Terminal) Width() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.isTTY {
		if tw, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && tw > 0 {
			t.width = tw
		}
	}
	return t.width
}

// Write implements io.Writer with mutex protection.
func (t *Terminal) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.out.Write(p)
}

// Println writes a full line with newline.
func (t *Terminal) Println(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprintln(t.out, text)
}

// Print writes text without a trailing newline.
func (t *Terminal) Print(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(t.out, text)
}

// OverwriteLine replaces the current line in-place (for progress updates).
// When not a TTY, this is a no-op — progress lines are only shown on completion.
func (t *Terminal) OverwriteLine(text string) {
	if !t.isTTY {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprintf(t.out, "\r\033[K%s", text)
}

// FinishLine replaces the current in-place line with a final version + newline.
// When not a TTY, just prints the line normally.
func (t *Terminal) FinishLine(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.isTTY {
		fmt.Fprintf(t.out, "\r\033[K%s\n", text)
	} else {
		fmt.Fprintln(t.out, text)
	}
}
