package logger

import (
	"bytes"
	"strings"
	"testing"

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

func TestNewLogger(t *testing.T) {
	l := NewLogger()
	assert.Equal(t, Info, l.Severity)
	assert.Equal(t, Development, l.Env)
	assert.Empty(t, l.Message)
	assert.Empty(t, l.Data)
}

func TestLog_development_severities(t *testing.T) {
	sevs := []struct {
		sev  Severity
		want string
	}{
		{Debug, "level=debug"},
		{Info, "level=info"},
		{Warning, "level=warning"},
		{Error, "level=error"},
		{Severity("OTHER"), "level=info"},
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
			s := buf.String()
			require.Contains(t, strings.ToLower(s), tc.want)
			require.Contains(t, s, "hello meta")
		})
	}
}

func TestLog_development_fatal(t *testing.T) {
	buf := captureLogrus(t)
	l := NewLogger()
	l.Env = Development
	l.Log(Fatal, "bye", "ctx")
	assert.Equal(t, Fatal, l.Severity)
	s := buf.String()
	require.Contains(t, strings.ToLower(s), "level=fatal")
	require.Contains(t, s, "bye ctx")
}

func TestLog_production(t *testing.T) {
	buf := captureLogrus(t)
	l := NewLogger()
	l.Env = Production
	l.Log(Error, "prodmsg", "ignored")
	assert.Equal(t, Error, l.Severity)
	s := buf.String()
	require.Contains(t, strings.ToLower(s), "level=info")
	require.Contains(t, s, "prodmsg")
	require.NotContains(t, s, "ignored")
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
	s := buf.String()
	require.Contains(t, strings.ToLower(s), "level=info")
	require.Contains(t, s, " ")
}
