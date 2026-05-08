package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
)

// ToolDeps provides dependencies for schedule tools.
// Scheduler is a function so it resolves after AgentService assigns s.scheduler
// (registerAgents runs before the scheduler is constructed).
type ToolDeps struct {
	Store *Store
	// Scheduler returns the live scheduler; nil if not yet wired (should not happen after init).
	Scheduler func() *Scheduler
	Memory    *memory.PostgresStore
	Logger    logger.Logger
	UserID    func(ctx context.Context) string
	OrgID     func(ctx context.Context) string
}

func CreateScheduleTool(deps ToolDeps) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "create_schedule",
		Description: "Create a recurring scheduled task. The agent will run the prompt on the given cron schedule and deliver results via the specified notification channel. Use standard 5-field cron expressions (minute hour day-of-month month day-of-week) or descriptors like @every 30m, @hourly, @daily.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":             {"type": "string", "description": "Human-readable name for this schedule (e.g. 'App-1 Log Monitor')"},
				"prompt":           {"type": "string", "description": "The task prompt the agent will execute each run. Be specific about what to check, what tools to use, and what to report."},
				"cron_expression":  {"type": "string", "description": "Cron schedule expression. Examples: '*/30 * * * *' (every 30 min), '0 9 * * *' (daily 9 AM), '@every 1h', '@daily'"},
				"timezone":         {"type": "string", "description": "IANA timezone for the schedule (default: UTC). Example: 'America/New_York'"},
				"delivery_channel": {"type": "string", "enum": ["slack", "discord", "email", "all"], "description": "Notification channel to deliver results. Use 'all' to send to all configured channels."},
				"description":      {"type": "string", "description": "Optional description of what this schedule does"},
				"max_runs":         {"type": "integer", "description": "Maximum number of runs before auto-completing (omit for unlimited)"},
				"timeout_seconds":  {"type": "integer", "description": "Maximum execution time per run in seconds (default: 300)"},
				"notify_on":        {"type": "string", "enum": ["always", "failure_only", "success_only", "smart", "never"], "description": "When to send notifications. 'smart' (recommended) = the agent decides per-run whether findings are worth alerting — it skips if nothing changed or no issues found. 'failure_only' = only on errors/timeouts. 'always' = every run. 'never' = silent. Default: 'smart'."}
			},
			"required": ["name", "prompt", "cron_expression", "delivery_channel"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return createScheduleHandler(ctx, args, deps)
		},
	}
}

