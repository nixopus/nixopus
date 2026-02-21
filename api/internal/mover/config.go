package mover

import (
	"sync"
	"time"
)

// MoverConfig holds CLI-side mover configuration (WebSocket client, sync engine, watcher).
// Set via Configure() from build-time cliconfig; nil/unset fields use defaults.
type MoverConfig struct {
	// WebSocket client
	SendBufferSize        int
	ReceiveBufferSize     int
	WriteWait             time.Duration
	PongWait              time.Duration
	PingPeriod            time.Duration
	MaxMessageSize        int64
	HandshakeTimeout      time.Duration
	InitialReconnectDelay time.Duration
	MaxReconnectDelay     time.Duration
	ReconnectBackoffRate  float64
	MaxReconnectAttempts  int
	CloseFlushDelay       time.Duration

	// Sync engine
	DebounceMs          int
	LargeSyncThreshold  int
	ChunkSize           int
	ManifestWaitTimeout time.Duration
	SyncWorkers         int
	SyncConcurrency     int

	// Watcher
	EventsBufferSize  int
	WatcherDebounceMs int
	GitCheckTimeout   time.Duration

	// Sync state
	SyncStateDebounceMs int
}

var (
	moverCfg   MoverConfig
	moverCfgMu sync.RWMutex
)

// Configure sets the mover configuration (call from CLI main/runSingleApp before creating Client/Engine).
func Configure(cfg MoverConfig) {
	moverCfgMu.Lock()
	moverCfg = cfg
	moverCfgMu.Unlock()
}

func getMoverCfg() MoverConfig {
	moverCfgMu.RLock()
	defer moverCfgMu.RUnlock()
	return moverCfg
}

// Default values when not configured
const (
	defaultSendBufferSize        = 8192
	defaultReceiveBufferSize     = 1024
	defaultWriteWait             = 10 * time.Second
	defaultPongWait              = 60 * time.Second
	defaultPingPeriod            = 25 * time.Second
	defaultMaxMessageSize        = 64 * 1024 * 1024 // 64MB
	defaultHandshakeTimeout      = 60 * time.Second
	defaultInitialReconnectDelay = 1 * time.Second
	defaultMaxReconnectDelay     = 30 * time.Second
	defaultReconnectBackoffRate  = 2.0
	defaultCloseFlushDelay       = 100 * time.Millisecond

	defaultDebounceMs          = 100
	defaultLargeSyncThreshold  = 50
	defaultChunkSize           = 64 * 1024
	defaultManifestWaitTimeout = 5 * time.Second
	defaultSyncWorkers         = 20
	defaultSyncConcurrency     = 50

	defaultEventsBufferSize  = 512
	defaultWatcherDebounceMs = 100
	defaultGitCheckTimeout   = 5 * time.Second

	defaultSyncStateDebounceMs = 500
)

func sendBufferSize() int {
	if v := getMoverCfg().SendBufferSize; v > 0 {
		return v
	}
	return defaultSendBufferSize
}

func receiveBufferSize() int {
	if v := getMoverCfg().ReceiveBufferSize; v > 0 {
		return v
	}
	return defaultReceiveBufferSize
}

func writeWait() time.Duration {
	if v := getMoverCfg().WriteWait; v > 0 {
		return v
	}
	return defaultWriteWait
}

func pongWait() time.Duration {
	if v := getMoverCfg().PongWait; v > 0 {
		return v
	}
	return defaultPongWait
}

func pingPeriod() time.Duration {
	if v := getMoverCfg().PingPeriod; v > 0 {
		return v
	}
	return defaultPingPeriod
}

func maxMessageSize() int64 {
	if v := getMoverCfg().MaxMessageSize; v > 0 {
		return v
	}
	return defaultMaxMessageSize
}

func handshakeTimeout() time.Duration {
	if v := getMoverCfg().HandshakeTimeout; v > 0 {
		return v
	}
	return defaultHandshakeTimeout
}

func initialReconnectDelay() time.Duration {
	if v := getMoverCfg().InitialReconnectDelay; v > 0 {
		return v
	}
	return defaultInitialReconnectDelay
}

func maxReconnectDelay() time.Duration {
	if v := getMoverCfg().MaxReconnectDelay; v > 0 {
		return v
	}
	return defaultMaxReconnectDelay
}

func reconnectBackoffRate() float64 {
	if v := getMoverCfg().ReconnectBackoffRate; v > 0 {
		return v
	}
	return defaultReconnectBackoffRate
}

func maxReconnectAttempts() int {
	return getMoverCfg().MaxReconnectAttempts
}

func closeFlushDelay() time.Duration {
	if v := getMoverCfg().CloseFlushDelay; v >= 0 {
		return v
	}
	return defaultCloseFlushDelay
}

func getMoverDebounceMs() int {
	if v := getMoverCfg().DebounceMs; v > 0 {
		return v
	}
	return defaultDebounceMs
}

func largeSyncThreshold() int {
	if v := getMoverCfg().LargeSyncThreshold; v > 0 {
		return v
	}
	return defaultLargeSyncThreshold
}

func chunkSize() int {
	if v := getMoverCfg().ChunkSize; v > 0 {
		return v
	}
	return defaultChunkSize
}

func manifestWaitTimeout() time.Duration {
	if v := getMoverCfg().ManifestWaitTimeout; v > 0 {
		return v
	}
	return defaultManifestWaitTimeout
}

func syncWorkers() int {
	if v := getMoverCfg().SyncWorkers; v > 0 {
		return v
	}
	return defaultSyncWorkers
}

func syncConcurrency() int {
	if v := getMoverCfg().SyncConcurrency; v > 0 {
		return v
	}
	return defaultSyncConcurrency
}

func eventsBufferSize() int {
	if v := getMoverCfg().EventsBufferSize; v > 0 {
		return v
	}
	return defaultEventsBufferSize
}

func watcherDebounceMs() int {
	if v := getMoverCfg().WatcherDebounceMs; v > 0 {
		return v
	}
	return defaultWatcherDebounceMs
}

func gitCheckTimeout() time.Duration {
	if v := getMoverCfg().GitCheckTimeout; v > 0 {
		return v
	}
	return defaultGitCheckTimeout
}

func syncStateDebounceMs() int {
	if v := getMoverCfg().SyncStateDebounceMs; v > 0 {
		return v
	}
	return defaultSyncStateDebounceMs
}
