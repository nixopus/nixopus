package health

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
)

func TestHealthEndpoint(t *testing.T) {
	Test(t,
		Description("Health check endpoint should return 200 OK"),
		Get(tests.GetHealthURL()),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".message").Equal("Server is up and running"),
	)
}

func TestHealthEndpointWithInvalidMethod(t *testing.T) {
	Test(t,
		Description("Health check endpoint should return 405 Method Not Allowed for POST"),
		Post(tests.GetHealthURL()),
		Expect().Status().Equal(http.StatusMethodNotAllowed),
	)
}

func TestHealthEndpointWithInvalidPath(t *testing.T) {
	Test(t,
		Description("Health check endpoint should return 404 Not Found for invalid path"),
		Get(tests.GetHealthURL()+"/invalid"),
		Expect().Status().Equal(http.StatusNotFound),
	)
}

func TestHealthEndpointWithHeaders(t *testing.T) {
	Test(t,
		Description("Health check endpoint should handle custom headers"),
		Get(tests.GetHealthURL()),
		Send().Headers("X-Custom-Header").Add("test-value"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Headers("Content-Type").Equal("application/json"),
	)
}

func TestHealthEndpointWithQueryParams(t *testing.T) {
	Test(t,
		Description("Health check endpoint should handle query parameters"),
		Get(tests.GetHealthURL()+"?format=json"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".message").Equal("Server is up and running"),
	)
}
