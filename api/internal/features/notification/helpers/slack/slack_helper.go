package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackManager struct {
}

func NewSlackManager() *SlackManager {
	return &SlackManager{}
}

type SlackMessage struct {
	Text string `json:"text"`
}

func (m *SlackManager) SendNotification(message string) error {
	slackMsg := SlackMessage{
		Text: message,
	}

	jsonData, err := json.Marshal(slackMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}
	resp, err := http.Post("", "application/json", bytes.NewBuffer(jsonData)) // TODO: add webhook url
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
