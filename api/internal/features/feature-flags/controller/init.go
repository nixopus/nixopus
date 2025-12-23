package controller

import (
	"context"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	"github.com/raghavyuva/nixopus-api/internal/features/feature-flags/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
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

func (c *FeatureFlagController) GetFeatureFlags(f fuego.ContextNoBody) (*types.Response, error) {
	organizationID := utils.GetOrganizationID(f.Request())
	flags, err := c.service.GetFeatureFlags(organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	return &types.Response{
		Status:  "success",
		Message: "Feature flags retrieved successfully",
		Data:    flags,
	}, nil
}

func (c *FeatureFlagController) UpdateFeatureFlag(f fuego.ContextWithBody[types.UpdateFeatureFlagRequest]) (*types.Response, error) {
	organizationID := utils.GetOrganizationID(f.Request())
	req, err := f.Body()

	if err != nil {
		return nil, err
	}

	if err = c.service.UpdateFeatureFlag(organizationID, req); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	// Invalidate the feature flag cache
	c.cache.InvalidateFeatureFlag(c.ctx, organizationID.String(), req.FeatureName)

	return &types.Response{
		Status:  "success",
		Message: "Feature flag updated successfully",
	}, nil
}

func (c *FeatureFlagController) IsFeatureEnabled(f fuego.ContextNoBody) (*types.Response, error) {
	organizationID := utils.GetOrganizationID(f.Request())
	featureName := f.Request().URL.Query().Get("feature_name")

	isEnabled, err := c.service.IsFeatureEnabled(organizationID, featureName)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	return &types.Response{
		Status:  "success",
		Message: "Feature flag status retrieved successfully",
		Data:    map[string]bool{"is_enabled": isEnabled},
	}, nil
}
