package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/features/auth/storage"
	auth_types "github.com/nixopus/nixopus/api/internal/features/auth/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type AuthService struct {
	storage storage.AuthRepository
	db      *bun.DB
	Cache   *cache.Cache
	Ctx     context.Context
	logger  logger.Logger
}

func NewAuthService(
	repo storage.AuthRepository,
	db *bun.DB,
	l logger.Logger,
	ctx context.Context,
	redisURL string,
) *AuthService {
	var authCache *cache.Cache
	if redisURL != "" {
		c, err := cache.NewCache(redisURL)
		if err != nil {
			l.Log(logger.Error, "failed to create auth cache, proceeding without cache", err.Error())
		} else {
			authCache = c
		}
	}

	return &AuthService{
		storage: repo,
		db:      db,
		Cache:   authCache,
		logger:  l,
		Ctx:     ctx,
	}
}

func (s *AuthService) GetUserByEmail(email string) (*shared_types.User, error) {
	return s.storage.FindUserByEmail(email)
}

func (s *AuthService) userStorage(ctx context.Context) (*storage.UserStorage, error) {
	if s.db == nil {
		return nil, fmt.Errorf("auth service: database not configured")
	}
	return &storage.UserStorage{DB: s.db, Ctx: ctx}, nil
}

// BuildBootstrap assembles bootstrap session payload (orgs, active org, servers, provisioning).
func (s *AuthService) BuildBootstrap(ctx context.Context, user *shared_types.User, sessionActiveOrgID *string) (*auth_types.BootstrapResponse, error) {
	st, err := s.userStorage(ctx)
	if err != nil {
		return nil, err
	}

	provisionStatus := "pending"
	if user.ProvisionStatus != nil && *user.ProvisionStatus != "" {
		provisionStatus = *user.ProvisionStatus
	}

	orgRows, err := st.ListBootstrapOrganizations(ctx, user.ID)
	if err != nil {
		s.logger.Log(logger.Error, "bootstrap: failed to query members", err.Error())
		orgRows = nil
	}

	orgs := make([]auth_types.BootstrapOrg, 0, len(orgRows))
	var firstOrgID string
	for _, row := range orgRows {
		orgs = append(orgs, auth_types.BootstrapOrg{
			ID:   row.ID.String(),
			Name: row.Name,
			Role: row.Role,
		})
		if firstOrgID == "" {
			firstOrgID = row.ID.String()
		}
	}

	activeOrgID := firstOrgID
	if sessionActiveOrgID != nil && *sessionActiveOrgID != "" {
		activeOrgID = *sessionActiveOrgID
	}
	var activeOrgIDPtr *string
	if activeOrgID != "" {
		activeOrgIDPtr = &activeOrgID
	}

	hasServers := false
	if activeOrgID != "" {
		if orgUUID, perr := uuid.Parse(activeOrgID); perr == nil {
			exists, errExists := st.OrgHasSSHKeys(ctx, orgUUID)
			if errExists == nil && exists {
				hasServers = true
			}
		}
	}

	var provisionID *string
	var provisionStep *string
	if provisionStatus == "provisioning" {
		upd, errUpd := st.GetLatestUserProvisionDetails(ctx, user.ID)
		if errUpd != nil {
			s.logger.Log(logger.Error, "bootstrap: failed to query user_provision_details", errUpd.Error())
			return nil, errUpd
		}
		id := upd.ID.String()
		provisionID = &id
		if upd.Step != nil {
			step := string(*upd.Step)
			provisionStep = &step
		}
	}

	return &auth_types.BootstrapResponse{
		User: auth_types.BootstrapUser{
			ID:              user.ID.String(),
			Name:            user.Name,
			Email:           user.Email,
			IsOnboarded:     user.IsOnboarded,
			ProvisionStatus: provisionStatus,
		},
		Organizations:        orgs,
		ActiveOrganizationID: activeOrgIDPtr,
		HasServers:           hasServers,
		ProvisionID:          provisionID,
		ProvisionStep:        provisionStep,
	}, nil
}

// GetAdminRegistered returns whether a password credential account exists (cache-aside).
func (s *AuthService) GetAdminRegistered(ctx context.Context) (bool, error) {
	if s.Cache != nil {
		registered, hit, err := s.Cache.GetAdminRegistered(ctx)
		if err != nil {
			s.logger.Log(logger.Error, "cache read failed, falling through to db", err.Error())
		}
		if hit && err == nil {
			return registered, nil
		}
	}

	st, err := s.userStorage(ctx)
	if err != nil {
		return false, err
	}

	count, err := st.CountAccountsWithPassword(ctx)
	if err != nil {
		return false, err
	}

	adminRegistered := count > 0

	if s.Cache != nil {
		if cacheErr := s.Cache.SetAdminRegistered(ctx, adminRegistered); cacheErr != nil {
			s.logger.Log(logger.Error, "failed to cache admin registration status", cacheErr.Error())
		}
	}

	return adminRegistered, nil
}
