package mover

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/config"
)

type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateClosed
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateReconnecting:
		return "reconnecting"
	case StateConnected:
		return "connected"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

type ConnectionEvent struct {
	State   ConnectionState
	Error   error
	Attempt int
}

type Client struct {
	serverURL string
	token     string

	conn   *websocket.Conn
	connMu sync.RWMutex

	state   ConnectionState
	stateMu sync.RWMutex

	send    chan SyncMessage
	receive chan SyncMessage
	done    chan struct{}

	reconnectAttempts int64 // Use atomic for thread-safe access
	reconnectMu       sync.Mutex
	reconnectNeeded   chan struct{}
	onStateChange     func(ConnectionEvent)

	pendingMessages []SyncMessage
	pendingMu       sync.Mutex

	wg sync.WaitGroup

	// Track active goroutines to prevent leaks during reconnection
	readLoopRunning  int32 // atomic
	writeLoopRunning int32 // atomic

	// Guard to prevent simultaneous connection attempts
	connectingMu sync.Mutex
	isConnecting bool
}

type ClientOption func(*Client)

func WithOnStateChange(callback func(ConnectionEvent)) ClientOption {
	return func(c *Client) {
		c.onStateChange = callback
	}
}

func NewClient(serverURL, token string, opts ...ClientOption) (*Client, error) {
	c := &Client{
		serverURL:       serverURL,
		token:           token,
		send:            make(chan SyncMessage, sendBufferSize()),
		receive:         make(chan SyncMessage, receiveBufferSize()),
		done:            make(chan struct{}),
		state:           StateDisconnected,
		pendingMessages: make([]SyncMessage, 0),
		reconnectNeeded: make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Start the reconnection manager (runs forever until Close())
	c.wg.Add(1)
	go c.runReconnectionManager()

	// Connect immediately
	if err := c.establishConnection(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Send(msg SyncMessage) error {
	// Check state atomically
	c.stateMu.RLock()
	state := c.state
	c.stateMu.RUnlock()

	if state == StateClosed {
		return fmt.Errorf("client closed")
	}

	// Queue messages if disconnected
	if state == StateReconnecting || state == StateDisconnected {
		c.queueMessage(msg)
		return nil
	}

	// Send immediately if connected
	// Use select to handle both send and done channels atomically
	select {
	case c.send <- msg:
		return nil
	case <-c.done:
		return fmt.Errorf("client closed")
	}
}

func (c *Client) Receive() <-chan SyncMessage {
	return c.receive
}

func (c *Client) State() ConnectionState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

func (c *Client) IsConnected() bool {
	return c.State() == StateConnected
}

func (c *Client) Close() error {
	c.stateMu.Lock()
	if c.state == StateClosed {
		c.stateMu.Unlock()
		return nil
	}
	c.state = StateClosed
	c.stateMu.Unlock()

	// Signal all goroutines to stop
	close(c.done)

	// Close connection (this will cause read/write loops to exit)
	c.closeWebSocketConnection()

	// Close reconnect channel
	close(c.reconnectNeeded)

	// Wait for all goroutines to finish
	c.wg.Wait()

	return nil
}

// establishConnection creates a new WebSocket connection and starts read/write loops
func (c *Client) establishConnection() error {
	// Prevent simultaneous connection attempts
	c.connectingMu.Lock()
	if c.isConnecting {
		c.connectingMu.Unlock()
		return fmt.Errorf("connection attempt already in progress")
	}
	c.isConnecting = true
	c.connectingMu.Unlock()

	defer func() {
		c.connectingMu.Lock()
		c.isConnecting = false
		c.connectingMu.Unlock()
	}()

	c.setState(StateConnecting)

	conn, err := c.dialWebSocket()
	if err != nil {
		c.setState(StateDisconnected)
		return err
	}

	// CRITICAL: Close old connection BEFORE setting new one to prevent races
	oldConn := c.replaceWebSocketConnection(conn)
	if oldConn != nil {
		oldConn.Close()
	}

	c.setState(StateConnected)
	c.resetReconnectAttempts()

	// Start read and write loops only if not already running
	// Use CompareAndSwap to atomically check and set (prevents duplicate goroutines)
	if atomic.CompareAndSwapInt32(&c.readLoopRunning, 0, 1) {
		c.wg.Add(1)
		go c.runReadLoop()
	}

	if atomic.CompareAndSwapInt32(&c.writeLoopRunning, 0, 1) {
		c.wg.Add(1)
		go c.runWriteLoop()
	}

	// Flush queued messages after a brief delay
	go c.flushQueuedMessagesAfterDelay()

	return nil
}

// dialWebSocket creates a new WebSocket connection with auth
func (c *Client) dialWebSocket() (*websocket.Conn, error) {
	// Create custom dialer with increased handshake timeout
	// This allows time for server-side database operations (session verification, app context lookup)
	dialer := &websocket.Dialer{
		HandshakeTimeout:  handshakeTimeout(),
		ReadBufferSize:    256 * 1024,
		WriteBufferSize:   256 * 1024,
		EnableCompression: false,
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)
	if orgID, err := config.GetOrganizationID(); err == nil && orgID != "" {
		header.Set("X-Organization-Id", orgID)
	}
	conn, _, err := dialer.Dial(c.serverURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	conn.SetReadLimit(maxMessageSize())
	return conn, nil
}

// replaceWebSocketConnection atomically replaces connection and returns old one
// Returns old connection so caller can close it safely
func (c *Client) replaceWebSocketConnection(newConn *websocket.Conn) *websocket.Conn {
	c.connMu.Lock()
	oldConn := c.conn
	c.conn = newConn
	c.connMu.Unlock()
	return oldConn
}

// closeWebSocketConnection safely closes the current connection
func (c *Client) closeWebSocketConnection() {
	c.connMu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()
}

// runReconnectionManager listens for reconnection requests and handles them
func (c *Client) runReconnectionManager() {
	defer c.wg.Done()

	for {
		select {
		case <-c.reconnectNeeded:
			c.attemptReconnectionWithBackoff()
		case <-c.done:
			return
		}
	}
}

// attemptReconnectionWithBackoff tries to reconnect with exponential backoff
func (c *Client) attemptReconnectionWithBackoff() {
	c.setState(StateReconnecting)

	for {
		// Check if we should stop
		select {
		case <-c.done:
			return
		default:
		}

		// Increment attempt counter atomically
		attempt := atomic.AddInt64(&c.reconnectAttempts, 1)

		// Check max attempts
		maxAttempts := maxReconnectAttempts()
		if maxAttempts > 0 && int(attempt) > maxAttempts {
			c.emitStateChange(ConnectionEvent{
				State:   StateDisconnected,
				Error:   fmt.Errorf("max reconnection attempts (%d) exceeded", maxAttempts),
				Attempt: int(attempt),
			})
			return
		}

		// Calculate and wait for backoff delay
		delay := c.calculateBackoffDelay(int(attempt))
		c.emitStateChange(ConnectionEvent{
			State:   StateReconnecting,
			Attempt: int(attempt),
		})

		select {
		case <-time.After(delay):
		case <-c.done:
			return
		}

		// Try to reconnect
		if err := c.establishConnection(); err != nil {
			continue
		}

		return
	}
}

// calculateBackoffDelay computes exponential backoff delay
func (c *Client) calculateBackoffDelay(attempt int) time.Duration {
	delay := float64(initialReconnectDelay()) * math.Pow(reconnectBackoffRate(), float64(attempt-1))
	if delay > float64(maxReconnectDelay()) {
		delay = float64(maxReconnectDelay())
	}
	return time.Duration(delay)
}

// requestReconnection signals that reconnection is needed (non-blocking)
func (c *Client) requestReconnection() {
	select {
	case c.reconnectNeeded <- struct{}{}:
	default:
		// Already requested, skip
	}
}

// resetReconnectAttempts resets the attempt counter after successful connection
func (c *Client) resetReconnectAttempts() {
	atomic.StoreInt64(&c.reconnectAttempts, 0)
}

// runReadLoop continuously reads messages from the WebSocket
func (c *Client) runReadLoop() {
	defer func() {
		atomic.StoreInt32(&c.readLoopRunning, 0)
		c.wg.Done()
	}()

	for {
		// Check for shutdown before each operation
		select {
		case <-c.done:
			return
		default:
		}

		// Get connection atomically (may be nil if closed)
		c.connMu.RLock()
		conn := c.conn
		if conn == nil {
			c.connMu.RUnlock()
			return
		}

		// Set read deadline and pong handler while holding lock
		// This ensures pong handler uses the same connection
		conn.SetReadDeadline(time.Now().Add(pongWait()))
		conn.SetPongHandler(func(string) error {
			// Pong handler needs to access connection atomically
			c.connMu.RLock()
			conn := c.conn
			if conn != nil {
				conn.SetReadDeadline(time.Now().Add(pongWait()))
			}
			c.connMu.RUnlock()
			return nil
		})
		c.connMu.RUnlock()

		// Read message (blocks until message or error)
		// Note: Connection may be replaced during read, but ReadMessage will return error
		_, message, err := conn.ReadMessage()
		if err != nil {
			c.handleReadError(err)
			return
		}

		// Parse and forward message
		if !c.parseAndForwardMessage(message) {
			return
		}
	}
}

// handleReadError processes read errors and triggers reconnection
func (c *Client) handleReadError(err error) {
	// Error handled via connection state callback
	c.closeWebSocketConnection()
	if c.shouldAttemptReconnection() {
		c.setState(StateDisconnected)
		c.requestReconnection()
	}
}

// recvEnvelope is the first-pass decode. Payload stays as json.RawMessage
// to avoid double Marshal/Unmarshal when forwarding to typed structs.
type recvEnvelope struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// parseAndForwardMessage parses JSON once and unmarshals Payload into the typed struct.
// Avoids double encode (was: Unmarshal full → Marshal payload → Unmarshal payload).
func (c *Client) parseAndForwardMessage(message []byte) bool {
	var env recvEnvelope
	if err := json.Unmarshal(message, &env); err != nil {
		return true // Continue on parse error
	}

	payload := c.unmarshalPayloadByType(env.Type, env.Payload)
	msg := SyncMessage{Type: env.Type, Timestamp: env.Timestamp, Payload: payload}

	select {
	case c.receive <- msg:
		return true
	case <-c.done:
		return false
	}
}

// unmarshalPayloadByType unmarshals RawMessage into the appropriate typed struct.
func (c *Client) unmarshalPayloadByType(typ MessageType, raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	switch typ {
	case MessageTypeManifest:
		var m ManifestPayload
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil
		}
		return m
	case MessageTypePipelineProgress:
		var m PipelineProgressPayload
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil
		}
		return m
	case MessageTypeBuildStatus:
		var m BuildStatusPayload
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil
		}
		return m
	case MessageTypeBuildLog:
		var m BuildLogPayload
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil
		}
		return m
	case MessageTypeDeploymentStatus:
		var m DeploymentStatusPayload
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil
		}
		return m
	default:
		var v interface{}
		_ = json.Unmarshal(raw, &v)
		return v
	}
}

