package routes

import (
	"github.com/go-fuego/fuego"
	machine_controller "github.com/nixopus/nixopus/api/internal/features/machine/controller"
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
		"/status",
		machineController.GetMachineStatus,
		fuego.OptionSummary("Get machine lifecycle status"),
		fuego.OptionDescription("Returns the current state of the provisioned machine instance (active, paused, etc)."),
	)
	fuego.Post(
		machineGroup,
		"/restart",
		machineController.RestartMachine,
		fuego.OptionSummary("Restart machine"),
		fuego.OptionDescription("Restarts the provisioned machine instance."),
	)
	fuego.Post(
		machineGroup,
		"/pause",
		machineController.PauseMachine,
		fuego.OptionSummary("Pause machine"),
		fuego.OptionDescription("Pauses the provisioned machine instance."),
	)
	fuego.Post(
		machineGroup,
		"/resume",
		machineController.ResumeMachine,
		fuego.OptionSummary("Resume machine"),
		fuego.OptionDescription("Resumes a paused machine instance."),
	)
	fuego.Post(
		machineGroup,
		"/backup",
		machineController.TriggerBackup,
		fuego.OptionSummary("Trigger machine backup"),
		fuego.OptionDescription("Initiates an async backup of the provisioned machine (snapshot + S3 upload). Returns immediately; poll GET /machine/backups for status."),
	)
	fuego.Get(
		machineGroup,
		"/backups",
		machineController.ListBackups,
		fuego.OptionSummary("List machine backups"),
		fuego.OptionDescription("Returns the backup history for the organization's provisioned machine."),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
		fuego.OptionQuery("search", "Search by machine name"),
		fuego.OptionQuery("sort_by", "Sort field (created_at, status, size_bytes)"),
		fuego.OptionQuery("sort_order", "Sort order (asc, desc)"),
		fuego.OptionQuery("status", "Filter by backup status"),
	)
	fuego.Get(
		machineGroup,
		"/backup/schedule",
		machineController.GetBackupSchedule,
		fuego.OptionSummary("Get backup schedule"),
		fuego.OptionDescription("Returns the automatic backup schedule configuration for the organization."),
	)
	fuego.Put(
		machineGroup,
		"/backup/schedule",
		machineController.UpdateBackupSchedule,
		fuego.OptionSummary("Update backup schedule"),
		fuego.OptionDescription("Updates the automatic backup schedule (enable/disable, frequency, time)."),
	)
	fuego.Get(
		machineGroup,
		"/metrics",
		machineController.GetMachineMetrics,
		fuego.OptionSummary("Get machine metrics"),
		fuego.OptionDescription("Returns time-series CPU, memory, and network metrics. Query params: from, to (RFC3339), limit."),
		fuego.OptionQuery("from", "Start time RFC3339 (default: 1 hour ago)"),
		fuego.OptionQuery("to", "End time RFC3339 (default: now)"),
		fuego.OptionQuery("limit", "Max rows (default 500, max 1000)"),
	)
	fuego.Get(
		machineGroup,
		"/events",
		machineController.GetMachineEvents,
		fuego.OptionSummary("Get machine events"),
		fuego.OptionDescription("Returns DoS and bandwidth breach events. Query params: from, to (RFC3339), limit."),
		fuego.OptionQuery("from", "Start time RFC3339 (default: 1 hour ago)"),
		fuego.OptionQuery("to", "End time RFC3339 (default: now)"),
		fuego.OptionQuery("limit", "Max rows (default 200)"),
	)
	fuego.Get(
		machineGroup,
		"/metrics/summary",
		machineController.GetMachineMetricsSummary,
		fuego.OptionSummary("Get machine metrics summary"),
		fuego.OptionDescription("Returns an LLM-friendly aggregate summary of machine health, traffic, and events."),
		fuego.OptionQuery("from", "Start time RFC3339 (default: 1 hour ago)"),
		fuego.OptionQuery("to", "End time RFC3339 (default: now)"),
	)
}

func (router *Router) RegisterMachineBillingRoutes(billingGroup *fuego.Server, machineController *machine_controller.MachineController) {
	fuego.Get(
		billingGroup,
		"/plans",
		machineController.ListMachinePlans,
		fuego.OptionSummary("List available machine plans"),
		fuego.OptionDescription("Returns all active machine plans with pricing, specs, and tier information."),
	)
	fuego.Post(
		billingGroup,
		"/plan/select",
		machineController.SelectMachinePlan,
		fuego.OptionSummary("Select a machine plan"),
		fuego.OptionDescription("Select a machine plan for the organization. Deducts the monthly cost from the wallet immediately. Requires sufficient wallet balance."),
	)
	fuego.Get(
		billingGroup,
		"/billing",
		machineController.GetMachineBilling,
		fuego.OptionSummary("Get machine billing status"),
		fuego.OptionDescription("Returns the current machine billing status, plan details, and any grace period warnings for the organization."),
	)
}
