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

func TestAccessValidator_SMTP(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	smtpUuid := uuid.New()
	smtpID := smtpUuid.String()

	smtpConfig := &shared_types.SMTPConfigs{
		ID:             smtpUuid,
		UserID:         userID,
		OrganizationID: orgID,
	}

	tests := []struct {
		name           string
		method         string
		url            string
		body           map[string]interface{}
		setupMock      func(*MockNotificationStorage)
		user           *shared_types.User
		expectedResult error
	}{
		{
			name:   "Create SMTP - Admin",
			method: http.MethodPost,
			url:    "/api/v1/notification/smtp",
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleAdmin,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
			},
			expectedResult: nil,
		},
		{
			name:   "Create SMTP - Non-Admin (Member)",
			method: http.MethodPost,
			url:    "/api/v1/notification/smtp",
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
			},
			expectedResult: notification.ErrAccessDenied,
		},
		{
			name:   "Create SMTP - Non-Admin (Viewer)",
			method: http.MethodPost,
			url:    "/api/v1/notification/smtp",
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleViewer,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
			},
			expectedResult: notification.ErrAccessDenied,
		},
		{
			name:   "Read SMTP - ID from Query",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   userID,
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: nil,
		},
		{
			name:   "Read SMTP - Admin Access",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(), 
				Type: shared_types.RoleAdmin,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: nil,
		},
		{
			name:   "Read SMTP - Member with Organization Access",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(), 
				Type: shared_types.RoleMember,
				Organizations: []shared_types.Organization{
					{ID: orgID},
				},
				OrganizationUsers: []shared_types.OrganizationUsers{
					{
						OrganizationID: orgID,
						Role: &shared_types.Role{
							Permissions: []shared_types.Permission{
								{Name: "read", Resource: "organization"},
							},
						},
					},
				},
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: nil,
		},
		{
			name:   "Read SMTP - Viewer with Organization Access",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(), 
				Type: shared_types.RoleViewer,
				Organizations: []shared_types.Organization{
					{ID: orgID},
				},
				OrganizationUsers: []shared_types.OrganizationUsers{
					{
						OrganizationID: orgID,
						Role: &shared_types.Role{
							Permissions: []shared_types.Permission{
								{Name: "read", Resource: "organization"},
							},
						},
					},
				},
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: nil,
		},
		{
			name:   "Read SMTP - Member Not in Organization",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleMember,
				Organizations: []shared_types.Organization{
					{ID: uuid.New()}, 
				},
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: notification.ErrUserDoesNotBelongToOrganization,
		},
		{
			name:   "Read SMTP - Member in Org but No Permission",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleMember,
				Organizations: []shared_types.Organization{
					{ID: orgID},
				},
				OrganizationUsers: []shared_types.OrganizationUsers{
					{
						OrganizationID: orgID,
						Role: &shared_types.Role{
							Permissions: []shared_types.Permission{
								{Name: "write", Resource: "organization"},
							},
						},
					},
				},
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: notification.ErrUserDoesNotHavePermissionForTheResource,
		},
		{
			name:   "Update SMTP - Creator",
			method: http.MethodPut,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   userID,
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: nil,
		},
		{
			name:   "Update SMTP - Non-Creator",
			method: http.MethodPut,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: notification.ErrPermissionDenied,
		},
		{
			name:   "Delete SMTP - Creator",
			method: http.MethodDelete,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   userID,
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: nil,
		},
		{
			name:   "Delete SMTP - Non-Creator",
			method: http.MethodDelete,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   uuid.New(),
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(smtpConfig, nil).Once()
			},
			expectedResult: notification.ErrPermissionDenied,
		},
		{
			name:   "Read SMTP - Storage Error",
			method: http.MethodGet,
			url:    "/api/v1/notification/smtp?id=" + smtpID,
			user: &shared_types.User{
				ID:   userID,
				Type: shared_types.RoleMember,
			},
			setupMock: func(mock *MockNotificationStorage) {
				mock.On("GetSmtp", "smtp").Return(nil, errors.New("not found")).Maybe()
				mock.On("GetSmtp", smtpID).Return(nil, errors.New("storage error")).Once()
			},
			expectedResult: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockNotificationStorage)
			validator := validation.NewValidator(mockStorage)

			if tt.setupMock != nil {
				tt.setupMock(mockStorage)
			}

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

			result := validator.AccessValidator(w, req, tt.user)

			if tt.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expectedResult.Error(), result.Error())
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
