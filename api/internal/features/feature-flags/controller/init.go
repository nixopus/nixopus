package controller

import (
	"context"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/features/feature-flags/service"
	"github.com/nixopus/nixopus/api/internal/features/feature-flags/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type FeatureFlagController struct {
	service *service.FeatureFlagService
	logger  logger.Logger
	ctx     context.Context
	cache   *cache.Cache
}

func NewFeatureFlagController(service *service.FeatureFlagService, logger logger.Logger, ctx context.Context, cache *cache.Cache) *FeatureFlagController {
	return &FeatureFlagController{
		service: service,
		logger:  logger,
		ctx:     ctx,
		cache:   cache,
	}
}

func (c *FeatureFlagController) GetFeatureFlags(f fuego.ContextNoBody) (*types.ListFeatureFlagsResponse, error) {
	organizationID := utils.GetOrganizationID(f.Request())
	ctxStr := fmt.Sprintf("org_id=%s", organizationID)
	c.logger.Log(logger.Info, "feature flags: GetFeatureFlags", ctxStr)

	flags, err := c.service.GetFeatureFlags(organizationID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("feature flags: GetFeatureFlags: %v", err), ctxStr)
		return nil, err
	}

	c.logger.Log(logger.Info, "feature flags: GetFeatureFlags ok", fmt.Sprintf("%s count=%d", ctxStr, len(flags)))
	return &types.ListFeatureFlagsResponse{
		Status:  "success",
		Message: "Feature flags retrieved successfully",
		Data:    flags,
	}, nil
}

func (c *FeatureFlagController) UpdateFeatureFlag(f fuego.ContextWithBody[shared_types.UpdateFeatureFlagRequest]) (*types.MessageResponse, error) {
	organizationID := utils.GetOrganizationID(f.Request())
	ctxStr := fmt.Sprintf("org_id=%s", organizationID)
	c.logger.Log(logger.Info, "feature flags: UpdateFeatureFlag", ctxStr)

	req, err := f.Body()

	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("feature flags: UpdateFeatureFlag body: %v", err), ctxStr)
		return nil, err
	}

	if len(req.FeatureName) > 255 {
		return nil, fuego.BadRequestError{Detail: "feature_name must not exceed 255 characters"}
	}

	ctxStr = fmt.Sprintf("org_id=%s feature_name=%s is_enabled=%t", organizationID, req.FeatureName, req.IsEnabled)
	if err = c.service.UpdateFeatureFlag(organizationID, req); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("feature flags: UpdateFeatureFlag: %v", err), ctxStr)
		return nil, err
	}

	// Invalidate the feature flag cache
	c.cache.InvalidateFeatureFlag(c.ctx, organizationID.String(), req.FeatureName)

	c.logger.Log(logger.Info, "feature flags: UpdateFeatureFlag ok", ctxStr)
	return &types.MessageResponse{
		Status:  "success",
		Message: "Feature flag updated successfully",
	}, nil
}

func (c *FeatureFlagController) IsFeatureEnabled(f fuego.ContextNoBody) (*types.IsFeatureEnabledResponse, error) {
	organizationID := utils.GetOrganizationID(f.Request())
	featureName := f.Request().URL.Query().Get("feature_name")
	ctxStr := fmt.Sprintf("org_id=%s feature_name=%s", organizationID, featureName)
	c.logger.Log(logger.Info, "feature flags: IsFeatureEnabled", ctxStr)

	isEnabled, err := c.service.IsFeatureEnabled(organizationID, featureName)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("feature flags: IsFeatureEnabled: %v", err), ctxStr)
		return nil, err
	}

	c.logger.Log(logger.Info, "feature flags: IsFeatureEnabled ok", fmt.Sprintf("%s enabled=%t", ctxStr, isEnabled))
	return &types.IsFeatureEnabledResponse{
		Status:  "success",
		Message: "Feature flag status retrieved successfully",
		Data:    types.IsFeatureEnabledData{IsEnabled: isEnabled},
	}, nil
}
