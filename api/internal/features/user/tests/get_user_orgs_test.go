package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/service"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/stretchr/testify/assert"
)

func TestGetUserOrganizations(t *testing.T) {
	t.Run("test user does not exist", func(t *testing.T) {
		mockStorage := &MockUserStorage{}
		mockStorage.On("GetUserOrganizationsWithRolesAndPermissions", "non-existent-user").Return([]types.UserOrganizationsResponse{}, types.ErrUserDoesNotExist)

		userService := service.NewUserService(nil, context.Background(), logger.NewLogger(), mockStorage)
		orgs, err := userService.GetUserOrganizations("non-existent-user")

		assert.Empty(t, orgs)
		assert.Equal(t, types.ErrUserDoesNotExist, err)
	})

	t.Run("test storage layer returns error", func(t *testing.T) {
		mockStorage := &MockUserStorage{}
		mockStorage.On("GetUserOrganizationsWithRolesAndPermissions", "user-id").Return([]types.UserOrganizationsResponse{}, errors.New("storage error"))

		userService := service.NewUserService(nil, context.Background(), logger.NewLogger(), mockStorage)
		orgs, err := userService.GetUserOrganizations("user-id")

		assert.Empty(t, orgs)
		assert.NotNil(t, err)
	})

	t.Run("test storage layer returns empty organizations", func(t *testing.T) {
		mockStorage := &MockUserStorage{}
		mockStorage.On("GetUserOrganizationsWithRolesAndPermissions", "user-id").Return([]types.UserOrganizationsResponse{}, nil)

		userService := service.NewUserService(nil, context.Background(), logger.NewLogger(), mockStorage)
		orgs, err := userService.GetUserOrganizations("user-id")

		assert.Empty(t, orgs)
		assert.Nil(t, err)
	})
}
