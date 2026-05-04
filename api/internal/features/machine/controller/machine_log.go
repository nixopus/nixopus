package controller

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *MachineController) logMachineDebug(handler, reason, data string) {
	c.logger.Log(logger.Debug, fmt.Sprintf("machine: %s: %s", handler, reason), data)
}
