package logger

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/sirupsen/logrus"
)

// Environment represents the operating environment
type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

// Severity represents log severity levels
type Severity string

const (
	Debug   Severity = "DEBUG"
	Info    Severity = "INFO"
	Warning Severity = "WARNING"
	Error   Severity = "ERROR"
	Fatal   Severity = "FATAL"
)

// Logger holds information for logging
type Logger struct {
	Severity Severity
	Message  string
	Env      Environment
	Data     string
}

// NewLogger creates a new logger with default values
func NewLogger() Logger {
	return Logger{
		Severity: Info,
		Message:  "",
		Env:      Development,
		Data:     "",
	}
}

// Log records a log entry with the specified severity, message, and data.
func (l *Logger) Log(sev Severity, msg, data string) {
	l.logWithEntry(logrus.NewEntry(logrus.StandardLogger()), sev, msg, data)
}

// LogCtx records a log entry and automatically attaches the request_id from
// the context set by RequestIDMiddleware. Use this instead of Log in any code
// that has access to a request context (handlers, middleware, services called
// from a request path).
func (l *Logger) LogCtx(ctx context.Context, sev Severity, msg, data string) {
	entry := logrus.NewEntry(logrus.StandardLogger())
	if ctx != nil {
		if id, ok := ctx.Value(types.RequestIDKey).(string); ok && id != "" {
			entry = entry.WithField("request_id", id)
		}
	}
	l.logWithEntry(entry, sev, msg, data)
}

func (l *Logger) logWithEntry(entry *logrus.Entry, sev Severity, msg, data string) {
	l.Severity, l.Message, l.Data = sev, msg, data
	logEntry := l.Message + " " + l.Data

	if l.Env == Development {
		switch sev {
		case Debug:
			entry.Debug(logEntry)
		case Info:
			entry.Info(logEntry)
		case Warning:
			entry.Warn(logEntry)
		case Error:
			entry.Error(logEntry)
		case Fatal:
			entry.Fatal(logEntry)
		default:
			entry.Info(logEntry)
		}
	} else if l.Env == Production {
		entry.Info(l.Message)
	}
}
