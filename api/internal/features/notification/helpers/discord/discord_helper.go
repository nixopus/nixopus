package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type DiscordManager struct {
}

func NewDiscordManager() *DiscordManager {
	return &DiscordManager{}
}

type DiscordMessage struct {
	Content string `json:"content"`
}

func (m *DiscordManager) SendNotification(message string, webhookURL string) error {
	discordMsg := DiscordMessage{
		Content: message,
	}

	jsonData, err := json.Marshal(discordMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord webhook returned non-200 status code: %d", resp.StatusCode)
	}

	log.Printf("Discord notification sent successfully")
	return nil
}
