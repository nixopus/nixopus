package routes

import (
	"github.com/go-fuego/fuego"
	trail "github.com/raghavyuva/nixopus-api/internal/features/trail/controller"
)

// RegisterTrailRoutes registers trail provisioning routes.
//
// Routes:
//   - POST /api/v1/trail/provision - Initiate trail provisioning
//   - GET /api/v1/trail/status/{sessionId} - Get provision status
func (router *Router) RegisterTrailRoutes(group *fuego.Server, controller *trail.TrailController) {
	fuego.Post(group, "/provision", controller.ProvisionTrail)
	fuego.Get(group, "/status/{sessionId}", controller.GetStatus)
}
