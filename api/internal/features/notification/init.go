package notification

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/uptrace/bun"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func NewNotificationPayload(payloadType NotificationPayloadType, userID string, data interface{}, category NotificationCategory) NotificationPayload {
	return NotificationPayload{
		Type:      payloadType,
		UserID:    userID,
		Timestamp: time.Now(),
		Data:      data,
		Category:  category,
	}
}

func NewNotificationManager(channels *NotificationChannels, db *bun.DB) *NotificationManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationManager{
		Channels:    channels,
		PayloadChan: make(chan NotificationPayload, 100),
		ctx:         ctx,
		cancel:      cancel,
		db:          db,
	}
}

// Start starts the notification manager to listen for notifications from a go routine
// based on the type of the notification that we receive we can check the notification preferences of the user
// and then send the notification to the respective channel
func (m *NotificationManager) Start() {
	go func() {
		for {
			select {
			case payload := <-m.PayloadChan:
				switch payload.Category {
				case NotificationCategoryAuthentication:
					fmt.Printf("Authentication Notification - %+v", payload)
					m.SendEmail(payload.UserID,"login successfully")
				case NotificationCategoryOrganization:
					fmt.Printf("Organization Notification - %+v", payload)
				}
			case <-m.ctx.Done():
				return
			}
		}
	}()
}

func (m *NotificationManager) Stop() {
	m.cancel()
}

// SendNotification sends a notification
func (m *NotificationManager) SendNotification(payload NotificationPayload) {
	m.PayloadChan <- payload
}

// here we can get the notification preferences of the user (like should send to slack/email/discord, how many times to send, what type of contents to send)
func (m *NotificationManager) CheckUserNotificationPreferences(userID string) {

}

// we will categorize the notifications based on the type of the notification
func (m *NotificationManager) GetPreferencesBasedOnCategory() {

}

func(m *NotificationManager) SendEmail(userId string,body string) {
	smtpConfig,err := m.GetSmtp(userId)
	fmt.Println(smtpConfig)
	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
    from :=  smtpConfig.Username
    pass := smtpConfig.Password
    to := smtpConfig.FromEmail

    msg := "From: " + from + "\n" +
        "To: " + to + "\n" +
        "Subject: Hello there\n\n" +
        body

    err = smtp.SendMail(smtpConfig.Host + ":" + fmt.Sprint(smtpConfig.Port),
        smtp.PlainAuth("", from, pass, smtpConfig.Host),
        from, []string{to}, []byte(msg))

    if err != nil {
        log.Printf("smtp error: %s", err)
        return
    }
    log.Println("Successfully sended to " + to)
}

func (s *NotificationManager) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	config := &shared_types.SMTPConfigs{}
	err := s.db.NewSelect().Model(config).Where("user_id = ?", ID).Scan(s.ctx)
	if err != nil {
		return nil, err
	}
	return config, nil
}