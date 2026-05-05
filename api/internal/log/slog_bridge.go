package log

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// newSlogLogrusHandler forwards Go's [log/slog] records to logrus so Fuego and other
// libraries using slog emit the same structured lines as the rest of the app.
func newSlogLogrusHandler(groups []string, attrs []slog.Attr) slog.Handler {
	return &slogLogrusHandler{groups: groups, attrs: attrs}
}

type slogLogrusHandler struct {
	groups []string
	attrs  []slog.Attr
}

func (h *slogLogrusHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *slogLogrusHandler) Handle(_ context.Context, r slog.Record) error {
	prefix := ""
	if len(h.groups) > 0 {
		prefix = strings.Join(h.groups, ".") + "."
	}
	fields := logrus.Fields{}
	for _, a := range h.attrs {
		fields[prefix+a.Key] = slogValueToAny(a.Value)
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[prefix+a.Key] = slogValueToAny(a.Value)
		return true
	})

	lvl := slogLevelToLogrusLevel(r.Level)
	logrus.WithFields(fields).WithTime(r.Time).Log(lvl, r.Message)
	return nil
}

func (h *slogLogrusHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := append(append([]slog.Attr{}, h.attrs...), attrs...)
	return newSlogLogrusHandler(h.groups, merged)
}

func (h *slogLogrusHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	g := append(append([]string{}, h.groups...), name)
	return newSlogLogrusHandler(g, h.attrs)
}

func slogLevelToLogrusLevel(l slog.Level) logrus.Level {
	switch {
	case l >= slog.LevelError:
		return logrus.ErrorLevel
	case l >= slog.LevelWarn:
		return logrus.WarnLevel
	case l >= slog.LevelInfo:
		return logrus.InfoLevel
	default:
		return logrus.DebugLevel
	}
}

func slogValueToAny(v slog.Value) interface{} {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339Nano)
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindAny:
		return v.Any()
	case slog.KindGroup:
		return v.String()
	default:
		return v.String()
	}
}
