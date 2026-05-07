package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
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

// Ecosystem detection patterns
var ecosystemPatterns = map[string]*regexp.Regexp{
	"nextjs":  regexp.MustCompile(`(?i)(next\.js|nextjs|next\.config)`),
	"nuxt":    regexp.MustCompile(`(?i)(nuxt|nuxt\.config)`),
	"react":   regexp.MustCompile(`(?i)(react|create-react-app|react-scripts)`),
	"vite":    regexp.MustCompile(`(?i)(vite|vite\.config)`),
	"django":  regexp.MustCompile(`(?i)(django|manage\.py|wsgi)`),
	"flask":   regexp.MustCompile(`(?i)(flask|gunicorn.*flask)`),
	"fastapi": regexp.MustCompile(`(?i)(fastapi|uvicorn)`),
	"rails":   regexp.MustCompile(`(?i)(rails|ruby on rails|Gemfile)`),
	"laravel": regexp.MustCompile(`(?i)(laravel|artisan|composer\.json.*laravel)`),
	"express": regexp.MustCompile(`(?i)(express|express\.js)`),
	"nestjs":  regexp.MustCompile(`(?i)(nestjs|@nestjs)`),
	"go":      regexp.MustCompile(`(?i)(go\.mod|golang|gin-gonic|fiber)`),
	"rust":    regexp.MustCompile(`(?i)(cargo\.toml|rust|actix|axum)`),
	"svelte":  regexp.MustCompile(`(?i)(svelte|sveltekit|svelte\.config)`),
	"astro":   regexp.MustCompile(`(?i)(astro|astro\.config)`),
}

// GetPatterns returns patterns for the given ecosystem with confidence >= 0.3.
// CreatePatternTables creates the deploy_patterns and deploy_outcomes tables if
// they do not yet exist. Called at AgentService startup so missing migrations
// don't cause repeated ERR log spam and lost outcome data.
func (s *AgentService) CreatePatternTables(ctx context.Context) {
	if _, err := s.store.DB.NewCreateTable().
		Model((*DeployPattern)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		s.logger.Log(logger.Error, "deploy_patterns: failed to create deploy_patterns table", err.Error())
	}
	if _, err := s.store.DB.NewCreateTable().
		Model((*DeployOutcome)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		s.logger.Log(logger.Error, "deploy_patterns: failed to create deploy_outcomes table", err.Error())
	}
}

