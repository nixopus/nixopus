package storage

import (
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func storageLog(l *logger.Logger, sev logger.Severity, msg, data string) {
	if l == nil {
		return
	}
	l.Log(sev, msg, data)
}
