package telemetry

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
)

// validPayload returns a minimal valid TrackInstallRequest body.
func validPayload() map[string]interface{} {
	return map[string]interface{}{
		"event_type": "install_success",
		"os":         "ubuntu",
		"arch":       "amd64",
		"version":    "1.0.0",
		"duration":   30,
	}
}

func TestTrackInstall_success(t *testing.T) {
	Test(t,
		Description("valid install event returns 201 with success response"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(validPayload()),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".message").Equal("event recorded"),
	)
}

func TestTrackInstall_allEventTypes(t *testing.T) {
	for _, eventType := range []string{"install_started", "install_success", "install_failure"} {
		eventType := eventType
		t.Run(eventType, func(t *testing.T) {
			payload := validPayload()
			payload["event_type"] = eventType
			Test(t,
				Description("event_type "+eventType+" is accepted"),
				Post(tests.GetTelemetryURL()),
				Send().Headers("Content-Type").Add("application/json"),
				Send().Body().JSON(payload),
				Expect().Status().Equal(http.StatusCreated),
			)
		})
	}
}

func TestTrackInstall_withError(t *testing.T) {
	payload := validPayload()
	payload["event_type"] = "install_failure"
	payload["error"] = "failed to download binary: connection refused"
	Test(t,
		Description("install_failure with non-empty error field is accepted"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusCreated),
	)
}

func TestTrackInstall_withPreReleaseVersion(t *testing.T) {
	payload := validPayload()
	payload["version"] = "2.0.0-beta.1"
	Test(t,
		Description("pre-release semver version is accepted"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusCreated),
	)
}

func TestTrackInstall_invalidEventType(t *testing.T) {
	payload := validPayload()
	payload["event_type"] = "not_valid"
	Test(t,
		Description("unknown event_type returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_invalidOS(t *testing.T) {
	payload := validPayload()
	payload["os"] = "windows"
	Test(t,
		Description("unsupported OS returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_invalidArch(t *testing.T) {
	payload := validPayload()
	payload["arch"] = "x86"
	Test(t,
		Description("unsupported arch returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_invalidVersion(t *testing.T) {
	for _, ver := range []string{"notaversion", "v1.0.0", "1.0", "1.0.0.0"} {
		ver := ver
		t.Run(ver, func(t *testing.T) {
			payload := validPayload()
			payload["version"] = ver
			Test(t,
				Description("invalid version format '"+ver+"' returns 400"),
				Post(tests.GetTelemetryURL()),
				Send().Headers("Content-Type").Add("application/json"),
				Send().Body().JSON(payload),
				Expect().Status().Equal(http.StatusBadRequest),
			)
		})
	}
}

func TestTrackInstall_durationNegative(t *testing.T) {
	payload := validPayload()
	payload["duration"] = -1
	Test(t,
		Description("negative duration returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_durationTooLarge(t *testing.T) {
	payload := validPayload()
	payload["duration"] = 7201
	Test(t,
		Description("duration above 7200 returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_errorTooLong(t *testing.T) {
	payload := validPayload()
	payload["error"] = strings.Repeat("x", 201)
	Test(t,
		Description("error message exceeding 200 chars returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_invalidJSON(t *testing.T) {
	Test(t,
		Description("malformed JSON body returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String("{invalid-json}"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_emptyBody(t *testing.T) {
	Test(t,
		Description("empty body returns 400"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrackInstall_getMethodNotAllowed(t *testing.T) {
	Test(t,
		Description("GET on telemetry endpoint returns 405"),
		Get(tests.GetTelemetryURL()),
		Expect().Status().Equal(http.StatusMethodNotAllowed),
	)
}

func TestTrackInstall_deleteMethodNotAllowed(t *testing.T) {
	Test(t,
		Description("DELETE on telemetry endpoint returns 405"),
		Delete(tests.GetTelemetryURL()),
		Expect().Status().Equal(http.StatusMethodNotAllowed),
	)
}

func TestTrackInstall_responseShape(t *testing.T) {
	Test(t,
		Description("success response contains both status and message fields"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(validPayload()),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".status").NotEqual(nil),
		Expect().Body().JSON().JQ(".message").NotEqual(nil),
	)
}

func TestTrackInstall_noAuthRequired(t *testing.T) {
	Test(t,
		Description("endpoint is public — no cookies or auth headers needed"),
		Post(tests.GetTelemetryURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(validPayload()),
		Expect().Status().Equal(http.StatusCreated),
	)
}
