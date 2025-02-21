package organization

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type PermissionsController struct {
	app *storage.App
}

func NewPermissionsController(app *storage.App) *PermissionsController {
	return &PermissionsController{
		app: app,
	}
}

func validateCreatePermissionRequest(permission types.CreatePermissionRequest) error {
	if permission.Name == "" {
		return types.ErrPermissionNameRequired
	}
	if permission.Resource == "" {
		return types.ErrPermissionResourceRequired
	}
	return nil
}

func validateGetPermissionRequest(id string) error {
	if id == "" {
		return types.ErrPermissionIDRequired
	}
	return nil
}

func validateUpdatePermissionRequest(permission types.UpdatePermissionRequest) error {
	if permission.Name == "" && permission.Description == "" {
		return types.ErrPermissionEmptyFields
	}
	return nil
}

func validateAddPermissionToRoleRequest(permission types.AddPermissionToRoleRequest) error {
	if permission.PermissionID == "" {
		return types.ErrPermissionIDRequired
	}
	if permission.RoleID == "" {
		return types.ErrRoleIDRequired
	}
	return nil
}

func validateRemovePermissionFromRoleRequest(permission types.RemovePermissionFromRoleRequest) error {
	if permission.PermissionID == "" {
		return types.ErrPermissionIDRequired
	}
	if permission.RoleID == "" {
		return types.ErrRoleIDRequired
	}
	return nil
}

func (c *PermissionsController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.CreatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := validateCreatePermissionRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingPermission, err := storage.GetPermissionByName(c.app.Store.DB, permission.Name, c.app.Ctx)
	if err == nil && existingPermission.ID != uuid.Nil {
		utils.SendErrorResponse(w, types.ErrPermissionAlreadyExists.Error(), http.StatusBadRequest)
		return
	}

	permissionToCreate := types.Permission{
		ID:          uuid.New(),
		Name:        permission.Name,
		Resource:    permission.Resource,
		Description: permission.Description,
	}

	if err := storage.CreatePermission(c.app.Store.DB, permissionToCreate, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToCreatePermission.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission created successfully", nil)
}

func (c *PermissionsController) GetPermissions(w http.ResponseWriter, r *http.Request) {
	permission, err := storage.GetPermissions(c.app.Store.DB, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermissions.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permissions fetched successfully", permission)
}

func (c *PermissionsController) GetPermission(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := validateGetPermissionRequest(id); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	permission, err := storage.GetPermission(c.app.Store.DB, id, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermission.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permission fetched successfully", permission)
}

func (c *PermissionsController) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.UpdatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := validateUpdatePermissionRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingPermission, err := storage.GetPermission(c.app.Store.DB, permission.ID, c.app.Ctx)
	if err == nil && existingPermission.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrPermissionDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	permissionToUpdate := types.Permission{
		ID:          existingPermission.ID,
		Name:        permission.Name,
		Resource:    permission.Resource,
		Description: permission.Description,
		UpdatedAt:   time.Now(),
	}

	if err := storage.UpdatePermission(c.app.Store.DB, &permissionToUpdate, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToUpdatePermission.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission updated successfully", nil)
}

func (c *PermissionsController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	var permission types.DeletePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := validateGetPermissionRequest(permission.ID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingPermission, err := storage.GetPermission(c.app.Store.DB, permission.ID, c.app.Ctx)
	if err == nil && existingPermission.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrPermissionDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	if err := storage.DeletePermission(c.app.Store.DB, permission.ID, c.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDeletePermission.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission deleted successfully", nil)
}

func (p *PermissionsController) AddPermissionToRole(w http.ResponseWriter, r *http.Request) {
	var permission types.AddPermissionToRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := validateAddPermissionToRoleRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingRole, err := storage.GetRole(p.app.Store.DB, permission.RoleID, p.app.Ctx)
	if err == nil && existingRole.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrRoleDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingPermission, err := storage.GetPermission(p.app.Store.DB, permission.PermissionID, p.app.Ctx)
	if err == nil && existingPermission.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrPermissionDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	rolePermissionToCreate := types.RolePermissions{
		ID:           uuid.New(),
		RoleID:       existingRole.ID,
		PermissionID: existingPermission.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := storage.AddPermissionToRole(p.app.Store.DB, rolePermissionToCreate, p.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToAddPermissionToRole.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission added to role successfully", nil)
}

func (p *PermissionsController) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	var permission types.RemovePermissionFromRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := validateRemovePermissionFromRoleRequest(permission); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingRole, err := storage.GetRole(p.app.Store.DB, permission.RoleID, p.app.Ctx)
	if err == nil && existingRole.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrRoleDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	existingPermission, err := storage.GetPermission(p.app.Store.DB, permission.PermissionID, p.app.Ctx)
	if err == nil && existingPermission.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrPermissionDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	if err := storage.RemovePermissionFromRole(p.app.Store.DB, permission.PermissionID, p.app.Ctx); err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToRemovePermissionFromRole.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Permission removed from role successfully", nil)
}

func (p *PermissionsController) GetPermissionsByRole(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := validateGetPermissionRequest(id); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingRole, err := storage.GetRole(p.app.Store.DB, id, p.app.Ctx)
	if err == nil && existingRole.ID == uuid.Nil {
		utils.SendErrorResponse(w, types.ErrRoleDoesNotExist.Error(), http.StatusBadRequest)
		return
	}

	permissions, err := storage.GetPermissionsByRole(p.app.Store.DB, id, p.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToGetPermissionsByRole.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Permissions fetched successfully", permissions)
}
