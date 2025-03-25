package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessValidator_ParseRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           map[string]interface{}
		expectedResult error
	}{
		{
			name:           "Invalid Resource",
			method:         http.MethodGet,
			url:            "/api/v1/notification/invalid",
			expectedResult: notification.ErrInvalidResource,
		},
		{
			name:           "Valid Preferences Resource",
			method:         http.MethodGet,
			url:            "/api/v1/notification/preferences",
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockNotificationStorage)
			validator := validation.NewValidator(mockStorage)

			// Add this to handle the "smtp" case that's causing issues
			// This mock will handle any potential call to GetSmtp with "smtp"
			mockStorage.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()

			var reqBody []byte
			var err error

			if tt.body != nil {
				reqBody, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(reqBody))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			user := &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleViewer,
			}

			result := validator.AccessValidator(w, req, user)
			assert.Equal(t, tt.expectedResult, result)

			mockStorage.AssertExpectations(t)
		})
	}
}
