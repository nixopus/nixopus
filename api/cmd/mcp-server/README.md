# MCP Server

The Model Context Protocol (MCP) server for Nixopus.

## Development

### Using Air (Live Reload)

Install Air if you haven't already:
```bash
go install github.com/air-verse/air@latest
```

Run the MCP server with live reload:
```bash
# From the api directory
make mcp-dev

# Or directly
air -c .air.mcp.toml
```

### Manual Development

Build and run:
```bash
make mcp-build
make mcp-run

# Or directly
go run ./cmd/mcp-server
```

## Testing

Use the MCP client to test the server:

```bash
# List available tools
make mcp-client

# Test a tool call
CONTAINER_ID=<id> ORGANIZATION_ID=<org-id> AUTH_TOKEN=<token> make mcp-client-test
```

## Features

- **Authentication**: SuperTokens JWT token verification
- **Authorization**: Organization-level access control
- **Tools**: Currently supports `get_container_logs`

## Configuration

The server requires:
- Database connection (via config)
- Redis connection (via config)
- SuperTokens initialization

All configuration is loaded from the standard Nixopus config files.

