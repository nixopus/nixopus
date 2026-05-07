package deploy

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/uptrace/bun"
)

// JSONMap is a generic JSON-serializable map for bun.
type JSONMap map[string]interface{}

var _ driver.Valuer = JSONMap{}
var _ sql.Scanner = &JSONMap{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *JSONMap) Scan(src interface{}) error {
	if src == nil {
		*m = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("deploy_patterns: cannot scan %T into JSONMap", src)
	}
	return json.Unmarshal(data, m)
}

// JSONStringSlice is a []string stored as JSONB.
type JSONStringSlice []string

var _ driver.Valuer = JSONStringSlice{}
var _ sql.Scanner = &JSONStringSlice{}

func (s JSONStringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *JSONStringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("deploy_patterns: cannot scan %T into JSONStringSlice", src)
	}
	return json.Unmarshal(data, s)
}

// DeployPattern represents a learned deployment pattern.
type DeployPattern struct {
	bun.BaseModel `bun:"table:deploy_patterns"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Ecosystem     string    `bun:"ecosystem,notnull"`
	Framework     *string   `bun:"framework"`
	PatternType   string    `bun:"pattern_type,notnull"`
	Signature     string    `bun:"signature,notnull"`
	Resolution    string    `bun:"resolution,notnull"`
	Confidence    float64   `bun:"confidence,notnull,default:0.5"`
	HitCount      int       `bun:"hit_count,notnull,default:0"`
	MissCount     int       `bun:"miss_count,notnull,default:0"`
	LastSeenAt    time.Time `bun:"last_seen_at"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
}

