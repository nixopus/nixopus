package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

func init() {
	logrus.SetReportCaller(true)
	formatter := &logrus.TextFormatter{
		TimestampFormat:        "02-01-2006 15:04:05",
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableColors:          false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf(" %s:%d ", formatFilePath(f.File), f.Line)
		},
		PadLevelText: true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
		},
	}
	logrus.SetFormatter(formatter)
}

func formatFilePath(path string) string {
	arr := strings.Split(path, "/")
	return arr[len(arr)-1]
}
