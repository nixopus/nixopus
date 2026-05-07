package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
)

type ThreadResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (c *AgentController) CreateThread(f fuego.ContextNoBody) (*ThreadResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	var body struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	thread, err := c.service.CreateThreadWithOpts(r.Context(), user.ID.String(), orgID.String(), body.ID, body.Title)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: CreateThread error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "failed to create thread", Status: http.StatusInternalServerError}
	}

	return &ThreadResponse{
		Status:  "success",
		Message: "Thread created",
		Data:    thread,
	}, nil
}

func (c *AgentController) ListThreads(f fuego.ContextNoBody) (*ThreadResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	threads, err := c.service.ListThreads(r.Context(), user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: ListThreads error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "failed to list threads", Status: http.StatusInternalServerError}
	}

	return &ThreadResponse{
		Status:  "success",
		Message: "Threads retrieved",
		Data:    threads,
	}, nil
}

func (c *AgentController) GetThreadMessages(f fuego.ContextNoBody) (*ThreadResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	threadID := f.PathParam("threadId")
	if threadID == "" {
		return nil, fuego.BadRequestError{Detail: "thread_id is required"}
	}

	messages, err := c.service.GetThreadMessages(r.Context(), threadID, user.ID.String(), 50)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: GetMessages error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "failed to get messages", Status: http.StatusInternalServerError}
	}

	return &ThreadResponse{
		Status:  "success",
		Message: "Messages retrieved",
		Data:    messages,
	}, nil
}

func (c *AgentController) UpdateThread(f fuego.ContextNoBody) (*ThreadResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	threadID := f.PathParam("threadId")
	if threadID == "" {
		return nil, fuego.BadRequestError{Detail: "thread_id is required"}
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid request body", Err: err}
	}
	if body.Title == "" {
		return nil, fuego.BadRequestError{Detail: "title is required"}
	}

	if err := c.service.UpdateThread(r.Context(), threadID, user.ID.String(), body.Title); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: UpdateThread error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "failed to update thread", Status: http.StatusInternalServerError}
	}

	return &ThreadResponse{
		Status:  "success",
		Message: "Thread updated",
	}, nil
}

func (c *AgentController) DeleteThread(f fuego.ContextNoBody) (*ThreadResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	threadID := f.PathParam("threadId")
	if threadID == "" {
		return nil, fuego.BadRequestError{Detail: "thread_id is required"}
	}

	if err := c.service.DeleteThread(r.Context(), threadID, user.ID.String()); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("agent: DeleteThread error: %v", err), "")
		return nil, fuego.HTTPError{Detail: "failed to delete thread", Status: http.StatusInternalServerError}
	}

	return &ThreadResponse{
		Status:  "success",
		Message: "Thread deleted",
	}, nil
}

// ThreadMessagesResponse is a typed response for the messages endpoint
type ThreadMessagesResponse struct {
	Status   string                 `json:"status"`
	Message  string                 `json:"message"`
	Messages []memory.StoredMessage `json:"messages"`
}
