package terminal

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/melbahja/goph"
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
	ssh        *sshpkg.SSH
	conn       *websocket.Conn
	done       chan struct{}
	outputBuf  []byte
	bufferTime time.Duration
	bufferTick *time.Ticker
	log        logger.Logger
	wsLock     sync.Mutex

	client  *goph.Client
	session *ssh.Session
	stdin   io.WriteCloser

	TerminalId string
}

func NewTerminal(conn *websocket.Conn, log *logger.Logger, terminalId string) (*Terminal, error) {
	ssh_client := sshpkg.NewSSH()
	terminal := &Terminal{
		ssh:        ssh_client,
		conn:       conn,
		done:       make(chan struct{}),
		outputBuf:  make([]byte, 0, 4096),
		bufferTime: 10 * time.Millisecond,
		log:        *log,
		TerminalId: terminalId,
	}

	terminal.bufferTick = time.NewTicker(terminal.bufferTime)
	terminal.log.Log(logger.Info, "Terminal created", ssh_client.Host)
	return terminal, nil
}

func (t *Terminal) Start() {
	go t.bufferFlusher()

	go func() {
		client, err := t.ssh.Connect()
		if err != nil {
			t.log.Log(logger.Error, "Failed to connect to SSH", err.Error())
			close(t.done)
			return
		}
		t.client = client

		session, err := client.NewSession()
		if err != nil {
			t.log.Log(logger.Error, "Failed to create session", err.Error())
			client.Close()
			close(t.done)
			return
		}
		t.session = session
		defer session.Close()

		stdin, err := session.StdinPipe()
		if err != nil {
			t.log.Log(logger.Error, "Failed to get stdin pipe", err.Error())
			close(t.done)
			return
		}
		t.stdin = stdin

		stdout, err := session.StdoutPipe()
		if err != nil {
			t.log.Log(logger.Error, "Failed to get stdout pipe", err.Error())
			close(t.done)
			return
		}

		stderr, err := session.StderrPipe()
		if err != nil {
			t.log.Log(logger.Error, "Failed to get stderr pipe", err.Error())
			close(t.done)
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
			t.log.Log(logger.Error, "Failed to request PTY", err.Error())
			close(t.done)
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
				t.log.Log(logger.Info, fmt.Sprintf("Failed to set environment variable %s: %s", env, err.Error()), "")
			}
		}

		if err = session.Shell(); err != nil {
			t.log.Log(logger.Error, "Failed to start shell", err.Error())
			close(t.done)
			return
		}

		session.Wait()
	}()
}

func (t *Terminal) readOutput(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-t.done:
			return
		default:
			n, err := r.Read(buf)
			if err != nil {
				if err == io.EOF {
					continue
				}
				t.log.Log(logger.Error, "Error reading from SSH", err.Error())
				return
			}

			func() {
				t.wsLock.Lock()
				defer t.wsLock.Unlock()
				t.outputBuf = append(t.outputBuf, buf[:n]...)
			}()

			msg := TerminalMessage{
				TerminalId: t.TerminalId,
				Type:       "stdout",
				Data:       string(buf[:n]),
			}
			t.wsLock.Lock()
			err = t.conn.WriteJSON(msg)
			t.wsLock.Unlock()

			if err != nil {
				t.log.Log(logger.Error, "Error writing to websocket", err.Error())
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
	select {
	case <-t.done:
	default:
		close(t.done)
	}

	if t.bufferTick != nil {
		t.bufferTick.Stop()
	}

	t.flushBuffer()

	if t.session != nil {
		t.session.Close()
	}

	if t.client != nil {
		t.client.Close()
	}

	t.wsLock.Lock()
	err := t.conn.Close()
	t.wsLock.Unlock()

	if err != nil {
		t.log.Log(logger.Error, "Error closing websocket connection", err.Error())
	}

	return nil
}

func (t *Terminal) WriteMessage(message string) error {
	if t.stdin == nil {
		return fmt.Errorf("terminal not started or already closed")
	}

	_, err := t.stdin.Write([]byte(message))
	return err
}

func (t *Terminal) ResizeTerminal(rows, cols uint16) error {
	if t.session == nil {
		return fmt.Errorf("terminal not started or already closed")
	}

	return t.session.WindowChange(int(rows), int(cols))
}
