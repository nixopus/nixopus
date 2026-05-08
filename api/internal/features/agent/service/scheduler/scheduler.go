package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/robfig/cron/v3"
)

var (
	// tickInterval is how often due schedules are polled (tests may shorten this).
	tickInterval = 30 * time.Second
)

const (
	defaultTimeout  = 5 * time.Minute
	maxResultLength = 4000

	smartNotifyInstruction = `[notification-decision]
You MUST begin your response with exactly one of these two tags on the first line:
- [ALERT] — if you found NEW issues, errors, anomalies, or meaningful changes since the last run. The user will be notified.
- [OK] — if everything looks normal, unchanged, or the same as your last check. The user will NOT be notified.

Compare your current findings with your previous responses in this thread. Only use [ALERT] when there is genuinely new or changed information worth the user's attention. Do NOT alert for the same ongoing situation you already reported unless it has worsened or changed.
[/notification-decision]`
)

// AgentRunner executes an agent turn for a scheduled task.
// Implemented by AgentService to avoid circular imports.
type AgentRunner interface {
	RunScheduledTask(ctx context.Context, prompt string, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error)
}

// NotificationSender sends delivery notifications for completed runs.
type NotificationSender interface {
	Emit(event types.NotificationEvent) error
}

// Scheduler runs scheduled agent tasks on a timer.
type Scheduler struct {
	store    *Store
	memory   *memory.PostgresStore
	runner   AgentRunner
	notifier NotificationSender
	logger   logger.Logger
	parser   cron.Parser

	mu      sync.Mutex
	running map[uuid.UUID]context.CancelFunc
	stopCh  chan struct{}
}

func New(store *Store, mem *memory.PostgresStore, runner AgentRunner, notifier NotificationSender, l logger.Logger) *Scheduler {
	return &Scheduler{
		store:    store,
		memory:   mem,
		runner:   runner,
		notifier: notifier,
		logger:   l,
		parser:   cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		running:  make(map[uuid.UUID]context.CancelFunc),
		stopCh:   make(chan struct{}),
	}
}

// Start begins the scheduler loop in a background goroutine.
func (s *Scheduler) Start(ctx context.Context) {
	go s.loop(ctx)
	s.logger.Log(logger.Info, "scheduler: started", fmt.Sprintf("tick_interval=%s", tickInterval))
}

// Stop signals the scheduler loop to exit.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.mu.Lock()
	for id, cancel := range s.running {
		cancel()
		delete(s.running, id)
	}
	s.mu.Unlock()
	s.logger.Log(logger.Info, "scheduler: stopped", "")
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	now := time.Now()
	schedules, err := s.store.GetDueSchedules(ctx, now)
	if err != nil {
		s.logger.Log(logger.Error, "scheduler: tick: failed to get due schedules", err.Error())
		return
	}

	if len(schedules) > 0 {
		s.logger.Log(logger.Debug, "scheduler: tick",
			fmt.Sprintf("due_count=%d now=%s", len(schedules), now.Format(time.RFC3339)))
	}

	for _, sched := range schedules {
		if sched.MaxRuns != nil && sched.RunCount >= *sched.MaxRuns {
			s.logger.Log(logger.Info, "scheduler: schedule completed (max_runs reached)",
				fmt.Sprintf("schedule=%s name=%q run_count=%d max_runs=%d", sched.ID, sched.Name, sched.RunCount, *sched.MaxRuns))
			_ = s.store.UpdateScheduleStatus(ctx, sched.ID, StatusCompleted)
			continue
		}
		if sched.ConsecutiveErrorCount >= sched.MaxConsecutiveErrors {
			s.logger.Log(logger.Warning, "scheduler: disabling schedule (too many consecutive errors)",
				fmt.Sprintf("schedule=%s name=%q consecutive_errors=%d max=%d",
					sched.ID, sched.Name, sched.ConsecutiveErrorCount, sched.MaxConsecutiveErrors))
			_ = s.store.UpdateScheduleStatus(ctx, sched.ID, StatusFailed)
			s.notifyFailure(sched, "schedule disabled after too many consecutive errors")
			continue
		}

		s.mu.Lock()
		_, alreadyRunning := s.running[sched.ID]
		s.mu.Unlock()
		if alreadyRunning {
			s.logger.Log(logger.Debug, "scheduler: skipping (already running)",
				fmt.Sprintf("schedule=%s name=%q", sched.ID, sched.Name))
			continue
		}

		s.logger.Log(logger.Info, "scheduler: dispatching run",
			fmt.Sprintf("schedule=%s name=%q cron=%q user_id=%s org_id=%s",
				sched.ID, sched.Name, sched.CronExpression, sched.UserID, sched.OrgID))
		go s.execute(ctx, sched)
	}
}

