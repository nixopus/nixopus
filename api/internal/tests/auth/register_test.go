package auth

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestSuccessfullyRegister(t *testing.T) {
	_ = testutils.NewTestSetup()

	Test(t,
		Description("Register a new user"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:        "test@tes.com",
			Password:     "Password123@",
			Username:     "NixopusUser",
			Type:         shared_types.UserTypeAdmin,
			Organization: "",
		}),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestAdminAlreadyRegistered(t *testing.T) {
	setup := testutils.NewTestSetup()
	setup.CreateTestUserAndOrg()

	Test(t,
		Description("Register a new user"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:        "test@example.com",
			Password:     "Password123@",
			Username:     "NixopusUser",
			Type:         shared_types.UserTypeAdmin,
			Organization: "",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRegisterWithOrganization(t *testing.T) {
	_ = testutils.NewTestSetup()
	Test(t,
		Description("Register a new user with organization"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:        "test@example.com",
			Password:     "Password123@",
			Username:     "NixopusUser",
			Type:         shared_types.UserTypeAdmin,
			Organization: "123e4567-e89b-12d3-a456-426614174000",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRegisterWithInvalidType(t *testing.T) {
	_ = testutils.NewTestSetup()
	Test(t,
		Description("Register a new user with invalid type"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:    "test@example.com",
			Password: "Password123@",
			Username: "NixopusUser",
			Type:     "member",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRegisterWithInvalidEmail(t *testing.T) {
	_ = testutils.NewTestSetup()
	Test(t,
		Description("Register a new user with invalid email"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:    "testexample.com",
			Password: "Password123@",
			Username: "NixopusUser",
			Type:     shared_types.UserTypeAdmin,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRegisterWithInvalidPassword(t *testing.T) {
	_ = testutils.NewTestSetup()
	Test(t,
		Description("Register a new user with invalid password"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:        "test@example.com",
			Password:     "password",
			Username:     "NixopusUser",
			Type:         shared_types.UserTypeAdmin,
			Organization: "",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRegisterWithInvalidUsername(t *testing.T) {
	_ = testutils.NewTestSetup()
	Test(t,
		Description("Register a new user with invalid username"),
		Post(tests.GetRegisterURL()),
		Send().Body().JSON(types.RegisterRequest{
			Email:        "test@example.com",
			Password:     "Password123@",
			Username:     "",
			Type:         shared_types.UserTypeAdmin,
			Organization: "",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}
