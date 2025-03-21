package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type TermSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

type TerminalMessage struct {
	Type string    `json:"type"`
	Data string    `json:"data,omitempty"`
	Size *TermSize `json:"size,omitempty"`
}

type Terminal struct {
	pty        *os.File
	cmd        *exec.Cmd
	conn       *websocket.Conn
	done       chan struct{}
	outputBuf  []byte
	bufferTime time.Duration
	bufferTick *time.Ticker
	log        logger.Logger
	wsLock     sync.Mutex
}

func NewTerminal(conn *websocket.Conn, log *logger.Logger) (*Terminal, error) {
	cmd := exec.Command("/bin/bash")

	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	setWinsize(ptmx, 24, 80)

	terminal := &Terminal{
		pty:        ptmx,
		cmd:        cmd,
		conn:       conn,
		done:       make(chan struct{}),
		outputBuf:  make([]byte, 0, 4096),
		bufferTime: 10 * time.Millisecond,
		log:        logger.NewLogger(),
	}

	terminal.bufferTick = time.NewTicker(terminal.bufferTime)
	terminal.log.Log(logger.Info, "Terminal created", terminal.cmd.Dir)
	return terminal, nil
}

func (t *Terminal) Start() {
	go t.readPtyOutput()
	go t.bufferFlusher()
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
		err := t.conn.WriteJSON(map[string]interface{}{
			"data": map[string]interface{}{
				"output_type": "stdout",
				"content":     string(t.outputBuf),
			},
		})

		if err != nil {
			t.log.Log("error writing websocket message", err.Error(), "")
			close(t.done)
		}

		t.outputBuf = t.outputBuf[:0]
	}
}

func (t *Terminal) readPtyOutput() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-t.done:
			return
		default:
			n, err := t.pty.Read(buf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("pty closed")
					continue
				}
				t.log.Log("error reading from pty", err.Error(), "")
				close(t.done)
				return
			}

			func() {
				t.wsLock.Lock()
				defer t.wsLock.Unlock()
				t.outputBuf = append(t.outputBuf, buf[:n]...)
			}()
			
			t.wsLock.Lock()
			err = t.conn.WriteMessage(websocket.TextMessage, buf[:n])
			t.wsLock.Unlock()
			
			if err != nil {
				t.log.Log("error writing to websocket", err.Error(), "")
				close(t.done)
				return
			}
		}
	}
}

func (t *Terminal) Close() error {
	close(t.done)

	if t.bufferTick != nil {
		t.bufferTick.Stop()
	}

	t.flushBuffer()

	if err := t.pty.Close(); err != nil {
		return err
	}

	if err := t.cmd.Process.Kill(); err != nil {
		return err
	}

	t.wsLock.Lock()
	err := t.conn.Close()
	t.wsLock.Unlock()
	
	if err != nil {
		t.log.Log("error closing websocket connection", err.Error(), "")
	}

	return nil
}

func (t *Terminal) WriteMessage(message string) error {
	_, err := t.pty.Write([]byte(message))
	return err
}

func setWinsize(f *os.File, rows, cols uint16) {
	ws := &struct {
		Rows uint16
		Cols uint16
		X    uint16
		Y    uint16
	}{rows, cols, 0, 0}

	syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)
}