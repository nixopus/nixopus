package realtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type PostgresListener struct {
	config           types.Config
	notificationChan chan *PostgresNotification
}

type PostgresNotification struct {
	Channel string
	Payload string
}

func NewPostgresListener() *PostgresListener {
	return &PostgresListener{
		config: types.Config{
			DB_PORT:     os.Getenv("DB_PORT"),
			Port:        os.Getenv("PORT"),
			HostName:    os.Getenv("HOST_NAME"),
			Password:    os.Getenv("PASSWORD"),
			DBName:      os.Getenv("DB_NAME"),
			Username:    os.Getenv("USERNAME"),
			SSLMode:     os.Getenv("SSL_MODE"),
			MaxOpenConn: 10,
			Debug:       true,
			MaxIdleConn: 5,
		},
		notificationChan: make(chan *PostgresNotification, 100),
	}
}

func getConnString(c types.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.HostName,
		c.DB_PORT,
		c.Username,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}

// ListenToApplicationChanges listens to Postgres notifications for the
// "application_changes" channel. This channel is used to send notifications
// about changes to the applications table, such as when a new application is
// added or when an existing application is updated.
//
// The function returns a channel that receives notifications as *PostgresNotification
// objects. The function also starts a goroutine that listens for notifications
// on the given channel. If the context is canceled, the goroutine is stopped
// and the channel is closed.
//
// You should call this function once per instance of the PostgresListener struct.
func (c *PostgresListener) ListenToApplicationChanges(ctx context.Context) (<-chan *PostgresNotification, error) {
	conn, err := pgx.Connect(ctx, getConnString(c.config))
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(ctx, "LISTEN application_changes")
	if err != nil {
		return nil, err
	}

	go func() {
		defer conn.Close(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				notification, err := conn.WaitForNotification(ctx)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}
				c.notificationChan <- &PostgresNotification{
					Channel: notification.Channel,
					Payload: notification.Payload,
				}
			}
		}
	}()

	return c.notificationChan, nil
}
