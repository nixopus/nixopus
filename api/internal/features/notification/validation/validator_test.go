package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newValidator(t *testing.T) *Validator {
	t.Helper()
	// ValidateRequest does not use storage; nil repository is sufficient.
	return NewValidator(nil)
}

func TestNewValidator_NotNil(t *testing.T) {
	v := NewValidator(nil)
	require.NotNil(t, v)
}

func TestValidateRequest_InvalidType(t *testing.T) {
	v := newValidator(t)
	err := v.ValidateRequest("not-a-request")
	require.ErrorIs(t, err, notification.ErrInvalidRequestType)
}

func TestValidateRequest_CreateSMTPConfig_Success(t *testing.T) {
	v := newValidator(t)
	err := v.ValidateRequest(&notification.CreateSMTPConfigRequest{
		Host:           "smtp.example.com",
		Port:           587,
		Username:       "user@example.com",
		Password:       "secret",
		OrganizationID: uuid.New(),
	})
	require.NoError(t, err)
}

func TestValidateRequest_CreateSMTPConfig_Errors(t *testing.T) {
	v := newValidator(t)
	orgID := uuid.New()

	t.Run("missing host", func(t *testing.T) {
		err := v.ValidateRequest(&notification.CreateSMTPConfigRequest{
			Port:           587,
			Username:       "u",
			Password:       "p",
			OrganizationID: orgID,
		})
		require.ErrorIs(t, err, notification.ErrMissingHost)
	})
	t.Run("missing port", func(t *testing.T) {
		err := v.ValidateRequest(&notification.CreateSMTPConfigRequest{
			Host:           "h",
			Username:       "u",
			Password:       "p",
			OrganizationID: orgID,
		})
		require.ErrorIs(t, err, notification.ErrMissingPort)
	})
	t.Run("missing username", func(t *testing.T) {
		err := v.ValidateRequest(&notification.CreateSMTPConfigRequest{
			Host:           "h",
			Port:           1,
			Password:       "p",
			OrganizationID: orgID,
		})
		require.ErrorIs(t, err, notification.ErrMissingUsername)
	})
	t.Run("missing password", func(t *testing.T) {
		err := v.ValidateRequest(&notification.CreateSMTPConfigRequest{
			Host:           "h",
			Port:           1,
			Username:       "u",
			OrganizationID: orgID,
		})
		require.ErrorIs(t, err, notification.ErrMissingPassword)
	})
	t.Run("missing organization", func(t *testing.T) {
		err := v.ValidateRequest(&notification.CreateSMTPConfigRequest{
			Host:     "h",
			Port:     1,
			Username: "u",
			Password: "p",
		})
		require.ErrorIs(t, err, notification.ErrMissingOrganization)
	})
}

func TestValidateRequest_GetSMTPConfig(t *testing.T) {
	v := newValidator(t)
	err := v.ValidateRequest(&notification.GetSMTPConfigRequest{ID: uuid.Nil})
	require.ErrorIs(t, err, notification.ErrMissingID)

	err = v.ValidateRequest(&notification.GetSMTPConfigRequest{ID: uuid.New()})
	require.NoError(t, err)
}

func TestValidateRequest_UpdateSMTPConfig(t *testing.T) {
	v := newValidator(t)
	err := v.ValidateRequest(&notification.UpdateSMTPConfigRequest{ID: uuid.Nil})
	require.ErrorIs(t, err, notification.ErrMissingID)

	err = v.ValidateRequest(&notification.UpdateSMTPConfigRequest{ID: uuid.New()})
	require.NoError(t, err)
}

func TestValidateRequest_DeleteSMTPConfig(t *testing.T) {
	v := newValidator(t)
	err := v.ValidateRequest(&notification.DeleteSMTPConfigRequest{ID: uuid.Nil})
	require.ErrorIs(t, err, notification.ErrMissingID)

	err = v.ValidateRequest(&notification.DeleteSMTPConfigRequest{ID: uuid.New()})
	require.NoError(t, err)
}

func TestValidateRequest_UpdatePreference(t *testing.T) {
	v := newValidator(t)

	err := v.ValidateRequest(&notification.UpdatePreferenceRequest{
		Category: "security",
		Type:     "",
	})
	require.ErrorIs(t, err, notification.ErrMissingType)

	err = v.ValidateRequest(&notification.UpdatePreferenceRequest{
		Category: "",
		Type:     "login-alerts",
	})
	require.ErrorIs(t, err, notification.ErrMissingCategory)

	err = v.ValidateRequest(&notification.UpdatePreferenceRequest{
		Category: "invalid",
		Type:     "x",
	})
	require.ErrorIs(t, err, notification.ErrInvalidRequestType)

	for _, cat := range []string{"activity", "security", "update"} {
		cat := cat
		t.Run(cat, func(t *testing.T) {
			err := v.ValidateRequest(&notification.UpdatePreferenceRequest{
				Category: cat,
				Type:     "login-alerts",
				Enabled:  true,
			})
			require.NoError(t, err)
		})
	}
}

func TestParseRequestBody_ReadError(t *testing.T) {
	v := newValidator(t)
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Body = io.NopCloser(&failingReader{})
	var out map[string]any
	err := v.ParseRequestBody(r, r.Body, &out)
	require.Error(t, err)
}

func TestParseRequestBody_JSONError(t *testing.T) {
	v := newValidator(t)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	err := v.ParseRequestBody(r, r.Body, new(map[string]any))
	require.Error(t, err)
}

func TestParseRequestBody_Success_RereadsBody(t *testing.T) {
	v := newValidator(t)
	payload := map[string]string{"hello": "world"}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(raw))
	var decoded map[string]string
	require.NoError(t, v.ParseRequestBody(r, r.Body, &decoded))
	assert.Equal(t, "world", decoded["hello"])

	again, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(raw), string(again))
}

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}
