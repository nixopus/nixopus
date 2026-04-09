package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// parseTimeRange reads 'from' and 'to' query params (RFC3339). Default window: last 1 hour.
func parseTimeRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now().UTC()
	from := now.Add(-time.Hour)
	to := now
	if s := r.URL.Query().Get("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			from = t
		}
	}
	if s := r.URL.Query().Get("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			to = t
		}
	}
	return from, to
}

func parseLimit(r *http.Request, def int) int {
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func (c *MachineController) GetMachineMetrics(f fuego.ContextNoBody) (*machine_types.MachineMetricsResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	from, to := parseTimeRange(r)
	limit := parseLimit(r, 500)
	serverID := parseServerID(r)

	resp, err := c.metricsService.GetMetrics(r.Context(), orgID, serverID, from, to, limit)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
	return resp, nil
}

func (c *MachineController) GetMachineEvents(f fuego.ContextNoBody) (*machine_types.MachineEventsResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	from, to := parseTimeRange(r)
	limit := parseLimit(r, 200)
	serverID := parseServerID(r)

	resp, err := c.metricsService.GetEvents(r.Context(), orgID, serverID, from, to, limit)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
	return resp, nil
}

func (c *MachineController) GetMachineMetricsSummary(f fuego.ContextNoBody) (*machine_types.MachineSummaryResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	from, to := parseTimeRange(r)
	serverID := parseServerID(r)

	resp, err := c.metricsService.GetSummary(r.Context(), orgID, serverID, from, to)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
	return resp, nil
}
