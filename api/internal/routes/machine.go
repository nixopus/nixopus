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
	fuego.Get(
		machineGroup,
		"/plans",
		machineController.ListMachinePlans,
		fuego.OptionSummary("List available machine plans"),
		fuego.OptionDescription("Returns all active machine plans with pricing, specs, and tier information."),
	)
	fuego.Post(
		machineGroup,
		"/plan/select",
		machineController.SelectMachinePlan,
		fuego.OptionSummary("Select a machine plan"),
		fuego.OptionDescription("Select a machine plan for the organization. Deducts the monthly cost from the wallet immediately. Requires sufficient wallet balance."),
	)
	fuego.Get(
		machineGroup,
		"/billing",
		machineController.GetMachineBilling,
		fuego.OptionSummary("Get machine billing status"),
		fuego.OptionDescription("Returns the current machine billing status, plan details, and any grace period warnings for the organization."),
	)
}
