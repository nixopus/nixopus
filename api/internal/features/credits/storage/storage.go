package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/usage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/uptrace/bun"
)

type CreditsStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func NewCreditsStorage(db *bun.DB, ctx context.Context) *CreditsStorage {
	return &CreditsStorage{DB: db, Ctx: ctx}
}

// GetWalletBalance returns the current balance for an org in cents.
func (s *CreditsStorage) GetWalletBalance(orgID uuid.UUID) (int, error) {
	var tx machine_types.WalletTransaction
	err := s.DB.NewSelect().
		Model(&tx).
		Column("balance_after_cents").
		Where("organization_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}
	return tx.BalanceAfterCents, nil
}

// WalletLedgerItem is a single transaction in the wallet ledger.
type WalletLedgerItem struct {
	ID                string  `json:"id"`
	AmountCents       int     `json:"amount_cents"`
	EntryType         string  `json:"entry_type"`
	Reason            *string `json:"reason,omitempty"`
	BalanceAfterCents int     `json:"balance_after_cents"`
	CreatedAt         string  `json:"created_at"`
}

type WalletLedger struct {
	Items      []WalletLedgerItem `json:"items"`
	TotalCount int                `json:"total_count"`
}

// GetWalletLedger returns paginated wallet transactions for an org.
func (s *CreditsStorage) GetWalletLedger(orgID uuid.UUID, limit, offset int) (*WalletLedger, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var txs []machine_types.WalletTransaction
	count, err := s.DB.NewSelect().
		Model(&txs).
		Where("organization_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(limit).
		Offset(offset).
		ScanAndCount(s.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet ledger: %w", err)
	}

	items := make([]WalletLedgerItem, len(txs))
	for i, tx := range txs {
		items[i] = WalletLedgerItem{
			ID:                tx.ID.String(),
			AmountCents:       tx.AmountCents,
			EntryType:         tx.EntryType,
			Reason:            tx.Reason,
			BalanceAfterCents: tx.BalanceAfterCents,
			CreatedAt:         tx.CreatedAt.UTC().Format(time.RFC3339),
		}
	}

	return &WalletLedger{Items: items, TotalCount: count}, nil
}

// UsageBreakdownItem is a single row in usage breakdown.
type UsageBreakdownItem struct {
	Key          string  `json:"key"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	TotalTokens  int64   `json:"total_tokens"`
	RequestCount int64   `json:"request_count"`
}

// DailyUsageItem is a single day of usage.
type DailyUsageItem struct {
	Date     string  `json:"date"`
	CostUSD  float64 `json:"cost_usd"`
	Tokens   int64   `json:"tokens"`
	Requests int64   `json:"requests"`
}

// UsageHistory is the aggregated usage response.
type UsageHistory struct {
	TotalCostUSD float64              `json:"total_cost_usd"`
	TotalTokens  int64                `json:"total_tokens"`
	Breakdown    []UsageBreakdownItem `json:"breakdown"`
	DailyUsage   []DailyUsageItem     `json:"daily_usage"`
}

// GetUsageHistory returns usage history for an org grouped by model, user, or day.
func (s *CreditsStorage) GetUsageHistory(orgID uuid.UUID, period, groupBy string) (*UsageHistory, error) {
	cutoff := periodToCutoff(period)

	type totalsRow struct {
		TotalCostUSD float64 `bun:"total_cost_usd"`
		TotalTokens  int64   `bun:"total_tokens"`
	}
	var totals totalsRow
	err := s.DB.NewRaw(
		`SELECT COALESCE(SUM(cost_usd::numeric), 0) AS total_cost_usd,
		        COALESCE(SUM(total_tokens), 0) AS total_tokens
		 FROM ai_usage_logs
		 WHERE organization_id = ? AND created_at >= ?`,
		orgID, cutoff,
	).Scan(s.Ctx, &totals)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage totals: %w", err)
	}

	groupExpr := groupByExpr(groupBy)

	type breakdownRow struct {
		Key          string  `bun:"key"`
		TotalCostUSD float64 `bun:"total_cost_usd"`
		TotalTokens  int64   `bun:"total_tokens"`
		RequestCount int64   `bun:"request_count"`
	}
	var breakdown []breakdownRow
	err = s.DB.NewRaw(
		fmt.Sprintf(`SELECT %s AS key,
		                    COALESCE(SUM(cost_usd::numeric), 0) AS total_cost_usd,
		                    COALESCE(SUM(total_tokens), 0) AS total_tokens,
		                    COUNT(*) AS request_count
		             FROM ai_usage_logs
		             WHERE organization_id = ? AND created_at >= ?
		             GROUP BY %s
		             ORDER BY total_cost_usd DESC`, groupExpr, groupExpr),
		orgID, cutoff,
	).Scan(s.Ctx, &breakdown)
	if err != nil {
		breakdown = nil
	}

	type dailyRow struct {
		Date     string  `bun:"date"`
		CostUSD  float64 `bun:"cost_usd"`
		Tokens   int64   `bun:"tokens"`
		Requests int64   `bun:"requests"`
	}
	var daily []dailyRow
	err = s.DB.NewRaw(
		`SELECT created_at::date::text AS date,
		        COALESCE(SUM(cost_usd::numeric), 0) AS cost_usd,
		        COALESCE(SUM(total_tokens), 0) AS tokens,
		        COUNT(*) AS requests
		 FROM ai_usage_logs
		 WHERE organization_id = ? AND created_at >= ?
		 GROUP BY created_at::date
		 ORDER BY created_at::date ASC`,
		orgID, cutoff,
	).Scan(s.Ctx, &daily)
	if err != nil {
		daily = nil
	}

	bItems := make([]UsageBreakdownItem, len(breakdown))
	for i, r := range breakdown {
		bItems[i] = UsageBreakdownItem{
			Key:          r.Key,
			TotalCostUSD: r.TotalCostUSD,
			TotalTokens:  r.TotalTokens,
			RequestCount: r.RequestCount,
		}
	}

	dItems := make([]DailyUsageItem, len(daily))
	for i, r := range daily {
		dItems[i] = DailyUsageItem{
			Date:     r.Date,
			CostUSD:  r.CostUSD,
			Tokens:   r.Tokens,
			Requests: r.Requests,
		}
	}

	return &UsageHistory{
		TotalCostUSD: totals.TotalCostUSD,
		TotalTokens:  totals.TotalTokens,
		Breakdown:    bItems,
		DailyUsage:   dItems,
	}, nil
}

// UsageLogItem is a single usage log entry.
type UsageLogItem struct {
	ID               string  `json:"id"`
	ModelID          string  `json:"model_id"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	CachedTokens     int     `json:"cached_tokens"`
	ReasoningTokens  int     `json:"reasoning_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	RequestType      string  `json:"request_type,omitempty"`
	AgentID          string  `json:"agent_id,omitempty"`
	SessionID        string  `json:"session_id,omitempty"`
	LatencyMs        *int    `json:"latency_ms,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UserID           *string `json:"user_id,omitempty"`
}

type UsageLogList struct {
	Items      []UsageLogItem `json:"items"`
	TotalCount int            `json:"total_count"`
}

// GetUsageLogs returns paginated detailed AI usage logs for an org.
func (s *CreditsStorage) GetUsageLogs(orgID uuid.UUID, limit, offset int) (*UsageLogList, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var logs []usage.AIUsageLog
	count, err := s.DB.NewSelect().
		Model(&logs).
		Where("organization_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(limit).
		Offset(offset).
		ScanAndCount(s.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage logs: %w", err)
	}

	items := make([]UsageLogItem, len(logs))
	for i, l := range logs {
		item := UsageLogItem{
			ID:               l.ID.String(),
			ModelID:          l.ModelID,
			PromptTokens:     l.PromptTokens,
			CompletionTokens: l.CompletionTokens,
			TotalTokens:      l.TotalTokens,
			CachedTokens:     l.CachedTokens,
			ReasoningTokens:  l.ReasoningTokens,
			CostUSD:          l.CostUsd,
			RequestType:      l.RequestType,
			AgentID:          l.AgentID,
			SessionID:        l.SessionID,
			LatencyMs:        l.LatencyMs,
			Status:           l.Status,
			CreatedAt:        l.CreatedAt.UTC().Format(time.RFC3339),
		}
		if l.UserID != nil {
			s := l.UserID.String()
			item.UserID = &s
		}
		items[i] = item
	}

	return &UsageLogList{Items: items, TotalCount: count}, nil
}

func periodToCutoff(period string) time.Time {
	now := time.Now().UTC()
	switch period {
	case "7d":
		return now.AddDate(0, 0, -7)
	case "90d":
		return now.AddDate(0, 0, -90)
	default:
		return now.AddDate(0, 0, -30)
	}
}

func groupByExpr(groupBy string) string {
	switch groupBy {
	case "model":
		return "model_id"
	case "user":
		return "user_id::text"
	default:
		return "created_at::date::text"
	}
}
