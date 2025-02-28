package types

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type UserOrganizationsResponse struct {
	Organization shared_types.Organization `json:"organization"`
	Role         shared_types.Role         `json:"role"`
}
