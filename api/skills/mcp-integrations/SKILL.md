---
name: mcp-integrations
description: MCP server discovery, tool invocation, and provider catalog integration. Load when a task involves external services, third-party tools, or when the user asks about MCP servers.
metadata:
  version: "1.0"
---

# MCP Integrations

When a task involves external services, third-party tools, or capabilities beyond core Nixopus (e.g. databases, monitoring, CI/CD, analytics, logging, storage, auth providers), proactively check whether an MCP integration can help.

**These are runtime tools, not skill files. Use `search_tools` / `load_tool` to find and load them — do NOT use `skill_read` or `skill_search`.**

1. Run `search_tools("mcp")` then `load_tool(...)` to load MCP tools.
2. Call `discover_mcp_tools` to list tools from all enabled MCP servers. Each tool entry includes server_id, tool name, description, and inputSchema.
3. Call `call_mcp_tool` to invoke a specific tool: pass server_id (UUID from discover_mcp_tools), tool_name (exact name string), and arguments (a JSON object matching the tool's inputSchema — use proper types: strings, numbers, booleans, not everything as strings).
4. If no relevant integration exists, call `list_mcp_provider_catalog` to show what integrations the user can enable.

Also use these tools when the user explicitly asks about MCP servers — list, add, update, delete, or test connections.
