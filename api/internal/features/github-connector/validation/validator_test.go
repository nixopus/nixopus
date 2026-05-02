package validation_test

import (
	"errors"
	"testing"

	gctypes "github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/validation"
)

func TestNewValidator(t *testing.T) {
	v := validation.NewValidator(nil)
	if v == nil {
		t.Fatal("expected non-nil Validator")
	}
}

func TestValidator_ValidateRequest_invalidType(t *testing.T) {
	v := validation.NewValidator(nil)
	err := v.ValidateRequest("not-a-request")
	if !errors.Is(err, gctypes.ErrInvalidRequestType) {
		t.Fatalf("ValidateRequest(wrong type): got %v, want ErrInvalidRequestType", err)
	}
	err = v.ValidateRequest(struct{}{})
	if !errors.Is(err, gctypes.ErrInvalidRequestType) {
		t.Fatalf("ValidateRequest(empty struct): got %v, want ErrInvalidRequestType", err)
	}
}

func TestValidator_ValidateRequest_create_noCredentials(t *testing.T) {
	v := validation.NewValidator(nil)
	req := &gctypes.CreateGithubConnectorRequest{}
	if err := v.ValidateRequest(req); err != nil {
		t.Fatalf("expected nil when all credential fields empty, got %v", err)
	}
}

func TestValidator_ValidateRequest_create_whitespaceOnlyTreatedAsEmpty(t *testing.T) {
	v := validation.NewValidator(nil)
	req := &gctypes.CreateGithubConnectorRequest{
		Slug:          " \t\n",
		Pem:           "   ",
		ClientID:      "",
		ClientSecret:  " ",
		WebhookSecret: "",
		AppID:         "  ",
	}
	if err := v.ValidateRequest(req); err != nil {
		t.Fatalf("expected nil for whitespace-only fields (no credentials), got %v", err)
	}
}

func TestValidator_ValidateRequest_create_partialCredentials(t *testing.T) {
	v := validation.NewValidator(nil)

	tests := []struct {
		name    string
		req     *gctypes.CreateGithubConnectorRequest
		wantErr error
	}{
		{
			name:    "only slug triggers full set; missing pem",
			req:     &gctypes.CreateGithubConnectorRequest{Slug: "my-app"},
			wantErr: gctypes.ErrMissingPem,
		},
		{
			name:    "only app_id triggers full set; missing slug first",
			req:     &gctypes.CreateGithubConnectorRequest{AppID: "123"},
			wantErr: gctypes.ErrMissingSlug,
		},
		{
			name: "slug and pem; missing client_id",
			req: &gctypes.CreateGithubConnectorRequest{
				Slug: "s",
				Pem:  "pem",
			},
			wantErr: gctypes.ErrMissingClientID,
		},
		{
			name: "slug pem client_id; missing client_secret",
			req: &gctypes.CreateGithubConnectorRequest{
				Slug:     "s",
				Pem:      "p",
				ClientID: "c",
			},
			wantErr: gctypes.ErrMissingClientSecret,
		},
		{
			name: "all but webhook_secret",
			req: &gctypes.CreateGithubConnectorRequest{
				Slug:         "s",
				Pem:          "p",
				ClientID:     "c",
				ClientSecret: "cs",
			},
			wantErr: gctypes.ErrMissingWebhookSecret,
		},
		{
			name: "missing slug when others set",
			req: &gctypes.CreateGithubConnectorRequest{
				Pem:           "p",
				ClientID:      "c",
				ClientSecret:  "cs",
				WebhookSecret: "w",
			},
			wantErr: gctypes.ErrMissingSlug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateRequest_create_allCredentialFieldsPresent(t *testing.T) {
	v := validation.NewValidator(nil)
	req := &gctypes.CreateGithubConnectorRequest{
		Slug:          "my-slug",
		Pem:           "-----BEGIN RSA PRIVATE KEY-----\n",
		ClientID:      "Iv1.abc",
		ClientSecret:  "secret",
		WebhookSecret: "whsec",
	}
	if err := v.ValidateRequest(req); err != nil {
		t.Fatalf("expected nil without app_id when other fields set, got %v", err)
	}

	reqWithApp := *req
	reqWithApp.AppID = "999"
	if err := v.ValidateRequest(&reqWithApp); err != nil {
		t.Fatalf("expected nil with app_id, got %v", err)
	}
}

func TestValidator_ValidateRequest_update(t *testing.T) {
	v := validation.NewValidator(nil)

	t.Run("missing installation_id", func(t *testing.T) {
		err := v.ValidateRequest(&gctypes.UpdateGithubConnectorRequest{})
		if !errors.Is(err, gctypes.ErrMissingInstallationID) {
			t.Fatalf("got %v, want ErrMissingInstallationID", err)
		}
	})

	t.Run("ok", func(t *testing.T) {
		err := v.ValidateRequest(&gctypes.UpdateGithubConnectorRequest{InstallationID: "12345"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestValidator_ValidateRequest_delete(t *testing.T) {
	v := validation.NewValidator(nil)

	t.Run("missing id", func(t *testing.T) {
		err := v.ValidateRequest(&gctypes.DeleteGithubConnectorRequest{})
		if !errors.Is(err, gctypes.ErrMissingID) {
			t.Fatalf("got %v, want ErrMissingID", err)
		}
	})

	t.Run("ok", func(t *testing.T) {
		err := v.ValidateRequest(&gctypes.DeleteGithubConnectorRequest{ID: "550e8400-e29b-41d4-a716-446655440000"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
