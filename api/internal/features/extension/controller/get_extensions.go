package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/extension/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (c *ExtensionsController) GetExtensions(ctx fuego.ContextNoBody) (*types.ListExtensionsResponse, error) {
	params := shared_types.ExtensionListParams{}

	categoryParam := ctx.QueryParam("category")
	if categoryParam != "" {
		cat := shared_types.ExtensionCategory(categoryParam)
		params.Category = &cat
	}

	searchParam := ctx.QueryParam("search")
	if searchParam != "" {
		params.Search = searchParam
	}

	if typeParam := ctx.QueryParam("type"); typeParam != "" {
		et := shared_types.ExtensionType(typeParam)
		params.Type = &et
	}

	sortByParam := ctx.QueryParam("sort_by")
	if sortByParam != "" {
		params.SortBy = shared_types.ExtensionSortField(sortByParam)
	}

	sortDirParam := ctx.QueryParam("sort_dir")
	if sortDirParam != "" {
		params.SortDir = shared_types.SortDirection(sortDirParam)
	}

	pageParam := ctx.QueryParam("page")
	if pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil && page > 0 {
			params.Page = page
		}
	}

	pageSizeParam := ctx.QueryParam("page_size")
	if pageSizeParam != "" {
		if pageSize, err := strconv.Atoi(pageSizeParam); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}

	var ctxParts []string
	ctxParts = append(ctxParts, fmt.Sprintf("page=%d", params.Page), fmt.Sprintf("page_size=%d", params.PageSize))
	ctxParts = append(ctxParts, fmt.Sprintf("sort_by=%s", params.SortBy), fmt.Sprintf("sort_dir=%s", params.SortDir))
	if params.Category != nil {
		ctxParts = append(ctxParts, fmt.Sprintf("category=%s", *params.Category))
	}
	if params.Type != nil {
		ctxParts = append(ctxParts, fmt.Sprintf("type=%s", *params.Type))
	}
	if params.Search != "" {
		ctxParts = append(ctxParts, fmt.Sprintf("search=%q", params.Search))
	}
	ctxStr := strings.Join(ctxParts, " ")
	c.logger.Log(logger.Info, "extension: GetExtensions", ctxStr)

	response, err := c.service.ListExtensions(params)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("extension: GetExtensions: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "extension: GetExtensions ok", fmt.Sprintf("%s total=%d total_pages=%d", ctxStr, response.Total, response.TotalPages))
	return &types.ListExtensionsResponse{
		Status:  "success",
		Message: "Extensions retrieved successfully",
		Data:    *response,
	}, nil
}

func (c *ExtensionsController) GetCategories(ctx fuego.ContextNoBody) (*types.CategoriesResponse, error) {
	c.logger.Log(logger.Info, "extension: GetCategories", "")
	cats, err := c.service.ListCategories()
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("extension: GetCategories: %v", err), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
	c.logger.Log(logger.Info, "extension: GetCategories ok", fmt.Sprintf("count=%d", len(cats)))
	return &types.CategoriesResponse{
		Status:  "success",
		Message: "Categories retrieved successfully",
		Data:    cats,
	}, nil
}

func (c *ExtensionsController) GetExtension(ctx fuego.ContextNoBody) (*types.ExtensionResponse, error) {
	id := ctx.PathParam("id")
	if id == "" {
		c.logger.Log(logger.Debug, "extension: GetExtension missing id", "")
		return nil, fuego.BadRequestError{
			Detail: "extension ID is required",
		}
	}

	ctxStr := fmt.Sprintf("id=%s", id)
	c.logger.Log(logger.Info, "extension: GetExtension", ctxStr)

	extension, err := c.service.GetExtension(id)
	if err != nil {
		if err.Error() == "extension not found" {
			return nil, fuego.NotFoundError{
				Detail: err.Error(),
				Err:    err,
			}
		}
		c.logger.Log(logger.Error, fmt.Sprintf("extension: GetExtension: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "extension: GetExtension ok", fmt.Sprintf("%s extension_id=%s", ctxStr, extension.ExtensionID))
	return &types.ExtensionResponse{
		Status:  "success",
		Message: "Extension retrieved successfully",
		Data:    *extension,
	}, nil
}

func (c *ExtensionsController) GetExtensionByExtensionID(ctx fuego.ContextNoBody) (*types.ExtensionResponse, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		c.logger.Log(logger.Debug, "extension: GetExtensionByExtensionID missing extension_id", "")
		return nil, fuego.BadRequestError{
			Detail: "extension ID is required",
		}
	}

	ctxStr := fmt.Sprintf("extension_id=%s", extensionID)
	c.logger.Log(logger.Info, "extension: GetExtensionByExtensionID", ctxStr)

	extension, err := c.service.GetExtensionByID(extensionID)
	if err != nil {
		if err.Error() == "extension not found" {
			return nil, fuego.NotFoundError{
				Detail: err.Error(),
				Err:    err,
			}
		}
		c.logger.Log(logger.Error, fmt.Sprintf("extension: GetExtensionByExtensionID: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "extension: GetExtensionByExtensionID ok", ctxStr)
	return &types.ExtensionResponse{
		Status:  "success",
		Message: "Extension retrieved successfully",
		Data:    *extension,
	}, nil
}
