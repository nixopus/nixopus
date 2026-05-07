package deploy

import (
	"testing"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/stretchr/testify/assert"
)

func TestDetectEcosystem_NextJS(t *testing.T) {
	messages := []memory.StoredMessage{
		{Role: llm.RoleUser, Content: "deploy my next.js app from github.com/org/frontend"},
	}
	eco := DetectEcosystem(messages, "")
	assert.Equal(t, "nextjs", eco)
}

func TestDetectEcosystem_Django(t *testing.T) {
	messages := []memory.StoredMessage{
		{Role: llm.RoleUser, Content: "I have a Django project with manage.py"},
	}
	eco := DetectEcosystem(messages, "")
	assert.Equal(t, "django", eco)
}

func TestDetectEcosystem_Go(t *testing.T) {
	eco := DetectEcosystem(nil, "deploy my go.mod project with gin-gonic")
	assert.Equal(t, "go", eco)
}

func TestDetectEcosystem_Rust(t *testing.T) {
	messages := []memory.StoredMessage{
		{Role: llm.RoleTool, Content: `{"files":["Cargo.toml","src/main.rs"]}`},
	}
	eco := DetectEcosystem(messages, "")
	assert.Equal(t, "rust", eco)
}

func TestDetectEcosystem_NoMatch(t *testing.T) {
	eco := DetectEcosystem(nil, "hello world")
	assert.Equal(t, "", eco)
}

func TestDetectEcosystem_FromCurrentInput(t *testing.T) {
	eco := DetectEcosystem(nil, "deploy my nuxt.config.ts app")
	assert.Equal(t, "nuxt", eco)
}

func TestDetectEcosystemFromLLM(t *testing.T) {
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "deploy vite.config.ts project"},
	}
	eco := DetectEcosystemFromLLM(messages, "")
	assert.Equal(t, "vite", eco)
}

func TestFormatPatterns_Empty(t *testing.T) {
	result := FormatPatterns("nextjs", nil)
	assert.Equal(t, "", result)
}

func TestFormatPatterns_WithFixes(t *testing.T) {
	patterns := []DeployPattern{
		{
			Ecosystem:   "nextjs",
			PatternType: "failure_fix",
			Signature:   "ENOENT .next/standalone",
			Resolution:  "add output: standalone to next.config.js",
			Confidence:  0.85,
			HitCount:    3,
			MissCount:   1,
		},
	}
	result := FormatPatterns("nextjs", patterns)
	assert.Contains(t, result, "[deploy-patterns] ecosystem:nextjs")
	assert.Contains(t, result, "known_fixes:")
	assert.Contains(t, result, "ENOENT .next/standalone")
	assert.Contains(t, result, "add output: standalone to next.config.js")
	assert.Contains(t, result, "confidence:85%")
	assert.Contains(t, result, "seen:4")
	assert.Contains(t, result, "[/deploy-patterns]")
}

func TestFormatPatterns_MultiplTypes(t *testing.T) {
	patterns := []DeployPattern{
		{PatternType: "failure_fix", Signature: "err1", Resolution: "fix1", Confidence: 0.9, HitCount: 5, MissCount: 0},
		{PatternType: "pitfall", Signature: "port conflict", Resolution: "use 3000", Confidence: 0.7, HitCount: 2, MissCount: 1},
		{PatternType: "fast_path", Signature: "has dockerfile", Resolution: "skip buildpack detection", Confidence: 0.95, HitCount: 10, MissCount: 0},
	}
	result := FormatPatterns("react", patterns)
	assert.Contains(t, result, "known_fixes:")
	assert.Contains(t, result, "pitfalls:")
	assert.Contains(t, result, "fast_paths:")
}

func TestClassifyOutcome_Success(t *testing.T) {
	result := ClassifyOutcome("running", nil)
	assert.Equal(t, "success", result)

	result = ClassifyOutcome("active", nil)
	assert.Equal(t, "success", result)
}

func TestClassifyOutcome_Failed(t *testing.T) {
	result := ClassifyOutcome("build_failed", nil)
	assert.Equal(t, "failed", result)

	result = ClassifyOutcome("failed", nil)
	assert.Equal(t, "failed", result)
}

func TestClassifyOutcome_FromMessages(t *testing.T) {
	messages := []memory.StoredMessage{
		{Role: llm.RoleAssistant, Content: "The application is now live at https://app.example.com"},
	}
	result := ClassifyOutcome("", messages)
	assert.Equal(t, "success", result)
}

func TestClassifyOutcome_RollbackFromMessages(t *testing.T) {
	messages := []memory.StoredMessage{
		{Role: llm.RoleAssistant, Content: "I've initiated a rollback to the previous version"},
	}
	result := ClassifyOutcome("", messages)
	assert.Equal(t, "rollback", result)
}

func TestClassifyOutcome_Unknown(t *testing.T) {
	messages := []memory.StoredMessage{
		{Role: llm.RoleAssistant, Content: "Sure, let me check that for you."},
	}
	result := ClassifyOutcome("", messages)
	assert.Equal(t, "", result)
}

func TestJSONStringSlice_ValueAndScan(t *testing.T) {
	s := JSONStringSlice{"a", "b", "c"}
	val, err := s.Value()
	assert.NoError(t, err)
	assert.Equal(t, `["a","b","c"]`, val)

	var scanned JSONStringSlice
	err = scanned.Scan([]byte(`["x","y"]`))
	assert.NoError(t, err)
	assert.Equal(t, JSONStringSlice{"x", "y"}, scanned)
}

func TestJSONMap_ValueAndScan(t *testing.T) {
	m := JSONMap{"key": "value"}
	val, err := m.Value()
	assert.NoError(t, err)
	assert.Contains(t, val, "key")

	var scanned JSONMap
	err = scanned.Scan([]byte(`{"foo":"bar"}`))
	assert.NoError(t, err)
	assert.Equal(t, "bar", scanned["foo"])
}

func TestNilIfEmpty(t *testing.T) {
	assert.Nil(t, NilIfEmpty(""))
	result := NilIfEmpty("hello")
	assert.NotNil(t, result)
	assert.Equal(t, "hello", *result)
}

func TestTruncateStr(t *testing.T) {
	assert.Equal(t, "hello", TruncateStr("hello", 10))
	assert.Equal(t, "hel", TruncateStr("hello", 3))
}
