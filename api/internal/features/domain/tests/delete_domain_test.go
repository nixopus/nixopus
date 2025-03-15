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
)

func TestDeleteDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainID       string
		existingDomain *shared_types.Domain
		getErr         error
		deleteErr      error
		expectedErr    error
	}{
		{
			name:           "domain exists and deletion is successful",
			domainID:       "123e4567-e89b-12d3-a456-426614174000",
			existingDomain: &shared_types.Domain{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")},
			expectedErr:    nil,
		},
		{
			name:           "domain does not exist",
			domainID:       "123e4567-e89b-12d3-a456-426614174000",
			existingDomain: nil,
			expectedErr:    types.ErrDomainNotFound,
		},
		{
			name:        "error occurs while retrieving domain",
			domainID:    "123e4567-e89b-12d3-a456-426614174000",
			getErr:      errors.New("get error"),
			expectedErr: errors.New("get error"),
		},
		{
			name:           "error occurs while deleting domain",
			domainID:       "123e4567-e89b-12d3-a456-426614174000",
			existingDomain: &shared_types.Domain{ID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")},
			deleteErr:      errors.New("delete error"),
			expectedErr:    errors.New("delete error"),
		},
		{
			name:        "invalid domain ID",
			domainID:    "123",
			expectedErr: types.ErrInvalidDomainID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := NewMockDomainStorage()

			if test.expectedErr != types.ErrInvalidDomainID {
				mockStorage.On("GetDomain", test.domainID).Return(test.existingDomain, test.getErr)

				if test.existingDomain != nil && test.getErr == nil {
					mockStorage.On("DeleteDomain", test.existingDomain).Return(test.deleteErr)
				}
			}

			s := service.NewDomainsService(nil, context.Background(), logger.NewLogger(), mockStorage)
			err := s.DeleteDomain(test.domainID)

			if test.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
