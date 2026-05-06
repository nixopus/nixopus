package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	deploy_types "github.com/nixopus/nixopus/api/internal/features/deploy/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func TestGetSourceResolver_PublicGit(t *testing.T) {
	svc := &TaskService{}

	resolver := svc.GetSourceResolver(shared_types.SourcePublicGit)

	if resolver == nil {
		t.Fatal("expected non-nil resolver for SourcePublicGit")
	}
	if _, ok := resolver.(*PublicGitSourceResolver); !ok {
		t.Errorf("expected *PublicGitSourceResolver, got %T", resolver)
	}
}

func TestGetSourceResolver_DefaultIsGithub(t *testing.T) {
	svc := &TaskService{}

	resolver := svc.GetSourceResolver(shared_types.SourceGithub)

	if resolver == nil {
		t.Fatal("expected non-nil resolver for SourceGithub")
	}
	if _, ok := resolver.(*GithubSourceResolver); !ok {
		t.Errorf("expected *GithubSourceResolver, got %T", resolver)
	}
}

func TestGetSourceResolver_AllSourceTypes(t *testing.T) {
	svc := &TaskService{}

	tests := []struct {
		source       shared_types.Source
		expectedType string
	}{
		{shared_types.SourceS3, "*tasks.S3SourceResolver"},
		{shared_types.SourceZip, "*tasks.ZipSourceResolver"},
		{shared_types.SourceStaging, "*tasks.StagingSourceResolver"},
		{shared_types.SourceTemplate, "*tasks.TemplateSourceResolver"},
		{shared_types.SourcePublicGit, "*tasks.PublicGitSourceResolver"},
		{shared_types.SourceGithub, "*tasks.GithubSourceResolver"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			resolver := svc.GetSourceResolver(tt.source)
			if resolver == nil {
				t.Fatalf("expected non-nil resolver for %s", tt.source)
			}
			actualType := fmt.Sprintf("%T", resolver)
			if actualType != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, actualType)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func sortedPorts(ports []int) []int {
	cp := make([]int, len(ports))
	copy(cp, ports)
	sort.Ints(cp)
	return cp
}

func portsEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	as, bs := sortedPorts(a), sortedPorts(b)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// ParseComposeFile
// ---------------------------------------------------------------------------

func TestParseComposeFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		content := []byte(`
services:
  web:
    ports:
      - "8080:80"
`)
		dir := t.TempDir()
		path := filepath.Join(dir, "docker-compose.yml")
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
		svcs, err := ParseComposeFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(svcs) != 1 {
			t.Fatalf("expected 1 service, got %d", len(svcs))
		}
		if svcs[0].ServiceName != "web" {
			t.Errorf("expected service name 'web', got %q", svcs[0].ServiceName)
		}
		if !portsEqual(svcs[0].Ports, []int{8080}) {
			t.Errorf("expected ports [8080], got %v", svcs[0].Ports)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := ParseComposeFile("/no/such/file/docker-compose.yml")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.yml")
		if err := os.WriteFile(path, []byte("{{bad yaml{{"), 0644); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
		_, err := ParseComposeFile(path)
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})
}

// ---------------------------------------------------------------------------
// ParseComposeYAML
// ---------------------------------------------------------------------------

