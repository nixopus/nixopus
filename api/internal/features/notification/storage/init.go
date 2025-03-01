package storage

import (
	"strconv"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type NotificationStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func (s NotificationStorage) AddSmtp(config *shared_types.SMTPConfigs) error {
	_, err := s.DB.NewInsert().Model(config).Exec(s.Ctx)
	return err
}

func (s NotificationStorage) UpdateSmtp(config *notification.UpdateSMTPConfigRequest) error {
	var smtp *shared_types.SMTPConfigs
	_, err := s.DB.NewUpdate().Model(smtp).
		Set("host = ?", config.Host).
		Set("port = ?", strconv.Itoa(config.Port)).
		Set("username = ?", config.Username).
		Set("password = ?", config.Password).
		Set("from_name = ?", config.FromName).
		Set("from_email = ?", config.FromEmail).
		Where("id = ?", config.ID).Exec(s.Ctx)
	return err
}

func (s NotificationStorage) DeleteSmtp(ID string) error {
	var config shared_types.SMTPConfigs
	_, err := s.DB.NewDelete().Model(config).Where("id = ?", ID).Exec(s.Ctx)
	return err
}

func (s NotificationStorage) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	config := &shared_types.SMTPConfigs{}
	err := s.DB.NewSelect().Model(config).Where("user_id = ?", ID).Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return config, nil
}
