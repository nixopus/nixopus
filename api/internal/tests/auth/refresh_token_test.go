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

// func TestRefreshToken(t *testing.T) {
// 	setup := testutils.NewTestSetup()
// 	user, _, err := setup.GetTestAuthResponse()
// 	if err != nil {
// 		t.Fatalf("failed to get test auth response: %v", err)
// 	}
// 	testCases := []struct {
// 		name           string
// 		request        types.RefreshTokenRequest
// 		expectedStatus int
// 	}{
// 		{
// 			name: "Successfully refresh a token",
// 			request: types.RefreshTokenRequest{
// 				RefreshToken: user.RefreshToken,
// 			},
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name: "Refresh token with invalid token format",
// 			request: types.RefreshTokenRequest{
// 				RefreshToken: "invalid",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "Refresh token with expired token",
// 			request: types.RefreshTokenRequest{
// 				RefreshToken: "6092f3b4-9b16-433f-9bd8-fd289347ac87",
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 		},
// 		{
// 			name: "Refresh token with empty refresh token",
// 			request: types.RefreshTokenRequest{
// 				RefreshToken: "",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			Test(t,
// 				Description(tc.name),
// 				Post(tests.GetRefreshTokenURL()),
// 				Send().Body().JSON(tc.request),
// 				Expect().Status().Equal(int64(tc.expectedStatus)),
// 			)
// 		})
// 	}
// }
