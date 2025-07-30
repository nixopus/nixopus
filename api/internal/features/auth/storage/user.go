package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"

	userTypes "github.com/raghavyuva/nixopus-api/internal/features/auth/types"
)

type UserStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type AuthRepository interface {
	FindUserByEmail(email string) (*types.User, error)
	FindUserByUsername(username string) (*types.User, error)
	FindUserByID(id string) (*types.User, error)
	CreateUser(user *types.User) error
	UpdateUser(user *types.User) error
	CreateRefreshToken(user_id uuid.UUID) (*types.RefreshToken, error)
	GetRefreshToken(refreshToken string) (*types.RefreshToken, error)
	GetResetToken(token string) (*types.User, error)
	RevokeRefreshToken(refreshToken string) error
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) AuthRepository
	StoreVerificationToken(userID string, token string, expiresAt time.Time) error
	GetVerificationToken(token string) (string, time.Time, error)
	DeleteVerificationToken(token string) error
	UpdateUserEmailVerification(userID string, verified bool) error
	FindUserByType(userType string) (*types.User, error)
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
		Relation("Role").
		Relation("Organization").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	for i, orgUser := range user.OrganizationUsers {
		if orgUser.Role != nil {
			err = u.getDB().NewSelect().
				Model(&user.OrganizationUsers[i].Role.Permissions).
				Join("JOIN role_permissions AS rp ON rp.permission_id = p.id").
				Where("rp.role_id = ?", orgUser.Role.ID).
				Scan(u.Ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return user, nil
}

// FindUserByUsername finds a user by username in the database.
//
// The function returns an error if the user does not exist or if the query
// fails.
func (u *UserStorage) FindUserByUsername(username string) (*types.User, error) {
	user := &types.User{}
	err := u.getDB().NewSelect().
		Model(user).
		Where("username = ?", username).
		Relation("Organizations").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	err = u.getDB().NewSelect().
		Model(&user.OrganizationUsers).
		Where("user_id = ?", user.ID).
		Relation("Role").
		Relation("Organization").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	for i, orgUser := range user.OrganizationUsers {
		if orgUser.Role != nil {
			err = u.getDB().NewSelect().
				Model(&user.OrganizationUsers[i].Role.Permissions).
				Join("JOIN role_permissions AS rp ON rp.permission_id = p.id").
				Where("rp.role_id = ?", orgUser.Role.ID).
				Scan(u.Ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return user, nil
}

// FindUserByID finds a user by id in the database.
//
// The function returns an error if the user does not exist or if the query
// fails.
func (u *UserStorage) FindUserByID(id string) (*types.User, error) {
	user := &types.User{}
	err := u.getDB().NewSelect().Model(user).Where("id = ?", id).Scan(u.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUser inserts a new user into the database.
//
// The function takes a pointer to a `types.User` struct as an argument.
// The function returns an error if the query fails.
func (u *UserStorage) CreateUser(user *types.User) error {
	_, err := u.getDB().NewInsert().Model(user).Exec(u.Ctx)
	return err
}

// UpdateUser updates an existing user's information in the database.
//
// The function takes a pointer to a `types.User` struct, which contains
// the updated user information. The update operation is performed based
// on the user's ID.
//
// Returns an error if the update query fails.
func (u *UserStorage) UpdateUser(user *types.User) error {
	_, err := u.getDB().NewUpdate().Model(user).Where("id = ?", user.ID).Exec(u.Ctx)
	return err
}

// GetResetToken retrieves a user by their reset token.
//
// This function takes a reset token as input and queries the database to find
// the associated user. If the token is found, it returns a pointer to the
// `types.User` object. If the token is not found or if the query fails, it
// returns an error.
//
// The function is used in scenarios where password reset operations are
// needed, as it helps verify the validity of the reset token.
func (u *UserStorage) GetResetToken(token string) (*types.User, error) {
	user := &types.User{}
	err := u.getDB().NewSelect().Model(user).Where("reset_token = ?", token).Scan(u.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateRefreshToken creates a new refresh token for a user in the database.
//
// The function takes a UUID as input, which is the ID of the user for whom
// the refresh token should be created. It generates a new refresh token with
// a random token string and an expiration time set to 30 days from the moment
// of creation. The function returns a pointer to the newly created `types.RefreshToken`
// object if the creation is successful, or an error if the insertion query fails.
func (u *UserStorage) CreateRefreshToken(userID uuid.UUID) (*types.RefreshToken, error) {
	refreshToken := &types.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		CreatedAt: time.Now(),
	}

	_, err := u.getDB().NewInsert().Model(refreshToken).Exec(u.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, nil
}

// GetRefreshToken retrieves a refresh token by its token string.
//
// The function takes a refresh token string as input and queries the database
// to find the associated refresh token. If the token is found and is valid
// (i.e., has not expired and has not been revoked), it returns a pointer to
// the `types.RefreshToken` object. If the token is not found, or if the query
// fails, or if the token has expired or been revoked, it returns an error.
func (u *UserStorage) GetRefreshToken(token string) (*types.RefreshToken, error) {
	var refreshToken types.RefreshToken
	err := u.getDB().NewSelect().Model(&refreshToken).Where("token = ?", token).Scan(u.Ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	if refreshToken.RevokedAt != nil {
		return nil, fmt.Errorf("refresh token revoked")
	}

	return &refreshToken, nil
}

// RevokeRefreshToken sets the RevokedAt field of a refresh token to the current time.
//
// This function takes a refresh token string as input and updates the corresponding
// entry in the database to mark it as revoked. If the operation is successful, it
// returns nil. Otherwise, it returns an error indicating the failure to update the
// database.
func (u *UserStorage) RevokeRefreshToken(token string) error {
	now := time.Now()
	_, err := u.getDB().NewUpdate().
		Model(&types.RefreshToken{}).
		Set("revoked_at = ?", now).
		Where("token = ?", token).
		Exec(u.Ctx)
	return err
}

func (u *UserStorage) StoreVerificationToken(userID string, token string, expiresAt time.Time) error {
	log.Printf("Attempting to store verification token for user %s with token %s", userID, token)

	verificationToken := &userTypes.VerificationToken{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	var existingToken userTypes.VerificationToken
	err := u.getDB().NewSelect().Model(&existingToken).Where("token = ?", token).Scan(u.Ctx)
	if err == nil {
		log.Printf("Token %s already exists for user %s", token, userID)
		return fmt.Errorf("token already exists")
	}

	_, err = u.getDB().NewInsert().Model(verificationToken).Exec(u.Ctx)
	if err != nil {
		log.Printf("Failed to store verification token for user %s: %v", userID, err)
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	var storedToken userTypes.VerificationToken
	err = u.getDB().NewSelect().Model(&storedToken).Where("token = ?", token).Scan(u.Ctx)
	if err != nil {
		log.Printf("Failed to verify token storage for user %s: %v", userID, err)
		return fmt.Errorf("failed to verify token storage: %w", err)
	}

	log.Printf("Successfully stored and verified token for user %s", userID)
	return nil
}

func (u *UserStorage) GetVerificationToken(token string) (string, time.Time, error) {
	log.Printf("Attempting to retrieve verification token %s", token)

	var verificationToken userTypes.VerificationToken
	err := u.getDB().NewSelect().
		Model(&verificationToken).
		Where("token = ?", token).
		Scan(u.Ctx)

	if err != nil {
		log.Printf("Failed to get verification token %s: %v", token, err)
		return "", time.Time{}, fmt.Errorf("verification token not found: %w", err)
	}

	if time.Now().After(verificationToken.ExpiresAt) {
		log.Printf("Token %s has expired for user %s", token, verificationToken.UserID)
		return "", time.Time{}, fmt.Errorf("verification token expired")
	}

	log.Printf("Successfully retrieved verification token for user %s", verificationToken.UserID)
	return verificationToken.UserID.String(), verificationToken.ExpiresAt, nil
}

func (u *UserStorage) DeleteVerificationToken(token string) error {
	_, err := u.getDB().NewDelete().Model(&userTypes.VerificationToken{}).Where("token = ?", token).Exec(u.Ctx)
	return err
}

func (u *UserStorage) UpdateUserEmailVerification(userID string, verified bool) error {
	log.Printf("Updating email verification status for user %s to %v", userID, verified)

	_, err := u.getDB().NewUpdate().
		Model(&types.User{}).
		Set("is_verified = ?", verified).
		Where("id = ?", userID).
		Exec(u.Ctx)

	if err != nil {
		log.Printf("Failed to update email verification status for user %s: %v", userID, err)
		return fmt.Errorf("failed to update email verification status: %w", err)
	}

	log.Printf("Successfully updated email verification status for user %s", userID)
	return nil
}

// UserBelongsToOrganization checks if a user belongs to a specific organization
func (u *UserStorage) UserBelongsToOrganization(userID string, organizationID string) (bool, error) {
	count, err := u.getDB().NewSelect().
		Model((*types.OrganizationUsers)(nil)).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Count(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to check organization membership: %w", err)
	}
	return count > 0, nil
}

// FindUserByType finds a user by their type in the database.
//
// The function takes a user type as input and queries the database to find
// the associated user. If the user type is found, it returns a pointer to the
// `types.User` object. If the user type is not found or if the query fails, it
// returns an error.
func (u *UserStorage) FindUserByType(userType string) (*types.User, error) {
	user := &types.User{}
	err := u.getDB().NewSelect().Model(user).Where("type = ?", userType).Scan(u.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
