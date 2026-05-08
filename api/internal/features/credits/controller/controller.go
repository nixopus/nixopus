package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	credits_storage "github.com/nixopus/nixopus/api/internal/features/credits/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/uptrace/bun"
)

// CreditsRepository is the minimal storage interface required by CreditsController.
// Satisfied by *credits_storage.CreditsStorage in production and mocks in tests.
type CreditsRepository interface {
	GetWalletBalance(orgID uuid.UUID) (int, error)
	GetWalletLedger(orgID uuid.UUID, limit, offset int) (*credits_storage.WalletLedger, error)
	GetUsageHistory(orgID uuid.UUID, period, groupBy string) (*credits_storage.UsageHistory, error)
	GetUsageLogs(orgID uuid.UUID, limit, offset int) (*credits_storage.UsageLogList, error)
}

type CreditsController struct {
	storage CreditsRepository
	logger  logger.Logger
}

func NewCreditsController(db *bun.DB, ctx context.Context, l logger.Logger) *CreditsController {
	return &CreditsController{
		storage: credits_storage.NewCreditsStorage(db, ctx),
		logger:  l,
	}
}

// NewCreditsControllerWithRepository creates a controller with an injectable repository,
// intended for use in tests.
func NewCreditsControllerWithRepository(repo CreditsRepository, l logger.Logger) *CreditsController {
	return &CreditsController{storage: repo, logger: l}
}

// BalanceResponse is returned by GET /credits/balance.
type BalanceResponse struct {
	BalanceUSDCents int    `json:"balance_usd_cents"`
	BalanceUSD      string `json:"balance_usd"`
	CreditsEnabled  bool   `json:"credits_enabled"`
}

func (c *CreditsController) GetBalance(f fuego.ContextNoBody) (*BalanceResponse, error) {
	if config.AppConfig.App.SelfHosted {
		return &BalanceResponse{CreditsEnabled: false}, nil
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	balance, err := c.storage.GetWalletBalance(orgID)
	if err != nil {
		c.logger.Log(logger.Error, "credits: GetBalance failed", err.Error())
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "failed to get wallet balance"}
	}

	dollars := balance / 100
	cents := balance % 100
	var balanceStr string
	if cents < 10 {
		balanceStr = strconv.Itoa(dollars) + ".0" + strconv.Itoa(cents)
	} else {
		balanceStr = strconv.Itoa(dollars) + "." + strconv.Itoa(cents)
	}

	return &BalanceResponse{
		BalanceUSDCents: balance,
		BalanceUSD:      balanceStr,
		CreditsEnabled:  true,
	}, nil
}

type UsageResponse struct {
	Status string                        `json:"status"`
	Data   *credits_storage.UsageHistory `json:"data"`
}

func (c *CreditsController) GetUsage(f fuego.ContextNoBody) (*UsageResponse, error) {
	if config.AppConfig.App.SelfHosted {
		return &UsageResponse{Status: "ok", Data: &credits_storage.UsageHistory{}}, nil
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	period := r.URL.Query().Get("period")
	if period != "7d" && period != "30d" && period != "90d" {
		period = "30d"
	}
	groupBy := r.URL.Query().Get("groupBy")
	if groupBy != "model" && groupBy != "user" && groupBy != "day" {
		groupBy = "day"
	}

	history, err := c.storage.GetUsageHistory(orgID, period, groupBy)
	if err != nil {
		c.logger.Log(logger.Error, "credits: GetUsage failed", err.Error())
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "failed to get usage history"}
	}

	return &UsageResponse{Status: "ok", Data: history}, nil
}

type TransactionsResponse struct {
	Status string                        `json:"status"`
	Data   *credits_storage.WalletLedger `json:"data"`
}

func (c *CreditsController) GetTransactions(f fuego.ContextNoBody) (*TransactionsResponse, error) {
	if config.AppConfig.App.SelfHosted {
		return &TransactionsResponse{Status: "ok", Data: &credits_storage.WalletLedger{}}, nil
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	ledger, err := c.storage.GetWalletLedger(orgID, limit, offset)
	if err != nil {
		c.logger.Log(logger.Error, "credits: GetTransactions failed", err.Error())
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "failed to get transactions"}
	}

	return &TransactionsResponse{Status: "ok", Data: ledger}, nil
}

type UsageLogsResponse struct {
	Status string                        `json:"status"`
	Data   *credits_storage.UsageLogList `json:"data"`
}

func (c *CreditsController) GetUsageLogs(f fuego.ContextNoBody) (*UsageLogsResponse, error) {
	if config.AppConfig.App.SelfHosted {
		return &UsageLogsResponse{Status: "ok", Data: &credits_storage.UsageLogList{}}, nil
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := c.storage.GetUsageLogs(orgID, limit, offset)
	if err != nil {
		c.logger.Log(logger.Error, "credits: GetUsageLogs failed", err.Error())
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "failed to get usage logs"}
	}

	return &UsageLogsResponse{Status: "ok", Data: logs}, nil
}
