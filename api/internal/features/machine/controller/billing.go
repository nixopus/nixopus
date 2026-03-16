package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/machine/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *MachineController) ListMachinePlans(f fuego.ContextNoBody) (*types.ListPlansResponse, error) {
	return c.billingService.ListPlans()
}

func (c *MachineController) SelectMachinePlan(f fuego.ContextWithBody[types.SelectPlanRequest]) (*types.SelectPlanResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	if body.PlanTier == "" {
		return nil, fuego.BadRequestError{Detail: "plan_tier is required"}
	}

	return c.billingService.SelectPlan(r.Context(), orgID, body.PlanTier)
}

func (c *MachineController) GetMachineBilling(f fuego.ContextNoBody) (*types.MachineBillingResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	return c.billingService.GetBillingStatus(orgID)
}
