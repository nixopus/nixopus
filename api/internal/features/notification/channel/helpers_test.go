package channel

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

// ---------------------------------------------------------------------------
// In-memory SQLite DB helpers
// ---------------------------------------------------------------------------

// newChannelTestDB opens a fresh in-memory SQLite database and creates all
// tables used by the channel adapters (smtp_configs, webhook_configs).
func newChannelTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", fmt.Sprintf("file:channeltest_%s?mode=memory&cache=shared", uuid.New().String()))
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	_, err = db.NewCreateTable().
		Model((*shared_types.SMTPConfigs)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewCreateTable().
		Model((*shared_types.WebhookConfig)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err)

	return db
}

// insertActiveSMTP inserts an active SMTP config for the given org and returns it.
func insertActiveSMTP(t *testing.T, db *bun.DB, host string, port int, orgID uuid.UUID) *shared_types.SMTPConfigs {
	t.Helper()
	cfg := &shared_types.SMTPConfigs{
		ID:             uuid.New(),
		Host:           host,
		Port:           port,
		Username:       "testuser",
		Password:       "testpass",
		FromEmail:      "from@example.com",
		FromName:       "Test",
		Security:       "none",
		IsActive:       true,
		UserID:         uuid.New(),
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := db.NewInsert().Model(cfg).Exec(context.Background())
	require.NoError(t, err)
	return cfg
}

// insertActiveWebhook inserts an active webhook config for the given org and type.
func insertActiveWebhook(t *testing.T, db *bun.DB, webhookType, webhookURL string, orgID uuid.UUID) {
	t.Helper()
	cfg := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           webhookType,
		WebhookURL:     webhookURL,
		ChannelID:      webhookType + ":" + orgID.String(),
		IsActive:       true,
		UserID:         uuid.New(),
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := db.NewInsert().Model(cfg).Exec(context.Background())
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Minimal fake SMTP server
// ---------------------------------------------------------------------------

// smtpResult collects the outcome of a single fake SMTP transaction.
type smtpResult struct {
	Received bool
	Body     string
}

// startFakeSMTPServer starts a minimal SMTP listener on a random localhost port
// that accepts one transaction then closes. It advertises AUTH PLAIN so that
// smtp.PlainAuth works without TLS (localhost is exempt from the TLS requirement).
// Returns the address ("host:port") and a channel that receives the body text.
func startFakeSMTPServer(t *testing.T) (addr string, results chan smtpResult) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	results = make(chan smtpResult, 1)

	go func() {
		defer l.Close()
		conn, err := l.Accept()
		if err != nil {
			results <- smtpResult{}
			return
		}
		defer conn.Close()

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		write := func(s string) {
			_, _ = rw.WriteString(s)
			_ = rw.Flush()
		}

		write("220 smtp.test ESMTP\r\n")

		var msgBuf strings.Builder
		inData := false

		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\r\n")

			if inData {
				if line == "." {
					inData = false
					write("250 OK: message queued\r\n")
					results <- smtpResult{Received: true, Body: msgBuf.String()}
					continue
				}
				// un-dot-stuff
				if strings.HasPrefix(line, ".") {
					line = line[1:]
				}
				msgBuf.WriteString(line + "\n")
				continue
			}

			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				write("250-smtp.test Hello\r\n250-AUTH PLAIN LOGIN\r\n250 SIZE 10240000\r\n")
			case strings.HasPrefix(upper, "AUTH"):
				write("235 Authentication successful\r\n")
			case strings.HasPrefix(upper, "MAIL FROM"):
				write("250 OK\r\n")
			case strings.HasPrefix(upper, "RCPT TO"):
				write("250 OK\r\n")
			case upper == "DATA":
				write("354 Start mail input; end with <CRLF>.<CRLF>\r\n")
				inData = true
			case upper == "QUIT":
				write("221 Bye\r\n")
				return
			default:
				write("500 Unknown command\r\n")
			}
		}
	}()

	t.Cleanup(func() { l.Close() })
	return l.Addr().String(), results
}

// startBadSMTPServer starts a server that accepts a connection but immediately
// sends garbage, causing smtp.SendMail to fail during the protocol handshake.
func startBadSMTPServer(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() {
		defer l.Close()
		conn, err := l.Accept()
		if err != nil {
			return
		}
		_, _ = conn.Write([]byte("garbage not smtp\r\n"))
		conn.Close()
	}()
	t.Cleanup(func() { l.Close() })
	return l.Addr().String()
}
