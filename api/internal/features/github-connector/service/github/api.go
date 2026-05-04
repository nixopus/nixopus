package gh

import (
	"github.com/nixopus/nixopus/api/internal/features/github-connector/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// APIBaseURL is the GitHub REST API base (tests may replace temporarily).
var APIBaseURL = "https://api.github.com"

// SetAPIBaseURL updates the GitHub REST base URL (used in tests).
func SetAPIBaseURL(url string) { APIBaseURL = url }

// API holds GitHub App / REST dependencies shared by all connector operations.
type API struct {
	Storage storage.GithubConnectorRepository
	Logger  logger.Logger
}

// NewAPI builds a GitHub API helper bound to storage and logger.
func NewAPI(s storage.GithubConnectorRepository, l logger.Logger) *API {
	return &API{Storage: s, Logger: l}
}
