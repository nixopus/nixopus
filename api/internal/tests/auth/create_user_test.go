package auth

// DEPRECATED : user is created through supertokens authentication
// import (
// 	"net/http"
// 	"testing"

// 	. "github.com/Eun/go-hit"
// 	"github.com/google/uuid"
// 	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
// 	"github.com/raghavyuva/nixopus-api/internal/tests"
// 	"github.com/raghavyuva/nixopus-api/internal/testutils"
// )

// func TestCreateUser(t *testing.T) {
// 	setup := testutils.NewTestSetup()
// 	user, org, err := setup.GetTestAuthResponse()
// 	if err != nil {
// 		t.Fatalf("failed to get test auth response: %v", err)
// 	}

// 	orgID := org.ID.String()

// 	testCases := []struct {
// 		name           string
// 		request        types.RegisterRequest
// 		expectedStatus int
// 		token          string
// 	}{
// 		{
// 			name: "Successfully create a new user",
// 			request: types.RegisterRequest{
// 				Email:        "newuser@example.com",
// 				Password:     "Password123@",
// 				Username:     "newuser",
// 				Type:         "viewer",
// 				Organization: orgID,
// 			},
// 			expectedStatus: http.StatusOK,
// 			token:          user.AccessToken,
// 		},
// 		{
// 			name: "Create user with duplicate email",
// 			request: types.RegisterRequest{
// 				Email:        "newuser@example.com",
// 				Password:     "Password123@",
// 				Username:     "newuser2",
// 				Type:         "viewer",
// 				Organization: orgID,
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 			token:          user.AccessToken,
// 		},
// 		{
// 			name: "Create user with invalid type",
// 			request: types.RegisterRequest{
// 				Email:        "invalid@example.com",
// 				Password:     "Password123@",
// 				Username:     "invalid",
// 				Type:         "invalid_type",
// 				Organization: orgID,
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 			token:          user.AccessToken,
// 		},
// 		{
// 			name: "Create user with duplicate username",
// 			request: types.RegisterRequest{
// 				Email:        "newuser3@example.com",
// 				Password:     "Password123@",
// 				Username:     "newuser",
// 				Type:         "viewer",
// 				Organization: orgID,
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 			token:          user.AccessToken,
// 		},
// 		{
// 			name: "Create user without token",
// 			request: types.RegisterRequest{
// 				Email:        "notoken@example.com",
// 				Password:     "Password123@",
// 				Username:     "notoken",
// 				Type:         "viewer",
// 				Organization: orgID,
// 			},
// 			expectedStatus: http.StatusUnauthorized,
// 			token:          "",
// 		},
// 		{
// 			name: "Create user with invalid organization ID format",
// 			request: types.RegisterRequest{
// 				Email:        "invalidorg@example.com",
// 				Password:     "Password123@",
// 				Username:     "invalidorg",
// 				Type:         "viewer",
// 				Organization: "invalid-org-id",
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 			token:          user.AccessToken,
// 		},
// 		{
// 			name: "Create user with invalid organization ID",
// 			request: types.RegisterRequest{
// 				Email:        "invalidorg@example.com",
// 				Password:     "Password123@",
// 				Username:     "invalidorg",
// 				Type:         "viewer",
// 				Organization: uuid.New().String(),
// 			},
// 			expectedStatus: http.StatusForbidden,
// 			token:          user.AccessToken,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			Test(t,
// 				Description(tc.name),
// 				Post(tests.GetCreateUserURL()),
// 				Send().Headers("Authorization").Add("Bearer "+tc.token),
// 				Send().Headers("X-Organization-Id").Add(tc.request.Organization),
// 				Send().Body().JSON(tc.request),
// 				Expect().Status().Equal(int64(tc.expectedStatus)),
// 			)
// 		})
// 	}
// }
