package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	// Optional headers added to every request (e.g. HTTP-Referer for OpenRouter).
	Headers map[string]string
	// HTTP client override; uses a default with 120s timeout if nil.
	HTTPClient *http.Client
}

type openaiProvider struct {
	config OpenAIConfig
	client *http.Client
}

func NewOpenAIProvider(cfg OpenAIConfig) Provider {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &openaiProvider{config: cfg, client: client}
}

func (p *openaiProvider) Complete(ctx context.Context, params CompletionParams) (*Response, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm: create request: %w", err)
	}
	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, p.parseError(resp)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("llm: decode response: %w", err)
	}
	return &result, nil
}

func (p *openaiProvider) Stream(ctx context.Context, params CompletionParams) (*StreamIterator, error) {
	streamReq := struct {
		CompletionParams
		Stream        bool                   `json:"stream"`
		StreamOptions map[string]interface{} `json:"stream_options,omitempty"`
	}{
		CompletionParams: params,
		Stream:           true,
		StreamOptions:    map[string]interface{}{"include_usage": true},
	}

	body, err := json.Marshal(streamReq)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal stream request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm: create stream request: %w", err)
	}
	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: stream request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, p.parseError(resp)
	}

	streamCtx, cancel := context.WithCancel(ctx)
	ch := make(chan StreamEvent, 16)

	go p.readSSE(streamCtx, resp.Body, ch)

	return newStreamIterator(ch, cancel), nil
}

func (p *openaiProvider) readSSE(ctx context.Context, body io.ReadCloser, ch chan<- StreamEvent) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- StreamEvent{Type: EventDone}
			return
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamEvent{Type: EventError, Err: fmt.Errorf("llm: decode chunk: %w", err)}
			return
		}

		ch <- StreamEvent{Type: EventChunk, Chunk: &chunk}
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{Type: EventError, Err: fmt.Errorf("llm: read stream: %w", err)}
	}
}

func (p *openaiProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	for k, v := range p.config.Headers {
		req.Header.Set(k, v)
	}
}

func (p *openaiProvider) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	var apiErr struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		return &ProviderError{
			StatusCode: resp.StatusCode,
			Type:       apiErr.Error.Type,
			Code:       apiErr.Error.Code,
			Message:    apiErr.Error.Message,
		}
	}

	return &ProviderError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}

type ProviderError struct {
	StatusCode int
	Type       string
	Code       string
	Message    string
}

func (e *ProviderError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("llm: api error %d [%s/%s]: %s", e.StatusCode, e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("llm: api error %d: %s", e.StatusCode, e.Message)
}