func createScheduleHandler(ctx context.Context, args json.RawMessage, deps ToolDeps) (json.RawMessage, error) {
	l := deps.Logger

	var input struct {
		Name            string `json:"name"`
		Prompt          string `json:"prompt"`
		CronExpression  string `json:"cron_expression"`
		Timezone        string `json:"timezone"`
		DeliveryChannel string `json:"delivery_channel"`
		Description     string `json:"description"`
		MaxRuns         *int   `json:"max_runs"`
		TimeoutSeconds  int    `json:"timeout_seconds"`
		NotifyOn        string `json:"notify_on"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		l.Log(logger.Error, "create_schedule: failed to unmarshal args", err.Error())
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	l.Log(logger.Debug, "create_schedule: invoked",
		fmt.Sprintf("name=%q cron=%q channel=%s", input.Name, input.CronExpression, input.DeliveryChannel))

	if input.Name == "" || input.Prompt == "" || input.CronExpression == "" {
		l.Log(logger.Warning, "create_schedule: missing required fields",
			fmt.Sprintf("name=%q prompt_len=%d cron=%q", input.Name, len(input.Prompt), input.CronExpression))
		return nil, fmt.Errorf("name, prompt, and cron_expression are required")
	}

	sch := deps.Scheduler()
	if sch == nil {
		l.Log(logger.Error, "create_schedule: scheduler is nil", "")
		return nil, fmt.Errorf("scheduler not initialized")
	}
	if err := sch.ValidateCron(input.CronExpression); err != nil {
		l.Log(logger.Warning, "create_schedule: invalid cron",
			fmt.Sprintf("expr=%q err=%s", input.CronExpression, err.Error()))
		return json.Marshal(map[string]interface{}{
			"error":   "invalid cron expression",
			"detail":  err.Error(),
			"example": "Use 5-field cron (e.g. '*/30 * * * *') or descriptors like '@every 30m', '@hourly', '@daily'",
		})
	}

	userID := deps.UserID(ctx)
	orgID := deps.OrgID(ctx)

	l.Log(logger.Debug, "create_schedule: resolved context",
		fmt.Sprintf("user_id=%q org_id=%q", userID, orgID))

	if userID == "" {
		l.Log(logger.Error, "create_schedule: empty user_id from context", "")
		return nil, fmt.Errorf("create schedule: missing user context; schedule tools need an authenticated chat session")
	}
	if orgID == "" {
		l.Log(logger.Error, "create_schedule: empty org_id from context", "")
		return nil, fmt.Errorf("create schedule: missing organization context")
	}
	if _, err := uuid.Parse(userID); err != nil {
		l.Log(logger.Error, "create_schedule: user_id is not a valid UUID",
			fmt.Sprintf("user_id=%q err=%s", userID, err.Error()))
		return nil, fmt.Errorf("create schedule: invalid user_id in context: %w", err)
	}
	if _, err := uuid.Parse(orgID); err != nil {
		l.Log(logger.Error, "create_schedule: org_id is not a valid UUID",
			fmt.Sprintf("org_id=%q err=%s", orgID, err.Error()))
		return nil, fmt.Errorf("create schedule: invalid org_id in context: %w", err)
	}

	if input.Timezone == "" {
		input.Timezone = "UTC"
	}
	if input.TimeoutSeconds == 0 {
		input.TimeoutSeconds = 300
	}
	switch input.NotifyOn {
	case "always", "failure_only", "success_only", "smart", "never":
		break
	default:
		input.NotifyOn = "smart"
	}

	threadID := uuid.New().String()
	thread := &memory.Thread{
		ID:     threadID,
		UserID: userID,
		Title:  fmt.Sprintf("Schedule: %s", input.Name),
		Metadata: memory.JSONMap{
			"org_id": orgID,
			"type":   "schedule",
		},
	}
	if err := deps.Memory.CreateThread(ctx, thread); err != nil {
		l.Log(logger.Error, "create_schedule: failed to create thread",
			fmt.Sprintf("thread_id=%s err=%s", threadID, err.Error()))
		return nil, fmt.Errorf("create schedule thread: %w", err)
	}

	l.Log(logger.Debug, "create_schedule: thread created", fmt.Sprintf("thread_id=%s", threadID))

	schedule := &Schedule{
		UserID:          userID,
		OrgID:           orgID,
		ThreadID:        threadID,
		Name:            input.Name,
		Description:     input.Description,
		Prompt:          input.Prompt,
		CronExpression:  input.CronExpression,
		Timezone:        input.Timezone,
		Status:          StatusActive,
		DeliveryChannel: input.DeliveryChannel,
		NotifyOn:        input.NotifyOn,
		MaxRuns:         input.MaxRuns,
		TimeoutSeconds:  input.TimeoutSeconds,
		Metadata:        JSONMetadata{},
	}

	nextRun := sch.computeNextRun(*schedule)
	schedule.NextRunAt = nextRun

	if err := deps.Store.CreateSchedule(ctx, schedule); err != nil {
		l.Log(logger.Error, "create_schedule: DB insert failed",
			fmt.Sprintf("name=%q user_id=%s org_id=%s err=%s", input.Name, userID, orgID, err.Error()))
		return nil, fmt.Errorf("create schedule: %w", err)
	}

	var nextRunStr string
	if nextRun != nil {
		nextRunStr = nextRun.Format(time.RFC3339)
	}

	l.Log(logger.Info, "create_schedule: success",
		fmt.Sprintf("schedule_id=%s name=%q cron=%q next_run=%s user_id=%s org_id=%s",
			schedule.ID, schedule.Name, schedule.CronExpression, nextRunStr, userID, orgID))

	return json.Marshal(map[string]interface{}{
		"schedule_id":     schedule.ID.String(),
		"name":            schedule.Name,
		"cron_expression": schedule.CronExpression,
		"timezone":        schedule.Timezone,
		"next_run_at":     nextRunStr,
		"delivery":        schedule.DeliveryChannel,
		"notify_on":       schedule.NotifyOn,
		"status":          "active",
		"message":         fmt.Sprintf("Schedule '%s' created. Next run at %s.", schedule.Name, nextRunStr),
	})
}

func ListSchedulesTool(deps ToolDeps) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "list_schedules",
		Description: "List all scheduled tasks for the current user. Shows active, paused, and recently completed schedules.",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return listSchedulesHandler(ctx, deps)
		},
	}
}

func listSchedulesHandler(ctx context.Context, deps ToolDeps) (json.RawMessage, error) {
	l := deps.Logger
	userID := deps.UserID(ctx)
	orgID := deps.OrgID(ctx)

	l.Log(logger.Debug, "list_schedules: invoked",
		fmt.Sprintf("user_id=%q org_id=%q", userID, orgID))

	schedules, err := deps.Store.ListSchedulesForUser(ctx, userID, orgID)
	if err != nil {
		l.Log(logger.Error, "list_schedules: query failed",
			fmt.Sprintf("user_id=%s org_id=%s err=%s", userID, orgID, err.Error()))
		return nil, fmt.Errorf("list schedules: %w", err)
	}

	l.Log(logger.Debug, "list_schedules: result", fmt.Sprintf("count=%d", len(schedules)))

	type scheduleInfo struct {
		ID              string  `json:"id"`
		Name            string  `json:"name"`
		CronExpression  string  `json:"cron_expression"`
		Status          string  `json:"status"`
		DeliveryChannel string  `json:"delivery_channel"`
		RunCount        int     `json:"run_count"`
		LastRunAt       *string `json:"last_run_at"`
		NextRunAt       *string `json:"next_run_at"`
		ErrorCount      int     `json:"error_count"`
	}

	items := make([]scheduleInfo, 0, len(schedules))
	for _, s := range schedules {
		info := scheduleInfo{
			ID:              s.ID.String(),
			Name:            s.Name,
			CronExpression:  s.CronExpression,
			Status:          string(s.Status),
			DeliveryChannel: s.DeliveryChannel,
			RunCount:        s.RunCount,
			ErrorCount:      s.ErrorCount,
		}
		if s.LastRunAt != nil {
			t := s.LastRunAt.Format(time.RFC3339)
			info.LastRunAt = &t
		}
		if s.NextRunAt != nil {
			t := s.NextRunAt.Format(time.RFC3339)
			info.NextRunAt = &t
		}
		items = append(items, info)
	}

	return json.Marshal(map[string]interface{}{
		"schedules": items,
		"count":     len(items),
	})
}

func DeleteScheduleTool(deps ToolDeps) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "delete_schedule",
		Description: "Delete a scheduled task by ID. This stops all future runs.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"schedule_id": {"type": "string", "description": "The schedule ID to delete"}
			},
			"required": ["schedule_id"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return deleteScheduleHandler(ctx, args, deps)
		},
	}
}

func deleteScheduleHandler(ctx context.Context, args json.RawMessage, deps ToolDeps) (json.RawMessage, error) {
	l := deps.Logger
	var input struct {
		ScheduleID string `json:"schedule_id"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		l.Log(logger.Error, "delete_schedule: failed to unmarshal args", err.Error())
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	l.Log(logger.Debug, "delete_schedule: invoked", fmt.Sprintf("schedule_id=%q", input.ScheduleID))

	scheduleID, err := uuid.Parse(input.ScheduleID)
	if err != nil {
		l.Log(logger.Warning, "delete_schedule: invalid UUID",
			fmt.Sprintf("schedule_id=%q err=%s", input.ScheduleID, err.Error()))
		return nil, fmt.Errorf("invalid schedule_id: %w", err)
	}

	userID := deps.UserID(ctx)
	if err := deps.Store.SoftDeleteSchedule(ctx, scheduleID, userID); err != nil {
		l.Log(logger.Error, "delete_schedule: DB delete failed",
			fmt.Sprintf("schedule_id=%s user_id=%s err=%s", scheduleID, userID, err.Error()))
		return nil, fmt.Errorf("delete schedule: %w", err)
	}

	l.Log(logger.Info, "delete_schedule: success",
		fmt.Sprintf("schedule_id=%s user_id=%s", scheduleID, userID))

	return json.Marshal(map[string]interface{}{
		"deleted":     true,
		"schedule_id": input.ScheduleID,
		"message":     "Schedule deleted. No further runs will be executed.",
	})
}

func PauseScheduleTool(deps ToolDeps) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "pause_schedule",
		Description: "Pause a scheduled task. The schedule can be resumed later.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"schedule_id": {"type": "string", "description": "The schedule ID to pause"}
			},
			"required": ["schedule_id"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return toggleScheduleHandler(ctx, args, deps, StatusPaused)
		},
	}
}

