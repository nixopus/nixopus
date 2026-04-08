package middleware

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	jwksCache    jwk.Set
	jwksMu       sync.RWMutex
	jwksExpiry   time.Time
	jwksCacheTTL = 15 * time.Minute
)

func isJWT(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 3
}

func validateM2MJWT(ctx context.Context, rawToken string, headerOrgID string) (orgID string, err error) {
	keySet, err := fetchJWKS(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	parseOpts := []jwt.ParseOption{
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	}
	if issuer := os.Getenv("AUTH_ISSUER"); issuer != "" {
		parseOpts = append(parseOpts, jwt.WithIssuer(issuer))
	}
	if audience := os.Getenv("AUTH_AUDIENCE"); audience != "" {
		parseOpts = append(parseOpts, jwt.WithAudience(audience))
	}

	parsed, err := jwt.Parse([]byte(rawToken), parseOpts...)
	if err != nil {
		return "", fmt.Errorf("JWT validation failed: %w", err)
	}

	if orgClaim, ok := parsed.PrivateClaims()["https://nixopus.com/org"]; ok {
		if orgStr, ok := orgClaim.(string); ok && orgStr != "" {
			return orgStr, nil
		}
	}

	if headerOrgID != "" {
		return headerOrgID, nil
	}

	return "", fmt.Errorf("missing organization: no https://nixopus.com/org claim and no X-Organization-Id header")
}

func fetchJWKS(ctx context.Context) (jwk.Set, error) {
	jwksMu.RLock()
	if jwksCache != nil && time.Now().Before(jwksExpiry) {
		cached := jwksCache
		jwksMu.RUnlock()
		return cached, nil
	}
	jwksMu.RUnlock()

	jwksMu.Lock()
	defer jwksMu.Unlock()

	if jwksCache != nil && time.Now().Before(jwksExpiry) {
		return jwksCache, nil
	}

	jwksURL := os.Getenv("AUTH_JWKS_URL")
	if jwksURL == "" {
		return nil, fmt.Errorf("AUTH_JWKS_URL not configured")
	}

	keySet, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS from %s: %w", jwksURL, err)
	}

	jwksCache = keySet
	jwksExpiry = time.Now().Add(jwksCacheTTL)

	return keySet, nil
}
