package channel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constructor + Type
// ---------------------------------------------------------------------------

func TestNewEmailChannel_ReturnsNonNil(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewEmailChannel(db, context.Background())
	require.NotNil(t, ch)
}

func TestEmailChannel_Type(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewEmailChannel(db, context.Background())
	assert.Equal(t, "email", ch.Type())
}

// ---------------------------------------------------------------------------
// Send — early validation errors
// ---------------------------------------------------------------------------

func TestEmailChannel_Send_MissingOrgID(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewEmailChannel(db, context.Background())

	err := ch.Send(context.Background(), Message{
		To:   "user@example.com",
		Body: "hello",
		// no Metadata
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization_id required")
}

func TestEmailChannel_Send_GetSmtpError(t *testing.T) {
	// Use a DB with no smtp_configs table so the lookup fails.
	db := newChannelTestDB(t)
	// Drop the smtp_configs table to force an error.
	_, err := db.ExecContext(context.Background(), "DROP TABLE smtp_configs")
	require.NoError(t, err)

	ch := NewEmailChannel(db, context.Background())
	err = ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Body:     "hello",
		Metadata: map[string]string{"organization_id": uuid.New().String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve SMTP config")
}

func TestEmailChannel_Send_EmptyRecipient(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	// Use the bad-SMTP address as host; we won't reach dial because recipient is empty.
	insertActiveSMTP(t, db, "127.0.0.1", 9, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:       "",
		Body:     "hello",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recipient address is required")
}

// ---------------------------------------------------------------------------
// Send — template rendering branches
// ---------------------------------------------------------------------------

func TestEmailChannel_Send_TemplateName_ParseError(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	insertActiveSMTP(t, db, "127.0.0.1", 9, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:           "user@example.com",
		Body:         "hello",
		TemplateName: "nonexistent_template.html",
		Metadata:     map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template render failed")
}

// TestEmailChannel_Send_TemplateName_ExecuteError covers the tmpl.Execute error
// path by using a template that references an undefined sub-template.
func TestEmailChannel_Send_TemplateName_ExecuteError(t *testing.T) {
	db := newChannelTestDB(t)
	orgID := uuid.New()
	insertActiveSMTP(t, db, "127.0.0.1", 9, orgID)

	// Build a temp directory tree that mirrors what renderTemplate expects:
	//   <root>/internal/features/notification/templates/<name>
	root := t.TempDir()
	tmplDir := filepath.Join(root, "internal", "features", "notification", "templates")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))

	// Template that calls an undefined sub-template → Execute will error.
	tmplName := "bad_execute.html"
	content := `{{template "undefined_subtmpl" .}}`
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, tmplName), []byte(content), 0o644))

	// Change working directory so os.Getwd() returns root.
	t.Chdir(root)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:           "user@example.com",
		Body:         "hello",
		TemplateName: tmplName,
		Metadata:     map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template render failed")
}

// ---------------------------------------------------------------------------
// Send — content-type branches
// ---------------------------------------------------------------------------

// HTMLBody set + no TemplateName → html content-type, body = HTMLBody.
func TestEmailChannel_Send_HTMLBody_NoTemplateName(t *testing.T) {
	smtpAddr, results := startFakeSMTPServer(t)

	db := newChannelTestDB(t)
	orgID := uuid.New()
	parts := strings.SplitN(smtpAddr, ":", 2)
	port, _ := strconv.Atoi(parts[1])
	insertActiveSMTP(t, db, parts[0], port, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		HTMLBody: "<b>hello</b>",
		Subject:  "Test",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
	select {
	case res := <-results:
		assert.True(t, res.Received)
		assert.Contains(t, res.Body, "text/html")
	default:
		t.Fatal("fake SMTP server did not receive a message")
	}
}

// TemplateName set → rendered html body (uses a real temp template file).
func TestEmailChannel_Send_TemplateName_Success(t *testing.T) {
	root := t.TempDir()
	tmplDir := filepath.Join(root, "internal", "features", "notification", "templates")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))
	tmplName := "welcome.html"
	require.NoError(t, os.WriteFile(
		filepath.Join(tmplDir, tmplName),
		[]byte(`<h1>Hello {{.Name}}</h1>`),
		0o644,
	))
	t.Chdir(root)

	smtpAddr, results := startFakeSMTPServer(t)
	db := newChannelTestDB(t)
	orgID := uuid.New()
	parts := strings.SplitN(smtpAddr, ":", 2)
	port, _ := strconv.Atoi(parts[1])
	insertActiveSMTP(t, db, parts[0], port, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:           "user@example.com",
		TemplateName: tmplName,
		TemplateData: map[string]interface{}{"Name": "Alice"},
		Subject:      "Welcome",
		Metadata:     map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
	select {
	case res := <-results:
		assert.True(t, res.Received)
		assert.Contains(t, res.Body, "text/html")
	default:
		t.Fatal("fake SMTP server did not receive a message")
	}
}

