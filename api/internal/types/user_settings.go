package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserSettings struct {
	bun.BaseModel `bun:"table:user_settings,alias:us" swaggerignore:"true"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID        uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
	FontFamily    string     `json:"font_family" bun:"font_family,notnull,default:'system-ui'"`
	FontSize      int        `json:"font_size" bun:"font_size,notnull,default:14"`
	Theme         string     `json:"theme" bun:"theme,notnull,default:'light'"`
	Language      string     `json:"language" bun:"language,notnull,default:'en'"`
	AutoUpdate    bool       `json:"auto_update" bun:"auto_update,notnull,default:false"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`

	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}
