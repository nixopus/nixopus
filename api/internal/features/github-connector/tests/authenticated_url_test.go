package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func TestCreateAuthenticatedRepoURL(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		accessToken string
		expectedURL string
		expectedErr string
	}{
		{
			name:        "Valid HTTPS repository URL",
			repoURL:     "https://github.com/user/repo",
			accessToken: "token",
			expectedURL: "https://oauth2:token@github.com/user/repo",
			expectedErr: "",
		},
		{
			name:        "Valid SSH repository URL",
			repoURL:     "git@github.com:user/repo.git",
			accessToken: "token",
			expectedURL: "https://oauth2:token@github.com/user/repo.git",
			expectedErr: "",
		},
		{
			name:        "Invalid repository URL format",
			repoURL:     "invalid-url",
			accessToken: "token",
			expectedURL: "",
			expectedErr: "unsupported repository URL format",
		},
		{
			name:        "Unsupported repository URL format",
			repoURL:     "ftp://github.com/user/repo",
			accessToken: "token",
			expectedURL: "",
			expectedErr: "unsupported repository URL format",
		},
		{
			name:        "Empty repository URL",
			repoURL:     "",
			accessToken: "token",
			expectedURL: "",
			expectedErr: "unsupported repository URL format",
		},
		{
			name:        "Empty access token",
			repoURL:     "https://github.com/user/repo",
			accessToken: "",
			expectedURL: "https://oauth2:@github.com/user/repo",
			expectedErr: "",
		},
	}

	mockStorage := NewMockGithubConnectorStorage()
	s := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockStorage)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualURL, actualErr := s.CreateAuthenticatedRepoURL(test.repoURL, test.accessToken)

			if actualURL != test.expectedURL {
				t.Errorf("expected URL %q, got %q", test.expectedURL, actualURL)
			}

			if test.expectedErr == "" && actualErr != nil {
				t.Errorf("expected no error, got %v", actualErr)
			} else if test.expectedErr != "" && actualErr == nil {
				t.Errorf("expected error %q, got nil", test.expectedErr)
			} else if test.expectedErr != "" && actualErr != nil && actualErr.Error() != test.expectedErr {
				t.Errorf("expected error %q, got %q", test.expectedErr, actualErr.Error())
			}
		})
	}
}

var ErrUnsupportedRepoURLFormat = errors.New("unsupported repository URL format")
