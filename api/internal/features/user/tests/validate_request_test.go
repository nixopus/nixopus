package tests

import (
	"errors"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/features/user/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestValidateRequest(t *testing.T) {
	v := &validation.Validator{}
	user := shared_types.User{}

	tests := []struct {
		name    string
		req     interface{}
		wantErr error
	}{
		{
			name: "Valid UpdateUserNameRequest",
			req: &types.UpdateUserNameRequest{
				Name: "newusername",
			},
			wantErr: nil,
		},
		{
			name:    "Invalid request type",
			req:     struct{}{},
			wantErr: types.ErrInvalidRequestType,
		},
		{
			name:    "Nil request",
			req:     nil,
			wantErr: types.ErrInvalidRequestType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(tt.req, user)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
