package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/agent/service/scheduler"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/usage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/middleware"
	"github.com/nixopus/nixopus/api/pkg/llm"
)

// RunScheduledTask implements scheduler.AgentRunner.
// It runs a non-streaming agent turn with internal auth context.
func (s *AgentService) RunScheduledTask(ctx context.Context, prompt string, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error) {
	s.logger.Log(logger.Info, "RunScheduledTask: starting",
		fmt.Sprintf("thread=%s user_id=%s org_id=%s model=%q history_len=%d prompt_len=%d",
			threadID, userID, orgID, model, len(history), len(prompt)))

	secret := os.Getenv("AUTH_SERVICE_SECRET")
	if secret == "" {
		secret = os.Getenv("BETTER_AUTH_SECRET")
	}
	if secret == "" {
		s.logger.Log(logger.Warning, "RunScheduledTask: no AUTH_SERVICE_SECRET or BETTER_AUTH_SECRET set", "")
	}
	sig := middleware.ComputeInternalSignature(secret, userID, orgID)

	ctx = context.WithValue(ctx, ctxKeyAuthToken, sig)
	ctx = context.WithValue(ctx, ctxKeyOrgID, orgID)
	ctx = context.WithValue(ctx, ctxKeySchedulerUserID, userID)
	ctx = context.WithValue(ctx, ctxKeyInternalSig, sig)
	ctx = context.WithValue(ctx, ctxKeyInternalUserID, userID)

	baseURL := "http://localhost:" + getEnvOrDefault("PORT", "2089")
	ctx = context.WithValue(ctx, ctxKeyBaseURL, baseURL)

	s.logger.Log(logger.Debug, "RunScheduledTask: context set",
		fmt.Sprintf("base_url=%s has_sig=%t user_id=%s org_id=%s", baseURL, sig != "", userID, orgID))

	if model == "" {
		model = getEnvOrDefault("AGENT_LIGHT_MODEL", "openai/gpt-4o-mini")
		s.logger.Log(logger.Debug, "RunScheduledTask: using default model", fmt.Sprintf("model=%s", model))
	}

	agent, ok := s.agents.Get("deploy")
	if !ok {
		s.logger.Log(logger.Error, "RunScheduledTask: deploy agent not registered", "")
		return nil, fmt.Errorf("deploy agent not registered")
	}

	s.logger.Log(logger.Debug, "RunScheduledTask: calling agent.RunWithOptions",
		fmt.Sprintf("thread=%s model=%s", threadID, model))

	start := time.Now()
	result, err := agent.RunWithOptions(ctx, llm.RunOptions{Model: model}, prompt, history...)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		s.logger.Log(logger.Error, "RunScheduledTask: agent run failed",
			fmt.Sprintf("thread=%s user_id=%s org_id=%s err=%s", threadID, userID, orgID, err.Error()))
		return nil, fmt.Errorf("scheduled agent run: %w", err)
	}

	costUsd := usage.ScheduledRunCostUsdFromUsage(result.TotalUsage)
	if s.usage != nil {
		bgCtx := context.Background()
		bal := s.usage.TrackUsage(bgCtx, usage.TrackingParams{
			OrgID:            orgID,
			UserID:           userID,
			ModelID:          usage.ResolveModelID(model),
			PromptTokens:     result.TotalUsage.PromptTokens,
			CompletionTokens: result.TotalUsage.CompletionTokens,
			TotalTokens:      result.TotalUsage.TotalTokens,
			CostUsd:          costUsd,
			RequestType:      "scheduled",
			SessionID:        threadID,
			LatencyMs:        &latencyMs,
		})
		if bal >= 0 {
			s.logger.Log(logger.Debug, "RunScheduledTask: usage tracked",
				fmt.Sprintf("org_id=%s user_id=%s cost_usd=%.4f credits_remaining_cents=%d", orgID, userID, costUsd, bal))
		}
	}

	s.logger.Log(logger.Info, "RunScheduledTask: completed",
		fmt.Sprintf("thread=%s steps=%d tokens=%d content_len=%d",
			threadID, result.Steps, result.TotalUsage.TotalTokens, len(result.Content)))

	return result, nil
}

const (
	ctxKeyInternalSig    contextKey = "internal_sig"
	ctxKeyInternalUserID contextKey = "internal_user_id"
)

// GetScheduleStore returns the scheduler store for external use (controller).
func (s *AgentService) GetScheduleStore() *scheduler.Store {
	return s.scheduleStore
}

// GetScheduler returns the scheduler instance.
func (s *AgentService) GetScheduler() *scheduler.Scheduler {
	return s.scheduler
}
