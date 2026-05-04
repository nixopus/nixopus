package controller

import (
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
)

func (c *MCPController) TestServer(f fuego.ContextWithBody[validation.TestServerRequest]) (*Response, error) {
	body, err := f.Body()
	if err != nil {
		c.logMCPDebug("TestServer", "invalid body", "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := validation.ValidateTestRequest(&body); err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("mcp: TestServer validation: %v", err), fmt.Sprintf("provider_id=%s", body.ProviderID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	result := c.service.TestServer(&body)

	return &Response{
		Status:  "success",
		Message: "Test complete",
		Data:    result,
	}, nil
}
