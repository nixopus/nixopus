package logger

import (
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

// Log records a log entry with the specified severity, message, and data
func (l *Logger) Log(sev Severity, msg, data string) {
	l.Severity, l.Message, l.Data = sev, msg, data
	logEntry := l.Message + " " + l.Data

	if l.Env == Development {
		switch l.Severity {
		case Debug:
			logrus.Debug(logEntry)
		case Info:
			logrus.Info(logEntry)
		case Warning:
			logrus.Warn(logEntry)
		case Error:
			logrus.Error(logEntry)
		case Fatal:
			logrus.Fatal(logEntry)
		default:
			logrus.Info(logEntry)
		}
	} else if l.Env == Production {
		// here we probably need to store in a database
		logrus.Info(l.Message)
	}
}
