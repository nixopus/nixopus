package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
)

var creditSkipPrefixes = []string{
	"/api/v1/credits/",
	"/api/internal/credits/",
	"/health",
	"/healthz",
	"/readyz",
	"/metrics",
}

var creditSkipIncludes = []string{
	"/threads",
	"/memory",
}

type MachineWarning struct {
	Status        string  `json:"status"`
	GraceDeadline *string `json:"grace_deadline"`
	DaysRemaining *int    `json:"days_remaining"`
	Message       string  `json:"message"`
}

type CreditCheckResult struct {
	Allowed        bool            `json:"allowed"`
	BalanceCents   int             `json:"balance_cents"`
	MachineWarning *MachineWarning `json:"machine_warning,omitempty"`
}

type MachineStatusResult struct {
	HasMachine    bool
	Status        string
	GraceDeadline *string
	DaysRemaining *int
	PlanCostCents int
}

type CreditGateDeps struct {
	GetWalletBalance func(ctx context.Context, orgID uuid.UUID) (int, error)
	GetMachineStatus func(ctx context.Context, orgID uuid.UUID) (*MachineStatusResult, error)
}

type creditContextKey struct{}

func GetCreditCheckResult(ctx context.Context) *CreditCheckResult {
	v, _ := ctx.Value(creditContextKey{}).(*CreditCheckResult)
	return v
}

func ShouldSkipCreditCheck(pathname string) bool {
	for _, prefix := range creditSkipPrefixes {
		if strings.HasPrefix(pathname, prefix) {
			return true
		}
	}
	for _, inc := range creditSkipIncludes {
		if strings.Contains(pathname, inc) {
			return true
		}
	}
	return false
}

func buildMachineWarning(ms *MachineStatusResult) *MachineWarning {
	if ms == nil || !ms.HasMachine {
		return nil
	}

	switch ms.Status {
	case "suspended":
		return &MachineWarning{
			Status:  "suspended",
			Message: "Your server was reset due to insufficient wallet balance. Top up your wallet to restore service.",
		}
	case "grace_period":
		days := 0
		if ms.DaysRemaining != nil {
			days = *ms.DaysRemaining
		}
		cost := "the monthly machine cost"
		if ms.PlanCostCents > 0 {
			cost = formatCents(ms.PlanCostCents)
		}
		suffix := "s"
		if days == 1 {
			suffix = ""
		}
		return &MachineWarning{
			Status:        "grace_period",
			GraceDeadline: ms.GraceDeadline,
			DaysRemaining: ms.DaysRemaining,
			Message:       "Your server will be reset in " + itoa(days) + " day" + suffix + ". Wallet balance is insufficient to cover " + cost + ". Top up now to keep your server.",
		}
	case "unbilled":
		return &MachineWarning{
			Status:  "upgrade_required",
			Message: "You are on a trial machine without a billing plan. Select a machine plan to keep your server running.",
		}
	}
	return nil
}

func formatCents(cents int) string {
	dollars := cents / 100
	remainder := cents % 100
	if remainder < 10 {
		return "$" + itoa(dollars) + ".0" + itoa(remainder)
	}
	return "$" + itoa(dollars) + "." + itoa(remainder)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func CreditGateMiddleware(next http.Handler, deps CreditGateDeps, l logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.AppConfig.App.SelfHosted {
			next.ServeHTTP(w, r)
			return
		}

		if ShouldSkipCreditCheck(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		orgIDStr, ok := r.Context().Value(types.OrganizationIDKey).(string)
		if !ok || orgIDStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		balance, err := deps.GetWalletBalance(r.Context(), orgID)
		if err != nil {
			l.LogCtx(r.Context(), logger.Error, "credit gate: failed to get wallet balance", err.Error())
			balance = 0
		}

		if balance <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"title":       "Payment Required",
				"status":      http.StatusPaymentRequired,
				"detail":      "No AI credits remaining. Please top up or upgrade your plan.",
				"upgrade_url": "/billing",
			})
			return
		}

		var machineWarning *MachineWarning
		if deps.GetMachineStatus != nil {
			ms, msErr := deps.GetMachineStatus(r.Context(), orgID)
			if msErr != nil {
				l.LogCtx(r.Context(), logger.Error, "credit gate: failed to get machine status", msErr.Error())
			} else {
				machineWarning = buildMachineWarning(ms)
			}
		}

		result := &CreditCheckResult{
			Allowed:        true,
			BalanceCents:   balance,
			MachineWarning: machineWarning,
		}

		ctx := context.WithValue(r.Context(), creditContextKey{}, result)
		r = r.WithContext(ctx)

		if machineWarning != nil {
			warningJSON, _ := json.Marshal(machineWarning)
			w.Header().Set("X-Credit-Warning", string(warningJSON))
		}

		next.ServeHTTP(w, r)
	})
}
