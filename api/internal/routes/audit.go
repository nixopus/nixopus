package routes

import (
	"github.com/go-fuego/fuego"
	audit "github.com/nixopus/nixopus/api/internal/features/audit/controller"
)

// RegisterAuditRoutes registers audit routes
func (router *Router) RegisterAuditRoutes(auditGroup *fuego.Server, auditController *audit.AuditController) {
	fuego.Get(
		auditGroup,
		"/logs",
		auditController.GetRecentAuditLogs,
		fuego.OptionSummary("List audit logs"),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
		fuego.OptionQuery("search", "Search text"),
		fuego.OptionQuery("resource_type", "Resource type filter"),
	)
}
