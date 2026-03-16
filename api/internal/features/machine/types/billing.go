package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MachineBillingStatus string

const (
	MachineBillingStatusActive      MachineBillingStatus = "active"
	MachineBillingStatusGracePeriod MachineBillingStatus = "grace_period"
	MachineBillingStatusSuspended   MachineBillingStatus = "suspended"
	MachineBillingStatusCancelled   MachineBillingStatus = "cancelled"
)

type MachinePlan struct {
	bun.BaseModel `bun:"table:machine_plans,alias:mp" swaggerignore:"true"`

	ID               uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Tier             string    `bun:"tier,notnull,unique" json:"tier"`
	Name             string    `bun:"name,notnull" json:"name"`
	RamMB            int       `bun:"ram_mb,notnull" json:"ram_mb"`
	Vcpu             int       `bun:"vcpu,notnull" json:"vcpu"`
	StorageMB        int       `bun:"storage_mb,notnull" json:"storage_mb"`
	MonthlyCostCents int       `bun:"monthly_cost_cents,notnull" json:"monthly_cost_cents"`
	IsActive         bool      `bun:"is_active,notnull,default:true" json:"is_active"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type OrgMachineBilling struct {
	bun.BaseModel `bun:"table:org_machine_billing,alias:omb" swaggerignore:"true"`

	ID                 uuid.UUID            `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID            `bun:"organization_id,notnull,type:uuid" json:"organization_id"`
	SSHKeyID           *uuid.UUID           `bun:"ssh_key_id,type:uuid" json:"ssh_key_id,omitempty"`
	MachinePlanID      uuid.UUID            `bun:"machine_plan_id,notnull,type:uuid" json:"machine_plan_id"`
	Status             MachineBillingStatus `bun:"status,notnull,default:'active'" json:"status"`
	CurrentPeriodStart time.Time            `bun:"current_period_start,notnull" json:"current_period_start"`
	CurrentPeriodEnd   time.Time            `bun:"current_period_end,notnull" json:"current_period_end"`
	GraceDeadline      *time.Time           `bun:"grace_deadline" json:"grace_deadline,omitempty"`
	LastChargedAt      *time.Time           `bun:"last_charged_at" json:"last_charged_at,omitempty"`
	CreatedAt          time.Time            `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time            `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type WalletTransaction struct {
	bun.BaseModel `bun:"table:wallet_transactions,alias:wt" swaggerignore:"true"`

	ID                uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	OrganizationID    uuid.UUID `bun:"organization_id,notnull,type:uuid" json:"organization_id"`
	AmountCents       int       `bun:"amount_cents,notnull" json:"amount_cents"`
	EntryType         string    `bun:"entry_type,notnull" json:"entry_type"`
	BalanceAfterCents int       `bun:"balance_after_cents,notnull" json:"balance_after_cents"`
	Reason            *string   `bun:"reason" json:"reason,omitempty"`
	ReferenceID       *string   `bun:"reference_id" json:"reference_id,omitempty"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type MachinePlanResponse struct {
	ID               string `json:"id"`
	Tier             string `json:"tier"`
	Name             string `json:"name"`
	RamMB            int    `json:"ram_mb"`
	Vcpu             int    `json:"vcpu"`
	StorageMB        int    `json:"storage_mb"`
	MonthlyCostCents int    `json:"monthly_cost_cents"`
	MonthlyCostUSD   string `json:"monthly_cost_usd"`
}

type ListPlansResponse struct {
	Status string                `json:"status"`
	Data   []MachinePlanResponse `json:"data"`
}

type SelectPlanRequest struct {
	PlanTier string `json:"plan_tier" validate:"required"`
}

type SelectPlanResponse struct {
	Status            string               `json:"status"`
	Message           string               `json:"message"`
	Plan              *MachinePlanResponse `json:"plan,omitempty"`
	ChargedCents      int                  `json:"charged_cents,omitempty"`
	BalanceAfterCents int                  `json:"balance_after_cents,omitempty"`
	PeriodEnd         string               `json:"period_end,omitempty"`
	Error             string               `json:"error,omitempty"`
}

type MachineBillingResponse struct {
	Status string                    `json:"status"`
	Data   *MachineBillingStatusData `json:"data"`
}

type MachineBillingStatusData struct {
	HasMachine       bool   `json:"has_machine"`
	PlanTier         string `json:"plan_tier,omitempty"`
	PlanName         string `json:"plan_name,omitempty"`
	MonthlyCostCents int    `json:"monthly_cost_cents,omitempty"`
	MonthlyCostUSD   string `json:"monthly_cost_usd,omitempty"`
	BillingStatus    string `json:"billing_status,omitempty"`
	PeriodEnd        string `json:"period_end,omitempty"`
	GraceDeadline    string `json:"grace_deadline,omitempty"`
	DaysRemaining    *int   `json:"days_remaining,omitempty"`
	Message          string `json:"message,omitempty"`
}
