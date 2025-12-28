package routes

import (
	"github.com/go-fuego/fuego"
	health "github.com/raghavyuva/nixopus-api/internal/features/health"
	// api "github.com/raghavyuva/nixopus-api/internal/version" // Commented out - version manager disabled
)

// RegisterHealthRoutes registers health check and version routes
func (router *Router) RegisterHealthRoutes(healthGroup *fuego.Server) {
	fuego.Get(healthGroup, "", health.HealthCheck)
	// Commented out - version manager related endpoint
	// versionGroup := fuego.Group(healthGroup, "/versions")
	// fuego.Get(versionGroup, "", func(c fuego.ContextNoBody) (interface{}, error) {
	// 	docs := api.NewVersionDocumentation()
	// 	if err := docs.Load("api/versions.json"); err != nil {
	// 		return nil, err
	// 	}
	// 	return docs, nil
	// })
}
