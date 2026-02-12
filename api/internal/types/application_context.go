package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ApplicationContext stores cached Merkle root and path→checksum for live dev manifest.
// Persisted to application_context table so manifest survives API restarts.
type ApplicationContext struct {
	bun.BaseModel `bun:"table:application_context,alias:ac" swaggerignore:"true"`

	ApplicationID uuid.UUID       `json:"application_id" bun:"application_id,pk,type:uuid"`
	RootHash      string          `json:"root_hash" bun:"root_hash"`
	Simhash       string          `json:"simhash" bun:"simhash"`
	Paths         PathChecksumMap `json:"paths" bun:"paths,type:jsonb,notnull"`
	UpdatedAt     time.Time       `json:"updated_at" bun:"updated_at,notnull"`
}

// PathChecksumMap is path → checksum for Merkle manifest.
// Implements driver.Valuer and sql.Scanner for jsonb.
type PathChecksumMap map[string]string

// Value implements driver.Valuer for DB writes.
// Returns string (not []byte) so pgx simple protocol sends it as text for jsonb parsing.
func (p PathChecksumMap) Value() (driver.Value, error) {
	if p == nil {
		return "{}", nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner for DB reads.
func (p *PathChecksumMap) Scan(value interface{}) error {
	if value == nil {
		*p = make(map[string]string)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		*p = make(map[string]string)
		return nil
	}
	return json.Unmarshal(b, p)
}
