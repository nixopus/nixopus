package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DeployController) HandleGithubWebhook(f fuego.ContextNoBody) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "handling github webhook", "")

	payload, err := io.ReadAll(f.Request().Body)
	if err != nil {
		c.logger.Log(logger.Error, "failed to read webhook payload", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	signature := f.Request().Header.Get("X-Hub-Signature-256")
	if signature == "" {
		c.logger.Log(logger.Error, "missing webhook signature", "")
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("missing webhook signature"),
			Status: http.StatusUnauthorized,
		}
	}

	eventType := f.Request().Header.Get("X-GitHub-Event")
	if eventType != "push" {
		c.logger.Log(logger.Info, "ignoring non-push event", eventType)
		return &shared_types.Response{
			Status:  "success",
			Message: "Ignored non-push event",
			Data:    nil,
		}, nil
	}

	fmt.Printf("payload: %+v\n", string(payload))

	var webhookPayload shared_types.WebhookPayload

	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		c.logger.Log(logger.Error, "failed to parse webhook payload", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// err = c.service.HandleGithubWebhook(webhookPayload)
	// if err != nil {
	// 	c.logger.Log(logger.Error, "failed to handle github webhook", err.Error())
	// 	return nil, fuego.HTTPError{
	// 		Err:    err,
	// 		Status: http.StatusInternalServerError,
	// 	}
	// }
	err = c.taskService.EnqueueWebhookTask(webhookPayload)
	if err != nil {
		c.logger.Log(logger.Error, "failed to enqueue webhook task", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github webhook handled successfully", webhookPayload.Repository.FullName)
	return &shared_types.Response{
		Status:  "success",
		Message: "Github webhook handled successfully",
		Data:    nil,
	}, nil
}
