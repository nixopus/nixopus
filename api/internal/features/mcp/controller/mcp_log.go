package controller

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *MCPController) logMCPDebug(handler, reason, data string) {
	c.logger.Log(logger.Debug, fmt.Sprintf("mcp: %s: %s", handler, reason), data)
}
