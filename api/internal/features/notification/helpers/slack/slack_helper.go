package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

func (m *SlackManager) SendNotification(message string, webhookUrl string) error {
	slackMsg := SlackMessage{
		Text: message,
	}

	jsonData, err := json.Marshal(slackMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned non-200 status code: %d", resp.StatusCode)
	}

	log.Printf("Slack notification sent successfully")
	return nil
}
