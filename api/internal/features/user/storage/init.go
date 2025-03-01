package storage

import (
	"context"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type UserStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func CreateNewUserStorage(db *bun.DB, ctx context.Context) *UserStorage {
	return &UserStorage{
		DB:  db,
		Ctx: ctx,
	}
}

func (s *UserStorage) GetUserById(id string) (*shared_types.User, error) {
	user := &shared_types.User{}
	err := s.DB.NewSelect().Model(user).Where("id = ?", id).Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserStorage) UpdateUserName(userID string, userName string, updatedAt time.Time) error {
	_, err := s.DB.NewUpdate().
		Table("users").
		Set("username = ?", userName).
		Set("updated_at = ?", updatedAt).
		Where("id = ?", userID).
		Exec(s.Ctx)

	return err
}

func (s *UserStorage) GetUserOrganizationsWithRolesAndPermissions(userID string) ([]types.UserOrganizationsResponse, error) {
	var organizationUsers []shared_types.OrganizationUsers

	query := s.DB.NewSelect().
		TableExpr("organization_users AS ou").
		ColumnExpr("ou.*").
		Join("LEFT JOIN organizations AS o ON o.id = ou.organization_id").
		Join("LEFT JOIN roles AS r ON r.id = ou.role_id").
		Where("ou.user_id = ?", userID).
		Where("ou.deleted_at IS NULL")

	err := query.Scan(s.Ctx, &organizationUsers)
	if err != nil {
		return nil, err
	}

	var response []types.UserOrganizationsResponse
	for _, ou := range organizationUsers {
		var organization shared_types.Organization
		err := s.DB.NewSelect().
			Model(&organization).
			Where("id = ?", ou.OrganizationID).
			Scan(s.Ctx)
		if err != nil {
			continue
		}

		var role shared_types.Role
		err = s.DB.NewSelect().
			Model(&role).
			Relation("Permissions").
			Where("id = ?", ou.RoleID).
			Scan(s.Ctx)
		if err != nil {
			continue
		}

		orgResponse := types.UserOrganizationsResponse{
			Organization: organization,
			Role:         role,
		}

		response = append(response, orgResponse)
	}

	return response, nil
}
