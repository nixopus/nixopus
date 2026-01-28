package betterauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// getBetterAuthURL returns the Better Auth URL from environment variables
// This reads the environment variable dynamically each time it's called,
// ensuring that secrets loaded after package initialization are picked up
func getBetterAuthURL() string {
	url := os.Getenv("BETTER_AUTH_URL")
	if url == "" {
		// Default fallback for development
		url = "http://localhost:9090"
	} else {
		// Ensure protocol is present (http:// or https://)
		// If URL doesn't start with a protocol, assume https for production
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			// Default to https if no protocol specified (production environments)
			url = "https://" + url
		}
	}
	return url
}

// getBetterAuthAPI returns the Better Auth API base URL
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

// VerifySession verifies a Better Auth session by calling the Better Auth API
// It forwards the cookies from the original request to verify the session
func VerifySession(r *http.Request) (*SessionResponse, error) {
	betterAuthAPI := getBetterAuthAPI()
	url := betterAuthAPI + "/get-session"

	// Create a new request to Better Auth get-session endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR VerifySession: Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward cookies from the original request
	// Better Auth requires cookies for session validation
	cookieHeader := r.Header.Get("Cookie")
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
		log.Printf("DEBUG VerifySession: Forwarding Cookie header (length: %d)", len(cookieHeader))
	} else {
		// Fallback: add cookies individually if Cookie header wasn't present
		cookies := r.Cookies()
		if len(cookies) > 0 {
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
			log.Printf("DEBUG VerifySession: Forwarding %d cookies individually", len(cookies))
		} else {
			log.Printf("WARN VerifySession: No cookies found in request - session validation will likely fail")
		}
	}

	// Forward Authorization header if present
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Forward Origin header - Better Auth validates origins against trustedOrigins
	// This is critical for Better Auth to accept the request
	var originHeader string
	if origin := r.Header.Get("Origin"); origin != "" {
		originHeader = origin
		req.Header.Set("Origin", origin)
	} else if referer := r.Header.Get("Referer"); referer != "" {
		// Fallback to extracting origin from Referer header
		// Extract scheme and host from Referer URL (format: https://domain.com/path)
		if strings.HasPrefix(referer, "http://") || strings.HasPrefix(referer, "https://") {
			// Determine scheme
			scheme := "https"
			if strings.HasPrefix(referer, "http://") {
				scheme = "http"
			}
			// Remove scheme prefix to get host + path
			withoutScheme := strings.TrimPrefix(strings.TrimPrefix(referer, "https://"), "http://")
			// Extract host (everything before the first slash, or entire string if no slash)
			host := withoutScheme
			if slashIndex := strings.Index(withoutScheme, "/"); slashIndex > 0 {
				host = withoutScheme[:slashIndex]
			}
			originHeader = scheme + "://" + host
			req.Header.Set("Origin", originHeader)
		}
		// Also forward Referer header - Better Auth might use it for validation
		req.Header.Set("Referer", referer)
	}

	// Forward other relevant headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", r.UserAgent())

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR VerifySession: HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to verify session: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR VerifySession: Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	bodyStr := strings.TrimSpace(string(body))

	// Handle null response (Better Auth returns null when session is invalid/missing)
	if bodyStr == "null" || bodyStr == "" {
		cookieHeader := req.Header.Get("Cookie")
		cookieInfo := "none"
		if cookieHeader != "" {
			// Log cookie names only (not values) for security
			cookieNames := []string{}
			for _, cookie := range r.Cookies() {
				cookieNames = append(cookieNames, cookie.Name)
			}
			cookieInfo = fmt.Sprintf("%d cookies: %v", len(cookieNames), cookieNames)
		}
		log.Printf("ERROR VerifySession: Better Auth returned null response. Status: %d, URL: %s, Origin: %s, Referer: %s, Cookies: %s, ResponseHeaders: %v",
			resp.StatusCode, url, req.Header.Get("Origin"), req.Header.Get("Referer"), cookieInfo, resp.Header)
		return nil, fmt.Errorf("invalid session: Better Auth returned null (no active session)")
	}

	// Parse response
	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		log.Printf("ERROR VerifySession: Failed to parse JSON response: %v, body: %s, status: %d", err, bodyStr, resp.StatusCode)
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, bodyStr)
	}

	// Check for errors
	if sessionResp.Error != nil {
		log.Printf("ERROR VerifySession: Better Auth returned error: %s (status: %d)", sessionResp.Error.Message, sessionResp.Error.Status)
		return nil, fmt.Errorf("session verification failed: %s (status: %d)", sessionResp.Error.Message, sessionResp.Error.Status)
	}

	// Check if session is valid (has user data)
	if sessionResp.User.ID == "" {
		log.Printf("ERROR VerifySession: Session has no user data, response: %s", string(body))
		return nil, fmt.Errorf("invalid session: no user data (response: %s)", string(body))
	}

	return &sessionResp, nil
}

// SendOTP sends an OTP to the user's email via Better Auth
func SendOTP(email string) error {
	betterAuthAPI := getBetterAuthAPI()

	payload := map[string]interface{}{
		"email": email,
		"type":  "sign-in",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", betterAuthAPI+"/email-otp/send-verification-otp", bytes.NewBuffer(jsonData))
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