func (s *Scheduler) execute(parentCtx context.Context, sched Schedule) {
	timeout := time.Duration(sched.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = defaultTimeout
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	s.mu.Lock()
	s.running[sched.ID] = cancel
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, sched.ID)
		s.mu.Unlock()
	}()

	s.logger.Log(logger.Info, "scheduler: execute: starting",
		fmt.Sprintf("schedule=%s name=%q thread=%s timeout=%s user_id=%s org_id=%s run_count=%d",
			sched.ID, sched.Name, sched.ThreadID, timeout, sched.UserID, sched.OrgID, sched.RunCount))

	run := &ScheduleRun{
		ScheduleID: sched.ID,
		ThreadID:   sched.ThreadID,
		Status:     RunRunning,
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		s.logger.Log(logger.Error, "scheduler: execute: failed to create run record",
			fmt.Sprintf("schedule=%s err=%s", sched.ID, err.Error()))
		return
	}

	s.logger.Log(logger.Debug, "scheduler: execute: run record created",
		fmt.Sprintf("run_id=%s schedule=%s", run.ID, sched.ID))

	nextRun := s.computeNextRun(sched)

	history, err := s.memory.GetMessages(ctx, sched.ThreadID, 50)
	if err != nil {
		s.logger.Log(logger.Warning, "scheduler: execute: failed to load thread history",
			fmt.Sprintf("schedule=%s thread=%s err=%s", sched.ID, sched.ThreadID, err.Error()))
	} else {
		s.logger.Log(logger.Debug, "scheduler: execute: loaded history",
			fmt.Sprintf("schedule=%s messages=%d", sched.ID, len(history)))
	}
	messages := storedToMessages(history)

	prompt := fmt.Sprintf("[scheduled-task]\nSchedule: %s\nName: %s\nRun #%d\n\n%s\n[/scheduled-task]",
		sched.CronExpression, sched.Name, sched.RunCount+1, sched.Prompt)

	if sched.NotifyOn == "smart" {
		prompt += "\n\n" + smartNotifyInstruction
	}

	s.logger.Log(logger.Debug, "scheduler: execute: calling RunScheduledTask",
		fmt.Sprintf("schedule=%s prompt_len=%d history_len=%d", sched.ID, len(prompt), len(messages)))

	start := time.Now()
	result, agentErr := s.runner.RunScheduledTask(ctx, prompt, sched.ThreadID, sched.UserID, sched.OrgID, "", messages)
	elapsed := time.Since(start)

	success := agentErr == nil && result != nil
	var resultText, errText string
	var tokensUsed int

	if success {
		resultText = result.Content
		if len(resultText) > maxResultLength {
			resultText = resultText[:maxResultLength] + "\n... [truncated]"
		}
		tokensUsed = result.TotalUsage.TotalTokens
	} else if agentErr != nil {
		errText = agentErr.Error()
		s.logger.Log(logger.Error, "scheduler: execute: agent error",
			fmt.Sprintf("schedule=%s name=%q err=%s elapsed=%s", sched.ID, sched.Name, errText, elapsed))
	}

	runStatus := RunSucceeded
	if !success {
		if ctx.Err() == context.DeadlineExceeded {
			runStatus = RunTimedOut
			s.logger.Log(logger.Warning, "scheduler: execute: timed out",
				fmt.Sprintf("schedule=%s name=%q timeout=%s", sched.ID, sched.Name, timeout))
		} else {
			runStatus = RunFailed
		}
	}

	if err := s.store.CompleteRun(parentCtx, run.ID, runStatus, resultText, errText, tokensUsed); err != nil {
		s.logger.Log(logger.Error, "scheduler: execute: failed to complete run",
			fmt.Sprintf("run_id=%s schedule=%s err=%s", run.ID, sched.ID, err.Error()))
	}
	if err := s.store.UpdateScheduleAfterRun(parentCtx, sched.ID, nextRun, success); err != nil {
		s.logger.Log(logger.Error, "scheduler: execute: failed to update schedule after run",
			fmt.Sprintf("schedule=%s err=%s", sched.ID, err.Error()))
	}

	if success && result != nil {
		bgCtx := context.Background()
		seq, _ := s.memory.GetMessageCount(bgCtx, sched.ThreadID)
		toStore := memory.MessagesFromLLM(sched.ThreadID, []llm.Message{
			{Role: llm.RoleUser, Content: prompt},
			{Role: llm.RoleAssistant, Content: result.Content},
		}, seq)
		if err := s.memory.AppendMessages(bgCtx, sched.ThreadID, toStore); err != nil {
			s.logger.Log(logger.Warning, "scheduler: execute: failed to persist thread messages",
				fmt.Sprintf("schedule=%s thread=%s err=%s", sched.ID, sched.ThreadID, err.Error()))
		}
	}

	if sched.DeliveryChannel != "" {
		shouldNotify := false
		deliverBody := resultText

		switch sched.NotifyOn {
		case "smart":
			if runStatus != RunSucceeded {
				shouldNotify = true
			} else {
				agentSaysAlert, cleanBody := parseSmartDecision(resultText)
				shouldNotify = agentSaysAlert
				deliverBody = cleanBody
				s.logger.Log(logger.Info, "scheduler: execute: smart decision",
					fmt.Sprintf("schedule=%s name=%q alert=%t body_len=%d",
						sched.ID, sched.Name, agentSaysAlert, len(cleanBody)))
			}
		case "failure_only":
			shouldNotify = runStatus != RunSucceeded
		case "success_only":
			shouldNotify = runStatus == RunSucceeded
		case "never":
			shouldNotify = false
		default:
			shouldNotify = true
		}

		if shouldNotify {
			s.deliver(sched, deliverBody, errText, runStatus)
		} else {
			s.logger.Log(logger.Debug, "scheduler: execute: skipping notification",
				fmt.Sprintf("schedule=%s name=%q status=%s notify_on=%s",
					sched.ID, sched.Name, runStatus, sched.NotifyOn))
		}
	}

	var nextRunStr string
	if nextRun != nil {
		nextRunStr = nextRun.Format(time.RFC3339)
	}

	s.logger.Log(logger.Info, "scheduler: run completed",
		fmt.Sprintf("schedule=%s name=%q status=%s tokens=%d elapsed=%s next_run=%s",
			sched.ID, sched.Name, runStatus, tokensUsed, elapsed, nextRunStr))
}

