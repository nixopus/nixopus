package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/agent/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *AgentController) StreamChat(f fuego.ContextNoBody) (any, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	var req service.StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid request body", Err: err}
	}
	if strings.TrimSpace(req.Input) == "" {
		return nil, fuego.BadRequestError{Detail: "input is required"}
	}

	if req.Model == "" {
		req.Model = resolveModel(r)
	}

	authToken := extractToken(r.Header.Get("Authorization"))

	// Disable the server's WriteTimeout for this long-lived SSE connection.
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		c.logger.Log(logger.Error, "failed to disable SSE write deadline: "+err.Error(), "")
		// Continue anyway; SSE may still work with timeouts.
	}

	rw := unwrapFlusher(w)
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	if err := c.service.StreamChat(r.Context(), rw, req, authToken, user.ID.String(), orgID.String()); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: StreamChat error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "stream failed", Status: http.StatusInternalServerError}
	}

	return nil, nil
}

type flushableWriter struct {
	http.ResponseWriter
}

func (fw *flushableWriter) Flush() {
	if f, ok := fw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func unwrapFlusher(w http.ResponseWriter) http.ResponseWriter {
	if _, ok := w.(http.Flusher); ok {
		return w
	}
	return &flushableWriter{ResponseWriter: w}
}
