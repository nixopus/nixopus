package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AgentChannel struct {
	webhookURL   string
	tokenURL     string
	clientID     string
	clientSecret string

	mu          sync.RWMutex
	cachedToken string
	tokenExpiry time.Time
	httpClient  *http.Client
}

type AgentEventPayload struct {
	EventID        string                 `json:"event_id"`
	Source         string                 `json:"source"`
	EventType      string                 `json:"event_type"`
	Timestamp      string                 `json:"timestamp"`
	OrganizationID string                 `json:"organization_id"`
	UserID         string                 `json:"user_id"`
	Payload        map[string]interface{} `json:"payload"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewAgentChannel(webhookURL, tokenURL, clientID, clientSecret string) *AgentChannel {
	return &AgentChannel{
		webhookURL:   webhookURL,
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *AgentChannel) Type() string { return "agent" }

func (a *AgentChannel) Send(ctx context.Context, msg Message) error {
	token, err := a.getToken()
	if err != nil {
		return fmt.Errorf("agent channel token acquisition failed: %w", err)
	}

	eventPayload := make(map[string]interface{})
	for k, v := range msg.Metadata {
		if k == "event_type" || k == "organization_id" || k == "user_id" {
			continue
		}
		eventPayload[k] = v
	}
	eventPayload["message"] = msg.Body

	payload := AgentEventPayload{
		EventID:        uuid.New().String(),
		Source:         "nixopus",
		EventType:      msg.Metadata["event_type"],
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		OrganizationID: msg.Metadata["organization_id"],
		UserID:         msg.Metadata["user_id"],
		Payload:        eventPayload,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("agent channel marshal failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("agent channel request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("agent channel webhook POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent channel webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (a *AgentChannel) getToken() (string, error) {
	a.mu.RLock()
	if a.cachedToken != "" && time.Now().Before(a.tokenExpiry) {
		token := a.cachedToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cachedToken != "" && time.Now().Before(a.tokenExpiry) {
		return a.cachedToken, nil
	}

	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", a.clientID)
	params.Set("client_secret", a.clientSecret)
	req, err := http.NewRequest(http.MethodPost, a.tokenURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token endpoint returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	a.cachedToken = tokenResp.AccessToken
	skew := 5 * time.Minute
	if time.Duration(tokenResp.ExpiresIn)*time.Second < skew*2 {
		skew = time.Duration(tokenResp.ExpiresIn) * time.Second / 2
	}
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - skew)

	return a.cachedToken, nil
}
