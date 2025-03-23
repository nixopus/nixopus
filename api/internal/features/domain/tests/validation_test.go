package tests

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateID(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "Empty ID",
			id:      "",
			wantErr: types.ErrMissingDomainID,
		},
		{
			name:    "Invalid UUID",
			id:      "not-a-uuid",
			wantErr: types.ErrInvalidDomainID,
		},
		{
			name:    "Valid UUID",
			id:      "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateID(tt.id)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateName(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	tests := []struct {
		name       string
		domainName string
		wantErr    error
	}{
		{
			name:       "Empty Name",
			domainName: "",
			wantErr:    types.ErrMissingDomainName,
		},
		{
			name:       "Name Too Short",
			domainName: "ab",
			wantErr:    types.ErrDomainNameTooShort,
		},
		{
			name:       "Name Too Long",
			domainName: strings.Repeat("a", 256),
			wantErr:    types.ErrDomainNameTooLong,
		},
		{
			name:       "Valid Name",
			domainName: "example.com",
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateName(tt.domainName)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestParseRequestBody(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	tests := []struct {
		name     string
		jsonBody string
		wantErr  bool
	}{
		{
			name:     "Valid JSON",
			jsonBody: `{"name": "example.com"}`,
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			jsonBody: `{"name": "example.com"`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(bytes.NewReader([]byte(tt.jsonBody)))
			var decoded map[string]string
			err := validator.ParseRequestBody(nil, body, &decoded)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "example.com", decoded["name"])
			}
		})
	}
}

