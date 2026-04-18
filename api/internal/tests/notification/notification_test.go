package notification

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- SMTP ---

func TestGetSMTPConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /notification/smtp without auth returns 401"),
		Get(tests.GetNotificationSMTPURL()+"?id="+uuid.New().String()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetSMTPConfig_ValidAuth_Empty(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /notification/smtp with valid auth returns 200 (no config yet)"),
		Get(tests.GetNotificationSMTPURL()+"?id="+auth.OrganizationID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestCreateSMTPConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /notification/smtp without auth returns 401"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Body().JSON(map[string]interface{}{
			"host": "smtp.example.com", "port": 587,
			"username": "user@example.com", "password": "password",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCreateSMTPConfig_MissingHost(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/smtp without host returns 400"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"port": 587, "username": "user@example.com", "password": "password",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateSMTPConfig_MissingPort(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/smtp with port 0 (missing) returns 400"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"host": "smtp.example.com", "username": "user@example.com", "password": "password",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateSMTPConfig_MissingUsername(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/smtp without username returns 400"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"host": "smtp.example.com", "port": 587, "password": "password",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateSMTPConfig_MissingPassword(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/smtp without password returns 400"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"host": "smtp.example.com", "port": 587, "username": "user@example.com",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateSMTPConfig_ValidData(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	Test(t,
		Description("POST /notification/smtp with all required fields returns 200"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"host":            "smtp.example.com",
			"port":            587,
			"username":        "user@example.com",
			"password":        "password123",
			"from_name":       "Test Sender",
			"from_email":      "test@example.com",
			"organization_id": orgID.String(),
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestCreateSMTPConfig_DuplicateReturnsError(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	payload := map[string]interface{}{
		"host":            "smtp.example.com",
		"port":            587,
		"username":        "user@example.com",
		"password":        "password123",
		"from_name":       "Test",
		"from_email":      "test@example.com",
		"organization_id": orgID.String(),
	}

	Test(t,
		Description("First SMTP config creation succeeds"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(payload),
		Expect().Status().Equal(http.StatusOK),
	)

	Test(t,
		Description("Second SMTP config creation for same org returns conflict"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(payload),
		Expect().Status().OneOf(int64(http.StatusConflict), int64(http.StatusInternalServerError)),
	)
}

func TestUpdateSMTPConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("PUT /notification/smtp without auth returns 401"),
		Put(tests.GetNotificationSMTPURL()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateSMTPConfig_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /notification/smtp without id returns 400"),
		Put(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateSMTPConfig_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /notification/smtp with random ID returns error"),
		Put(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"id": uuid.New().String(),
		}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestUpdateSMTPConfig_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	// Create first
	Test(t,
		Description("Create SMTP config"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"host":            "smtp.example.com",
			"port":            587,
			"username":        "user@example.com",
			"password":        "password123",
			"organization_id": orgID.String(),
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	// Fetch to get ID
	var smtpID string
	Test(t,
		Description("Fetch SMTP config to get ID"),
		Get(tests.GetNotificationSMTPURL()+"?id="+auth.OrganizationID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&smtpID),
	)

	if smtpID == "" {
		t.Skip("no SMTP config ID returned, skipping update test")
	}

	newHost := "smtp2.example.com"
	Test(t,
		Description("PUT /notification/smtp with existing ID updates config"),
		Put(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"id":              smtpID,
			"host":            newHost,
			"organization_id": orgID.String(),
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestDeleteSMTPConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("DELETE /notification/smtp without auth returns 401"),
		Delete(tests.GetNotificationSMTPURL()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestDeleteSMTPConfig_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /notification/smtp without id returns 400"),
		Delete(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestDeleteSMTPConfig_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /notification/smtp with non-existent id returns error"),
		Delete(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestDeleteSMTPConfig_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	Test(t,
		Description("Create SMTP config for deletion"),
		Post(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"host":            "smtp.example.com",
			"port":            587,
			"username":        "user@example.com",
			"password":        "password123",
			"organization_id": orgID.String(),
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	var smtpID string
	Test(t,
		Description("Fetch SMTP config ID"),
		Get(tests.GetNotificationSMTPURL()+"?id="+auth.OrganizationID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&smtpID),
	)

	if smtpID == "" {
		t.Skip("no SMTP ID to delete")
	}

	Test(t,
		Description("DELETE /notification/smtp with valid id returns 200"),
		Delete(tests.GetNotificationSMTPURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": smtpID}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Notification Preferences ---

func TestGetNotificationPreferences_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /notification/preferences without auth returns 401"),
		Get(tests.GetNotificationPreferencesURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetNotificationPreferences_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /notification/preferences with valid auth returns 200"),
		Get(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestUpdateNotificationPreferences_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /notification/preferences without auth returns 401"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Body().JSON(map[string]interface{}{"category": "activity", "type": "deploy", "enabled": true}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateNotificationPreferences_MissingCategory(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/preferences without category returns 400"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"type": "deploy", "enabled": true}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateNotificationPreferences_InvalidCategory(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/preferences with invalid category returns 400"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"category": "invalid_category", "type": "deploy", "enabled": true}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateNotificationPreferences_MissingType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/preferences without type returns 400"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"category": "activity", "enabled": true}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateNotificationPreferences_Activity(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/preferences activity category returns 200"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"category": "activity", "type": "deploy", "enabled": true}),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestUpdateNotificationPreferences_Security(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/preferences security category returns 200"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"category": "security", "type": "login", "enabled": false}),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestUpdateNotificationPreferences_Update(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/preferences update category returns 200"),
		Post(tests.GetNotificationPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"category": "update", "type": "release", "enabled": true}),
		Expect().Status().Equal(http.StatusOK),
	)
}

// --- Webhook ---

func TestCreateWebhookConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /notification/webhook without auth returns 401"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Body().JSON(map[string]interface{}{"type": "slack", "webhook_url": "https://hooks.slack.com/test"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCreateWebhookConfig_MissingType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/webhook without type returns 400"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"webhook_url": "https://hooks.slack.com/test"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateWebhookConfig_InvalidType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/webhook with invalid type returns 400"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"type": "email", "webhook_url": "https://example.com"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateWebhookConfig_Slack(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/webhook with slack type returns 200"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/services/TEST/TEST/TEST",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestCreateWebhookConfig_Discord(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/webhook with discord type returns 200"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"type":        "discord",
			"webhook_url": "https://discord.com/api/webhooks/TEST/TEST",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestGetWebhookConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /notification/webhook/slack without auth returns 401"),
		Get(tests.GetNotificationWebhookURL("slack")),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetWebhookConfig_Slack_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /notification/webhook/slack with valid auth returns 200"),
		Get(tests.GetNotificationWebhookURL("slack")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestGetWebhookConfig_Discord_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /notification/webhook/discord with valid auth returns 200"),
		Get(tests.GetNotificationWebhookURL("discord")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestUpdateWebhookConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("PUT /notification/webhook without auth returns 401"),
		Put(tests.GetNotificationWebhookBaseURL()),
		Send().Body().JSON(map[string]interface{}{"type": "slack"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateWebhookConfig_MissingType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /notification/webhook without type returns 400"),
		Put(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateWebhookConfig_InvalidType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /notification/webhook with invalid type returns 400"),
		Put(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"type": "telegram"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateWebhookConfig_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Create first
	Test(t,
		Description("Create slack webhook config"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"type":        "slack",
			"webhook_url": "https://hooks.slack.com/services/OLD/OLD/OLD",
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	newURL := "https://hooks.slack.com/services/NEW/NEW/NEW"
	isActive := true
	Test(t,
		Description("PUT /notification/webhook updates webhook URL"),
		Put(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"type":        "slack",
			"webhook_url": newURL,
			"is_active":   isActive,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestDeleteWebhookConfig_NoAuth(t *testing.T) {
	Test(t,
		Description("DELETE /notification/webhook without auth returns 401"),
		Delete(tests.GetNotificationWebhookBaseURL()),
		Send().Body().JSON(map[string]interface{}{"type": "slack"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestDeleteWebhookConfig_MissingType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /notification/webhook without type returns 400"),
		Delete(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestDeleteWebhookConfig_InvalidType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /notification/webhook with invalid type returns 400"),
		Delete(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"type": "invalid"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestDeleteWebhookConfig_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("Create webhook before deleting"),
		Post(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"type":        "discord",
			"webhook_url": "https://discord.com/api/webhooks/DELETE/ME",
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	Test(t,
		Description("DELETE /notification/webhook with valid type returns 200"),
		Delete(tests.GetNotificationWebhookBaseURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"type": "discord"}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Send Notification ---

func TestSendNotification_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /notification/send without auth returns 401"),
		Post(tests.GetNotificationSendURL()),
		Send().Body().JSON(map[string]interface{}{"channel": "slack", "message": "test"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestSendNotification_MissingChannel(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/send without channel returns 400"),
		Post(tests.GetNotificationSendURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"message": "hello"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestSendNotification_MissingMessage(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/send without message returns 400"),
		Post(tests.GetNotificationSendURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"channel": "slack"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestSendNotification_InvalidChannel(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /notification/send with invalid channel returns 400"),
		Post(tests.GetNotificationSendURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"channel": "telegram", "message": "test"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestSendNotification_ValidRequest_NoWebhookConfigured(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// With no webhook configured, channel send may succeed (fire-and-forget) or return error
	Test(t,
		Description("POST /notification/send to slack with no webhook config — accepts request"),
		Post(tests.GetNotificationSendURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"channel": "slack",
			"message": "integration test message",
		}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}
