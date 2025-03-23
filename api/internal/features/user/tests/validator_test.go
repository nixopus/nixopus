package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	"github.com/raghavyuva/nixopus-api/internal/features/user/validation"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestAccessValidator(t *testing.T) {
	v := validation.NewValidator()

	tests := []struct {
		name       string
		path       string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "allowed endpoint",
			path:       "/api/v1/user",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "allowed endpoint",
			path:       "/api/v1/user/name",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "allowed endpoint",
			path:       "/api/v1/user/organizations",
			wantErr:    false,
			errMessage: "",
		},
		{
			name:       "disallowed endpoint",
			path:       "/api/v1/user/other",
			wantErr:    true,
			errMessage: types.ErrInvalidAccess.Error(),
		},
		{
			name:       "empty path",
			path:       "",
			wantErr:    true,
			errMessage: types.ErrInvalidAccess.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			user := &shared_types.User{}

			err = v.AccessValidator(w, req, user)

			if (err != nil) != tt.wantErr {
				t.Errorf("AccessValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("AccessValidator() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}
