package routes

import (
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	agentController "github.com/nixopus/nixopus/api/internal/features/agent/controller"
)

func (router *Router) RegisterAgentRoutes(agentGroup *fuego.Server, controller *agentController.AgentController) {
	chatGroup := fuego.Group(agentGroup, "/chat", option.Tags("Agent"))
	fuego.Post(chatGroup, "", controller.Chat, fuego.OptionSummary("Chat with agent (blocking)"))
	fuego.Post(chatGroup, "/stream", controller.StreamChat, fuego.OptionSummary("Chat with agent (streaming SSE)"))
	fuego.Post(chatGroup, "/cancel", controller.CancelStream, fuego.OptionSummary("Cancel an active streaming response"))

	threadGroup := fuego.Group(agentGroup, "/threads", option.Tags("Agent"))
	fuego.Post(threadGroup, "", controller.CreateThread, fuego.OptionSummary("Create a new conversation thread"))
	fuego.Get(threadGroup, "", controller.ListThreads, fuego.OptionSummary("List conversation threads"))
	fuego.Get(threadGroup, "/{threadId}/messages", controller.GetThreadMessages, fuego.OptionSummary("Get thread messages"))
	fuego.Patch(threadGroup, "/{threadId}", controller.UpdateThread, fuego.OptionSummary("Update thread title"))
	fuego.Delete(threadGroup, "/{threadId}", controller.DeleteThread, fuego.OptionSummary("Delete a conversation thread"))

	scheduleGroup := fuego.Group(agentGroup, "/schedules", option.Tags("Agent"))
	fuego.Get(scheduleGroup, "", controller.ListSchedules, fuego.OptionSummary("List scheduled tasks"))
	fuego.Get(scheduleGroup, "/{scheduleId}", controller.GetSchedule, fuego.OptionSummary("Get a scheduled task"))
	fuego.Get(scheduleGroup, "/{scheduleId}/runs", controller.GetScheduleRuns, fuego.OptionSummary("Get schedule run history"))
	fuego.Patch(scheduleGroup, "/{scheduleId}", controller.UpdateScheduleStatus, fuego.OptionSummary("Update schedule status (pause/resume)"))
	fuego.Delete(scheduleGroup, "/{scheduleId}", controller.DeleteSchedule, fuego.OptionSummary("Delete a scheduled task"))
}