func (s *Scheduler) deliver(sched Schedule, resultText, errText string, status RunStatus) {
	if s.notifier == nil {
		s.logger.Log(logger.Warning, "scheduler: deliver: notifier is nil, skipping delivery",
			fmt.Sprintf("schedule=%s name=%q", sched.ID, sched.Name))
		return
	}

	body := resultText
	if status != RunSucceeded {
		body = fmt.Sprintf("Scheduled task '%s' failed: %s", sched.Name, errText)
	}
	if body == "" {
		body = fmt.Sprintf("Scheduled task '%s' completed with no output.", sched.Name)
	}

	subject := fmt.Sprintf("Nixopus Scheduled Task: %s", sched.Name)

	channels := []string{sched.DeliveryChannel}
	if sched.DeliveryChannel == "all" {
		channels = []string{"slack", "discord", "email"}
	}

	s.logger.Log(logger.Debug, "scheduler: deliver: emitting notification",
		fmt.Sprintf("schedule=%s name=%q status=%s channels=%v user_id=%s org_id=%s body_len=%d",
			sched.ID, sched.Name, status, channels, sched.UserID, sched.OrgID, len(body)))

	if err := s.notifier.Emit(types.NotificationEvent{
		Type:           types.EventAgentScheduleRun,
		UserID:         sched.UserID,
		OrganizationID: sched.OrgID,
		Data: map[string]interface{}{
			"schedule_name": sched.Name,
			"schedule_id":   sched.ID.String(),
			"status":        string(status),
			"body":          body,
			"subject":       subject,
		},
		Channels: channels,
	}); err != nil {
		s.logger.Log(logger.Error, "scheduler: deliver: emit failed",
			fmt.Sprintf("schedule=%s err=%s", sched.ID, err.Error()))
	}
}

