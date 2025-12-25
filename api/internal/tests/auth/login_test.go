package auth

// Deprecated: Makes use of supertokens authentication
// import (
// 	"net/http"
// 	"testing"

// 	. "github.com/Eun/go-hit"
// 	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
// 	"github.com/raghavyuva/nixopus-api/internal/tests"
// 	"github.com/raghavyuva/nixopus-api/internal/testutils"
// )

// func TestLogin(t *testing.T) {
// 	setup := testutils.NewTestSetup()
// 	setup.CreateTestUserAndOrg()

// 	testCases := []struct {
// 		name           string
// 		request        types.LoginRequest
// 		expectedStatus int
// 	}{
// 		{
// 			name: "Successfully login a user",
// 			request: types.LoginRequest{
// 				Email:    "test@example.com",
// 				Password: "Password123@",
// 			},
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name: "Login with wrong email",
// 			request: types.LoginRequest{
// 				Email:    "invalid@example.com",
// 				Password: "Password123@",
// 			},
// 			expectedStatus: http.StatusUnauthorized,
// 		},
// 		{
// 			name: "Login with wrong password",
// 			request: types.LoginRequest{
// 				Email:    "test@example.com",
// 				Password: "InvalidPassword@123",
// 			},
// 			expectedStatus: http.StatusUnauthorized,
// 		},
// 		{
// 			name: "Login with invalid password",
// 			request: types.LoginRequest{
// 				Email:    "test@example.com",
// 				Password: "invalidpassword",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "Login with invalid email and password",
// 			request: types.LoginRequest{
// 				Email:    "invalid@example.com",
// 				Password: "invalidpassword",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "Login with no email",
// 			request: types.LoginRequest{
// 				Password: "Password123@",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "Login with no password",
// 			request: types.LoginRequest{
// 				Email: "test@example.com",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			Test(t,
// 				Description(tc.name),
// 				Post(tests.GetLoginURL()),
// 				Send().Body().JSON(tc.request),
// 				Expect().Status().Equal(int64(tc.expectedStatus)),
// 			)
// 		})
// 	}
// }
