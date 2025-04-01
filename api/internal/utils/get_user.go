package utils

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// GetUser retrieves the current user from the request context.
//
// This method extracts the user from the request context using the UserContextKey.
// If the user cannot be retrieved or is of the wrong type, an error is logged,
// an error response is sent to the client, and the method returns nil.
//
// Parameters:
//
//	w - the HTTP response writer used to send error responses.
//	r - the HTTP request containing the context from which to retrieve the user.
//
// Returns:
//
//	*types.User - a pointer to the User object retrieved from the context,
//	or nil if the user could not be retrieved.
func GetUser(w http.ResponseWriter, r *http.Request) *types.User {
	userAny := r.Context().Value(types.UserContextKey)
	user, ok := userAny.(*types.User)

	if !ok {
		SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return nil
	}

	return user
}

func GetOrganizationID(r *http.Request) uuid.UUID {
	organizationIDAny := r.Context().Value(types.OrganizationIDKey)
	if organizationIDAny == nil {
		return uuid.Nil
	}

	if strID, ok := organizationIDAny.(string); ok {
		if id, err := uuid.Parse(strID); err == nil {
			return id
		}
	}

	if id, ok := organizationIDAny.(uuid.UUID); ok {
		return id
	}

	return uuid.Nil
}
