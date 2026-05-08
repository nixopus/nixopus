package routes

import (
	"github.com/go-fuego/fuego"
	credits_controller "github.com/nixopus/nixopus/api/internal/features/credits/controller"
)

func (router *Router) RegisterCreditsRoutes(group *fuego.Server, ctrl *credits_controller.CreditsController) {
	fuego.Get(group, "/balance", ctrl.GetBalance,
		fuego.OptionSummary("Get credit wallet balance"),
		fuego.OptionDescription("Returns the organization's current AI credit wallet balance in cents and USD. Returns credits_enabled: false in self-hosted mode."),
	)
	fuego.Get(group, "/usage", ctrl.GetUsage,
		fuego.OptionSummary("Get AI usage history"),
		fuego.OptionDescription("Returns aggregated AI usage history. Query params: period (7d|30d|90d, default 30d), groupBy (model|user|day, default day)."),
		fuego.OptionQuery("period", "Time range: 7d, 30d, or 90d"),
		fuego.OptionQuery("groupBy", "Group by: model, user, or day"),
	)
	fuego.Get(group, "/transactions", ctrl.GetTransactions,
		fuego.OptionSummary("Get wallet transaction ledger"),
		fuego.OptionDescription("Returns paginated wallet credits and debits. Query params: limit (default 20, max 100), offset (default 0)."),
		fuego.OptionQueryInt("limit", "Page size (default 20, max 100)"),
		fuego.OptionQueryInt("offset", "Page offset (default 0)"),
	)
	fuego.Get(group, "/usage-logs", ctrl.GetUsageLogs,
		fuego.OptionSummary("Get detailed AI usage logs"),
		fuego.OptionDescription("Returns paginated per-request AI usage logs with token counts and cost. Query params: limit (default 20, max 100), offset (default 0)."),
		fuego.OptionQueryInt("limit", "Page size (default 20, max 100)"),
		fuego.OptionQueryInt("offset", "Page offset (default 0)"),
	)
}
