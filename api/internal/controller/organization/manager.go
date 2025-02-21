package organization

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	organization, err := storage.GetOrganizations(c.app.Store.DB, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetOrganizations.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Organizations fetched successfully", organization)
}

func (c *OrganizationsController) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := c.validator.ValidateID(id, types.ErrMissingOrganizationID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	organization, err := storage.GetOrganization(c.app.Store.DB, id, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetOrganization.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Organization fetched successfully", organization)
}

func (c *OrganizationsController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.CreateOrganizationRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &organization); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(organization); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingOrganization, err := storage.GetOrganizationByName(c.app.Store.DB, organization.Name, c.app.Ctx)
	if err == nil && existingOrganization.ID != uuid.Nil {
		utils.SendErrorResponse(w, types.ErrOrganizationAlreadyExists.Error(), http.StatusBadRequest)
		return
	}

	organizationToCreate := types.Organization{
		Name:        organization.Name,
		Description: organization.Description,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
		DeletedAt:   nil,
		ID:          uuid.New(),
	}

	if err := storage.CreateOrganization(c.app.Store.DB, organizationToCreate, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToCreateOrganization.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Organization created successfully", nil)
}

func (c *OrganizationsController) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.UpdateOrganizationRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &organization); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(organization); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingOrganization, err := storage.GetOrganization(c.app.Store.DB, organization.ID, c.app.Ctx)
	if err == nil && existingOrganization.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrOrganizationDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	organizationToUpdate := types.Organization{
		Name:        organization.Name,
		Description: organization.Description,
		UpdatedAt:   time.Now(),
		CreatedAt:   existingOrganization.CreatedAt,
		DeletedAt:   existingOrganization.DeletedAt,
		ID:          existingOrganization.ID,
	}

	if err := storage.UpdateOrganization(c.app.Store.DB, &organizationToUpdate, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToUpdateOrganization.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Organization updated successfully", nil)
}

func (c *OrganizationsController) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.DeleteOrganizationRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &organization); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(organization); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingOrganization, err := storage.GetOrganization(c.app.Store.DB, organization.ID, c.app.Ctx)
	if err == nil && existingOrganization.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrOrganizationDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	if err := storage.DeleteOrganization(c.app.Store.DB, organization.ID, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDeleteOrganization.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Organization deleted successfully", nil)
}

func (c *OrganizationsController) AddUserToOrganization(w http.ResponseWriter, r *http.Request) {
	var user types.AddUserToOrganizationRequest
	if err := c.validator.ParseRequestBody(r, r.Body, &user); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(user); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	roleId, err := uuid.Parse(user.RoleId)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrInvalidRoleID.Error(), http.StatusBadRequest)
		return
	}

	existingOrganization, err := storage.GetOrganization(c.app.Store.DB, user.OrganizationID, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingOrganization.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrOrganizationDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingUser, err := storage.FindUserByID(c.app.Store.DB, user.UserID, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUser.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrUserDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingRole, err := storage.GetRole(c.app.Store.DB, roleId.String(), c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingRole.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrRoleDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingUserInOrganization, err := storage.FindUserInOrganization(c.app.Store.DB, user.UserID, user.OrganizationID, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUserInOrganization.ID != uuid.Nil {
		utils.SendErrorResponse(w, types.ErrUserAlreadyInOrganization.Error(), http.StatusBadRequest)
		return
	}

	organizationUser := types.OrganizationUsers{
		UserID:         existingUser.ID,
		OrganizationID: existingOrganization.ID,
		RoleID:         roleId,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
		ID:             uuid.New(),
	}

	if err := storage.AddUserToOrganization(c.app.Store.DB, organizationUser, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToAddUserToOrganization.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User added to organization successfully", nil)
}

func (c *OrganizationsController) RemoveUserFromOrganization(w http.ResponseWriter, r *http.Request) {
	var user types.RemoveUserFromOrganizationRequest

	if err := c.validator.ParseRequestBody(r, r.Body, &user); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(user); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingOrganization, err := storage.GetOrganization(c.app.Store.DB, user.OrganizationID, c.app.Ctx)
	if err == nil && existingOrganization.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrOrganizationDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingUser, err := storage.FindUserByID(c.app.Store.DB, user.UserID, c.app.Ctx)
	if err == nil && existingUser.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrUserDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingUserInOrganization, err := storage.FindUserInOrganization(c.app.Store.DB, user.UserID, user.OrganizationID, c.app.Ctx)
	if err == nil && existingUserInOrganization.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrUserNotInOrganization.Error(), http.StatusBadRequest)
		return
	}

	if err := storage.RemoveUserFromOrganization(c.app.Store.DB, user.UserID, user.OrganizationID, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToRemoveUserFromOrganization.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User removed from organization successfully", nil)
}

func (c *OrganizationsController) GetOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := c.validator.ValidateID(id, types.ErrMissingOrganizationID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	users, err := storage.GetOrganizationUsers(c.app.Store.DB, id, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetOrganizationUsers.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Organization users fetched successfully", users)
}
