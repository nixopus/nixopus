package middleware

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	betterauth "github.com/nixopus/nixopus/api/internal/features/auth"
	"github.com/stretchr/testify/require"
)

func Test_getBetterAuthOrganizationMember_newRequestError(t *testing.T) {
	prev := rbacNewRequestWithContext
	rbacNewRequestWithContext = func(context.Context, string, string, io.Reader) (*http.Request, error) {
		return nil, errors.New("nr fail")
	}
	t.Cleanup(func() { rbacNewRequestWithContext = prev })

	t.Setenv("AUTH_SERVICE_URL", "http://example.invalid")
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "o")
	require.ErrorContains(t, err, "failed to create request")
}

func Test_getBetterAuthOrganizationMember_httpDoError(t *testing.T) {
	orig := betterauth.HTTPClient
	betterauth.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial fail")
		}),
	}
	t.Cleanup(func() { betterauth.HTTPClient = orig })

	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to fetch organization members")
}

func Test_getBetterAuthOrganizationMember_nonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("nope"))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Better Auth API returned status 403")
}

func Test_getBetterAuthOrganizationMember_readBodyError(t *testing.T) {
	prev := rbacReadResponseBody
	rbacReadResponseBody = func(io.Reader) ([]byte, error) {
		return nil, errors.New("read fail")
	}
	t.Cleanup(func() { rbacReadResponseBody = prev })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.ErrorContains(t, err, "failed to read response")
}

func Test_getBetterAuthOrganizationMember_directArray(t *testing.T) {
	uid := "11111111-1111-1111-1111-111111111111"
	oid := "22222222-2222-2222-2222-222222222222"
	body := fmt.Sprintf(`[{"userId":"%s","organizationId":"%s","role":"owner","user":{"id":"%s","email":"e","name":"n"}}]`, uid, oid, uid)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	m, err := getBetterAuthOrganizationMember(context.Background(), nil, uid, oid)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, uid, m.UserID)
}

func Test_getBetterAuthOrganizationMember_dataWrapper(t *testing.T) {
	uid := "11111111-1111-1111-1111-111111111111"
	oid := "22222222-2222-2222-2222-222222222222"
	inner := fmt.Sprintf(`[{"userId":"%s","organizationId":"%s","role":"admin","user":{"id":"%s"}}]`, uid, oid, uid)
	body := `{"data":` + inner + `}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	m, err := getBetterAuthOrganizationMember(context.Background(), nil, uid, oid)
	require.NoError(t, err)
	rs, ok := m.Role.(string)
	require.True(t, ok)
	require.Equal(t, "admin", rs)
}

func Test_getBetterAuthOrganizationMember_membersWrapper(t *testing.T) {
	uid := "11111111-1111-1111-1111-111111111111"
	oid := "22222222-2222-2222-2222-222222222222"
	inner := fmt.Sprintf(`[{"userId":"%s","organizationId":"%s","role":"viewer","user":{"id":"%s"}}]`, uid, oid, uid)
	body := `{"members":` + inner + `}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	m, err := getBetterAuthOrganizationMember(context.Background(), nil, uid, oid)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func Test_getBetterAuthOrganizationMember_dataUnmarshalError(t *testing.T) {
	body := `{"data":"not-an-array"}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse data array")
}

func Test_getBetterAuthOrganizationMember_membersUnmarshalError(t *testing.T) {
	body := `{"members":123}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse members array")
}

func Test_getBetterAuthOrganizationMember_singleMemberObject(t *testing.T) {
	uid := "11111111-1111-1111-1111-111111111111"
	oid := "22222222-2222-2222-2222-222222222222"
	body := fmt.Sprintf(`{"userId":"%s","organizationId":"%s","role":"member","user":{"id":"%s"}}`, uid, oid, uid)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	m, err := getBetterAuthOrganizationMember(context.Background(), nil, uid, oid)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func Test_getBetterAuthOrganizationMember_singleMemberUnmarshalNoUserID(t *testing.T) {
	body := `{"foo":1,"name":"orphan"}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	require.Contains(t, err.Error(), "response does not contain array or single member")
}

func Test_getBetterAuthOrganizationMember_parseBodyNeitherArrayNorObject(t *testing.T) {
	body := `not json`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "u", "00000000-0000-0000-0000-000000000001")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to parse response") || strings.Contains(err.Error(), "failed to parse"))
}

func Test_getBetterAuthOrganizationMember_userNotInList(t *testing.T) {
	body := `[{"userId":"other","organizationId":"o","role":"owner","user":{"id":"other"}}]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	_, err := getBetterAuthOrganizationMember(context.Background(), nil, "me", "o")
	require.Error(t, err)
	require.Contains(t, err.Error(), "is not a member")
}

func Test_getBetterAuthOrganizationMember_forwardsHeaders(t *testing.T) {
	uid := "11111111-1111-1111-1111-111111111111"
	oid := "22222222-2222-2222-2222-222222222222"
	body := fmt.Sprintf(`[{"userId":"%s","organizationId":"%s","role":"owner","user":{"id":"%s"}}]`, uid, oid, uid)

	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization") == "Bearer x"
		require.NotEmpty(t, r.Header.Get("Cookie"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("AUTH_SERVICE_URL", srv.URL)
	t.Cleanup(func() { _ = os.Unsetenv("AUTH_SERVICE_URL") })

	origReq := httptest.NewRequest(http.MethodGet, "/", nil)
	origReq.Header.Set("Authorization", "Bearer x")
	origReq.Header.Set("x-api-key", "k")
	origReq.AddCookie(&http.Cookie{Name: "a", Value: "b"})

	m, err := getBetterAuthOrganizationMember(context.Background(), origReq, uid, oid)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.True(t, sawAuth)
}
