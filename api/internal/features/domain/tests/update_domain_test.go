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

func TestUpdateDomain(t *testing.T) {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	
	tests := []struct {
		name            string
		domainID        string
		userID          string
		newName         string
		domainExists    bool
		getDomainErr    error
		updateDomainErr error
		expectedErr     error
	}{
		{
			name:            "success",
			domainID:        validUUID,
			userID:          "456",
			newName:         "new-name",
			domainExists:    true,
			getDomainErr:    nil,
			updateDomainErr: nil,
			expectedErr:     nil,
		},
		{
			name:            "domain not found",
			domainID:        validUUID,
			userID:          "456",
			newName:         "new-name",
			domainExists:    false,
			getDomainErr:    nil,
			updateDomainErr: nil,
			expectedErr:     types.ErrDomainNotFound,
		},
		{
			name:            "storage get domain error",
			domainID:        validUUID,
			userID:          "456",
			newName:         "new-name",
			domainExists:    false,
			getDomainErr:    types.ErrDomainNotFound,
			updateDomainErr: nil,
			expectedErr:     types.ErrDomainNotFound,
		},
		{
			name:            "storage update domain error",
			domainID:        validUUID,
			userID:          "456",
			newName:         "new-name",
			domainExists:    true,
			getDomainErr:    nil,
			updateDomainErr: errors.New("update domain error"),
			expectedErr:     errors.New("update domain error"),
		},
		{
			name:         "invalid domain ID",
			domainID:     "123",
			userID:       "456",
			newName:      "new-name",
			expectedErr:  types.ErrInvalidDomainID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := NewMockDomainStorage()

			if test.expectedErr != types.ErrInvalidDomainID {
				var domain *shared_types.Domain
				if test.domainExists {
					domain = &shared_types.Domain{ID: uuid.MustParse(test.domainID)}
				}
				
				mockStorage.WithGetDomain(test.domainID, domain, test.getDomainErr)
				
				if test.domainExists && test.getDomainErr == nil {
					mockStorage.WithUpdateDomain(test.domainID, test.newName, test.updateDomainErr)
					
					if test.updateDomainErr == nil {
						mockStorage.WithGetDomain(test.domainID, domain, nil)
					}
				}
			}

			s := service.NewDomainsService(nil, context.Background(), logger.NewLogger(), mockStorage)
			_, err := s.UpdateDomain(test.newName, test.userID, test.domainID)

			if test.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			}
			
			mockStorage.AssertExpectations(t)
		})
	}
}