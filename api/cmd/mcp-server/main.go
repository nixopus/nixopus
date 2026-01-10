package main

import (
	"context"
	"io"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/supertokens"
	mcp_pkg "github.com/raghavyuva/nixopus-api/internal/mcp"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	"github.com/raghavyuva/nixopus-api/internal/redisclient"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/vmihailenco/taskq/v3"
)

func main() {
	store := config.Init()
	ctx := context.Background()
	l := logger.NewLogger()

	// Initialize Redis client and queue
	redisClient, err := redisclient.New(config.AppConfig.Redis.URL)
	if err != nil {
		log.Fatalf("failed to create redis client for queue: %v", err)
	}

	taskq.SetLogger(log.New(io.Discard, "", 0))
	queue.Init(redisClient)

	// Initialize SuperTokens for authentication
	app := &storage.App{
		Store: store,
		Ctx:   ctx,
	}
	supertokens.Init(app)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nixopus",
		Version: "1.0.0",
	}, nil)

	// Register all tools
	mcp_pkg.RegisterTools(server, store, ctx, l)

	// Run the server over stdin/stdout, until the client disconnects
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
