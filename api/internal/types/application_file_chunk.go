package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ApplicationFileChunk stores a semantic chunk of file content for RAG/context.
// Persisted to application_file_chunks table.
type ApplicationFileChunk struct {
	bun.BaseModel `bun:"table:application_file_chunks,alias:afc" swaggerignore:"true"`

	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid"`
	Path          string    `json:"path" bun:"path,notnull"`
	StartLine     int       `json:"start_line" bun:"start_line,notnull"`
	EndLine       int       `json:"end_line" bun:"end_line,notnull"`
	Content       string    `json:"content" bun:"content,notnull"`
	ChunkHash     string    `json:"chunk_hash" bun:"chunk_hash,notnull"`
	Language      string    `json:"language" bun:"language"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,notnull"`
}
