package terminal

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Terminal struct {
	pty     *os.File
	cmd     *exec.Cmd
	conn    *websocket.Conn
	writeMu sync.Mutex
}

// NewTerminal creates a new WebSocket-enabled terminal
func NewTerminal(conn *websocket.Conn) (*Terminal, error) {
	// Start bash command
	cmd := exec.Command("/bin/bash")

	// Start the command with a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &Terminal{
		pty:  ptmx,
		cmd:  cmd,
		conn: conn,
	}, nil
}

// Start begins the terminal session
func (t *Terminal) Start() {
	// Read from the pty and write to websocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := t.pty.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				// Log error and terminate
				return
			}

			t.writeMu.Lock()
			err = t.conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			t.writeMu.Unlock()

			if err != nil {
				return
			}
		}
	}()

	// Read from websocket and write to the pty
	go func() {
		for {
			messageType, p, err := t.conn.ReadMessage()
			if err != nil {
				return
			}

			if messageType == websocket.TextMessage {
				var msg shared_types.Payload

				if err := json.Unmarshal(p, &msg); err != nil {
					return
				}

				if msg.Action == "terminal" {
					var dataObj struct {
						Type string `json:"type"`
						Rows int    `json:"rows"`
						Cols int    `json:"cols"`
					}

					if err := json.Unmarshal([]byte(msg.Data.(string)), &dataObj); err == nil {
						if dataObj.Type == "resize" {
							t.resize(dataObj.Rows, dataObj.Cols)
						} else {
							if data, ok := msg.Data.(string); ok {
								_, err = t.pty.Write([]byte(data))
								if err != nil {
									return
								}
							}
						}
					}
				}
			} else if messageType == websocket.BinaryMessage {
				_, err = t.pty.Write(p)
				if err != nil {
					return
				}
			}
		}
	}()
}

// resize changes the size of the terminal
func (t *Terminal) resize(rows, cols int) error {
	// Define the winsize struct
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{
		Row:    uint16(rows),
		Col:    uint16(cols),
		Xpixel: 0,
		Ypixel: 0,
	}

	// Set the window size
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		t.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)

	if errno != 0 {
		return errno
	}

	return nil
}

// Close cleans up resources
func (t *Terminal) Close() error {
	if err := t.pty.Close(); err != nil {
		return err
	}

	// Kill the process if it's still running
	if t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}

	return nil
}
