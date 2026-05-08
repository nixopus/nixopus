package usage

import (
	"context"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	billingstorage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/pkg/llm"
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

type TrackingParams struct {
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

type Deps struct {
	DebitWallet      func(orgID uuid.UUID, amountCents int, reason string, referenceID string) (bool, error)
	GetWalletBalance func(orgID uuid.UUID) (int, error)
	InsertLog        func(ctx context.Context, log *AIUsageLog) error
}

type Tracker struct {
	deps   Deps
	logger logger.Logger
}

func NewTracker(db *bun.DB, ctx context.Context, l logger.Logger) *Tracker {
	billing := billingstorage.NewBillingStorage(db, ctx)
	return &Tracker{
		deps: Deps{
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

func NewTrackerWithDeps(deps Deps, l logger.Logger) *Tracker {
	return &Tracker{deps: deps, logger: l}
}

// TrackUsage debits the organization wallet and inserts a usage log row.
// Fault-tolerant: failures are logged but never propagated to the caller.
// Returns the remaining balance in cents, or -1 if unknown/skipped.
func (t *Tracker) TrackUsage(ctx context.Context, params TrackingParams) int {
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

func ResolveModelID(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	if m := os.Getenv("AGENT_MODEL"); m != "" {
		return m
	}
	return "anthropic/claude-sonnet-4"
}

// billableTokens returns the token count we charge for: aggregated total from the
// provider when positive, otherwise prompt + completion (some APIs omit total_tokens).
func billableTokensFromUsage(u llm.Usage) int {
	if u.TotalTokens > 0 {
		return u.TotalTokens
	}
	return u.PromptTokens + u.CompletionTokens
}

// ScheduledRunCostUsdFromUsage converts agent-reported usage into dollars for TrackUsage.
// Charge is linear: billableTokens × SCHEDULED_RUN_USD_PER_TOKEN.
// Default rate is 1e-5 ($0.00001/token, i.e. $10 per 1M tokens). Set SCHEDULED_RUN_USD_PER_TOKEN=0 to skip debits.
func ScheduledRunCostUsdFromUsage(u llm.Usage) float64 {
	n := billableTokensFromUsage(u)
	if n <= 0 {
		return 0
	}
	per := 1e-5
	if v, ok := os.LookupEnv("SCHEDULED_RUN_USD_PER_TOKEN"); ok {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil && f >= 0 {
			per = f
		}
	}
	if per == 0 {
		return 0
	}
	return float64(n) * per
}
