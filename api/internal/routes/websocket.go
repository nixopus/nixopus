package routes

import (
	"log"

	"github.com/go-fuego/fuego"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	"github.com/raghavyuva/nixopus-api/internal/realtime"
	"github.com/raghavyuva/nixopus-api/internal/scheduler"
)

// RegisterWebSocketRoutes registers WebSocket routes
func (router *Router) RegisterWebSocketRoutes(server *fuego.Server, deployController *deploy.DeployController, healthCheckScheduler *scheduler.HealthCheckScheduler) {
	wsServer, err := realtime.NewSocketServer(deployController, router.app.Store.DB, router.app.Ctx)
	if err != nil {
		log.Fatal(err)
	}
	router.socketServer = wsServer

	// Set SocketServer on health check scheduler for broadcasting results
	if healthCheckScheduler != nil {
		healthCheckScheduler.SetSocketServer(wsServer)
	}

	wsHandler := func(c fuego.ContextNoBody) (interface{}, error) {
		log.Printf("WebSocket connection attempt from: %s", c.Request().RemoteAddr)

		wsServer.HandleHTTP(c.Response(), c.Request())
		return nil, nil
	}

	fuego.Get(server, "/ws", wsHandler)
}

// GetSocketServer returns the WebSocket server instance
func (router *Router) GetSocketServer() *realtime.SocketServer {
	return router.socketServer
}
