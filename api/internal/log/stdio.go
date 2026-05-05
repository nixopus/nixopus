package log

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Printf, Println, and Fatal variants delegate to logrus so output matches the
// global structured formatter (same path as logger.Logger / middleware).

func Printf(format string, args ...interface{}) {
	logrus.Info(fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	logrus.Info(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Infof(format string, args ...interface{}) {
	logrus.Info(fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	logrus.Info(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Warnf(format string, args ...interface{}) {
	logrus.Warn(fmt.Sprintf(format, args...))
}

func Warn(args ...interface{}) {
	logrus.Warn(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Errorf(format string, args ...interface{}) {
	logrus.Error(fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	logrus.Error(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}