func TestValidateCreateDomainRequest(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	tests := []struct {
		name    string
		req     types.CreateDomainRequest
		wantErr error
	}{
		{
			name:    "Valid Request",
			req:     types.CreateDomainRequest{Name: "example.com"},
			wantErr: nil,
		},
		{
			name:    "Empty Name",
			req:     types.CreateDomainRequest{Name: ""},
			wantErr: types.ErrMissingDomainName,
		},
		{
			name:    "Name Too Short",
			req:     types.CreateDomainRequest{Name: "ab"},
			wantErr: types.ErrDomainNameTooShort,
		},
		{
			name:    "Name Too Long",
			req:     types.CreateDomainRequest{Name: strings.Repeat("a", 256)},
			wantErr: types.ErrDomainNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreateDomainRequest(tt.req)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateUpdateDomainRequest(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	validID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	validUUID, _ := uuid.Parse(validID)

	nonExistentID := "a47ac10b-58cc-4372-a567-0e02b2c3d480"

	errorID := "b47ac10b-58cc-4372-a567-0e02b2c3d481"

	adminUser := shared_types.User{
		ID:   validUUID,
		Type: "admin",
	}

	ownerUser := shared_types.User{
		ID:   validUUID,
		Type: "regular",
	}

	otherUser := shared_types.User{
		ID:   uuid.New(),
		Type: "regular",
	}

	mockDomain := &shared_types.Domain{
		ID:     validUUID,
		Name:   "example.com",
		UserID: validUUID,
	}

	mockStorage.WithGetDomain(validID, mockDomain, nil)
	mockStorage.WithGetDomain(nonExistentID, nil, nil)
	mockStorage.WithGetDomain(errorID, nil, assert.AnError)
	mockStorage.WithUpdateDomain(mockDomain.ID.String(), mockDomain.Name, nil)

	tests := []struct {
		name    string
		req     types.UpdateDomainRequest
		user    shared_types.User
		wantErr error
	}{
		{
			name:    "Valid Request by Admin",
			req:     types.UpdateDomainRequest{ID: validID, Name: "example.com"},
			user:    adminUser,
			wantErr: nil,
		},
		{
			name:    "Valid Request by Owner",
			req:     types.UpdateDomainRequest{ID: validID, Name: "example.com"},
			user:    ownerUser,
			wantErr: nil,
		},
		{
			name:    "Not Allowed - Not Owner or Admin",
			req:     types.UpdateDomainRequest{ID: validID, Name: "example.com"},
			user:    otherUser,
			wantErr: types.ErrNotAllowed,
		},
		{
			name:    "Invalid ID",
			req:     types.UpdateDomainRequest{ID: "not-a-uuid", Name: "example.com"},
			user:    adminUser,
			wantErr: types.ErrInvalidDomainID,
		},
		{
			name:    "Empty Name",
			req:     types.UpdateDomainRequest{ID: validID, Name: ""},
			user:    adminUser,
			wantErr: types.ErrMissingDomainName,
		},
		{
			name:    "Domain Not Found",
			req:     types.UpdateDomainRequest{ID: nonExistentID, Name: "example.com"},
			user:    adminUser,
			wantErr: types.ErrDomainNotFound,
		},
		{
			name:    "Storage Error",
			req:     types.UpdateDomainRequest{ID: errorID, Name: "example.com"},
			user:    adminUser,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdateDomainRequest(tt.req, tt.user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateDeleteDomainRequest(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	validID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	validUUID, _ := uuid.Parse(validID)

	nonExistentID := "a47ac10b-58cc-4372-a567-0e02b2c3d480"

	errorID := "b47ac10b-58cc-4372-a567-0e02b2c3d481"

	adminUser := shared_types.User{
		ID:   validUUID,
		Type: "admin",
	}

	ownerUser := shared_types.User{
		ID:   validUUID,
		Type: "regular",
	}

	otherUser := shared_types.User{
		ID:   uuid.New(),
		Type: "regular",
	}

	mockDomain := &shared_types.Domain{
		ID:     validUUID,
		Name:   "example.com",
		UserID: validUUID,
	}

	mockStorage.WithGetDomain(mockDomain.ID.String(), mockDomain, nil)
	mockStorage.WithGetDomain(nonExistentID, nil, nil)
	mockStorage.WithGetDomain(errorID, nil, assert.AnError)

	tests := []struct {
		name    string
		req     types.DeleteDomainRequest
		user    shared_types.User
		wantErr error
	}{
		{
			name:    "Valid Request by Admin",
			req:     types.DeleteDomainRequest{ID: validID},
			user:    adminUser,
			wantErr: nil,
		},
		{
			name:    "Valid Request by Owner",
			req:     types.DeleteDomainRequest{ID: validID},
			user:    ownerUser,
			wantErr: nil,
		},
		{
			name:    "Not Allowed - Not Owner or Admin",
			req:     types.DeleteDomainRequest{ID: validID},
			user:    otherUser,
			wantErr: types.ErrNotAllowed,
		},
		{
			name:    "Invalid ID",
			req:     types.DeleteDomainRequest{ID: "not-a-uuid"},
			user:    adminUser,
			wantErr: types.ErrInvalidDomainID,
		},
		{
			name:    "Domain Not Found",
			req:     types.DeleteDomainRequest{ID: nonExistentID},
			user:    adminUser,
			wantErr: types.ErrDomainNotFound,
		},
		{
			name:    "Storage Error",
			req:     types.DeleteDomainRequest{ID: errorID},
			user:    adminUser,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDeleteDomainRequest(tt.req, tt.user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateRequest(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	validator := validation.NewValidator(mockStorage)

	validID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	validUUID, _ := uuid.Parse(validID)

	adminUser := shared_types.User{
		ID:   validUUID,
		Type: "admin",
	}

	mockDomain := &shared_types.Domain{
		ID:     validUUID,
		Name:   "example.com",
		UserID: validUUID,
	}

	mockStorage.WithGetDomain(validID, mockDomain, nil)

	tests := []struct {
		name    string
		req     interface{}
		user    shared_types.User
		wantErr error
	}{
		{
			name:    "Valid Create Request",
			req:     &types.CreateDomainRequest{Name: "example.com"},
			user:    adminUser,
			wantErr: nil,
		},
		{
			name:    "Valid Update Request",
			req:     &types.UpdateDomainRequest{ID: validID, Name: "example.com"},
			user:    adminUser,
			wantErr: nil,
		},
		{
			name:    "Valid Delete Request",
			req:     &types.DeleteDomainRequest{ID: validID},
			user:    adminUser,
			wantErr: nil,
		},
		{
			name:    "Invalid Request Type",
			req:     "not-a-valid-request",
			user:    adminUser,
			wantErr: types.ErrInvalidRequestType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRequest(tt.req, tt.user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
