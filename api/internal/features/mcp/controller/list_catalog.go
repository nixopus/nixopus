package controller

import (
	"github.com/go-fuego/fuego"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
)

func (c *MCPController) ListCatalog(f fuego.ContextNoBody) (*Response, error) {
	entries := make([]catalogEntryWithLogo, len(mcp.Catalog))
	for i, p := range mcp.Catalog {
		entries[i] = withLogoURL(p)
	}
	return &Response{
		Status:  "success",
		Message: "Catalog fetched successfully",
		Data:    entries,
	}, nil
}
