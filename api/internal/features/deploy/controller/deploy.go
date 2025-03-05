package controller

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DeployController) HandleDeploy(conn *websocket.Conn, data interface{}, user *shared_types.User) {
	c.logger.Log(logger.Info, "deploying", "")

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Log(logger.Error, "failed to marshal data", err.Error())
		conn.WriteJSON(map[string]string{"status": "error", "message": "Invalid request format"})
		return
	}

	var payload types.CreateDeploymentRequest
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		c.logger.Log(logger.Error, "failed to unmarshal data", err.Error())
		conn.WriteJSON(map[string]string{"status": "error", "message": "Invalid request format"})
		return
	}

	if err := c.validator.ValidateRequest(&payload); err != nil {
		c.logger.Log(logger.Error, "validation failed", err.Error())
		conn.WriteJSON(map[string]string{"status": "error", "message": err.Error()})
		return
	}

	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		conn.WriteJSON(map[string]string{"status": "error", "message": "User not found"})
		return
	}

	err = c.service.CreateDeployment(&payload, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, "failed to create deployment", err.Error())
		conn.WriteJSON(map[string]string{"status": "error", "message": "Failed to create deployment"})
		return
	}

	conn.WriteJSON(map[string]string{"status": "success", "message": "Deployment started"})
}
