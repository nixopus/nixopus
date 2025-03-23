package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/service"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDomainSuccess(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	mockStorage.On("GetDomainByName", mock.Anything).Return(nil, nil)
	mockStorage.On("CreateDomain", mock.Anything).Return(nil)

	service := service.NewDomainsService(nil, context.Background(), logger.NewLogger(), mockStorage)

	req := types.CreateDomainRequest{Name: "example.com"}
	userID := uuid.New().String()
	_, err := service.CreateDomain(req, userID)
	assert.NoError(t, err)
}

func TestCreateDomainAlreadyExists(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	mockStorage.On("GetDomainByName", "example.com").Return(&shared_types.Domain{}, nil)

	service := service.NewDomainsService(nil, context.Background(), logger.NewLogger(), mockStorage)

	req := types.CreateDomainRequest{Name: "example.com"}
	userID := uuid.New().String()
	_, err := service.CreateDomain(req, userID)
	assert.Equal(t, types.ErrDomainAlreadyExists, err)
}

func TestCreateDomainInvalidUserID(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	mockStorage.On("GetDomainByName", "example.com").Return(nil, nil)

	service := service.NewDomainsService(nil, context.Background(), logger.NewLogger(), mockStorage)

	req := types.CreateDomainRequest{Name: "example.com"}
	_, err := service.CreateDomain(req, "invalid-user-id")
	assert.NotNil(t, err)
}

func TestCreateDomainStorageError(t *testing.T) {
	mockStorage := NewMockDomainStorage()
	mockStorage.On("GetDomainByName", mock.Anything).Return(nil, nil)
	mockStorage.On("CreateDomain", mock.Anything).Return(errors.New("storage error"))

	service := service.NewDomainsService(nil, context.Background(), logger.NewLogger(), mockStorage)

	req := types.CreateDomainRequest{Name: "example.com"}
	userID := uuid.New().String()
	_, err := service.CreateDomain(req, userID)
	assert.NotNil(t, err)
}
