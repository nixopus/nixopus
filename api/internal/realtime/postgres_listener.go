package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// StartListeningAndNotify starts listening for PostgreSQL notifications and notifies the server.
//
// Parameters:
//
//	pgListener - the PostgresListener instance to use
//	ctx - the context to use
//	server - the server to notify
func StartListeningAndNotify(pgListener *PostgresListener, ctx context.Context, server *SocketServer) error {
	notificationChan, err := pgListener.ListenToApplicationChanges(ctx)
	if err != nil {
		return fmt.Errorf("failed to listen for PostgreSQL notifications: %w", err)
	}

	// we will start a new goroutine to handle the notification we receive from the database changes here
	go server.handleNotifications(notificationChan)
	return nil
}

// handleNotifications processes incoming PostgreSQL notifications from the notification channel.
//
// This method listens on the provided channel for notifications related to application changes.
// Upon receiving a notification, it checks if the notification is from the "application_changes"
// channel. If it is, the method attempts to parse the JSON payload to extract the table, action,
// and ID details. If parsing is successful, it constructs a message containing the parsed data
// and broadcasts it to the appropriate topic using the BroadcastToTopic method.
//
// Parameters:
//
//	notificationChan - a channel that receives *PostgresNotification objects.
//
// Errors and any issues parsing the notification payload are logged, but do not stop the
// processing of further notifications.
func (s *SocketServer) handleNotifications(notificationChan <-chan *PostgresNotification) {
	for notification := range notificationChan {
		// fmt.Printf("Received notification on channel %s: %s\n",
		// 	notification.Channel, notification.Payload)

		if notification.Channel == "application_changes" {
			var parsedPayload struct {
				Table         string                 `json:"table"`
				Action        string                 `json:"action"`
				ApplicationID string                 `json:"application_id"`
				Data          map[string]interface{} `json:"data"`
			}

			if err := json.Unmarshal([]byte(notification.Payload), &parsedPayload); err != nil {
				log.Printf("Error parsing notification payload: %v", err)
				continue
			}

			resourceID := parsedPayload.ApplicationID

			messageData := map[string]interface{}{
				"table":          parsedPayload.Table,
				"action":         parsedPayload.Action,
				"application_id": parsedPayload.ApplicationID,
				"data":           parsedPayload.Data,
			}

			// we will broadcast the message to the topic here so all the clients who are subscribed to the topic will receive the message
			s.BroadcastToTopic(MonitorApplicationDeployment, resourceID, messageData)
		}
	}
}
