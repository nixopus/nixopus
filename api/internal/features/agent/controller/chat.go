package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/agent/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *AgentController) Chat(f fuego.ContextWithBody[service.ChatRequest]) (*service.ChatResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if strings.TrimSpace(body.Input) == "" {
		return nil, fuego.BadRequestError{Detail: "input is required"}
	}

	if body.Model == "" {
		body.Model = resolveModel(r)
	}

	authToken := extractToken(r.Header.Get("Authorization"))

	resp, err := c.service.Chat(r.Context(), body, authToken, user.ID.String(), orgID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: Chat error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "agent execution failed", Status: http.StatusInternalServerError}
	}

	if resp.BalanceCents != nil {
		w.Header().Set("X-Credits-Remaining", strconv.Itoa(*resp.BalanceCents))
	}

	return resp, nil
}

func extractToken(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return header
}

// resolveModel extracts model from X-Model-Id header or ?model query param.
// Header takes priority over query param (matches Mastra behavior).
func resolveModel(r *http.Request) string {
	if m := r.Header.Get("X-Model-Id"); m != "" {
		return m
	}
	return r.URL.Query().Get("model")
}
