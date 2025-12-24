package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

	response, err := c.service.ListExtensions(params)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ListExtensionsResponse{
		Status:  "success",
		Message: "Extensions retrieved successfully",
		Data:    *response,
	}, nil
}

func (c *ExtensionsController) GetCategories(ctx fuego.ContextNoBody) (*types.CategoriesResponse, error) {
	cats, err := c.service.ListCategories()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	return &types.CategoriesResponse{
		Status:  "success",
		Message: "Categories retrieved successfully",
		Data:    cats,
	}, nil
}

func (c *ExtensionsController) GetExtension(ctx fuego.ContextNoBody) (*types.ExtensionResponse, error) {
	id := ctx.PathParam("id")
	if id == "" {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	extension, err := c.service.GetExtension(id)
	if err != nil {
		if err.Error() == "extension not found" {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ExtensionResponse{
		Status:  "success",
		Message: "Extension retrieved successfully",
		Data:    *extension,
	}, nil
}

func (c *ExtensionsController) GetExtensionByExtensionID(ctx fuego.ContextNoBody) (*types.ExtensionResponse, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	extension, err := c.service.GetExtensionByID(extensionID)
	if err != nil {
		if err.Error() == "extension not found" {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ExtensionResponse{
		Status:  "success",
		Message: "Extension retrieved successfully",
		Data:    *extension,
	}, nil
}

func (c *ExtensionsController) GetExecution(ctx fuego.ContextNoBody) (*types.ExecutionResponse, error) {
	id := ctx.PathParam("execution_id")
	if id == "" {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	exec, err := c.service.GetExecutionByID(id)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}
	return &types.ExecutionResponse{
		Status:  "success",
		Message: "Execution retrieved successfully",
		Data:    exec,
	}, nil
}

func (c *ExtensionsController) ListExecutionsByExtensionID(ctx fuego.ContextNoBody) (*types.ListExecutionsResponse, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}
	execs, err := c.service.ListExecutionsByExtensionID(extensionID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}
	return &types.ListExecutionsResponse{
		Status:  "success",
		Message: "Executions retrieved successfully",
		Data:    execs,
	}, nil
}
