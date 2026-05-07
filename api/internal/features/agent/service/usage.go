package service

import (
	"context"
	"math"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	billingstorage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/uptrace/bun"
)

type AIUsageLog struct {
	bun.BaseModel `bun:"table:ai_usage_logs,alias:aul"`

	ID               uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	OrganizationID   uuid.UUID  `bun:"organization_id,notnull,type:uuid"`
	UserID           *uuid.UUID `bun:"user_id,type:uuid"`
	ModelID          string     `bun:"model_id,notnull"`
	PromptTokens     int        `bun:"prompt_tokens,notnull,default:0"`
	CompletionTokens int        `bun:"completion_tokens,notnull,default:0"`
	TotalTokens      int        `bun:"total_tokens,notnull,default:0"`
	CachedTokens     int        `bun:"cached_tokens,notnull,default:0"`
	ReasoningTokens  int        `bun:"reasoning_tokens,notnull,default:0"`
	CostUsd          float64    `bun:"cost_usd,notnull,default:0"`
	RequestType      string     `bun:"request_type"`
	AgentID          string     `bun:"agent_id"`
	WorkflowID       string     `bun:"workflow_id"`
	SessionID        string     `bun:"session_id"`
	LatencyMs        *int       `bun:"latency_ms"`
	Status           string     `bun:"status,default:'success'"`
	ErrorMessage     string     `bun:"error_message"`
	Metadata         string     `bun:"metadata,type:jsonb"`
	CreatedAt        time.Time  `bun:"created_at,notnull,default:now()"`
}

type UsageTrackingParams struct {
	OrgID            string
	UserID           string
	ModelID          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CostUsd          float64
	RequestType      string
	SessionID        string
	LatencyMs        *int
	Status           string
	ErrorMessage     string
}

type UsageDeps struct {
	DebitWallet      func(orgID uuid.UUID, amountCents int, reason string, referenceID string) (bool, error)
	GetWalletBalance func(orgID uuid.UUID) (int, error)
	InsertLog        func(ctx context.Context, log *AIUsageLog) error
}

type UsageTracker struct {
	deps   UsageDeps
	logger logger.Logger
}

func NewUsageTracker(db *bun.DB, ctx context.Context, l logger.Logger) *UsageTracker {
	billing := billingstorage.NewBillingStorage(db, ctx)
	return &UsageTracker{
		deps: UsageDeps{
			DebitWallet:      billing.DebitWallet,
			GetWalletBalance: billing.GetWalletBalance,
			InsertLog: func(ctx context.Context, log *AIUsageLog) error {
				_, err := db.NewInsert().Model(log).Exec(ctx)
				return err
			},
		},
		logger: l,
	}
}

func newUsageTrackerWithDeps(deps UsageDeps, l logger.Logger) *UsageTracker {
	return &UsageTracker{deps: deps, logger: l}
}

// TrackUsage debits the organization wallet and inserts a usage log row.
// Fault-tolerant: failures are logged but never propagated to the caller.
// Returns the remaining balance in cents, or -1 if unknown/skipped.
func (t *UsageTracker) TrackUsage(ctx context.Context, params UsageTrackingParams) int {
	if config.AppConfig.App.SelfHosted {
		return -1
	}

	if params.OrgID == "" {
		return -1
	}

	orgUUID, err := uuid.Parse(params.OrgID)
	if err != nil {
		t.logger.Log(logger.Error, "usage: invalid org ID", err.Error())
		return -1
	}

	costCents := int(math.Ceil(params.CostUsd * 100))
	if costCents > 0 {
		refID := uuid.New().String()
		ok, err := t.deps.DebitWallet(orgUUID, costCents, "ai_usage", refID)
		if err != nil {
			t.logger.Log(logger.Error, "usage: wallet debit failed", err.Error())
		} else if !ok {
			t.logger.Log(logger.Warning, "usage: wallet debit returned false (insufficient balance or duplicate)", "")
		}
	}

	status := params.Status
	if status == "" {
		status = "success"
	}

	log := &AIUsageLog{
		OrganizationID:   orgUUID,
		ModelID:          params.ModelID,
		PromptTokens:     params.PromptTokens,
		CompletionTokens: params.CompletionTokens,
		TotalTokens:      params.TotalTokens,
		CostUsd:          params.CostUsd,
		RequestType:      params.RequestType,
		SessionID:        params.SessionID,
		LatencyMs:        params.LatencyMs,
		Status:           status,
		ErrorMessage:     params.ErrorMessage,
	}

	if params.UserID != "" {
		if uid, err := uuid.Parse(params.UserID); err == nil {
			log.UserID = &uid
		}
	}

	if err := t.deps.InsertLog(ctx, log); err != nil {
		t.logger.Log(logger.Error, "usage: failed to insert usage log", err.Error())
	}

	balance, err := t.deps.GetWalletBalance(orgUUID)
	if err != nil {
		t.logger.Log(logger.Error, "usage: failed to get remaining balance", err.Error())
		return -1
	}

	return balance
}

func resolveModelID(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	if m := os.Getenv("AGENT_MODEL"); m != "" {
		return m
	}
	return "anthropic/claude-sonnet-4"
}
