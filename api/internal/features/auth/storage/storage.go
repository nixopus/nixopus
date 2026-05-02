package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type UserStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type AuthRepository interface {
	FindUserByEmail(email string) (*types.User, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) AuthRepository
}

func (u *UserStorage) WithTx(tx bun.Tx) AuthRepository {
	return &UserStorage{
		DB:  u.DB,
		Ctx: u.Ctx,
		tx:  &tx,
	}
}

func (u *UserStorage) BeginTx() (bun.Tx, error) {
	return u.DB.BeginTx(u.Ctx, nil)
}

func (u *UserStorage) getDB() bun.IDB {
	if u.tx != nil {
		return *u.tx
	}
	return u.DB
}

func (u *UserStorage) FindUserByEmail(email string) (*types.User, error) {
	user := &types.User{}
	err := u.getDB().NewSelect().
		Model(user).
		Where("email = ?", email).
		Relation("Organizations").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	err = u.getDB().NewSelect().
		Model(&user.OrganizationUsers).
		Where("user_id = ?", user.ID).
		Relation("Organization").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	user.ComputeCompatibilityFields()

	return user, nil
}

// Account is the Better Auth `account` table (credentials: email/password, OAuth, etc.).
type Account struct {
	bun.BaseModel `bun:"table:account,alias:a"`
	ID            uuid.UUID `bun:"id,pk,type:uuid"`
	UserID        uuid.UUID `bun:"user_id,type:uuid,notnull"`
	AccountID     string    `bun:"account_id,type:text"`
	ProviderID    string    `bun:"provider_id,type:text,notnull"`
	Password      *string   `bun:"password,type:text"`
	CreatedAt     time.Time `bun:"created_at,type:timestamp,notnull"`
	UpdatedAt     time.Time `bun:"updated_at,type:timestamp,notnull"`
}

// CountAccountsWithPassword returns how many rows have a password set (email/password signups).
func (u *UserStorage) CountAccountsWithPassword(ctx context.Context) (int, error) {
	return u.getDB().NewSelect().
		Model((*Account)(nil)).
		Where("password IS NOT NULL").
		Count(ctx)
}

// BootstrapOrgRow is one organization membership for bootstrap (member + organization).
type BootstrapOrgRow struct {
	ID   uuid.UUID
	Name string
	Role string
}

// ListBootstrapOrganizations returns the user's orgs via member, ordered by membership created_at.
func (u *UserStorage) ListBootstrapOrganizations(ctx context.Context, userID uuid.UUID) ([]BootstrapOrgRow, error) {
	var members []types.Member
	err := u.getDB().NewSelect().
		Model(&members).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]BootstrapOrgRow, 0, len(members))
	for _, m := range members {
		var org types.Organization
		errOrg := u.getDB().NewSelect().Model(&org).Where("id = ?", m.OrganizationID).Scan(ctx)
		if errOrg != nil {
			continue
		}
		out = append(out, BootstrapOrgRow{
			ID:   org.ID,
			Name: org.Name,
			Role: m.Role,
		})
	}
	return out, nil
}

// OrgHasSSHKeys reports whether the organization has at least one non-deleted ssh_keys row.
func (u *UserStorage) OrgHasSSHKeys(ctx context.Context, organizationID uuid.UUID) (bool, error) {
	return u.getDB().NewSelect().
		Table("ssh_keys").
		ColumnExpr("1").
		Where("organization_id = ?", organizationID).
		Where("deleted_at IS NULL").
		Limit(1).
		Exists(ctx)
}

// GetLatestUserProvisionDetails returns the most recent user provision row for the user.
func (u *UserStorage) GetLatestUserProvisionDetails(ctx context.Context, userID uuid.UUID) (*types.UserProvisionDetails, error) {
	var upd types.UserProvisionDetails
	err := u.getDB().NewSelect().
		Model(&upd).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &upd, nil
}
