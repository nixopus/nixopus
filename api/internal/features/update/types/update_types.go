package types

import "time"

type UpdateCheckResponse struct {
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	LastChecked     time.Time `json:"last_checked"`
	Environment     string    `json:"environment"`
}

type UpdateRequest struct {
	Force bool `json:"force"`
}

type UpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
