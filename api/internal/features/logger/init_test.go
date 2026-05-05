package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	_ "github.com/nixopus/nixopus/api/internal/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureLogrus(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	std := logrus.StandardLogger()
	prevOut := std.Out
	prevLevel := std.Level
	prevExit := std.ExitFunc
	std.SetOutput(&buf)
	std.SetLevel(logrus.TraceLevel)
	std.ExitFunc = func(int) {}
	t.Cleanup(func() {
		std.SetOutput(prevOut)
		std.SetLevel(prevLevel)
		std.ExitFunc = prevExit
	})
	return &buf
}

func firstJSONRecord(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	require.NotEmpty(t, buf.String())
	var m map[string]interface{}
	decoder := json.NewDecoder(buf)
	require.NoError(t, decoder.Decode(&m))
	return m
}

func TestNewLogger(t *testing.T) {
	l := NewLogger()
	assert.Equal(t, Info, l.Severity)
	assert.Equal(t, Development, l.Env)
	assert.Empty(t, l.Message)
	assert.Empty(t, l.Data)
}

func TestLog_development_severities(t *testing.T) {
	sevs := []struct {
		sev       Severity
		wantLevel float64
	}{
		{Debug, 20},
		{Info, 30},
		{Warning, 40},
		{Error, 50},
		{Severity("OTHER"), 30},
	}
	for _, tc := range sevs {
		t.Run(string(tc.sev), func(t *testing.T) {
			buf := captureLogrus(t)
			l := NewLogger()
			l.Env = Development
			l.Log(tc.sev, "hello", "meta")
			assert.Equal(t, tc.sev, l.Severity)
			assert.Equal(t, "hello", l.Message)
			assert.Equal(t, "meta", l.Data)
			m := firstJSONRecord(t, buf)
			require.Equal(t, tc.wantLevel, m["level"])
			require.Contains(t, m["msg"].(string), "hello meta")
		})
	}
}

func TestLog_development_fatal(t *testing.T) {
	buf := captureLogrus(t)
	l := NewLogger()
	l.Env = Development
	l.Log(Fatal, "bye", "ctx")
	assert.Equal(t, Fatal, l.Severity)
	m := firstJSONRecord(t, buf)
	require.Equal(t, float64(60), m["level"])
	require.Contains(t, m["msg"].(string), "bye ctx")
}

func TestLog_production(t *testing.T) {
	buf := captureLogrus(t)
	l := NewLogger()
	l.Env = Production
	l.Log(Error, "prodmsg", "ignored")
	assert.Equal(t, Error, l.Severity)
	m := firstJSONRecord(t, buf)
	require.Equal(t, float64(30), m["level"])
	require.Contains(t, m["msg"].(string), "prodmsg")
	require.NotContains(t, m["msg"].(string), "ignored")
}

func TestLog_unknownEnvironment(t *testing.T) {
	buf := captureLogrus(t)
	l := NewLogger()
	l.Env = Environment("staging")
	l.Log(Info, "nope", "data")
	assert.Equal(t, Info, l.Severity)
	assert.Empty(t, buf.String())
}

func TestLog_development_emptyMessageAndData(t *testing.T) {
	buf := captureLogrus(t)
	l := NewLogger()
	l.Env = Development
	l.Log(Info, "", "")
	m := firstJSONRecord(t, buf)
	require.Equal(t, float64(30), m["level"])
	require.Contains(t, m["msg"].(string), " ")
}
