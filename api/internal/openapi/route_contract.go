package openapi

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/go-fuego/fuego"
)

var (
	operationIDMu     sync.Mutex
	operationIDCounts = map[string]int{}
	wordSplitRe       = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// RouteContractOption applies stable operation IDs and concise descriptions
// across all routes to improve OpenAPI usability for LLM clients.
func RouteContractOption() fuego.RouteOption {
	return func(r *fuego.BaseRoute) {
		summary := strings.TrimSpace(r.Operation.Summary)
		if summary == "" {
			summary = fallbackSummary(r.Method, r.Path)
			fuego.OptionSummary(summary)(r)
		}

		opID := toLowerCamel(summary)
		if opID == "" {
			opID = toLowerCamel(fallbackSummary(r.Method, r.Path))
		}
		if opID == "" {
			opID = strings.ToLower(r.Method) + "Operation"
		}
		opID = ensureUniqueOperationID(opID)
		fuego.OptionOperationID(opID)(r)

		if len(r.Operation.Tags) == 0 {
			tag := inferPrimaryTag(r.Path)
			if tag != "" {
				fuego.OptionTags(tag)(r)
			}
		}

		description := buildDescription(summary, r.Method, r.Path)
		fuego.OptionOverrideDescription(description)(r)
	}
}

func ensureUniqueOperationID(base string) string {
	operationIDMu.Lock()
	defer operationIDMu.Unlock()

	operationIDCounts[base]++
	if operationIDCounts[base] == 1 {
		return base
	}
	return fmt.Sprintf("%s%d", base, operationIDCounts[base])
}

func fallbackSummary(method, rawPath string) string {
	action := actionForMethod(method)
	resource := inferResourceName(rawPath)
	if resource == "" {
		resource = "endpoint"
	}
	return strings.TrimSpace(action + " " + resource)
}

func actionForMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "Get"
	case "POST":
		return "Create"
	case "PUT":
		return "Update"
	case "PATCH":
		return "Patch"
	case "DELETE":
		return "Delete"
	default:
		return "Use"
	}
}

func inferResourceName(rawPath string) string {
	segments := splitPath(rawPath)
	if len(segments) == 0 {
		return ""
	}

	// Drop API prefix when present.
	if len(segments) >= 2 && segments[0] == "api" && strings.HasPrefix(segments[1], "v") {
		segments = segments[2:]
	}
	if len(segments) == 0 {
		return ""
	}

	last := segments[len(segments)-1]
	if isPathParam(last) && len(segments) > 1 {
		last = segments[len(segments)-2]
	}

	last = strings.ReplaceAll(last, "-", " ")
	last = strings.ReplaceAll(last, "_", " ")
	return strings.TrimSpace(last)
}

func inferPrimaryTag(rawPath string) string {
	segments := splitPath(rawPath)
	if len(segments) >= 3 && segments[0] == "api" && strings.HasPrefix(segments[1], "v") {
		return segments[2]
	}
	if len(segments) > 0 {
		return segments[0]
	}
	return ""
}

func buildDescription(summary, method, rawPath string) string {
	auth := "Required (bearer token)."
	if isPublicPath(rawPath) {
		auth = "Public endpoint."
	}

	scope := "Organization-scoped in authenticated context."
	if strings.Contains(rawPath, "/auth/") || isPublicPath(rawPath) {
		scope = "No organization scope required."
	}

	sideEffects := "Read-only operation."
	switch strings.ToUpper(method) {
	case "POST", "PUT", "PATCH", "DELETE":
		sideEffects = "May mutate server state."
	}

	return fmt.Sprintf(
		"%s\n\nAuth: %s\nScope: %s\nSide effects: %s",
		strings.TrimSpace(summary)+".",
		auth,
		scope,
		sideEffects,
	)
}

func isPublicPath(rawPath string) bool {
	publicPrefixes := []string{"/api/v1/health/", "/api/v1/webhook/", "/api/v1/live/", "/ws/"}
	publicExact := map[string]bool{
		"/api/v1/health":  true,
		"/api/v1/webhook": true,
		"/api/v1/live":    true,
		"/ws":             true,
	}
	if publicExact[rawPath] {
		return true
	}
	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(rawPath, prefix) {
			return true
		}
	}
	if rawPath == "/api/v1/auth/is-admin-registered" {
		return true
	}
	return false
}

func splitPath(rawPath string) []string {
	clean := path.Clean(rawPath)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || clean == "" {
		return nil
	}
	return strings.Split(clean, "/")
}

func isPathParam(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
}

func toLowerCamel(input string) string {
	parts := wordSplitRe.Split(input, -1)
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		return ""
	}

	for i := range filtered {
		part := filtered[i]
		if i == 0 {
			filtered[i] = strings.ToLower(part[:1]) + part[1:]
		} else {
			filtered[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(filtered, "")
}
