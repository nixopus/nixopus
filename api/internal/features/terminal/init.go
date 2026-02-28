package terminal

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"golang.org/x/crypto/ssh"
)

type TermSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

type TerminalMessage struct {
	TerminalId string    `json:"terminal_id"`
	Type       string    `json:"type"`
	Data       string    `json:"data,omitempty"`
	Size       *TermSize `json:"size,omitempty"`
}

type Terminal struct {
	sshManager *sshpkg.SSHManager
	conn       *websocket.Conn
	done       chan struct{}
	doneOnce   sync.Once     // ensures done is closed exactly once
	ready      chan struct{} // closed when stdin/session are wired up
	readyOnce  sync.Once
	outputBuf  []byte
	bufferTime time.Duration
	bufferTick *time.Ticker
	log        logger.Logger
	wsLock     *sync.Mutex // shared per-connection write mutex

	session   *ssh.Session
	stdin     io.WriteCloser
	release   func() // decrements pool inUse counter; called on Close
	startedAt time.Time

	TerminalId string
}

// signalDone closes the done channel exactly once, regardless of which
// goroutine calls it first (readOutput, session.Wait, or Close).
func (t *Terminal) signalDone() {
	t.doneOnce.Do(func() {
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] signalDone: closing done channel (uptime=%s)", t.TerminalId, time.Since(t.startedAt).Round(time.Second)), "")
		close(t.done)
	})
}

// IsDone returns true if the terminal session has ended.
func (t *Terminal) IsDone() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

// signalReady closes the ready channel, unblocking any WriteMessage or
// ResizeTerminal calls that arrived before Start() finished setup.
func (t *Terminal) signalReady() {
	t.readyOnce.Do(func() { close(t.ready) })
}

const readyTimeout = 10 * time.Second

// waitReady blocks until the terminal is ready (stdin wired up) or the
// terminal is shutting down. Returns false if the terminal died before
// becoming ready or the timeout elapsed.
func (t *Terminal) waitReady() bool {
	select {
	case <-t.ready:
		return true
	case <-t.done:
		return false
	case <-time.After(readyTimeout):
		t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] waitReady: timed out after %s", t.TerminalId, readyTimeout), "")
		return false
	}
}

func NewTerminal(ctx context.Context, conn *websocket.Conn, wsMu *sync.Mutex, log *logger.Logger, terminalId string) (*Terminal, error) {
	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}
	sshClient, err := sshManager.GetDefaultSSH()
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH client: %w", err)
	}
	terminal := &Terminal{
		sshManager: sshManager,
		conn:       conn,
		done:       make(chan struct{}),
		ready:      make(chan struct{}),
		outputBuf:  make([]byte, 0, 4096),
		bufferTime: 10 * time.Millisecond,
		log:        *log,
		wsLock:     wsMu,
		TerminalId: terminalId,
		startedAt:  time.Now(),
	}

	terminal.bufferTick = time.NewTicker(terminal.bufferTime)
	terminal.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] created, host=%s", terminalId, sshClient.Host), "")
	return terminal, nil
}

func (t *Terminal) Start() {
	go t.bufferFlusher()

	go func() {
		tid := t.TerminalId
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] start: borrowing SSH connection", tid), "")

		const maxRetries = 2
		var session *ssh.Session

		for attempt := 0; attempt < maxRetries; attempt++ {
			client, release, err := t.sshManager.Borrow("")
			if err != nil {
				t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] borrow failed (attempt %d/%d): %s", tid, attempt+1, maxRetries, err.Error()), "")
				t.signalDone()
				return
			}
			t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] borrowed SSH connection (attempt %d/%d)", tid, attempt+1, maxRetries), "")

			sess, err := client.NewSession()
			if err != nil {
				release()
				if sshpkg.IsClosedConnectionError(err) && attempt < maxRetries-1 {
					t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] stale connection, closing and retrying: %s", tid, err.Error()), "")
					t.sshManager.CloseConnection("")
					continue
				}
				t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] session creation failed: %s", tid, err.Error()), "")
				t.signalDone()
				return
			}

			t.release = release
			session = sess
			break
		}

		if session == nil {
			t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] failed to create session after %d retries", tid, maxRetries), "")
			t.signalDone()
			return
		}

		t.session = session
		defer session.Close()
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] SSH session created, setting up pipes", tid), "")

		stdin, err := session.StdinPipe()
		if err != nil {
			t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] stdin pipe failed: %s", tid, err.Error()), "")
			t.signalDone()
			return
		}
		t.stdin = stdin

		stdout, err := session.StdoutPipe()
		if err != nil {
			t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] stdout pipe failed: %s", tid, err.Error()), "")
			t.signalDone()
			return
		}

		stderr, err := session.StderrPipe()
		if err != nil {
			t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] stderr pipe failed: %s", tid, err.Error()), "")
			t.signalDone()
			return
		}

		go t.readOutput(stdout)
		go t.readOutput(stderr)

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
			ssh.ICANON:        1,
			ssh.ISIG:          1,
			ssh.ICRNL:         1,
		}

		if err = session.RequestPty("xterm-256color", 40, 100, modes); err != nil {
			t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] PTY request failed: %s", tid, err.Error()), "")
			t.signalDone()
			return
		}

		envVars := []string{
			"TERM=xterm-256color",
			"COLORTERM=truecolor",
			"LANG=en_US.UTF-8",
			"LC_ALL=en_US.UTF-8",
		}

		for _, env := range envVars {
			if err := session.Setenv(strings.Split(env, "=")[0], strings.Split(env, "=")[1]); err != nil {
				t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] setenv %s failed (non-fatal): %s", tid, strings.Split(env, "=")[0], err.Error()), "")
			}
		}

		if err = session.Shell(); err != nil {
			t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] shell start failed: %s", tid, err.Error()), "")
			t.signalDone()
			return
		}

		t.signalReady()
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] shell active, ready for input, waiting for session to end", tid), "")
		session.Wait()
		uptime := time.Since(t.startedAt).Round(time.Second)
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] session.Wait() returned, terminal lived %s", tid, uptime), "")

		t.notifyExit()
		t.signalDone()
	}()
}

