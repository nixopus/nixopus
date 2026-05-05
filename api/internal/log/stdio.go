package log

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Printf, Println, and Fatal variants delegate to logrus so output matches the
// global structured formatter (same path as logger.Logger / middleware).

// logAt logs a message at the specified level with the actual caller location.
func logAt(level logrus.Level, msg string) {
	if _, file, line, ok := runtime.Caller(2); ok {
		entry := logrus.WithField("caller", formatFilePath(file)+":"+strconv.Itoa(line))
		entry.Log(level, msg)
	} else {
		logrus.NewEntry(logrus.StandardLogger()).Log(level, msg)
	}
}

func Printf(format string, args ...interface{}) {
	logAt(logrus.InfoLevel, fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	logAt(logrus.InfoLevel, strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Infof(format string, args ...interface{}) {
	logAt(logrus.InfoLevel, fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	logAt(logrus.InfoLevel, strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Warnf(format string, args ...interface{}) {
	logAt(logrus.WarnLevel, fmt.Sprintf(format, args...))
}

func Warn(args ...interface{}) {
	logAt(logrus.WarnLevel, strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Errorf(format string, args ...interface{}) {
	logAt(logrus.ErrorLevel, fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	logAt(logrus.ErrorLevel, strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Fatalf(format string, args ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		entry := logrus.WithField("caller", formatFilePath(file)+":"+strconv.Itoa(line))
		entry.Fatalf(format, args...)
	} else {
		logrus.Fatalf(format, args...)
	}
}

func Fatal(args ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		entry := logrus.WithField("caller", formatFilePath(file)+":"+strconv.Itoa(line))
		entry.Fatal(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
	} else {
		logrus.Fatal(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
	}
}
