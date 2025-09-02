package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/audit/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// AuditMiddleware is a middleware that captures audit logs for all authenticated requests
// resourceType explicitly defines what type of resource this route group operates on
func AuditMiddleware(next http.Handler, app *storage.App, l logger.Logger, resourceType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auditService := service.NewAuditService(app.Store.DB, app.Ctx, l)

		user, ok := r.Context().Value(types.UserContextKey).(*types.User)
		if !ok {
			l.Log(logger.Debug, "Audit middleware skipped", fmt.Sprintf("No user in context, path: %s", r.URL.Path))
			next.ServeHTTP(w, r)
			return
		}

		orgIDStr, ok := r.Context().Value(types.OrganizationIDKey).(string)
		if !ok {
			l.Log(logger.Debug, "Audit middleware skipped", fmt.Sprintf("No organization ID in context, path: %s, user_id: %s", r.URL.Path, user.ID))
			next.ServeHTTP(w, r)
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			l.Log(logger.Warning, "Audit middleware skipped", fmt.Sprintf("Invalid organization ID, path: %s, user_id: %s, org_id: %s, error: %s", r.URL.Path, user.ID, orgIDStr, err.Error()))
			next.ServeHTTP(w, r)
			return
		}

		auditAction := getAuditActionFromMethod(r.Method)

		// Skip audit logging for GET requests (access operations)
		if auditAction == types.AuditActionAccess {
			next.ServeHTTP(w, r)
			return
		}

		var requestBody map[string]interface{}
		if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				if len(bodyBytes) > 0 {
					json.Unmarshal(bodyBytes, &requestBody)
				}
			}
		}

		resourceID := extractResourceIDFromPath(r.URL.Path)

		auditResourceType := mapResourceType(resourceType)

		rw := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		if rw.statusCode >= 200 && rw.statusCode < 300 {
			auditReq := &service.AuditLogRequest{
				UserID:         user.ID,
				OrganizationID: orgID,
				Action:         auditAction,
				ResourceType:   auditResourceType,
				ResourceID:     resourceID,
				OldValues:      nil,
				NewValues:      requestBody,
				Metadata: map[string]interface{}{
					"path":          r.URL.Path,
					"method":        r.Method,
					"status":        rw.statusCode,
					"resource_type": resourceType,
					"endpoint":      getEndpointName(r.URL.Path),
				},
				IPAddress: r.RemoteAddr,
				UserAgent: r.UserAgent(),
				RequestID: uuid.New(),
			}

			// Fire audit logging asynchronously to not affect API response time
			go func() {
				if err := auditService.LogAction(auditReq); err != nil {
					l.Log(logger.Warning, "Failed to create audit log", fmt.Sprintf("path: %s, method: %s, user_id: %s, org_id: %s, error: %s",
						r.URL.Path, r.Method, user.ID, orgID, err.Error()))
				} else {
					l.Log(logger.Debug, "Audit log created", fmt.Sprintf("path: %s, method: %s, user_id: %s, org_id: %s",
						r.URL.Path, r.Method, user.ID, orgID))
				}
			}()
		}
	})
}

// auditResponseWriter captures the status code
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *auditResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getAuditActionFromMethod maps HTTP methods to audit actions
func getAuditActionFromMethod(method string) types.AuditAction {
	switch method {
	case http.MethodGet:
		return types.AuditActionAccess
	case http.MethodPost:
		return types.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return types.AuditActionUpdate
	case http.MethodDelete:
		return types.AuditActionDelete
	default:
		return types.AuditActionAccess
	}
}

// extractResourceIDFromPath extracts any UUID from the path (simple approach)
func extractResourceIDFromPath(path string) uuid.UUID {
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if id, err := uuid.Parse(segment); err == nil {
			return id
		}
	}
	return uuid.Nil
}

// mapResourceType converts string resource types to audit resource types
func mapResourceType(resourceType string) types.AuditResourceType {
	switch resourceType {
	case "user":
		return types.AuditResourceUser
	case "organization":
		return types.AuditResourceOrganization
	case "role":
		return types.AuditResourceRole
	case "permission":
		return types.AuditResourcePermission
	case "application":
		return types.AuditResourceApplication
	case "deploy", "deployment":
		return types.AuditResourceDeployment
	case "domain":
		return types.AuditResourceDomain
	case "github-connector", "github_connector":
		return types.AuditResourceGithubConnector
	case "smtp", "smtp_config":
		return types.AuditResourceSmtpConfig
	case "notification", "notifications":
		return types.AuditResourceNotification
	case "feature_flags", "feature-flags", "feature_flag":
		return types.AuditResourceFeatureFlag
	case "file-manager", "file_manager":
		return types.AuditResourceFileManager
	case "container":
		return types.AuditResourceContainer
	case "audit":
		return types.AuditResourceAudit
	case "terminal":
		return types.AuditResourceTerminal
	case "integration", "integrations":
		return types.AuditResourceIntegration
	default:
		return types.AuditResourceOrganization // Default fallback
	}
}

// getEndpointName extracts a human-readable endpoint name from path
func getEndpointName(path string) string {
	// Remove /api/v1/ prefix
	path = strings.TrimPrefix(path, "/api/v1/")

	// Take the first segment as endpoint name
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] != "" {
		return segments[0]
	}
	return "unknown"
}
