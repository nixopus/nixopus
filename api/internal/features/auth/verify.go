package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

var (
	authReadAll = io.ReadAll

	// forwardCookiesList returns cookies to forward when the raw Cookie header is empty.
	// Overridable in tests: net/http never exposes cookies without a Cookie header, but
	// callers may still want the AddCookie path when using custom request shims.
	forwardCookiesList = func(r *http.Request) []*http.Cookie { return r.Cookies() }

	// VerifySessionLogger, when set (e.g. from routes.NewRouter), is used for VerifySession logs.
	// If nil, logging falls back to a default Development logger.NewLogger().
	VerifySessionLogger *logger.Logger
)

func verifySessionLog(sev logger.Severity, msg, data string) {
	if VerifySessionLogger != nil {
		VerifySessionLogger.Log(sev, msg, data)
		return
	}
	l := logger.NewLogger()
	l.Log(sev, msg, data)
}

// HTTPClient is a shared HTTP client for Better Auth API calls.
// Uses connection pooling and timeouts to reduce latency from repeated connection setup.
var HTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

// getBetterAuthURL returns the Better Auth URL from config with fallback to localhost for development.
func getBetterAuthURL() string {
	url := config.AppConfig.BetterAuth.URL
	if url == "" {
		return "http://localhost:9090"
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return url
}

func getBetterAuthAPI() string {
	return getBetterAuthURL() + "/api/auth"
}

// SessionResponse represents the response from Better Auth get-session endpoint
type SessionResponse struct {
	Session struct {
		ID                   string  `json:"id"`
		UserID               string  `json:"userId"`
		ExpiresAt            string  `json:"expiresAt"`
		Token                string  `json:"token"`
		ActiveOrganizationID *string `json:"activeOrganizationId"` // Can be null
	} `json:"session"`
	User struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"emailVerified"`
	} `json:"user"`
	Error *struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"error"`
}

// forwardCookies forwards cookies from the original request to the Better Auth request.
// Better Auth requires cookies for session validation.
func forwardCookies(originalReq *http.Request, targetReq *http.Request) {
	cookieHeader := originalReq.Header.Get("Cookie")
	if cookieHeader != "" {
		targetReq.Header.Set("Cookie", cookieHeader)
		verifySessionLog(logger.Debug, "auth: VerifySession forward Cookie header", fmt.Sprintf("length=%d", len(cookieHeader)))
		return
	}

	cookies := forwardCookiesList(originalReq)
	if len(cookies) > 0 {
		for _, cookie := range cookies {
			targetReq.AddCookie(cookie)
		}
		verifySessionLog(logger.Debug, "auth: VerifySession forward cookies individually", fmt.Sprintf("count=%d", len(cookies)))
	} else {
		verifySessionLog(logger.Warning, "auth: VerifySession no cookies on request", "session validation may fail")
	}
}

func extractOriginFromReferer(referer string) string {
	if !strings.HasPrefix(referer, "http://") && !strings.HasPrefix(referer, "https://") {
		return ""
	}

	scheme := "https"
	if strings.HasPrefix(referer, "http://") {
		scheme = "http"
	}

	withoutScheme := strings.TrimPrefix(strings.TrimPrefix(referer, "https://"), "http://")

	host := withoutScheme
	if slashIndex := strings.Index(withoutScheme, "/"); slashIndex > 0 {
		host = withoutScheme[:slashIndex]
	}

	return scheme + "://" + host
}

// forwardHeaders forwards relevant headers from the original request to Better Auth request.
// Better Auth validates origins against trustedOrigins and uses X-Forwarded-Proto to
// determine secure context for cookie name resolution (__Secure- prefix vs plain).
func forwardHeaders(originalReq *http.Request, targetReq *http.Request) {
	if authHeader := originalReq.Header.Get("Authorization"); authHeader != "" {
		targetReq.Header.Set("Authorization", authHeader)
	}

	origin := originalReq.Header.Get("Origin")
	if origin == "" {
		if referer := originalReq.Header.Get("Referer"); referer != "" {
			origin = extractOriginFromReferer(referer)
			if origin != "" {
				targetReq.Header.Set("Origin", origin)
			}
			targetReq.Header.Set("Referer", referer)
		}
	} else {
		targetReq.Header.Set("Origin", origin)
	}

	if apiKey := originalReq.Header.Get("x-api-key"); apiKey != "" {
		targetReq.Header.Set("x-api-key", apiKey)
	}

	// Forward proxy/protocol headers so Better Auth resolves the correct cookie names.
	// Internal calls go over HTTP but the original client request came through TLS/Caddy.
	// Without this, Better Auth looks for "better-auth.session_token" (HTTP) while the
	// browser sent "__Secure-better-auth.session_token" (HTTPS), causing session lookup to fail.
	if proto := originalReq.Header.Get("X-Forwarded-Proto"); proto != "" {
		targetReq.Header.Set("X-Forwarded-Proto", proto)
	} else if originalReq.TLS != nil || strings.HasPrefix(origin, "https://") {
		targetReq.Header.Set("X-Forwarded-Proto", "https")
	}

	if host := originalReq.Header.Get("X-Forwarded-Host"); host != "" {
		targetReq.Header.Set("X-Forwarded-Host", host)
	}

	targetReq.Header.Set("Content-Type", "application/json")
	targetReq.Header.Set("User-Agent", originalReq.UserAgent())
}

func parseSessionResponse(body []byte, statusCode int, url string, req *http.Request, originalReq *http.Request) (*SessionResponse, error) {
	bodyStr := strings.TrimSpace(string(body))

	if bodyStr == "null" || bodyStr == "" {
		cookieInfo := "none"
		if cookieHeader := req.Header.Get("Cookie"); cookieHeader != "" {
			cookieNames := make([]string, 0, len(originalReq.Cookies()))
			for _, cookie := range originalReq.Cookies() {
				cookieNames = append(cookieNames, cookie.Name)
			}
			cookieInfo = fmt.Sprintf("%d cookies: %v", len(cookieNames), cookieNames)
		}
		verifySessionLog(logger.Error, "auth: VerifySession Better Auth returned null",
			fmt.Sprintf("status=%d url=%s origin=%s referer=%s cookies=%s",
				statusCode, url, req.Header.Get("Origin"), req.Header.Get("Referer"), cookieInfo))
		return nil, fmt.Errorf("invalid session: Better Auth returned null (no active session)")
	}

	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		verifySessionLog(logger.Error, fmt.Sprintf("auth: VerifySession parse JSON: %v", err),
			fmt.Sprintf("status=%d body=%s", statusCode, bodyStr))
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, bodyStr)
	}

	if sessionResp.Error != nil {
		verifySessionLog(logger.Error, "auth: VerifySession Better Auth error response",
			fmt.Sprintf("message=%s status=%d", sessionResp.Error.Message, sessionResp.Error.Status))
		return nil, fmt.Errorf("session verification failed: %s (status: %d)", sessionResp.Error.Message, sessionResp.Error.Status)
	}

	if sessionResp.User.ID == "" {
		verifySessionLog(logger.Error, "auth: VerifySession missing user in response", fmt.Sprintf("body=%s", bodyStr))
		return nil, fmt.Errorf("invalid session: no user data (response: %s)", bodyStr)
	}

	return &sessionResp, nil
}

// VerifySession verifies a Better Auth session by calling the Better Auth API.
// It forwards cookies and headers from the original request to verify the session.
func VerifySession(r *http.Request) (*SessionResponse, error) {
	betterAuthAPI := getBetterAuthAPI()
	url := betterAuthAPI + "/get-session"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		verifySessionLog(logger.Error, fmt.Sprintf("auth: VerifySession build request: %v", err), url)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	forwardCookies(r, req)
	forwardHeaders(r, req)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		verifySessionLog(logger.Error, fmt.Sprintf("auth: VerifySession HTTP: %v", err), url)
		return nil, fmt.Errorf("failed to verify session: %w", err)
	}
	defer resp.Body.Close()

	body, err := authReadAll(resp.Body)
	if err != nil {
		verifySessionLog(logger.Error, fmt.Sprintf("auth: VerifySession read body: %v", err), fmt.Sprintf("url=%s", url))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseSessionResponse(body, resp.StatusCode, url, req, r)
}
