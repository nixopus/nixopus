package live

import (
	"net/http"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/config"
)

// Live config accessors read from config.AppConfig.Live (populated by config.Init at startup).
// Durations are parsed on first access; string fields are used as-is.

func chunkSize() int64 {
	c := config.AppConfig.Live
	if c.ChunkSize > 0 {
		return c.ChunkSize
	}
	return 64 * 1024
}

func completionWorkers() int {
	c := config.AppConfig.Live
	if c.CompletionWorkers > 0 {
		return c.CompletionWorkers
	}
	return 4
}

func completionBuffer() int {
	c := config.AppConfig.Live
	if c.CompletionBuffer > 0 {
		return c.CompletionBuffer
	}
	return 256
}

func pendingCompletionsTick() time.Duration {
	c := config.AppConfig.Live
	if c.PendingCompletionsTick != "" {
		if d, err := time.ParseDuration(c.PendingCompletionsTick); err == nil && d > 0 {
			return d
		}
	}
	return 10 * time.Millisecond
}

func readBufferSize() int {
	c := config.AppConfig.Live
	if c.ReadBufferSize > 0 {
		return c.ReadBufferSize
	}
	return 256 * 1024
}

func writeBufferSize() int {
	c := config.AppConfig.Live
	if c.WriteBufferSize > 0 {
		return c.WriteBufferSize
	}
	return 256 * 1024
}

func readDeadline() time.Duration {
	c := config.AppConfig.Live
	if c.ReadDeadline != "" {
		if d, err := time.ParseDuration(c.ReadDeadline); err == nil && d > 0 {
			return d
		}
	}
	return 5 * time.Minute
}

func writeDeadline() time.Duration {
	c := config.AppConfig.Live
	if c.WriteDeadline != "" {
		if d, err := time.ParseDuration(c.WriteDeadline); err == nil && d > 0 {
			return d
		}
	}
	return 10 * time.Second
}

func fileInjectRetries() int {
	c := config.AppConfig.Live
	if c.FileInjectRetries > 0 {
		return c.FileInjectRetries
	}
	return 3
}

func buildDebounce() time.Duration {
	c := config.AppConfig.Live
	if c.BuildDebounce != "" {
		if d, err := time.ParseDuration(c.BuildDebounce); err == nil && d > 0 {
			return d
		}
	}
	return 3 * time.Second
}

func generatedDockerfileName() string {
	c := config.AppConfig.Live
	if c.GeneratedDockerfileName != "" {
		return c.GeneratedDockerfileName
	}
	return "Dockerfile.nixopus.dev"
}

func maxIndexableSize() int {
	c := config.AppConfig.Live
	if c.MaxIndexableSize > 0 {
		return c.MaxIndexableSize
	}
	return 512000
}

func allowedOriginsSlice() []string {
	c := config.AppConfig.Live
	if c.AllowedOrigins == "" {
		return nil
	}
	var out []string
	for _, o := range strings.Split(c.AllowedOrigins, ",") {
		if o := strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

func checkOriginFunc() func(*http.Request) bool {
	if !config.AppConfig.Live.CheckOrigin {
		return func(*http.Request) bool { return true }
	}
	allowed := allowedOriginsSlice()
	if len(allowed) == 0 {
		return func(*http.Request) bool { return true }
	}
	allowedMap := make(map[string]bool)
	for _, o := range allowed {
		allowedMap[o] = true
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin != "" && allowedMap[origin]
	}
}