// DeployOutcome records the result of a deploy conversation.
type DeployOutcome struct {
	bun.BaseModel     `bun:"table:deploy_outcomes"`
	ID                uuid.UUID       `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	OrgID             *string         `bun:"org_id"`
	ApplicationID     *string         `bun:"application_id"`
	Ecosystem         string          `bun:"ecosystem,notnull"`
	Framework         *string         `bun:"framework"`
	Source            *string         `bun:"source"`
	Outcome           string          `bun:"outcome,notnull"`
	StepsCount        int             `bun:"steps_count"`
	SelfHealAttempts  int             `bun:"self_heal_attempts"`
	FailureSignatures JSONStringSlice `bun:"failure_signatures,type:jsonb"`
	FixesApplied      JSONStringSlice `bun:"fixes_applied,type:jsonb"`
	Metadata          JSONMap         `bun:"metadata,type:jsonb"`
	CreatedAt         time.Time       `bun:"created_at,default:current_timestamp"`
}

// Store handles deploy pattern persistence.
type Store struct {
	DB     *bun.DB
	Logger logger.Logger
}

func NewStore(db *bun.DB, l logger.Logger) *Store {
	return &Store{DB: db, Logger: l}
}

// CreateTables creates the deploy_patterns and deploy_outcomes tables if
// they do not yet exist.
func (s *Store) CreateTables(ctx context.Context) {
	if _, err := s.DB.NewCreateTable().
		Model((*DeployPattern)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		s.Logger.Log(logger.Error, "deploy_patterns: failed to create deploy_patterns table", err.Error())
	}
	if _, err := s.DB.NewCreateTable().
		Model((*DeployOutcome)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		s.Logger.Log(logger.Error, "deploy_patterns: failed to create deploy_outcomes table", err.Error())
	}
}

func (s *Store) GetPatterns(ctx context.Context, ecosystem string) []DeployPattern {
	var patterns []DeployPattern
	err := s.DB.NewSelect().
		Model(&patterns).
		Where("ecosystem = ?", ecosystem).
		Where("confidence >= 0.3").
		OrderExpr("confidence DESC").
		Limit(15).
		Scan(ctx)
	if err != nil {
		return nil
	}
	return patterns
}

// RecordOutcome inserts a deploy outcome record.
func (s *Store) RecordOutcome(ctx context.Context, outcome *DeployOutcome) {
	_, err := s.DB.NewInsert().Model(outcome).Exec(ctx)
	if err != nil {
		s.Logger.Log(logger.Error, "deploy_patterns: failed to record outcome", err.Error())
	}
}

// UpsertPattern creates or updates a deploy pattern with confidence recalculation.
func (s *Store) UpsertPattern(ctx context.Context, ecosystem, patternType, signature, resolution string, succeeded bool) {
	var existing DeployPattern
	err := s.DB.NewSelect().
		Model(&existing).
		Where("ecosystem = ?", ecosystem).
		Where("signature = ?", signature).
		Where("pattern_type = ?", patternType).
		Scan(ctx)

	if err != nil {
		conf := 0.6
		hit, miss := 1, 0
		if !succeeded {
			conf = 0.4
			hit, miss = 0, 1
		}
		pattern := &DeployPattern{
			Ecosystem:   ecosystem,
			PatternType: patternType,
			Signature:   signature,
			Resolution:  resolution,
			Confidence:  conf,
			HitCount:    hit,
			MissCount:   miss,
			LastSeenAt:  time.Now(),
		}
		_, _ = s.DB.NewInsert().Model(pattern).Exec(ctx)
		return
	}

	if succeeded {
		existing.HitCount++
	} else {
		existing.MissCount++
	}
	total := existing.HitCount + existing.MissCount
	if total > 0 {
		existing.Confidence = float64(existing.HitCount) / float64(total)
	}
	existing.LastSeenAt = time.Now()
	if succeeded && resolution != "" {
		existing.Resolution = resolution
	}

	_, _ = s.DB.NewUpdate().
		Model(&existing).
		Column("hit_count", "miss_count", "confidence", "last_seen_at", "resolution").
		WherePK().
		Exec(ctx)
}

// RecordDeployOutcome scans finished conversation for deploy activity and records outcome.
func (s *Store) RecordDeployOutcome(ctx context.Context, orgID string, messages []memory.StoredMessage) {
	ecosystem := DetectEcosystem(messages, "")
	if ecosystem == "" {
		return
	}

	state := &State{}
	var failureSignatures []string
	var fixesApplied []string
	stepsCount := 0
	selfHealAttempts := 0

	for _, msg := range messages {
		if msg.Role == llm.RoleAssistant {
			stepsCount++
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					ExtractIDsFromArgs(tc.Function.Arguments, state)
					if tc.Function.Name == "redeploy_application" || tc.Function.Name == "restart_deployment" {
						selfHealAttempts++
					}
				}
			}
		}
		if msg.Role == llm.RoleTool && msg.Content != "" {
			ExtractIDsFromResult(msg.Content, state)

			var parsed map[string]interface{}
			if json.Unmarshal([]byte(msg.Content), &parsed) == nil {
				if errMsg, ok := parsed["error"].(string); ok && errMsg != "" {
					failureSignatures = append(failureSignatures, TruncateStr(errMsg, 200))
				}
			}
		}
		if msg.Role == llm.RoleAssistant && msg.Content != "" {
			if strings.Contains(strings.ToLower(msg.Content), "fixed") ||
				strings.Contains(strings.ToLower(msg.Content), "resolved") {
				fixesApplied = append(fixesApplied, TruncateStr(msg.Content, 100))
			}
		}
	}

	outcome := ClassifyOutcome(state.Status, messages)
	if outcome == "" {
		return
	}

	record := &DeployOutcome{
		OrgID:             NilIfEmpty(orgID),
		ApplicationID:     NilIfEmpty(state.ApplicationID),
		Ecosystem:         ecosystem,
		Source:            NilIfEmpty("agent"),
		Outcome:           outcome,
		StepsCount:        stepsCount,
		SelfHealAttempts:  selfHealAttempts,
		FailureSignatures: failureSignatures,
		FixesApplied:      fixesApplied,
		Metadata:          JSONMap{"deployment_id": state.DeploymentID},
	}

	s.RecordOutcome(ctx, record)

	if outcome == "success" && len(failureSignatures) > 0 && len(fixesApplied) > 0 {
		for _, sig := range failureSignatures {
			s.UpsertPattern(ctx, ecosystem, "failure_fix", sig, fixesApplied[0], true)
		}
	} else if outcome == "failed" && len(failureSignatures) > 0 {
		for _, sig := range failureSignatures {
			s.UpsertPattern(ctx, ecosystem, "failure_fix", sig, "", false)
		}
	}
}

func TruncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func NilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
