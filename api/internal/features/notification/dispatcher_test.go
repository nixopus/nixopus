package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification/channel"
	"github.com/nixopus/nixopus/api/internal/features/notification/tasks"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

type stubChannel struct {
	typeName string
	sendErr  error
	lastMsg  channel.Message
}

func (s *stubChannel) Type() string { return s.typeName }

func (s *stubChannel) Send(ctx context.Context, msg channel.Message) error {
	_ = ctx
	s.lastMsg = msg
	return s.sendErr
}

func dispatcherTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared&_busy_timeout=5000")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	for _, m := range []any{
		(*shared_types.NotificationPreferences)(nil),
		(*shared_types.PreferenceItem)(nil),
	} {
		_, err = db.NewCreateTable().Model(m).IfNotExists().Exec(ctx)
		require.NoError(t, err)
	}
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS "user" (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL,
	email_verified INTEGER NOT NULL DEFAULT 0,
	is_onboarded INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS organization_settings (
	id TEXT PRIMARY KEY,
	organization_id TEXT NOT NULL UNIQUE,
	settings TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS notification_preferences_one_per_user
		ON notification_preferences (user_id)
	`)
	require.NoError(t, err)
	return db
}

func insertUser(t *testing.T, db *bun.DB, id uuid.UUID, email string) {
	t.Helper()
	now := time.Now()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO "user" (id, name, email, email_verified, is_onboarded, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id.String(), "Test", email, 1, 0, now, now)
	require.NoError(t, err)
}

func insertOrgSettingsAI(t *testing.T, db *bun.DB, orgID uuid.UUID, aiEnabled bool) {
	t.Helper()
	data := shared_types.DefaultOrganizationSettingsData()
	data.AIIncidentsEnabled = &aiEnabled
	raw, err := json.Marshal(data)
	require.NoError(t, err)
	now := time.Now()
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO organization_settings (id, organization_id, settings, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), orgID.String(), string(raw), now, now)
	require.NoError(t, err)
}

func insertPrefEnabled(t *testing.T, db *bun.DB, userID uuid.UUID, category, typ string, enabled bool) {
	t.Helper()
	ctx := context.Background()
	prefID := uuid.New()
	now := time.Now()
	_, err := db.NewInsert().Model(&shared_types.NotificationPreferences{
		ID:        prefID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&shared_types.PreferenceItem{
		ID:           uuid.New(),
		PreferenceID: prefID,
		Category:     category,
		Type:         typ,
		Enabled:      enabled,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Exec(ctx)
	require.NoError(t, err)
}

func fullChannelMap() map[string]channel.Channel {
	return map[string]channel.Channel{
		"email":        &stubChannel{typeName: "email"},
		"slack":        &stubChannel{typeName: "slack"},
		"discord":      &stubChannel{typeName: "discord"},
		"agent":        &stubChannel{typeName: "agent"},
		"system_email": &stubChannel{typeName: "system_email"},
	}
}

func testQueue(t *testing.T) {
	t.Helper()
	mr := miniredis.RunT(t)
	opt, err := redis.ParseURL("redis://" + mr.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	queue.Init(rdb)
	t.Cleanup(func() { _ = queue.Close() })
}

func TestNewDispatcher(t *testing.T) {
	db := dispatcherTestDB(t)
	d := NewDispatcher(db, context.Background(), logger.NewLogger(), map[string]channel.Channel{})
	require.NotNil(t, d)
}

func TestDispatcher_buildMessage(t *testing.T) {
	ctx := context.Background()
	db := dispatcherTestDB(t)
	d := &Dispatcher{db: db, ctx: ctx, logger: logger.NewLogger()}
	uid := uuid.New().String()
	org := uuid.New().String()

	t.Run("email_with_template", func(t *testing.T) {
		ev := shared_types.NotificationEvent{
			Type:           shared_types.EventLoginAlert,
			UserID:         uid,
			OrganizationID: org,
			Data:           map[string]interface{}{"ip": "1.2.3.4"},
		}
		msg := d.buildMessage(ev, "email", "u@x.com")
		assert.Equal(t, "u@x.com", msg.To)
		assert.NotEmpty(t, msg.Subject)
		assert.NotEmpty(t, msg.TemplateName)
	})

	t.Run("email_without_template", func(t *testing.T) {
		ev := shared_types.NotificationEvent{
			Type:           shared_types.EventDeploySuccess,
			UserID:         uid,
			OrganizationID: org,
		}
		msg := d.buildMessage(ev, "email", "u@x.com")
		assert.Contains(t, msg.Body, "deploy.success")
	})

	t.Run("system_email_trial", func(t *testing.T) {
		ev := shared_types.NotificationEvent{
			Type:           shared_types.EventTrialExpired,
			UserID:         uid,
			OrganizationID: org,
			Data:           map[string]interface{}{"k": "v"},
		}
		msg := d.buildMessage(ev, "system_email", "u@x.com")
		assert.Equal(t, "trial-expired", msg.Metadata["resend_template_id"])
	})

	t.Run("slack_plain", func(t *testing.T) {
		ev := shared_types.NotificationEvent{
			Type:           shared_types.EventUserAddedToOrg,
			UserID:         uid,
			OrganizationID: org,
			Data: map[string]interface{}{
				"UserName": "A", "UserEmail": "a@b.com", "OrganizationName": "Org",
			},
		}
		msg := d.buildMessage(ev, "slack", "")
		assert.Contains(t, msg.Body, "New user")
	})

	t.Run("agent_metadata_strings", func(t *testing.T) {
		ev := shared_types.NotificationEvent{
			Type:           shared_types.EventDeployFailed,
			UserID:         uid,
			OrganizationID: org,
			Data: map[string]interface{}{
				"trace_id": "abc",
				"skip":     42,
				"empty":    "",
			},
		}
		msg := d.buildMessage(ev, "agent", "")
		assert.Equal(t, "abc", msg.Metadata["trace_id"])
		_, hasSkip := msg.Metadata["skip"]
		assert.False(t, hasSkip)
	})
}

func TestDispatcher_buildPlainTextBody(t *testing.T) {
	d := &Dispatcher{ctx: context.Background(), logger: logger.NewLogger()}
	data := func(kv ...string) map[string]interface{} {
		m := map[string]interface{}{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return m
	}
	t.Run("user_added", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventUserAddedToOrg,
			Data: data("UserName", "N", "UserEmail", "e@e.com", "OrganizationName", "O"),
		})
		assert.Contains(t, body, "N")
	})
	t.Run("user_removed", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventUserRemovedFromOrg,
			Data: data("UserName", "N", "UserEmail", "e@e.com", "OrganizationName", "O"),
		})
		assert.Contains(t, body, "removed")
	})
	t.Run("deploy_success", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventDeploySuccess,
			Data: data("app_name", "app1"),
		})
		assert.Contains(t, body, "app1")
	})
	t.Run("deploy_failed", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventDeployFailed,
			Data: data("app_name", "app1"),
		})
		assert.Contains(t, body, "failed")
	})
	t.Run("build_failed", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventBuildFailed,
			Data: data("app_name", "app1", "error_message", "boom"),
		})
		assert.Contains(t, body, "boom")
	})
	t.Run("health_critical", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventHealthCheckCritical,
			Data: data("app_id", "a", "endpoint", "/h", "consecutive_fails", "3"),
		})
		assert.Contains(t, body, "critical")
	})
	t.Run("default", func(t *testing.T) {
		body := d.buildPlainTextBody(shared_types.NotificationEvent{
			Type: shared_types.EventContainerCrashed,
		})
		assert.Contains(t, body, "container.crashed")
	})
}

func TestGetDataStr(t *testing.T) {
	assert.Equal(t, "", getDataStr(nil, "k"))
	assert.Equal(t, "", getDataStr(map[string]interface{}{"k": 1}, "k"))
	assert.Equal(t, "x", getDataStr(map[string]interface{}{"k": "x"}, "k"))
}

func TestDispatcher_resolveUserEmail(t *testing.T) {
	db := dispatcherTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	insertUser(t, db, uid, "want@example.com")
	d := NewDispatcher(db, ctx, logger.NewLogger(), nil)

	email, err := d.resolveUserEmail(uid.String())
	require.NoError(t, err)
	assert.Equal(t, "want@example.com", email)

	_, err = d.resolveUserEmail(uuid.New().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestDispatcher_isAgentEnabledForOrg(t *testing.T) {
	db := dispatcherTestDB(t)
	ctx := context.Background()
	orgID := uuid.New()
	insertOrgSettingsAI(t, db, orgID, true)
	orgDisabled := uuid.New()
	insertOrgSettingsAI(t, db, orgDisabled, false)
	d := NewDispatcher(db, ctx, logger.NewLogger(), nil)

	assert.False(t, d.isAgentEnabledForOrg("not-a-uuid"))
	assert.False(t, d.isAgentEnabledForOrg(uuid.New().String()))
	assert.True(t, d.isAgentEnabledForOrg(orgID.String()))
	assert.False(t, d.isAgentEnabledForOrg(orgDisabled.String()))
}

func TestCreateSMTPConfigRequest_String(t *testing.T) {
	orgID := uuid.New()
	r := CreateSMTPConfigRequest{
		Host:           "smtp.example.com",
		Port:           587,
		Username:       "user@example.com",
		FromName:       "Nixopus",
		FromEmail:      "noreply@example.com",
		OrganizationID: orgID,
	}
	s := r.String()
	assert.Contains(t, s, "smtp.example.com")
	assert.Contains(t, s, "587")
	assert.Contains(t, s, "user@example.com")
	assert.Contains(t, s, orgID.String())
}

func TestNewSMTPConfig(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	t.Run("with_from_email_and_from_name", func(t *testing.T) {
		req := &CreateSMTPConfigRequest{
			Host:           "smtp.example.com",
			Port:           587,
			Username:       "user@example.com",
			Password:       "secret",
			FromName:       "Nixopus",
			FromEmail:      "noreply@example.com",
			OrganizationID: orgID,
		}
		cfg := NewSMTPConfig(req, userID)
		assert.Equal(t, "noreply@example.com", cfg.FromEmail)
		assert.Equal(t, "Nixopus", cfg.FromName)
		assert.Equal(t, userID, cfg.UserID)
		assert.Equal(t, orgID, cfg.OrganizationID)
	})

	t.Run("defaults_from_email_to_username", func(t *testing.T) {
		req := &CreateSMTPConfigRequest{
			Host:           "smtp.example.com",
			Port:           587,
			Username:       "user@example.com",
			Password:       "secret",
			OrganizationID: orgID,
		}
		cfg := NewSMTPConfig(req, userID)
		assert.Equal(t, "user@example.com", cfg.FromEmail)
		assert.Equal(t, "user", cfg.FromName)
	})
}

func TestCategory_OpenAPI(t *testing.T) {
	var c Category = ActivityCategory

	schemaTypes := c.OpenAPISchemaType()
	assert.Equal(t, []string{"string"}, schemaTypes)

	enums := c.OpenAPISchemaEnum()
	assert.Contains(t, enums, string(ActivityCategory))
	assert.Contains(t, enums, string(SecurityCategory))
	assert.Contains(t, enums, string(UpdateCategory))

	desc := c.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "activity")
}

func TestDispatcher_SendDirect(t *testing.T) {
	db := dispatcherTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	insertUser(t, db, uid, "direct@example.com")
	emailCh := &stubChannel{typeName: "email"}
	d := NewDispatcher(db, ctx, logger.NewLogger(), map[string]channel.Channel{"email": emailCh})

	t.Run("unsupported_channel", func(t *testing.T) {
		resp := d.SendDirect(SendNotificationRequest{Channel: "sms", Message: "m"}, uid.String(), "org")
		require.False(t, resp.Success)
		assert.Contains(t, resp.Error, "unsupported")
	})

	t.Run("email_resolve_recipient_error", func(t *testing.T) {
		resp := d.SendDirect(SendNotificationRequest{Channel: "email", Message: "m"}, uuid.New().String(), "org")
		require.False(t, resp.Success)
		assert.Contains(t, resp.Error, "failed to resolve recipient")
	})

	t.Run("email_uses_db_when_to_empty", func(t *testing.T) {
		resp := d.SendDirect(SendNotificationRequest{Channel: "email", Message: "hello"}, uid.String(), "org-1")
		require.True(t, resp.Success)
		assert.Equal(t, "direct@example.com", emailCh.lastMsg.To)
	})

	t.Run("default_subject", func(t *testing.T) {
		_ = d.SendDirect(SendNotificationRequest{Channel: "email", To: "x@y.com", Message: "b"}, uid.String(), "org")
		assert.Equal(t, "Notification from Nixopus", emailCh.lastMsg.Subject)
	})

	t.Run("send_error", func(t *testing.T) {
		errCh := &stubChannel{typeName: "email", sendErr: errors.New("send failed")}
		dd := NewDispatcher(db, ctx, logger.NewLogger(), map[string]channel.Channel{"email": errCh})
		resp := dd.SendDirect(SendNotificationRequest{Channel: "email", To: "a@b.com", Message: "x"}, uid.String(), "org")
		require.False(t, resp.Success)
		assert.Contains(t, resp.Error, "send failed")
	})
}

func TestDispatcher_Emit(t *testing.T) {
	testQueue(t)
	db := dispatcherTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	insertUser(t, db, userID, "emit@example.com")
	insertPrefEnabled(t, db, userID, "activity", "team-updates", true)
	d := NewDispatcher(db, ctx, logger.NewLogger(), fullChannelMap())
	d.SetupQueue()
	require.NotNil(t, tasks.NotificationQueue)
	require.NotNil(t, tasks.TaskSendNotification)

	prevQ, prevT := tasks.NotificationQueue, tasks.TaskSendNotification
	t.Cleanup(func() { tasks.NotificationQueue, tasks.TaskSendNotification = prevQ, prevT })

	t.Run("skip_preference_password_reset", func(t *testing.T) {
		err := d.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventPasswordReset,
			UserID:         userID.String(),
			OrganizationID: orgID.String(),
			Channels:       []string{"email"},
		})
		require.NoError(t, err)
	})

	t.Run("preference_suppressed", func(t *testing.T) {
		u := uuid.New()
		insertUser(t, db, u, "suppress@example.com")
		insertPrefEnabled(t, db, u, "security", "login-alerts", false)
		err := d.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventLoginAlert,
			UserID:         u.String(),
			OrganizationID: orgID.String(),
			Channels:       []string{"email"},
		})
		require.NoError(t, err)
	})

	t.Run("preference_check_error_still_suppressed", func(t *testing.T) {
		bad := NewDispatcher(db, ctx, logger.NewLogger(), fullChannelMap())
		err := bad.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventLoginAlert,
			UserID:         "not-a-uuid",
			OrganizationID: orgID.String(),
			Channels:       []string{"email"},
		})
		require.NoError(t, err)
	})

	t.Run("unknown_event_uses_default_channels", func(t *testing.T) {
		err := d.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventContainerCrashed,
			UserID:         userID.String(),
			OrganizationID: orgID.String(),
		})
		require.NoError(t, err)
	})

	t.Run("unregistered_channel_skipped", func(t *testing.T) {
		err := d.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventPasswordReset,
			UserID:         userID.String(),
			OrganizationID: orgID.String(),
			Channels:       []string{"unknown_x"},
		})
		require.NoError(t, err)
	})

	t.Run("agent_skipped_when_disabled", func(t *testing.T) {
		err := d.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventDeployFailed,
			UserID:         userID.String(),
			OrganizationID: orgID.String(),
		})
		require.NoError(t, err)
	})

	t.Run("agent_enqueued_when_enabled", func(t *testing.T) {
		orgAI := uuid.New()
		insertOrgSettingsAI(t, db, orgAI, true)
		err := d.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventDeployFailed,
			UserID:         userID.String(),
			OrganizationID: orgAI.String(),
			Channels:       []string{"agent"},
		})
		require.NoError(t, err)
	})

	t.Run("enqueue_error", func(t *testing.T) {
		tasks.NotificationQueue, tasks.TaskSendNotification = nil, nil
		t.Cleanup(func() { tasks.NotificationQueue, tasks.TaskSendNotification = prevQ, prevT })
		dd := NewDispatcher(db, ctx, logger.NewLogger(), map[string]channel.Channel{
			"email": &stubChannel{typeName: "email"},
		})
		err := dd.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventPasswordReset,
			UserID:         userID.String(),
			OrganizationID: orgID.String(),
			Channels:       []string{"email"},
		})
		require.Error(t, err)
	})

	t.Run("resolve_email_failure_continues", func(t *testing.T) {
		emptyUserDB := dispatcherTestDB(t)
		dd := NewDispatcher(emptyUserDB, context.Background(), logger.NewLogger(), fullChannelMap())
		tasks.NotificationQueue, tasks.TaskSendNotification = prevQ, prevT
		dd.SetupQueue()
		err := dd.Emit(shared_types.NotificationEvent{
			Type:           shared_types.EventPasswordReset,
			UserID:         uuid.New().String(),
			OrganizationID: orgID.String(),
			Channels:       []string{"email"},
		})
		require.NoError(t, err)
	})
}