func (s *AgentService) GetPatterns(ctx context.Context, ecosystem string) []DeployPattern {
	var patterns []DeployPattern
	err := s.store.DB.NewSelect().
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
func (s *AgentService) RecordOutcome(ctx context.Context, outcome *DeployOutcome) {
	_, err := s.store.DB.NewInsert().Model(outcome).Exec(ctx)
	if err != nil {
		s.logger.Log(logger.Error, "deploy_patterns: failed to record outcome", err.Error())
	}
}

// UpsertPattern creates or updates a deploy pattern with confidence recalculation.
func (s *AgentService) UpsertPattern(ctx context.Context, ecosystem, patternType, signature, resolution string, succeeded bool) {
	var existing DeployPattern
	err := s.store.DB.NewSelect().
		Model(&existing).
		Where("ecosystem = ?", ecosystem).
		Where("signature = ?", signature).
		Where("pattern_type = ?", patternType).
		Scan(ctx)

	if err != nil {
		// Pattern doesn't exist; create it
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
		_, _ = s.store.DB.NewInsert().Model(pattern).Exec(ctx)
		return
	}

	// Update existing pattern
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

	_, _ = s.store.DB.NewUpdate().
		Model(&existing).
		Column("hit_count", "miss_count", "confidence", "last_seen_at", "resolution").
		WherePK().
		Exec(ctx)
}

// detectEcosystem scans messages for ecosystem indicators.
func detectEcosystem(messages []memory.StoredMessage, currentInput string) string {
	text := currentInput
	for _, m := range messages {
		text += " " + m.Content
	}

	for eco, pattern := range ecosystemPatterns {
		if pattern.MatchString(text) {
			return eco
		}
	}
	return ""
}

// detectEcosystemFromLLM works with llm.Message slice.
func detectEcosystemFromLLM(messages []llm.Message, currentInput string) string {
	text := currentInput
	for _, m := range messages {
		text += " " + m.Content
	}
	for eco, pattern := range ecosystemPatterns {
		if pattern.MatchString(text) {
			return eco
		}
	}
	return ""
}

// formatPatterns builds the [deploy-patterns] block for injection.
func formatPatterns(ecosystem string, patterns []DeployPattern) string {
	if len(patterns) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[deploy-patterns] ecosystem:%s\n", ecosystem))

	// Group by pattern type
	fixes := filterByType(patterns, "failure_fix")
	pitfalls := filterByType(patterns, "pitfall")
	fastPaths := filterByType(patterns, "fast_path")

	if len(fixes) > 0 {
		sb.WriteString("known_fixes:\n")
		for _, p := range fixes {
			sb.WriteString(fmt.Sprintf("- \"%s\" → %s (confidence:%d%%, seen:%d)\n",
				p.Signature, p.Resolution, int(p.Confidence*100), p.HitCount+p.MissCount))
		}
	}
	if len(pitfalls) > 0 {
		sb.WriteString("pitfalls:\n")
		for _, p := range pitfalls {
			sb.WriteString(fmt.Sprintf("- %s → %s (confidence:%d%%)\n",
				p.Signature, p.Resolution, int(p.Confidence*100)))
		}
	}
	if len(fastPaths) > 0 {
		sb.WriteString("fast_paths:\n")
		for _, p := range fastPaths {
			sb.WriteString(fmt.Sprintf("- %s → %s (confidence:%d%%)\n",
				p.Signature, p.Resolution, int(p.Confidence*100)))
		}
	}

	sb.WriteString("[/deploy-patterns]")
	return sb.String()
}

func filterByType(patterns []DeployPattern, patternType string) []DeployPattern {
	var result []DeployPattern
	for _, p := range patterns {
		if p.PatternType == patternType {
			result = append(result, p)
		}
	}
	return result
}

// recordDeployOutcome scans finished conversation for deploy activity and records outcome.
func (s *AgentService) recordDeployOutcome(ctx context.Context, orgID string, messages []memory.StoredMessage) {
	ecosystem := detectEcosystem(messages, "")
	if ecosystem == "" {
		return
	}

	state := &deployState{}
	var failureSignatures []string
	var fixesApplied []string
	stepsCount := 0
	selfHealAttempts := 0

	for _, msg := range messages {
		if msg.Role == llm.RoleAssistant {
			stepsCount++
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					extractIDsFromArgs(tc.Function.Arguments, state)
					if tc.Function.Name == "redeploy_application" || tc.Function.Name == "restart_deployment" {
						selfHealAttempts++
					}
				}
			}
		}
		if msg.Role == llm.RoleTool && msg.Content != "" {
			extractIDsFromResult(msg.Content, state)

			// Check for failure signatures in tool results
			var parsed map[string]interface{}
			if json.Unmarshal([]byte(msg.Content), &parsed) == nil {
				if errMsg, ok := parsed["error"].(string); ok && errMsg != "" {
					failureSignatures = append(failureSignatures, truncateStr(errMsg, 200))
				}
			}
		}
		if msg.Role == llm.RoleAssistant && msg.Content != "" {
			// Detect fixes mentioned in assistant responses
			if strings.Contains(strings.ToLower(msg.Content), "fixed") ||
				strings.Contains(strings.ToLower(msg.Content), "resolved") {
				fixesApplied = append(fixesApplied, truncateStr(msg.Content, 100))
			}
		}
	}

	// Classify outcome
	outcome := classifyOutcome(state.Status, messages)
	if outcome == "" {
		return
	}

	record := &DeployOutcome{
		OrgID:             nilIfEmpty(orgID),
		ApplicationID:     nilIfEmpty(state.ApplicationID),
		Ecosystem:         ecosystem,
		Source:            nilIfEmpty("agent"),
		Outcome:           outcome,
		StepsCount:        stepsCount,
		SelfHealAttempts:  selfHealAttempts,
		FailureSignatures: failureSignatures,
		FixesApplied:      fixesApplied,
		Metadata:          JSONMap{"deployment_id": state.DeploymentID},
	}

	s.RecordOutcome(ctx, record)

	// Update patterns based on outcome
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

func classifyOutcome(lastStatus string, messages []memory.StoredMessage) string {
	switch lastStatus {
	case "running", "active", "healthy":
		return "success"
	case "build_failed", "failed", "error", "crashed":
		return "failed"
	}

	// Fallback: scan last few messages for indicators
	for i := len(messages) - 1; i >= 0 && i >= len(messages)-5; i-- {
		content := strings.ToLower(messages[i].Content)
		if strings.Contains(content, "deployed successfully") ||
			strings.Contains(content, "is now live") ||
			strings.Contains(content, "running at") {
			return "success"
		}
		if strings.Contains(content, "failed") ||
			strings.Contains(content, "error") ||
			strings.Contains(content, "could not deploy") {
			return "failed"
		}
		if strings.Contains(content, "rollback") ||
			strings.Contains(content, "rolled back") {
			return "rollback"
		}
	}
	return ""
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
