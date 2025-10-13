package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ExtensionsController) GetExtensions(ctx fuego.ContextNoBody) (*types.ExtensionListResponse, error) {
	params := types.ExtensionListParams{}

	categoryParam := ctx.QueryParam("category")
	if categoryParam != "" {
		cat := types.ExtensionCategory(categoryParam)
		params.Category = &cat
	}

	searchParam := ctx.QueryParam("search")
	if searchParam != "" {
		params.Search = searchParam
	}

	if typeParam := ctx.QueryParam("type"); typeParam != "" {
		et := types.ExtensionType(typeParam)
		params.Type = &et
	}

	sortByParam := ctx.QueryParam("sort_by")
	if sortByParam != "" {
		params.SortBy = types.ExtensionSortField(sortByParam)
	}

	sortDirParam := ctx.QueryParam("sort_dir")
	if sortDirParam != "" {
		params.SortDir = types.SortDirection(sortDirParam)
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

	return response, nil
}

func (c *ExtensionsController) GetCategories(ctx fuego.ContextNoBody) ([]types.ExtensionCategory, error) {
	cats, err := c.service.ListCategories()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	return cats, nil
}

func (c *ExtensionsController) GetExtension(ctx fuego.ContextNoBody) (types.Extension, error) {
	id := ctx.PathParam("id")
	if id == "" {
		return types.Extension{}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	extension, err := c.service.GetExtension(id)
	if err != nil {
		if err.Error() == "extension not found" {
			return types.Extension{}, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
		c.logger.Log(logger.Error, err.Error(), "")
		return types.Extension{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return *extension, nil
}

func (c *ExtensionsController) GetExtensionByExtensionID(ctx fuego.ContextNoBody) (types.Extension, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		return types.Extension{}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	extension, err := c.service.GetExtensionByID(extensionID)
	if err != nil {
		if err.Error() == "extension not found" {
			return types.Extension{}, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
		c.logger.Log(logger.Error, err.Error(), "")
		return types.Extension{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return *extension, nil
}

func (c *ExtensionsController) GetExecution(ctx fuego.ContextNoBody) (*types.ExtensionExecution, error) {
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
	return exec, nil
}

func (c *ExtensionsController) ListExecutionsByExtensionID(ctx fuego.ContextNoBody) ([]types.ExtensionExecution, error) {
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
	return execs, nil
}
