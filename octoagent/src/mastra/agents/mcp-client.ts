import { MCPClient } from "@mastra/mcp";
import { fileURLToPath } from "url";
import { dirname, join, resolve } from "path";
import { config } from "../../config";


// TODO: Remove this once we have a proper way to get the MCP server path even in production mode
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const apiDir = resolve(__dirname, "../../../api");
const mcpServerPath = config.mcpServerPath || join(apiDir, "nixopus-mcp-server");

export const nixopusMcpClient = new MCPClient({
  id: "nixopus-mcp-client",
  servers: {
    nixopus: {
      command: mcpServerPath,
      args: [],
      env: {
        AUTH_TOKEN: config.authToken,
        ...process.env,
      },
    },
  },
});