func ResumeScheduleTool(deps ToolDeps) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "resume_schedule",
		Description: "Resume a paused scheduled task.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"schedule_id": {"type": "string", "description": "The schedule ID to resume"}
			},
			"required": ["schedule_id"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return toggleScheduleHandler(ctx, args, deps, StatusActive)
		},
	}
}

func toggleScheduleHandler(ctx context.Context, args json.RawMessage, deps ToolDeps, target ScheduleStatus) (json.RawMessage, error) {
	l := deps.Logger
	var input struct {
		ScheduleID string `json:"schedule_id"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		l.Log(logger.Error, "toggle_schedule: failed to unmarshal args", err.Error())
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	l.Log(logger.Debug, "toggle_schedule: invoked",
		fmt.Sprintf("schedule_id=%q target=%s", input.ScheduleID, target))

	scheduleID, err := uuid.Parse(input.ScheduleID)
	if err != nil {
		l.Log(logger.Warning, "toggle_schedule: invalid UUID",
			fmt.Sprintf("schedule_id=%q err=%s", input.ScheduleID, err.Error()))
		return nil, fmt.Errorf("invalid schedule_id: %w", err)
	}

	userID := deps.UserID(ctx)
	sched, err := deps.Store.GetScheduleForUser(ctx, scheduleID, userID)
	if err != nil {
		l.Log(logger.Error, "toggle_schedule: schedule not found",
			fmt.Sprintf("schedule_id=%s user_id=%s err=%s", scheduleID, userID, err.Error()))
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	if err := deps.Store.UpdateScheduleStatus(ctx, sched.ID, target); err != nil {
		l.Log(logger.Error, "toggle_schedule: status update failed",
			fmt.Sprintf("schedule_id=%s target=%s err=%s", sched.ID, target, err.Error()))
		return nil, fmt.Errorf("update schedule: %w", err)
	}

	if target == StatusActive {
		sch := deps.Scheduler()
		if sch == nil {
			l.Log(logger.Error, "toggle_schedule: scheduler is nil on resume", "")
			return nil, fmt.Errorf("scheduler not initialized")
		}
		nextRun := sch.computeNextRun(*sched)
		if nextRun != nil {
			deps.Store.DB.NewUpdate().
				Model((*Schedule)(nil)).
				Set("next_run_at = ?", nextRun).
				Set("consecutive_error_count = 0").
				Where("id = ?", sched.ID).
				Exec(ctx)
			l.Log(logger.Debug, "toggle_schedule: recomputed next_run",
				fmt.Sprintf("schedule_id=%s next_run=%s", sched.ID, nextRun.Format(time.RFC3339)))
		}
	}

	action := "paused"
	if target == StatusActive {
		action = "resumed"
	}

	l.Log(logger.Info, fmt.Sprintf("toggle_schedule: %s", action),
		fmt.Sprintf("schedule_id=%s name=%q user_id=%s", sched.ID, sched.Name, userID))

	return json.Marshal(map[string]interface{}{
		"schedule_id": input.ScheduleID,
		"status":      string(target),
		"message":     fmt.Sprintf("Schedule '%s' %s.", sched.Name, action),
	})
}