func (s *Scheduler) notifyFailure(sched Schedule, reason string) {
	if s.notifier == nil || sched.DeliveryChannel == "" {
		s.logger.Log(logger.Warning, "scheduler: notifyFailure: skipped (no notifier or channel)",
			fmt.Sprintf("schedule=%s", sched.ID))
		return
	}

	channels := []string{sched.DeliveryChannel}
	if sched.DeliveryChannel == "all" {
		channels = []string{"slack", "discord", "email"}
	}

	s.logger.Log(logger.Warning, "scheduler: notifyFailure: emitting disable notification",
		fmt.Sprintf("schedule=%s name=%q reason=%q channels=%v", sched.ID, sched.Name, reason, channels))

	if err := s.notifier.Emit(types.NotificationEvent{
		Type:           types.EventAgentScheduleDisabled,
		UserID:         sched.UserID,
		OrganizationID: sched.OrgID,
		Data: map[string]interface{}{
			"schedule_name": sched.Name,
			"schedule_id":   sched.ID.String(),
			"reason":        reason,
		},
		Channels: channels,
	}); err != nil {
		s.logger.Log(logger.Error, "scheduler: notifyFailure: emit failed",
			fmt.Sprintf("schedule=%s err=%s", sched.ID, err.Error()))
	}
}

func (s *Scheduler) computeNextRun(sched Schedule) *time.Time {
	cronExpr := sched.CronExpression
	// Descriptors like @every 30m parse the same as 5-field crons via robfig/cron.

	parsed, err := s.parser.Parse(cronExpr)
	if err != nil {
		s.logger.Log(logger.Error, "scheduler: invalid cron expression", fmt.Sprintf("schedule=%s expr=%s err=%s", sched.ID, cronExpr, err.Error()))
		return nil
	}

	loc := time.UTC
	if sched.Timezone != "" && sched.Timezone != "UTC" {
		if l, err := time.LoadLocation(sched.Timezone); err == nil {
			loc = l
		}
	}

	next := parsed.Next(time.Now().In(loc))
	utcNext := next.UTC()
	return &utcNext
}

// ValidateCron checks whether a cron expression is parseable.
func (s *Scheduler) ValidateCron(expr string) error {
	_, err := s.parser.Parse(expr)
	return err
}

// parseSmartDecision checks whether the agent's response starts with [ALERT] or [OK].
// Returns (shouldAlert, cleanedBody). Defaults to alert if no tag is found — safe fallback.
func parseSmartDecision(body string) (bool, string) {
	trimmed := strings.TrimSpace(body)

	if strings.HasPrefix(trimmed, "[ALERT]") {
		cleaned := strings.TrimSpace(strings.TrimPrefix(trimmed, "[ALERT]"))
		return true, cleaned
	}
	if strings.HasPrefix(trimmed, "[OK]") {
		cleaned := strings.TrimSpace(strings.TrimPrefix(trimmed, "[OK]"))
		return false, cleaned
	}

	// No tag found — default to alert (fail-open: better to notify than silently miss)
	return true, body
}

func storedToMessages(stored []memory.StoredMessage) []llm.Message {
	messages := make([]llm.Message, 0, len(stored))
	for _, m := range stored {
		msg := llm.Message{
			Role:    llm.Role(m.Role),
			Content: m.Content,
		}
		if len(m.ToolCalls) > 0 {
			msg.ToolCalls = m.ToolCalls
		}
		messages = append(messages, msg)
	}
	return messages
}
