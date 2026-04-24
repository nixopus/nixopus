package log

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFilePath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"github.com/x/pkg/internal/foo/bar.go", "bar.go"},
		{"/abs/path/onlyone.go", "onlyone.go"},
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
// line drives the [logrus.TextFormatter] caller prettyfier that references formatFilePath.
func TestInitLogrusEmitsWithCaller(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("init coverage ping")

	s := buf.String()
	require.Contains(t, s, "init coverage ping")
	// file:line from CallerPrettyfier
	require.Contains(t, s, ".go:")
	// timestamp from FullTimestamp
	require.Contains(t, s, "time=")
}
