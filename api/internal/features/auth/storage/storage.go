package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type UserStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	tx     *bun.Tx
	Logger *logger.Logger // optional; nil disables storage logs
}

func (u *UserStorage) storageLog(sev logger.Severity, msg, data string) {
	if u.Logger == nil {
		return
	}
	u.Logger.Log(sev, msg, data)
}

type AuthRepository interface {
	FindUserByEmail(email string) (*types.User, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) AuthRepository
}

func (u *UserStorage) WithTx(tx bun.Tx) AuthRepository {
	return &UserStorage{
		DB:     u.DB,
		Ctx:    u.Ctx,
		tx:     &tx,
		Logger: u.Logger,
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
	u.storageLog(logger.Debug, "storage: FindUserByEmail", fmt.Sprintf("email=%q", email))

	user := &types.User{}
	err := u.getDB().NewSelect().
		Model(user).
		Where("email = ?", email).
		Relation("Organizations").
		Scan(u.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.storageLog(logger.Debug, "storage: FindUserByEmail not found", fmt.Sprintf("email=%q", email))
		} else {
			u.storageLog(logger.Error, fmt.Sprintf("storage: FindUserByEmail user scan: %v", err), fmt.Sprintf("email=%q", email))
		}
		return nil, err
	}

	err = u.getDB().NewSelect().
		Model(&user.OrganizationUsers).
		Where("user_id = ?", user.ID).
		Relation("Organization").
		Scan(u.Ctx)
	if err != nil {
		u.storageLog(logger.Error, fmt.Sprintf("storage: FindUserByEmail org users scan: %v", err), fmt.Sprintf("email=%q user_id=%s", email, user.ID))
		return nil, err
	}

	user.ComputeCompatibilityFields()

	u.storageLog(logger.Debug, "storage: FindUserByEmail ok", fmt.Sprintf("email=%q user_id=%s", email, user.ID))
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
	u.storageLog(logger.Debug, "storage: CountAccountsWithPassword", "")
	n, err := u.getDB().NewSelect().
		Model((*Account)(nil)).
		Where("password IS NOT NULL").
		Count(ctx)
	if err != nil {
		u.storageLog(logger.Error, fmt.Sprintf("storage: CountAccountsWithPassword: %v", err), "")
		return 0, err
	}
	u.storageLog(logger.Debug, "storage: CountAccountsWithPassword ok", fmt.Sprintf("count=%d", n))
	return n, nil
}

// BootstrapOrgRow is one organization membership for bootstrap (member + organization).
type BootstrapOrgRow struct {
	ID   uuid.UUID
	Name string
	Role string
}

// ListBootstrapOrganizations returns the user's orgs via member, ordered by membership created_at.
func (u *UserStorage) ListBootstrapOrganizations(ctx context.Context, userID uuid.UUID) ([]BootstrapOrgRow, error) {
	ctxStr := fmt.Sprintf("user_id=%s", userID)
	u.storageLog(logger.Debug, "storage: ListBootstrapOrganizations", ctxStr)

	var members []types.Member
	err := u.getDB().NewSelect().
		Model(&members).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		u.storageLog(logger.Error, fmt.Sprintf("storage: ListBootstrapOrganizations members: %v", err), ctxStr)
		return nil, err
	}

	out := make([]BootstrapOrgRow, 0, len(members))
	for _, m := range members {
		var org types.Organization
		errOrg := u.getDB().NewSelect().Model(&org).Where("id = ?", m.OrganizationID).Scan(ctx)
		if errOrg != nil {
			u.storageLog(logger.Debug, fmt.Sprintf("storage: ListBootstrapOrganizations skip org: %v", errOrg), fmt.Sprintf("%s org_id=%s", ctxStr, m.OrganizationID))
			continue
		}
		out = append(out, BootstrapOrgRow{
			ID:   org.ID,
			Name: org.Name,
			Role: m.Role,
		})
	}
	u.storageLog(logger.Debug, "storage: ListBootstrapOrganizations ok", fmt.Sprintf("%s count=%d", ctxStr, len(out)))
	return out, nil
}

// OrgHasSSHKeys reports whether the organization has at least one non-deleted ssh_keys row.
func (u *UserStorage) OrgHasSSHKeys(ctx context.Context, organizationID uuid.UUID) (bool, error) {
	ctxStr := fmt.Sprintf("org_id=%s", organizationID)
	u.storageLog(logger.Debug, "storage: OrgHasSSHKeys", ctxStr)

	exists, err := u.getDB().NewSelect().
		Table("ssh_keys").
		ColumnExpr("1").
		Where("organization_id = ?", organizationID).
		Where("deleted_at IS NULL").
		Limit(1).
		Exists(ctx)
	if err != nil {
		u.storageLog(logger.Error, fmt.Sprintf("storage: OrgHasSSHKeys: %v", err), ctxStr)
		return false, err
	}
	u.storageLog(logger.Debug, "storage: OrgHasSSHKeys ok", fmt.Sprintf("%s exists=%v", ctxStr, exists))
	return exists, nil
}

// GetLatestUserProvisionDetails returns the most recent user provision row for the user.
func (u *UserStorage) GetLatestUserProvisionDetails(ctx context.Context, userID uuid.UUID) (*types.UserProvisionDetails, error) {
	ctxStr := fmt.Sprintf("user_id=%s", userID)
	u.storageLog(logger.Debug, "storage: GetLatestUserProvisionDetails", ctxStr)

	var upd types.UserProvisionDetails
	err := u.getDB().NewSelect().
		Model(&upd).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.storageLog(logger.Debug, "storage: GetLatestUserProvisionDetails not found", ctxStr)
		} else {
			u.storageLog(logger.Error, fmt.Sprintf("storage: GetLatestUserProvisionDetails: %v", err), ctxStr)
		}
		return nil, err
	}
	u.storageLog(logger.Debug, "storage: GetLatestUserProvisionDetails ok", fmt.Sprintf("%s provision_id=%s", ctxStr, upd.ID))
	return &upd, nil
}
