package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/supertokens"
	mcp_pkg "github.com/raghavyuva/nixopus-api/internal/mcp"
	"github.com/raghavyuva/nixopus-api/internal/storage"
)

func main() {
	store := config.Init()
	ctx := context.Background()
	l := logger.NewLogger()

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
