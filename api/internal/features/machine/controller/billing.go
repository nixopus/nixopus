package controller

import (
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MachineController) ListMachinePlans(f fuego.ContextNoBody) (*types.ListPlansResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("ListMachinePlans", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("ListMachinePlans", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	return c.billingService.ListPlans()
}

func (c *MachineController) SelectMachinePlan(f fuego.ContextWithBody[types.SelectPlanRequest]) (*types.SelectPlanResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("SelectMachinePlan", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("SelectMachinePlan", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMachineDebug("SelectMachinePlan", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid request body"}
	}

	if body.PlanTier == "" {
		c.logMachineDebug("SelectMachinePlan", "plan_tier required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "plan_tier is required"}
	}

	return c.billingService.SelectPlan(r.Context(), orgID, body.PlanTier)
}

func (c *MachineController) GetMachineBilling(f fuego.ContextNoBody) (*types.MachineBillingResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		c.logMachineDebug("GetMachineBilling", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMachineDebug("GetMachineBilling", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	return c.billingService.GetBillingStatus(orgID)
}
