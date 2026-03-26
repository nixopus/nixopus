package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
)

func (c *MCPController) TestServer(f fuego.ContextWithBody[validation.TestServerRequest]) (*Response, error) {
	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := validation.ValidateTestRequest(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	result := c.service.TestServer(&body)

	return &Response{
		Status:  "success",
		Message: "Test complete",
		Data:    result,
	}, nil
}
