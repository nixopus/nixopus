package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFilePath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"github.com/x/pkg/internal/foo/bar.go", "pkg/internal/foo/bar.go"},
		{"/Users/dev/nixopus/api/internal/scheduler/machines/init.go", "internal/scheduler/machines/init.go"},
		{"/abs/path/onlyone.go", "abs/path/onlyone.go"},
		{"single.go", "single.go"},
		{"", ""},
		{"noseparator", "noseparator"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.want, formatFilePath(tc.path), "input %q", tc.path)
		})
	}
}

// Importing this test package runs init() (logrus formatter, ReportCaller). Exercising a log
// line drives the pino-style formatter and caller path shortening via formatFilePath.
func TestInitLogrusEmitsPinoJSONWithCaller(t *testing.T) {
	t.Setenv("LOG_COLOR", "false")
	t.Setenv("LOG_PRETTY", "false")
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("init coverage ping")

	line := strings.TrimSpace(buf.String())
	require.NotEmpty(t, line)

	var rec map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(line), &rec))

	require.Equal(t, "init coverage ping", rec["msg"])
	require.Equal(t, float64(30), rec["level"])
	require.Contains(t, rec["caller"].(string), ".go:")
	require.NotNil(t, rec["time"])
	require.IsType(t, float64(0), rec["time"])
	require.NotNil(t, rec["pid"])
	require.NotNil(t, rec["hostname"])
}

func TestEncodeLogRecord_prettyEnv(t *testing.T) {
	t.Setenv("LOG_COLOR", "false")
	t.Setenv("LOG_PRETTY", "true")
	out, err := encodeLogRecord(map[string]interface{}{"level": 30, "msg": "hi"})
	require.NoError(t, err)
	require.Contains(t, string(out), "\n  \"msg\"")

	t.Setenv("LOG_PRETTY", "false")
	out, err = encodeLogRecord(map[string]interface{}{"level": 30, "msg": "hi"})
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &m))
	require.Equal(t, float64(30), m["level"])
	require.Equal(t, "hi", m["msg"])
}

func TestFormatLogTime_prettyVsCompact(t *testing.T) {
	ts := time.Unix(1700000000, 123456789).UTC()

	t.Setenv("LOG_PRETTY", "true")
	out := formatLogTime(ts)
	s, ok := out.(string)
	require.True(t, ok)
	_, err := time.Parse(time.RFC3339Nano, s)
	require.NoError(t, err)

	t.Setenv("LOG_PRETTY", "false")
	out = formatLogTime(ts)
	n, ok := out.(int64)
	require.True(t, ok)
	require.Equal(t, ts.UnixMilli(), n)

	t.Setenv("LOG_TIME", "rfc3339")
	t.Setenv("LOG_PRETTY", "false")
	out = formatLogTime(ts)
	s2, ok := out.(string)
	require.True(t, ok)
	_, err = time.Parse(time.RFC3339Nano, s2)
	require.NoError(t, err)
}

func TestAppendLevelBadge_toggleAndNoColor(t *testing.T) {
	payload := []byte(`{"level":30}`)

	t.Setenv("NO_COLOR", "")
	t.Setenv("LOG_COLOR", "true")
	out := appendLevelBadge(nil, 30, payload)
	s := string(out)
	require.Contains(t, s, "\033[32mINF\033[0m ")
	require.True(t, strings.HasSuffix(s, `{"level":30}`))

	t.Setenv("LOG_COLOR", "false")
	require.Equal(t, string(payload), string(appendLevelBadge(nil, 30, payload)))

	t.Setenv("LOG_COLOR", "true")
	t.Setenv("NO_COLOR", "1")
	require.Equal(t, string(payload), string(appendLevelBadge(nil, 30, payload)))
}
