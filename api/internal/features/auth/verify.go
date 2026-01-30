package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/config"
)

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
		log.Printf("DEBUG VerifySession: Forwarding Cookie header (length: %d)", len(cookieHeader))
		return
	}

	cookies := originalReq.Cookies()
	if len(cookies) > 0 {
		for _, cookie := range cookies {
			targetReq.AddCookie(cookie)
		}
		log.Printf("DEBUG VerifySession: Forwarding %d cookies individually", len(cookies))
	} else {
		log.Printf("WARN VerifySession: No cookies found in request - session validation will likely fail")
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
// Better Auth validates origins against trustedOrigins.
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
		log.Printf("ERROR VerifySession: Better Auth returned null response. Status: %d, URL: %s, Origin: %s, Referer: %s, Cookies: %s",
			statusCode, url, req.Header.Get("Origin"), req.Header.Get("Referer"), cookieInfo)
		return nil, fmt.Errorf("invalid session: Better Auth returned null (no active session)")
	}

	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		log.Printf("ERROR VerifySession: Failed to parse JSON response: %v, body: %s, status: %d", err, bodyStr, statusCode)
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, bodyStr)
	}

	if sessionResp.Error != nil {
		log.Printf("ERROR VerifySession: Better Auth returned error: %s (status: %d)", sessionResp.Error.Message, sessionResp.Error.Status)
		return nil, fmt.Errorf("session verification failed: %s (status: %d)", sessionResp.Error.Message, sessionResp.Error.Status)
	}

	if sessionResp.User.ID == "" {
		log.Printf("ERROR VerifySession: Session has no user data, response: %s", bodyStr)
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
		log.Printf("ERROR VerifySession: Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	forwardCookies(r, req)
	forwardHeaders(r, req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR VerifySession: HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to verify session: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR VerifySession: Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseSessionResponse(body, resp.StatusCode, url, req, r)
}

// SendOTP sends an OTP to the user's email via Better Auth for passwordless authentication.
func SendOTP(email string) error {
	betterAuthAPI := getBetterAuthAPI()
	url := betterAuthAPI + "/email-otp/send-verification-otp"

	payload := map[string]interface{}{
		"email": email,
		"type":  "sign-in",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send OTP: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
