package routes

import (
	"github.com/go-fuego/fuego"
	machine_controller "github.com/raghavyuva/nixopus-api/internal/features/machine/controller"
)

func (router *Router) RegisterMachineRoutes(machineGroup *fuego.Server, machineController *machine_controller.MachineController) {
	fuego.Get(
		machineGroup,
		"/stats",
		machineController.GetSystemStats,
		fuego.OptionSummary("Get machine system stats"),
	)
	fuego.Post(
		machineGroup,
		"/exec",
		machineController.ExecCommand,
		fuego.OptionSummary("Execute a command on the host machine"),
	)
}
