package routes

import (
	"github.com/go-fuego/fuego"
	audit "github.com/raghavyuva/nixopus-api/internal/features/audit/controller"
)

// RegisterAuditRoutes registers audit routes
func (router *Router) RegisterAuditRoutes(auditGroup *fuego.Server, auditController *audit.AuditController) {
	fuego.Get(auditGroup, "/logs", auditController.GetRecentAuditLogs)
}