// ---------------------------------------------------------------------------
// Send — subject default branch
// ---------------------------------------------------------------------------

func TestEmailChannel_Send_DefaultSubject(t *testing.T) {
	smtpAddr, results := startFakeSMTPServer(t)

	db := newChannelTestDB(t)
	orgID := uuid.New()
	parts := strings.SplitN(smtpAddr, ":", 2)
	port, _ := strconv.Atoi(parts[1])
	insertActiveSMTP(t, db, parts[0], port, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Body:     "plain body",
		Subject:  "", // triggers default
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
	select {
	case res := <-results:
		assert.True(t, res.Received)
		assert.Contains(t, res.Body, "Notification from Nixopus")
	default:
		t.Fatal("fake SMTP server did not receive a message")
	}
}

// ---------------------------------------------------------------------------
// Send — SMTP dial failure
// ---------------------------------------------------------------------------

func TestEmailChannel_Send_SMTPDialFail(t *testing.T) {
	badAddr := startBadSMTPServer(t)

	db := newChannelTestDB(t)
	orgID := uuid.New()
	parts := strings.SplitN(badAddr, ":", 2)
	port, _ := strconv.Atoi(parts[1])
	insertActiveSMTP(t, db, parts[0], port, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:       "user@example.com",
		Body:     "hello",
		Subject:  "Test",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "smtp send failed")
}

// ---------------------------------------------------------------------------
// Send — plain text success (no template, no HTMLBody, subject provided)
// ---------------------------------------------------------------------------

func TestEmailChannel_Send_PlainText_Success(t *testing.T) {
	smtpAddr, results := startFakeSMTPServer(t)

	db := newChannelTestDB(t)
	orgID := uuid.New()
	parts := strings.SplitN(smtpAddr, ":", 2)
	port, _ := strconv.Atoi(parts[1])
	insertActiveSMTP(t, db, parts[0], port, orgID)

	ch := NewEmailChannel(db, context.Background())
	err := ch.Send(context.Background(), Message{
		To:       "recipient@example.com",
		Body:     "plain text body",
		Subject:  "Hello",
		Metadata: map[string]string{"organization_id": orgID.String()},
	})
	require.NoError(t, err)
	select {
	case res := <-results:
		assert.True(t, res.Received)
		assert.Contains(t, res.Body, "text/plain")
	default:
		t.Fatal("fake SMTP server did not receive a message")
	}
}

// ---------------------------------------------------------------------------
// getSmtpByOrg — error path (no row in smtp_configs)
// ---------------------------------------------------------------------------

func TestEmailChannel_getSmtpByOrg_NoRow(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewEmailChannel(db, context.Background())

	cfg, err := ch.getSmtpByOrg("no-such-org")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active SMTP config for org no-such-org")
	assert.Nil(t, cfg)
}

// ---------------------------------------------------------------------------
// renderTemplate — parse error (non-existent template file)
// ---------------------------------------------------------------------------

func TestEmailChannel_renderTemplate_ParseError(t *testing.T) {
	db := newChannelTestDB(t)
	ch := NewEmailChannel(db, context.Background())

	_, err := ch.renderTemplate("definitely_does_not_exist.html", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template parse error")
}

// ---------------------------------------------------------------------------
// renderTemplate — success
// ---------------------------------------------------------------------------

func TestEmailChannel_renderTemplate_Success(t *testing.T) {
	root := t.TempDir()
	tmplDir := filepath.Join(root, "internal", "features", "notification", "templates")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmplDir, "greet.html"),
		[]byte(`Hello {{.Name}}!`),
		0o644,
	))
	t.Chdir(root)

	db := newChannelTestDB(t)
	ch := NewEmailChannel(db, context.Background())

	out, err := ch.renderTemplate("greet.html", map[string]interface{}{"Name": "Bob"})
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("Hello %s!", "Bob"), out)
}
