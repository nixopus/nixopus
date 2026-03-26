package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
)

func (c *MCPController) GetProviderIcon(f fuego.ContextNoBody) (any, error) {
	providerID := f.PathParam("provider_id")
	if mcp.GetProvider(providerID) == nil {
		return nil, fuego.NotFoundError{Detail: "provider not found"}
	}

	data, err := mcp.IconFS.ReadFile(fmt.Sprintf("icons/%s.svg", providerID))
	if err != nil {
		return nil, fuego.NotFoundError{Detail: "icon not found"}
	}

	w := f.Response()
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return nil, nil
}
