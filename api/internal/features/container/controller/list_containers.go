package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	containertypes "github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *ContainerController) ListContainers(fuegoCtx fuego.ContextNoBody) (*containertypes.ListContainersResponse, error) {
	params := parseContainerListParams(fuegoCtx.Request())
	ctx := fuegoCtx.Request().Context()

	ctxStr := fmt.Sprintf("page=%d page_size=%d status=%q search=%q", params.Page, params.PageSize, params.Status, params.Search)
	c.logger.Log(logger.Info, "container: ListContainers", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: ListContainers docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	resp, err := service.ListContainers(dockerService, c.logger, params)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: ListContainers: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: ListContainers ok", fmt.Sprintf("%s group_count=%d total_count=%d", ctxStr, resp.Data.GroupCount, resp.Data.TotalCount))
	return &resp, nil
}

func parseContainerListParams(r *http.Request) containertypes.ContainerListParams {
	q := r.URL.Query()
	pageStr := q.Get("page")
	pageSizeStr := q.Get("page_size")
	sortBy := strings.ToLower(strings.TrimSpace(q.Get("sort_by")))
	sortOrder := strings.ToLower(strings.TrimSpace(q.Get("sort_order")))

	if pageStr == "" {
		pageStr = "1"
	}
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	if sortBy == "" {
		sortBy = "name"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	return containertypes.ContainerListParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    strings.TrimSpace(q.Get("search")),
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Status:    strings.TrimSpace(q.Get("status")),
		Name:      strings.TrimSpace(q.Get("name")),
		Image:     strings.TrimSpace(q.Get("image")),
	}
}
