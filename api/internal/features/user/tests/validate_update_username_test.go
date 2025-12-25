package tests

import (
	"errors"
	"strings"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/features/user/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestValidateUpdateUserNameRequest(t *testing.T) {
	v := &validation.Validator{}
	user := shared_types.User{Username: "currentuser"}

	tests := []struct {
		name    string
		req     types.UpdateUserNameRequest
		wantErr error
	}{
		{
			name:    "Empty username",
			req:     types.UpdateUserNameRequest{Name: ""},
			wantErr: types.ErrUserNameIsEmpty,
		},
		{
			name:    "Same username as current user",
			req:     types.UpdateUserNameRequest{Name: "currentuser"},
			wantErr: types.ErrSameUserName,
		},
		{
			name:    "Username too long",
			req:     types.UpdateUserNameRequest{Name: strings.Repeat("a", 51)},
			wantErr: types.ErrUserNameTooLong,
		},
		{
			name:    "Username contains spaces",
			req:     types.UpdateUserNameRequest{Name: "username with spaces"},
			wantErr: types.ErrUserNameContainsSpaces,
		},
		{
			name:    "Username too short",
			req:     types.UpdateUserNameRequest{Name: "ab"},
			wantErr: types.ErrUsernameTooShort,
		},
		{
			name:    "Valid username",
			req:     types.UpdateUserNameRequest{Name: "newusername"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUpdateUserNameRequest(tt.req, user)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validateUpdateUserNameRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
