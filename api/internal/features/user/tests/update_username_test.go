package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/service"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/stretchr/testify/mock"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestUpdateUsername(t *testing.T) {
	storage := &MockUserStorage{}
	s := service.NewUserService(nil, context.Background(), logger.NewLogger(), storage)

	validUUID := uuid.New()
	nilUUID := uuid.Nil

	tests := []struct {
		name          string
		id            string
		req           *types.UpdateUserNameRequest
		wantErr       error
		mockUser      *shared_types.User
		mockGetErr    error
		mockUpdateErr error
	}{
		{
			name: "User exists and update is successful",
			id:   validUUID.String(),
			req: &types.UpdateUserNameRequest{
				Name: "new-username",
			},
			wantErr:       nil,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
		{
			name: "User does not exist",
			id:   nilUUID.String(),
			req: &types.UpdateUserNameRequest{
				Name: "new-username",
			},
			wantErr:       types.ErrUserDoesNotExist,
			mockUser:      &shared_types.User{ID: nilUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
		{
			name: "Update fails due to storage error",
			id:   validUUID.String(),
			req: &types.UpdateUserNameRequest{
				Name: "new-username",
			},
			wantErr:       types.ErrFailedToUpdateUser,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: errors.New("storage error"),
		},
		{
			name:          "Empty request",
			id:            validUUID.String(),
			req:           &types.UpdateUserNameRequest{},
			wantErr:       nil,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
		{
			name:          "Nil request",
			id:            validUUID.String(),
			req:           nil,
			wantErr:       types.ErrInvalidRequestType,
			mockUser:      &shared_types.User{ID: validUUID},
			mockGetErr:    nil,
			mockUpdateErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.ExpectedCalls = nil

			if tt.req != nil {
				storage.On("GetUserById", tt.id).Return(tt.mockUser, tt.mockGetErr)
			}

			if tt.req != nil && tt.mockUser.ID != uuid.Nil && tt.mockGetErr == nil {
				storage.On("UpdateUserName", tt.mockUser.ID.String(), tt.req.Name, mock.Anything).Return(tt.mockUpdateErr)
			}

			err := s.UpdateUsername(tt.id, tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