// runWriteLoop continuously writes messages and sends pings
func (c *Client) runWriteLoop() {
	defer func() {
		atomic.StoreInt32(&c.writeLoopRunning, 0)
		c.wg.Done()
	}()

	// Send initial ping
	if !c.sendPingMessage() {
		return
	}

	ticker := time.NewTicker(pingPeriod())
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.writeCloseMessage()
				return
			}
			if !c.writeTextMessageToConnection(msg) {
				return
			}

		case <-ticker.C:
			if !c.sendPingMessage() {
				return
			}

		case <-c.done:
			return
		}
	}
}

// writeTextMessageToConnection marshals and writes a text message
func (c *Client) writeTextMessageToConnection(msg SyncMessage) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		return true // Skip invalid messages
	}
	return c.writeMessageToConnection(websocket.TextMessage, data)
}

// sendPingMessage sends a ping to keep connection alive
func (c *Client) sendPingMessage() bool {
	return c.writeMessageToConnection(websocket.PingMessage, []byte{})
}

// writeCloseMessage sends a close message to the server
func (c *Client) writeCloseMessage() {
	c.writeMessageToConnection(websocket.CloseMessage, []byte{})
}

// writeMessageToConnection writes a message to the WebSocket (thread-safe)
func (c *Client) writeMessageToConnection(messageType int, data []byte) bool {
	c.connMu.RLock()
	conn := c.conn
	if conn == nil {
		c.connMu.RUnlock()
		return false
	}

	conn.SetWriteDeadline(time.Now().Add(writeWait()))
	err := conn.WriteMessage(messageType, data)
	c.connMu.RUnlock()

	if err != nil {
		c.handleWriteError(err)
		return false
	}

	return true
}

