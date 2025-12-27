package version_manager

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	CurrentVersion = "v1"
	DefaultVersion = CurrentVersion
	VersionHeader  = "X-API-Version"
	VersionParam   = "api-version"
)

type Version struct {
	Version string
	Path    string
}

func NewVersion(version string) Version {
	return Version{
		Version: version,
		Path:    fmt.Sprintf("/api/%s", version),
	}
}

// GetVersionFromRequest extracts the API version from the request
func GetVersionFromRequest(r *http.Request) string {
	if version := r.Header.Get(VersionHeader); version != "" {
		return version
	}

	if version := r.URL.Query().Get(VersionParam); version != "" {
		return version
	}

	path := r.URL.Path
	if strings.HasPrefix(path, "/api/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	return DefaultVersion
}

// IsValidVersion checks if the version is valid
func IsValidVersion(version string) bool {
	return version != ""
}

// GetVersionedPath returns the versioned path for a given endpoint
func GetVersionedPath(version, endpoint string) string {
	return fmt.Sprintf("/api/%s%s", version, endpoint)
}

// VersionMiddleware is a middleware that handles API versioning
func VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := GetVersionFromRequest(r)

		if !IsValidVersion(version) {
			version = DefaultVersion
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "api_version", version)
		r = r.WithContext(ctx)

		w.Header().Set(VersionHeader, version)

		next.ServeHTTP(w, r)
	})
}