func TestParseComposeYAML(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		envVars     map[string]string
		wantErr     bool
		checkResult func(t *testing.T, svcs []ParsedComposeService)
	}{
		{
			name: "empty YAML braces",
			yaml: "{}",
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if len(svcs) != 0 {
					t.Errorf("expected 0 services, got %d", len(svcs))
				}
			},
		},
		{
			name: "services with no ports",
			yaml: `
services:
  app:
    image: nginx
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if len(svcs) != 1 {
					t.Fatalf("expected 1 service, got %d", len(svcs))
				}
				if len(svcs[0].Ports) != 0 {
					t.Errorf("expected no ports, got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "short port syntax — single port",
			yaml: `
services:
  app:
    ports:
      - "80"
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{80}) {
					t.Errorf("expected [80], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "short port syntax — host:container",
			yaml: `
services:
  app:
    ports:
      - "8080:80"
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{8080}) {
					t.Errorf("expected [8080], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "short port syntax — range",
			yaml: `
services:
  app:
    ports:
      - "8080-8082:8080-8082"
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{8080}) {
					t.Errorf("expected [8080], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "long port syntax",
			yaml: `
services:
  app:
    ports:
      - target: 80
        published: 9090
        protocol: tcp
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{9090}) {
					t.Errorf("expected [9090], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "expose ports fallback",
			yaml: `
services:
  app:
    expose:
      - "3000"
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{3000}) {
					t.Errorf("expected [3000], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "env var substitution in port",
			yaml: `
services:
  app:
    ports:
      - "${APP_PORT:-5000}:5000"
`,
			envVars: map[string]string{"APP_PORT": "4000"},
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{4000}) {
					t.Errorf("expected [4000], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name: "env var default used when absent",
			yaml: `
services:
  app:
    ports:
      - "${MISSING_PORT:-6000}:6000"
`,
			checkResult: func(t *testing.T, svcs []ParsedComposeService) {
				if !portsEqual(svcs[0].Ports, []int{6000}) {
					t.Errorf("expected [6000], got %v", svcs[0].Ports)
				}
			},
		},
		{
			name:    "invalid YAML",
			yaml:    "services: [\nunclosed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var svcs []ParsedComposeService
			var err error
			if tt.envVars != nil {
				svcs, err = ParseComposeYAML([]byte(tt.yaml), tt.envVars)
			} else {
				svcs, err = ParseComposeYAML([]byte(tt.yaml))
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, svcs)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// resolveEnvVars
// ---------------------------------------------------------------------------

func TestResolveEnvVars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		env   map[string]string
		want  string
	}{
		{
			name:  "var present in env",
			input: "${PORT:-3000}",
			env:   map[string]string{"PORT": "8080"},
			want:  "8080",
		},
		{
			name:  "var absent, default used",
			input: "${PORT:-3000}",
			env:   map[string]string{},
			want:  "3000",
		},
		{
			name:  "var absent, no default",
			input: "${PORT}",
			env:   map[string]string{},
			want:  "",
		},
		{
			name:  "no vars in string",
			input: "8080:80",
			env:   map[string]string{"PORT": "9090"},
			want:  "8080:80",
		},
		{
			name:  "multiple vars",
			input: "${HOST:-localhost}:${PORT:-3000}",
			env:   map[string]string{"HOST": "myhost", "PORT": "4000"},
			want:  "myhost:4000",
		},
		{
			name:  "nil env uses default",
			input: "${VAR:-fallback}",
			env:   nil,
			want:  "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveEnvVars(tt.input, tt.env)
			if got != tt.want {
				t.Errorf("resolveEnvVars(%q, ...) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseShortPortSyntax
// ---------------------------------------------------------------------------

func TestParseShortPortSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		env   map[string]string
		want  []int
	}{
		{
			name:  "single port",
			input: "80",
			want:  []int{80},
		},
		{
			name:  "host:container",
			input: "3000:3000",
			want:  []int{3000},
		},
		{
			name:  "range — first port of range taken",
			input: "8080-8082:8080-8082",
			want:  []int{8080},
		},
		{
			name:  "ip:host:container",
			input: "127.0.0.1:9000:80",
			want:  []int{9000},
		},
		{
			name:  "protocol suffix stripped",
			input: "5000:5000/tcp",
			want:  []int{5000},
		},
		{
			name:  "env var substitution",
			input: "${PORT:-7000}:7000",
			env:   map[string]string{"PORT": "7777"},
			want:  []int{7777},
		},
		{
			name:  "env var default when absent",
			input: "${PORT:-7000}:7000",
			want:  []int{7000},
		},
		{
			name:  "invalid non-numeric",
			input: "abc:80",
			want:  nil,
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseShortPortSyntax(tt.input, tt.env)
			if !portsEqual(got, tt.want) {
				t.Errorf("parseShortPortSyntax(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parsePortOrRangeFirst
// ---------------------------------------------------------------------------

func TestParsePortOrRangeFirst(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"80", 80},
		{"8080-8090", 8080},
		{"0", 0},
		{"abc", 0},
		{"", 0},
		{"65535", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parsePortOrRangeFirst(tt.input)
			if got != tt.want {
				t.Errorf("parsePortOrRangeFirst(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseLongPortSyntax
// ---------------------------------------------------------------------------

func TestParseLongPortSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  []int
	}{
		{
			name:  "published as int",
			input: map[string]interface{}{"published": 8080, "target": 80},
			want:  []int{8080},
		},
		{
			name:  "published as float64",
			input: map[string]interface{}{"published": float64(9090), "target": 80},
			want:  []int{9090},
		},
		{
			name:  "published as string",
			input: map[string]interface{}{"published": "7070", "target": 80},
			want:  []int{7070},
		},
		{
			name:  "published as string range — first taken",
			input: map[string]interface{}{"published": "7000-7010", "target": 80},
			want:  []int{7000},
		},
		{
			name:  "no published, target as int",
			input: map[string]interface{}{"target": 3000},
			want:  []int{3000},
		},
		{
			name:  "no published, target as float64",
			input: map[string]interface{}{"target": float64(4000)},
			want:  []int{4000},
		},
		{
			name:  "neither published nor target",
			input: map[string]interface{}{"protocol": "tcp"},
			want:  nil,
		},
		{
			name:  "empty map",
			input: map[string]interface{}{},
			want:  nil,
		},
		{
			name:  "published invalid string",
			input: map[string]interface{}{"published": "abc"},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLongPortSyntax(tt.input)
			if !portsEqual(got, tt.want) {
				t.Errorf("parseLongPortSyntax(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractExposePorts
// ---------------------------------------------------------------------------

func TestExtractExposePorts(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
		want  []int
	}{
		{
			name:  "int port",
			input: []interface{}{3000},
			want:  []int{3000},
		},
		{
			name:  "string port",
			input: []interface{}{"8080"},
			want:  []int{8080},
		},
		{
			name:  "float64 port",
			input: []interface{}{float64(5432)},
			want:  []int{5432},
		},
		{
			name:  "invalid string — ignored",
			input: []interface{}{"notaport"},
			want:  nil,
		},
		{
			name:  "mixed valid and invalid",
			input: []interface{}{80, "bad", 443},
			want:  []int{80, 443},
		},
		{
			name:  "duplicate ports deduplicated",
			input: []interface{}{80, "80", 80},
			want:  []int{80},
		},
		{
			name:  "port 0 ignored",
			input: []interface{}{0},
			want:  nil,
		},
		{
			name:  "port above 65535 ignored",
			input: []interface{}{65536},
			want:  nil,
		},
		{
			name:  "empty slice",
			input: []interface{}{},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractExposePorts(tt.input)
			if !portsEqual(got, tt.want) {
				t.Errorf("extractExposePorts(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractPorts
// ---------------------------------------------------------------------------

func TestExtractPorts(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
		env   map[string]string
		want  []int
	}{
		{
			name:  "string short syntax",
			input: []interface{}{"8080:80"},
			want:  []int{8080},
		},
		{
			name:  "int port",
			input: []interface{}{3000},
			want:  []int{3000},
		},
		{
			name:  "float64 port",
			input: []interface{}{float64(5000)},
			want:  []int{5000},
		},
		{
			name:  "long map syntax",
			input: []interface{}{map[string]interface{}{"published": 9000, "target": 80}},
			want:  []int{9000},
		},
		{
			name:  "invalid type ignored",
			input: []interface{}{true},
			want:  nil,
		},
		{
			name:  "deduplication",
			input: []interface{}{"80", 80, float64(80)},
			want:  []int{80},
		},
		{
			name:  "env var in string port",
			input: []interface{}{"${SVCPORT:-2000}:2000"},
			env:   map[string]string{"SVCPORT": "2222"},
			want:  []int{2222},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPorts(tt.input, tt.env)
			if !portsEqual(got, tt.want) {
				t.Errorf("extractPorts(%v, ...) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetStringFromMap / GetMapFromString (util.go)
// ---------------------------------------------------------------------------

func TestGetStringFromMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
	}{
		{
			name:  "empty map returns empty string",
			input: map[string]string{},
		},
		{
			name:  "single key-value pair",
			input: map[string]string{"KEY": "VALUE"},
		},
		{
			name:  "multiple pairs",
			input: map[string]string{"A": "1", "B": "2", "C": "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := GetStringFromMap(tt.input)
			if len(tt.input) == 0 {
				if s != "" {
					t.Errorf("expected empty string for empty map, got %q", s)
				}
				return
			}
			// Round-trip: the string must deserialize back to the same map
			got := GetMapFromString(s)
			for k, v := range tt.input {
				if got[k] != v {
					t.Errorf("round-trip: key %q: got %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestGetMapFromString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "empty string",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "JSON format",
			input: `{"FOO":"bar","BAZ":"qux"}`,
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "legacy space-delimited format",
			input: "KEY1=val1 KEY2=val2",
			want:  map[string]string{"KEY1": "val1", "KEY2": "val2"},
		},
		{
			name:  "legacy single pair",
			input: "ONLY=one",
			want:  map[string]string{"ONLY": "one"},
		},
		{
			name:  "value with equals sign",
			input: "K=v=with=equals",
			want:  map[string]string{"K": "v=with=equals"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMapFromString(tt.input)
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q: got %q, want %q", k, got[k], v)
				}
			}
			if len(got) != len(tt.want) {
				t.Errorf("map length: got %d, want %d; map=%v", len(got), len(tt.want), got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// containsSensitiveKeyword (run.go)
// ---------------------------------------------------------------------------

func TestContainsSensitiveKeyword(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"PASSWORD", true},
		{"DB_PASSWORD", true},
		{"secret_key", true},
		{"API_TOKEN", true},
		{"AUTH_HEADER", true},
		{"PRIVATE_KEY", true},
		{"AWS_SECRET_ACCESS_KEY", true},
		{"CREDENTIAL_FILE", true},
		{"APP_NAME", false},
		{"PORT", false},
		{"DATABASE_HOST", false},
		{"LOG_LEVEL", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := containsSensitiveKeyword(tt.key)
			if got != tt.want {
				t.Errorf("containsSensitiveKeyword(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// sanitizeEnvVars (run.go) - method on *TaskService (no infra used)
// ---------------------------------------------------------------------------

func TestSanitizeEnvVars(t *testing.T) {
	svc := &TaskService{}

	tests := []struct {
		name    string
		envVars map[string]string
		check   func(t *testing.T, result []string)
	}{
		{
			name:    "empty map",
			envVars: map[string]string{},
			check: func(t *testing.T, result []string) {
				if len(result) != 0 {
					t.Errorf("expected empty slice, got %v", result)
				}
			},
		},
		{
			name:    "non-sensitive key is passed through",
			envVars: map[string]string{"APP_ENV": "production"},
			check: func(t *testing.T, result []string) {
				if len(result) != 1 || result[0] != "APP_ENV=production" {
					t.Errorf("unexpected result: %v", result)
				}
			},
		},
		{
			name:    "sensitive key is masked",
			envVars: map[string]string{"DB_PASSWORD": "s3cr3t"},
			check: func(t *testing.T, result []string) {
				if len(result) != 1 || result[0] != "DB_PASSWORD=********" {
					t.Errorf("unexpected result: %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.sanitizeEnvVars(tt.envVars)
			tt.check(t, got)
		})
	}
}

// ---------------------------------------------------------------------------
// webhookDedupKey (webhook.go)
// ---------------------------------------------------------------------------

func TestWebhookDedupKey(t *testing.T) {
	tests := []struct {
		appID      string
		commitHash string
		want       string
	}{
		{"app-123", "abc123", "webhook:dedup:app-123:abc123"},
		{"", "", "webhook:dedup::"},
		{"myapp", "deadbeef", "webhook:dedup:myapp:deadbeef"},
	}

	for _, tt := range tests {
		t.Run(tt.appID+":"+tt.commitHash, func(t *testing.T) {
			got := webhookDedupKey(tt.appID, tt.commitHash)
			if got != tt.want {
				t.Errorf("webhookDedupKey(%q, %q) = %q, want %q", tt.appID, tt.commitHash, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseComposeImagesOutput / extractImageTags (s3_export.go)
// ---------------------------------------------------------------------------

func TestParseComposeImagesOutput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantNil bool
	}{
		{
			name:    "empty output",
			input:   "",
			wantNil: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantNil: true,
		},
		{
			name:  "JSON array format",
			input: `[{"Repository":"myapp","Tag":"latest"},{"Repository":"redis","Tag":"7"}]`,
			want:  []string{"myapp:latest", "redis:7"},
		},
		{
			name:  "JSON lines format",
			input: `{"Repository":"myapp","Tag":"latest"}` + "\n" + `{"Repository":"redis","Tag":"7"}`,
			want:  []string{"myapp:latest", "redis:7"},
		},
		{
			name:  "JSON lines deduplication",
			input: `{"Repository":"myapp","Tag":"latest"}` + "\n" + `{"Repository":"myapp","Tag":"latest"}`,
			want:  []string{"myapp:latest"},
		},
		{
			name:  "JSON array deduplication",
			input: `[{"Repository":"web","Tag":"v1"},{"Repository":"web","Tag":"v1"}]`,
			want:  []string{"web:v1"},
		},
		{
			name:  "JSON lines skips malformed lines",
			input: "not-json\n" + `{"Repository":"good","Tag":"tag"}`,
			want:  []string{"good:tag"},
		},
		{
			name:  "entry with empty repo and tag skipped",
			input: `[{"Repository":"","Tag":""}]`,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseComposeImagesOutput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("length mismatch: got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExtractImageTags(t *testing.T) {
	tests := []struct {
		name    string
		entries []composeImageEntry
		want    []string
	}{
		{
			name:    "nil input",
			entries: nil,
			want:    nil,
		},
		{
			name:    "empty slice",
			entries: []composeImageEntry{},
			want:    nil,
		},
		{
			name: "single entry",
			entries: []composeImageEntry{
				{Repository: "myapp", Tag: "latest"},
			},
			want: []string{"myapp:latest"},
		},
		{
			name: "deduplication",
			entries: []composeImageEntry{
				{Repository: "myapp", Tag: "latest"},
				{Repository: "myapp", Tag: "latest"},
			},
			want: []string{"myapp:latest"},
		},
		{
			name: "empty repo:tag skipped",
			entries: []composeImageEntry{
				{Repository: "", Tag: ""},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractImageTags(tt.entries)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// generateComposeLabelsOverrideYAML (compose.go)
// ---------------------------------------------------------------------------

func TestGenerateComposeLabelsOverrideYAML(t *testing.T) {
	tests := []struct {
		name         string
		serviceNames []string
		labels       map[string]string
		checkResult  func(t *testing.T, output string)
	}{
		{
			name:         "no services",
			serviceNames: []string{},
			labels:       map[string]string{"key": "val"},
			checkResult: func(t *testing.T, output string) {
				if output != "services:\n" {
					t.Errorf("expected 'services:\\n', got %q", output)
				}
			},
		},
		{
			name:         "single service single label",
			serviceNames: []string{"web"},
			labels:       map[string]string{"traefik.enable": "true"},
			checkResult: func(t *testing.T, output string) {
				if len(output) == 0 {
					t.Fatal("output should not be empty")
				}
				for _, substr := range []string{"services:", "web:", "labels:", "traefik.enable", "true"} {
					if !strings.Contains(output, substr) {
						t.Errorf("output missing %q; got:\n%s", substr, output)
					}
				}
			},
		},
		{
			name:         "labels sorted alphabetically",
			serviceNames: []string{"app"},
			labels:       map[string]string{"z_label": "z", "a_label": "a"},
			checkResult: func(t *testing.T, output string) {
				aIdx := strings.Index(output, "a_label")
				zIdx := strings.Index(output, "z_label")
				if aIdx == -1 || zIdx == -1 {
					t.Fatalf("expected both labels present; got:\n%s", output)
				}
				if aIdx > zIdx {
					t.Errorf("expected a_label before z_label; got:\n%s", output)
				}
			},
		},
		{
			name:         "quotes escaped in label value",
			serviceNames: []string{"svc"},
			labels:       map[string]string{"rule": `Host("example.com")`},
			checkResult: func(t *testing.T, output string) {
				if !strings.Contains(output, `Host(\"example.com\")`) {
					t.Errorf("expected escaped quotes in output; got:\n%s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := generateComposeLabelsOverrideYAML(tt.serviceNames, tt.labels)
			tt.checkResult(t, string(out))
		})
	}
}

// ---------------------------------------------------------------------------
// filterServers (fanout.go)
// ---------------------------------------------------------------------------

func TestFilterServers(t *testing.T) {
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	id3 := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	all := []shared_types.ApplicationServer{
		{ServerID: id1},
		{ServerID: id2},
		{ServerID: id3},
	}

	tests := []struct {
		name      string
		all       []shared_types.ApplicationServer
		targetIDs []uuid.UUID
		wantLen   int
		wantIDs   []uuid.UUID
	}{
		{
			name:      "empty targetIDs returns all",
			all:       all,
			targetIDs: []uuid.UUID{},
			wantLen:   3,
		},
		{
			name:      "nil targetIDs returns all",
			all:       all,
			targetIDs: nil,
			wantLen:   3,
		},
		{
			name:      "single target filter",
			all:       all,
			targetIDs: []uuid.UUID{id2},
			wantLen:   1,
			wantIDs:   []uuid.UUID{id2},
		},
		{
			name:      "multiple target filter",
			all:       all,
			targetIDs: []uuid.UUID{id1, id3},
			wantLen:   2,
			wantIDs:   []uuid.UUID{id1, id3},
		},
		{
			name:      "target not in all returns empty",
			all:       all,
			targetIDs: []uuid.UUID{uuid.MustParse("00000000-0000-0000-0000-000000000099")},
			wantLen:   0,
		},
		{
			name:      "empty all returns empty",
			all:       []shared_types.ApplicationServer{},
			targetIDs: []uuid.UUID{id1},
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterServers(tt.all, tt.targetIDs)
			if len(got) != tt.wantLen {
				t.Fatalf("filterServers: got %d servers, want %d; result=%v", len(got), tt.wantLen, got)
			}
			for _, wantID := range tt.wantIDs {
				found := false
				for _, s := range got {
					if s.ServerID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected server %v in result, but not found", wantID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// shortHash (build.go)
// ---------------------------------------------------------------------------

func TestShortHash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdef1234567890", "abcdef12"},
		{"abcdefgh", "abcdefgh"},
		{"abc", "abc"},
		{"", ""},
		{"12345678extra", "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := shortHash(tt.input)
			if got != tt.want {
				t.Errorf("shortHash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// prepareBuildArgs / prepareLabels (build.go) - no infra calls
// ---------------------------------------------------------------------------

func TestPrepareBuildArgs(t *testing.T) {
	svc := &TaskService{}

	tests := []struct {
		name       string
		buildVars  string
		wantKeys   []string
		wantValues map[string]string
	}{
		{
			name:       "empty build variables",
			buildVars:  "",
			wantKeys:   []string{},
			wantValues: map[string]string{},
		},
		{
			name:       "JSON format build variables",
			buildVars:  `{"NODE_ENV":"production","DEBUG":"false"}`,
			wantKeys:   []string{"NODE_ENV", "DEBUG"},
			wantValues: map[string]string{"NODE_ENV": "production", "DEBUG": "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := BuildConfig{
				TaskPayload: shared_types.TaskPayload{
					Application: shared_types.Application{
						BuildVariables: tt.buildVars,
					},
				},
			}
			got := svc.prepareBuildArgs(cfg)
			for _, k := range tt.wantKeys {
				ptr, ok := got[k]
				if !ok {
					t.Errorf("expected key %q in buildArgs", k)
					continue
				}
				if *ptr != tt.wantValues[k] {
					t.Errorf("buildArgs[%q] = %q, want %q", k, *ptr, tt.wantValues[k])
				}
			}
			if len(got) != len(tt.wantKeys) {
				t.Errorf("buildArgs length: got %d, want %d", len(got), len(tt.wantKeys))
			}
		})
	}
}

func TestPrepareLabels(t *testing.T) {
	svc := &TaskService{}

	appID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	depID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	userID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	cfg := BuildConfig{
		TaskPayload: shared_types.TaskPayload{
			Application: shared_types.Application{
				ID:     appID,
				Name:   "myapp",
				UserID: userID,
			},
			ApplicationDeployment: shared_types.ApplicationDeployment{
				ID:         depID,
				CommitHash: "deadbeef",
			},
		},
	}

	labels := svc.prepareLabels(cfg)

	checks := map[string]string{
		"com.application.id":   appID.String(),
		"com.application.name": "myapp",
		"com.deployment.id":    depID.String(),
		"com.commit_hash":      "deadbeef",
		"com.user_id":          userID.String(),
	}

	for k, want := range checks {
		if got := labels[k]; got != want {
			t.Errorf("labels[%q] = %q, want %q", k, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// GetApplicationData / GetDeploymentConfig / mergeDeploymentUpdates (context.go)
// ---------------------------------------------------------------------------

func TestGetApplicationData(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	orgID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	ct := &ContextTask{
		UserId:         userID,
		OrganizationId: orgID,
	}

	req := &deploy_types.CreateDeploymentRequest{
		Name:        "my-app",
		Environment: "production",
		BuildPack:   shared_types.DockerFile,
		Repository:  "github.com/org/repo",
		Branch:      "main",
		Port:        3000,
		Source:      shared_types.SourceGithub,
	}

	now := time.Now()
	app := ct.GetApplicationData(req, &now)

	if app.Name != "my-app" {
		t.Errorf("Name: got %q, want %q", app.Name, "my-app")
	}
	if app.UserID != userID {
		t.Errorf("UserID mismatch")
	}
	if app.OrganizationID != orgID {
		t.Errorf("OrganizationID mismatch")
	}
	if app.Port != 3000 {
		t.Errorf("Port: got %d, want 3000", app.Port)
	}
	if app.Source != shared_types.SourceGithub {
		t.Errorf("Source: got %q, want %q", app.Source, shared_types.SourceGithub)
	}
	if !app.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", app.CreatedAt, now)
	}
}

func TestGetApplicationData_DefaultSource(t *testing.T) {
	ct := &ContextTask{
		UserId:         uuid.New(),
		OrganizationId: uuid.New(),
	}

	req := &deploy_types.CreateDeploymentRequest{
		Name:        "app",
		Environment: "production",
		BuildPack:   shared_types.DockerFile,
		Repository:  "repo",
		Branch:      "main",
		Port:        8080,
		Source:      "",
	}

	app := ct.GetApplicationData(req, nil)
	if app.Source != shared_types.SourceGithub {
		t.Errorf("expected default source %q, got %q", shared_types.SourceGithub, app.Source)
	}
}

func TestGetDeploymentConfig(t *testing.T) {
	ct := &ContextTask{}
	appID := uuid.New()

	dep := ct.GetDeploymentConfig(appID)

	if dep.ApplicationID != appID {
		t.Errorf("ApplicationID: got %v, want %v", dep.ApplicationID, appID)
	}
	if dep.ID == uuid.Nil {
		t.Error("ID should not be nil UUID")
	}
	if dep.ContainerStatus != "" {
		t.Errorf("ContainerStatus should be empty, got %q", dep.ContainerStatus)
	}
	if dep.CommitHash != "" {
		t.Errorf("CommitHash should be empty, got %q", dep.CommitHash)
	}
}

func TestMergeDeploymentUpdates(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	orgID := uuid.New()

	original := &shared_types.Application{
		ID:             appID,
		Name:           "original-name",
		Environment:    "production",
		BuildPack:      shared_types.DockerFile,
		Port:           3000,
		DockerfilePath: "Dockerfile",
		UserID:         userID,
		OrganizationID: orgID,
	}

	updateReq := &deploy_types.UpdateDeploymentRequest{
		Name:           "updated-name",
		Port:           4000,
		DockerfilePath: "docker/Dockerfile.prod",
		BasePath:       "/app",
	}

	ct := &ContextTask{
		Application:   original,
		ContextConfig: updateReq,
	}

	updated := ct.mergeDeploymentUpdates()

	if updated.Name != "updated-name" {
		t.Errorf("Name: got %q, want %q", updated.Name, "updated-name")
	}
	if updated.Port != 4000 {
		t.Errorf("Port: got %d, want 4000", updated.Port)
	}
	if updated.DockerfilePath != "docker/Dockerfile.prod" {
		t.Errorf("DockerfilePath: got %q, want %q", updated.DockerfilePath, "docker/Dockerfile.prod")
	}
	if updated.BasePath != "/app" {
		t.Errorf("BasePath: got %q, want /app", updated.BasePath)
	}
	if updated.ID != appID {
		t.Errorf("ID should be unchanged")
	}
}

func TestMergeDeploymentUpdates_EmptyDockerfilePath(t *testing.T) {
	original := &shared_types.Application{
		ID:             uuid.New(),
		DockerfilePath: "Dockerfile",
	}

	updateReq := &deploy_types.UpdateDeploymentRequest{
		DockerfilePath: "",
	}

	ct := &ContextTask{
		Application:   original,
		ContextConfig: updateReq,
	}

	updated := ct.mergeDeploymentUpdates()
	if updated.DockerfilePath != "Dockerfile" {
		t.Errorf("empty DockerfilePath update should default to 'Dockerfile', got %q", updated.DockerfilePath)
	}
}

func TestMergeDeploymentUpdates_AllFields(t *testing.T) {
	original := &shared_types.Application{
		ID:             uuid.New(),
		DockerfilePath: "Dockerfile",
	}

	updateReq := &deploy_types.UpdateDeploymentRequest{
		Name:                 "new-name",
		Environment:          "staging",
		BuildPack:            shared_types.DockerCompose,
		BuildVariables:       map[string]string{"NODE_ENV": "prod"},
		EnvironmentVariables: map[string]string{"DEBUG": "false"},
		PreRunCommand:        "npm ci",
		PostRunCommand:       "npm run seed",
		Port:                 5000,
		DockerfilePath:       "docker/Dockerfile",
		BasePath:             "/app",
	}

	ct := &ContextTask{
		Application:   original,
		ContextConfig: updateReq,
	}

	updated := ct.mergeDeploymentUpdates()

	if updated.Name != "new-name" {
		t.Errorf("Name: got %q, want new-name", updated.Name)
	}
	if updated.Environment != "staging" {
		t.Errorf("Environment: got %q, want staging", updated.Environment)
	}
	if updated.BuildPack != shared_types.DockerCompose {
		t.Errorf("BuildPack: got %q, want DockerCompose", updated.BuildPack)
	}
	if updated.PreRunCommand != "npm ci" {
		t.Errorf("PreRunCommand: got %q, want 'npm ci'", updated.PreRunCommand)
	}
	if updated.PostRunCommand != "npm run seed" {
		t.Errorf("PostRunCommand: got %q, want 'npm run seed'", updated.PostRunCommand)
	}
	if updated.Port != 5000 {
		t.Errorf("Port: got %d, want 5000", updated.Port)
	}
}

// Extra edge cases for GetMapFromString to improve branch coverage.
func TestGetMapFromString_EdgeCases(t *testing.T) {
	t.Run("trailing space in legacy format", func(t *testing.T) {
		got := GetMapFromString("KEY=val ")
		if got["KEY"] != "val" {
			t.Errorf("got %q, want val", got["KEY"])
		}
	})

	t.Run("pair without equals sign is ignored", func(t *testing.T) {
		got := GetMapFromString("NOEQUALS")
		if len(got) != 0 {
			t.Errorf("expected empty map for pair without '=', got %v", got)
		}
	})

	t.Run("multiple spaces between pairs", func(t *testing.T) {
		got := GetMapFromString("A=1  B=2")
		if got["A"] != "1" || got["B"] != "2" {
			t.Errorf("got %v, want A=1 B=2", got)
		}
	})
}

// Extra edge case for parseComposeImagesOutput — JSON-lines with blank lines.
func TestParseComposeImagesOutput_BlankLinesInJSONLines(t *testing.T) {
	input := `{"Repository":"img","Tag":"v1"}` + "\n\n" + `{"Repository":"img2","Tag":"v2"}`
	got, err := parseComposeImagesOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %v", got)
	}
}