// notifyExit sends a terminal_exit message over the WebSocket so the frontend
// knows the session ended and can prompt a reconnection.
func (t *Terminal) notifyExit() {
	msg := TerminalMessage{
		TerminalId: t.TerminalId,
		Type:       "exit",
		Data:       "session ended",
	}
	t.wsLock.Lock()
	err := t.conn.WriteJSON(msg)
	t.wsLock.Unlock()
	if err != nil {
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] notifyExit: failed to send exit message: %s", t.TerminalId, err.Error()), "")
	} else {
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] notifyExit: exit message sent to frontend", t.TerminalId), "")
	}
}

func (t *Terminal) readOutput(r io.Reader) {
	tid := t.TerminalId
	buf := make([]byte, 4096)
	totalBytes := 0
	reads := 0
	lastLogAt := time.Now()

	for {
		select {
		case <-t.done:
			t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] readOutput: done signal received, totalBytes=%d reads=%d uptime=%s", tid, totalBytes, reads, time.Since(t.startedAt).Round(time.Second)), "")
			return
		default:
			n, err := r.Read(buf)
			if n > 0 {
				reads++
				totalBytes += n
				if time.Since(lastLogAt) > 30*time.Second {
					t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] readOutput: alive, totalBytes=%d reads=%d uptime=%s", tid, totalBytes, reads, time.Since(t.startedAt).Round(time.Second)), "")
					lastLogAt = time.Now()
				}

				msg := TerminalMessage{
					TerminalId: t.TerminalId,
					Type:       "stdout",
					Data:       string(buf[:n]),
				}
				t.wsLock.Lock()
				writeErr := t.conn.WriteJSON(msg)
				t.wsLock.Unlock()

				if writeErr != nil {
					t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] readOutput: websocket write failed: %s (totalBytes=%d uptime=%s)", tid, writeErr.Error(), totalBytes, time.Since(t.startedAt).Round(time.Second)), "")
					t.signalDone()
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] readOutput: SSH read error: %s (totalBytes=%d uptime=%s)", tid, err.Error(), totalBytes, time.Since(t.startedAt).Round(time.Second)), "")
				} else {
					t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] readOutput: EOF (totalBytes=%d reads=%d uptime=%s)", tid, totalBytes, reads, time.Since(t.startedAt).Round(time.Second)), "")
				}
				t.signalDone()
				return
			}
		}
	}
}

func (t *Terminal) bufferFlusher() {
	for {
		select {
		case <-t.done:
			if t.bufferTick != nil {
				t.bufferTick.Stop()
			}
			return
		case <-t.bufferTick.C:
			t.flushBuffer()
		}
	}
}

func (t *Terminal) flushBuffer() {
	t.wsLock.Lock()
	defer t.wsLock.Unlock()

	if len(t.outputBuf) > 0 {
		msg := TerminalMessage{
			TerminalId: t.TerminalId,
			Type:       "stdout",
			Data:       string(t.outputBuf),
		}
		err := t.conn.WriteJSON(msg)
		if err != nil {
			t.log.Log(logger.Error, "Error writing websocket message", err.Error())
		}
		t.outputBuf = t.outputBuf[:0]
	}
}

func (t *Terminal) Close() error {
	tid := t.TerminalId
	t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] Close() called, uptime=%s", tid, time.Since(t.startedAt).Round(time.Second)), "")
	t.signalDone()

	if t.bufferTick != nil {
		t.bufferTick.Stop()
	}

	t.flushBuffer()

	if t.session != nil {
		if err := t.session.Close(); err != nil {
			t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] session.Close error (expected if already ended): %s", tid, err.Error()), "")
		}
	}

	if t.release != nil {
		t.release()
		t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] pool borrow released", tid), "")
	}

	t.log.Log(logger.Info, fmt.Sprintf("[terminal:%s] closed cleanly", tid), "")
	return nil
}

func (t *Terminal) WriteMessage(message string) error {
	if !t.waitReady() {
		t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] WriteMessage: terminal never became ready", t.TerminalId), "")
		return fmt.Errorf("terminal not ready or already closed")
	}

	_, err := t.stdin.Write([]byte(message))
	if err != nil {
		t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] WriteMessage: stdin write failed: %s (uptime=%s)", t.TerminalId, err.Error(), time.Since(t.startedAt).Round(time.Second)), "")
	}
	return err
}

func (t *Terminal) ResizeTerminal(rows, cols uint16) error {
	if !t.waitReady() {
		t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] ResizeTerminal: terminal never became ready", t.TerminalId), "")
		return fmt.Errorf("terminal not ready or already closed")
	}

	err := t.session.WindowChange(int(rows), int(cols))
	if err != nil {
		t.log.Log(logger.Error, fmt.Sprintf("[terminal:%s] ResizeTerminal: WindowChange(%d,%d) failed: %s", t.TerminalId, rows, cols, err.Error()), "")
	}
	return err
}