// handleWriteError processes write errors and triggers reconnection
func (c *Client) handleWriteError(err error) {
	// Error handled via connection state callback

	c.closeWebSocketConnection()
	if c.shouldAttemptReconnection() {
		c.setState(StateDisconnected)
		c.requestReconnection()
	}
}

// queueMessage adds a message to the pending queue
func (c *Client) queueMessage(msg SyncMessage) {
	c.pendingMu.Lock()
	c.pendingMessages = append(c.pendingMessages, msg)
	c.pendingMu.Unlock()
}

// flushQueuedMessagesAfterDelay waits briefly then flushes queued messages
func (c *Client) flushQueuedMessagesAfterDelay() {
	time.Sleep(closeFlushDelay())
	c.flushQueuedMessages()
}

// flushQueuedMessages sends all queued messages to the send channel
func (c *Client) flushQueuedMessages() {
	c.pendingMu.Lock()
	messages := c.pendingMessages
	c.pendingMessages = make([]SyncMessage, 0)
	c.pendingMu.Unlock()

	for _, msg := range messages {
		select {
		case c.send <- msg:
		case <-c.done:
			return
		}
	}
}

func (c *Client) shouldAttemptReconnection() bool {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state != StateClosed && c.state != StateReconnecting
}

func (c *Client) setState(state ConnectionState) {
	c.stateMu.Lock()
	c.state = state
	c.stateMu.Unlock()
	c.emitStateChange(ConnectionEvent{State: state})
}

func (c *Client) emitStateChange(event ConnectionEvent) {
	if c.onStateChange != nil {
		c.onStateChange(event)
	}
}
