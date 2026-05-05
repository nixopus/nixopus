package log

import (
	"bytes"
	"encoding/json"
	stdlog "log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

var (
	pinoPID      int
	pinoHostname string
)

func init() {
	var err error
	pinoHostname, err = os.Hostname()
	if err != nil {
		pinoHostname = "unknown"
	}
	pinoPID = os.Getpid()

	logrus.SetReportCaller(true)
	logrus.SetFormatter(&pinoFormatter{})

	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
	stdlog.SetOutput(&stdlibLogWriter{})

	slog.SetDefault(slog.New(newSlogLogrusHandler(nil, nil)))
}

type pinoFormatter struct{}

func (f *pinoFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	rec := make(map[string]interface{}, len(entry.Data)+8)
	for k, v := range entry.Data {
		rec[k] = v
	}

	rec["level"] = pinoLevel(entry.Level)
	rec["time"] = formatLogTime(entry.Time)
	rec["pid"] = pinoPID
	rec["hostname"] = pinoHostname
	rec["msg"] = entry.Message

	// Prefer explicit "caller" field from entry.Data (set by stdio.go wrappers)
	// Otherwise fall back to entry.Caller if HasCaller() is true
	if _, hasExplicitCaller := rec["caller"]; !hasExplicitCaller && entry.HasCaller() {
		rec["caller"] = formatFilePath(entry.Caller.File) + ":" + strconv.Itoa(entry.Caller.Line)
	}

	b, err := encodeLogRecord(rec)
	if err != nil {
		return nil, err
	}
	line := appendLevelBadge(nil, rec["level"], b)
	return append(line, '\n'), nil
}

func pinoLevel(l logrus.Level) int {
	switch l {
	case logrus.TraceLevel:
		return 10
	case logrus.DebugLevel:
		return 20
	case logrus.InfoLevel:
		return 30
	case logrus.WarnLevel:
		return 40
	case logrus.ErrorLevel:
		return 50
	case logrus.FatalLevel, logrus.PanicLevel:
		return 60
	default:
		return 30
	}
}

type stdlibLogWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (w *stdlibLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	nn, err := w.buf.Write(p)
	n += nn
	for {
		data := w.buf.Bytes()
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			break
		}
		line := bytes.TrimSpace(data[:idx])
		w.buf.Next(idx + 1)
		if len(line) == 0 {
			continue
		}
		emitStdlibPinoLine(string(line))
	}
	return n, err
}

func emitStdlibPinoLine(msg string) {
	lvl := 30
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "warning") || strings.Contains(lower, "warn:") {
		lvl = 40
	}

	rec := map[string]interface{}{
		"level":    lvl,
		"time":     formatLogTime(time.Now()),
		"pid":      pinoPID,
		"hostname": pinoHostname,
		"msg":      msg,
	}
	b, err := encodeLogRecord(rec)
	if err != nil {
		return
	}
	line := appendLevelBadge(nil, rec["level"], b)
	_, _ = os.Stderr.Write(append(line, '\n'))
}

// prettyLogOutput selects indented JSON for humans. Override with LOG_PRETTY:
// 1/true/yes/on or 0/false/no/off. When unset, uses a TTY on stderr (local dev)
// vs non-TTY (Docker/CI/pipes) for compact single-line records.
func prettyLogOutput() bool {
	switch strings.TrimSpace(strings.ToLower(os.Getenv("LOG_PRETTY"))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return term.IsTerminal(int(os.Stderr.Fd()))
	}
}

// formatLogTime chooses a timestamp encoding: RFC3339Nano when logs are pretty-printed
// for the terminal, or when LOG_TIME=rfc3339|iso; Unix milliseconds otherwise (Pino/Loki-friendly).
func formatLogTime(t time.Time) interface{} {
	switch strings.TrimSpace(strings.ToLower(os.Getenv("LOG_TIME"))) {
	case "rfc3339", "iso", "iso8601":
		return t.UTC().Format(time.RFC3339Nano)
	case "unix_ms", "millis", "epoch_ms":
		return t.UnixMilli()
	}
	if prettyLogOutput() {
		return t.Local().Format(time.RFC3339Nano)
	}
	return t.UnixMilli()
}

func encodeLogRecord(rec map[string]interface{}) ([]byte, error) {
	if prettyLogOutput() {
		return json.MarshalIndent(rec, "", "  ")
	}
	return json.Marshal(rec)
}

const ansiReset = "\033[0m"

// logColorsEnabled adds a short level badge before JSON when stderr is a color-capable TTY.
// LOG_COLOR=1/true/on or 0/false/off overrides; NO_COLOR (any non-empty value) disables.
func logColorsEnabled() bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	switch strings.TrimSpace(strings.ToLower(os.Getenv("LOG_COLOR"))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	if strings.TrimSpace(os.Getenv("TERM")) == "dumb" {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func appendLevelBadge(dst []byte, level interface{}, jsonBody []byte) []byte {
	if !logColorsEnabled() {
		return append(dst, jsonBody...)
	}
	lvl, ok := coercePinoLevel(level)
	if !ok {
		return append(dst, jsonBody...)
	}
	label, color := badgeStyleForPinoLevel(lvl)
	dst = append(dst, color...)
	dst = append(dst, label...)
	dst = append(dst, ansiReset...)
	dst = append(dst, ' ')
	return append(dst, jsonBody...)
}

func coercePinoLevel(level interface{}) (int, bool) {
	switch v := level.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func badgeStyleForPinoLevel(n int) (label string, ansiPrefix string) {
	switch n {
	case 10:
		return "TRC", "\033[90m"
	case 20:
		return "DBG", "\033[36m"
	case 30:
		return "INF", "\033[32m"
	case 40:
		return "WRN", "\033[33m"
	case 50:
		return "ERR", "\033[31m"
	case 60:
		return "FTL", "\033[1;31m"
	default:
		return "???", "\033[35m"
	}
}

// formatFilePath returns a short caller path that still identifies the file.
// Many packages use init.go; basename-only callers are ambiguous, so we prefer
// the path under …/api/ (e.g. internal/scheduler/machines/init.go) or the last
// few slash-separated segments on other layouts.
func formatFilePath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	if path == "" {
		return ""
	}
	const apiMarker = "/api/"
	if i := strings.LastIndex(path, apiMarker); i >= 0 {
		return path[i+len(apiMarker):]
	}
	segs := strings.Split(path, "/")
	var clean []string
	for _, s := range segs {
		if s != "" {
			clean = append(clean, s)
		}
	}
	if len(clean) == 0 {
		return path
	}
	const maxSegs = 4
	if len(clean) <= maxSegs {
		return strings.Join(clean, "/")
	}
	return strings.Join(clean[len(clean)-maxSegs:], "/")
}
