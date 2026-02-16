package live

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultAgentEndpoint = "http://localhost:4117"
	agentID              = "deployAgent"
	streamPath           = "/api/agents/" + agentID + "/stream"
	httpTimeout          = 120 * time.Second
)

// AgentClient communicates with the remote Mastra deploy agent via HTTP streaming.
type AgentClient struct {
	endpoint    string
	accessToken string
	threadID    string
	httpClient  *http.Client
}

// AgentMessage is a message in the Mastra conversation format.
type AgentMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// agentStreamRequest is the POST body for /api/agents/:agentId/stream.
type agentStreamRequest struct {
	Messages   []AgentMessage `json:"messages"`
	ThreadID   string         `json:"threadId,omitempty"`
	ResourceID string         `json:"resourceId,omitempty"`
}

// NewAgentClient creates a client for the deploy agent.
func NewAgentClient(accessToken string, threadID string) *AgentClient {
	endpoint := os.Getenv("AGENT_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultAgentEndpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")

	return &AgentClient{
		endpoint:    endpoint,
		accessToken: accessToken,
		threadID:    threadID,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// Stream sends messages to the deploy agent and returns a channel of text chunks.
// The channel is closed when the stream ends. Errors are sent as a final chunk
// prefixed with "[error]".
func (c *AgentClient) Stream(ctx context.Context, messages []AgentMessage) (<-chan string, error) {
	body := agentStreamRequest{
		Messages:   messages,
		ThreadID:   c.threadID,
		ResourceID: "cli",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.endpoint + streamPath
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent returned %d: %s", resp.StatusCode, string(errBody))
	}

	ch := make(chan string, 32)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		c.readMastraStream(ctx, resp.Body, ch)
	}()

	return ch, nil
}

// mastraSSEEvent represents a parsed SSE event from the Mastra agent stream.
// The stream format is:
//
//	data: {"type":"text-delta","from":"AGENT","payload":{"text":"chunk"}}
//	data: {"type":"finish",...}
//	data: [DONE]
type mastraSSEEvent struct {
	Type    string          `json:"type"`
	From    string          `json:"from"`
	Payload json.RawMessage `json:"payload"`
}

type textDeltaPayload struct {
	Text string `json:"text"`
}

// readMastraStream parses the Mastra agent SSE stream.
// Each line is `data: <json>` or `data: [DONE]`.
// We extract text from "text-delta" events and send them to the channel.
func (c *AgentClient) readMastraStream(ctx context.Context, body io.Reader, ch chan<- string) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		// SSE lines start with "data: "
		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		data = strings.TrimPrefix(data, "data:")
		data = strings.TrimSpace(data)

		if data == "" || data == "[DONE]" {
			if data == "[DONE]" {
				return
			}
			continue
		}

		var event mastraSSEEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "text-delta":
			var payload textDeltaPayload
			if err := json.Unmarshal(event.Payload, &payload); err == nil && payload.Text != "" {
				select {
				case ch <- payload.Text:
				case <-ctx.Done():
					return
				}
			}
		case "error":
			var payload struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err == nil {
				errMsg := payload.Message
				if errMsg == "" {
					errMsg = payload.Error
				}
				if errMsg != "" {
					select {
					case ch <- fmt.Sprintf("[error] %s", errMsg):
					case <-ctx.Done():
						return
					}
				}
			}
		case "finish":
			return
		}
	}
}
