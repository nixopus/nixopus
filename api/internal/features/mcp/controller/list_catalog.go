package controller

import (
	"sort"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
)

func (c *MCPController) ListCatalog(f fuego.ContextNoBody) (*Response, error) {
	q := f.Request().URL.Query()
	search := strings.ToLower(q.Get("q"))
	sortBy := q.Get("sort_by")   // "name" | "id"
	sortDir := q.Get("sort_dir") // "asc" | "desc"

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))

	// Filter
	filtered := make([]catalogEntryWithLogo, 0, len(mcp.Catalog))
	for _, p := range mcp.Catalog {
		if search == "" ||
			strings.Contains(strings.ToLower(p.Name), search) ||
			strings.Contains(strings.ToLower(p.ID), search) {
			filtered = append(filtered, withLogoURL(p))
		}
	}

	// Sort
	sort.Slice(filtered, func(i, j int) bool {
		var a, b string
		if sortBy == "id" {
			a, b = filtered[i].ID, filtered[j].ID
		} else {
			a, b = filtered[i].Name, filtered[j].Name
		}
		if sortDir == "desc" {
			return a > b
		}
		return a < b
	})

	totalCount := len(filtered)

	// Paginate
	if limit > 0 {
		offset := (page - 1) * limit
		if offset >= totalCount {
			filtered = []catalogEntryWithLogo{}
		} else {
			end := offset + limit
			if end > totalCount {
				end = totalCount
			}
			filtered = filtered[offset:end]
		}
	}

	return &Response{
		Status:  "success",
		Message: "Catalog fetched successfully",
		Data: PaginatedData[catalogEntryWithLogo]{
			Items:      filtered,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   limit,
		},
	}, nil
}
