package strategies

import "context"

type NextJSStrategy struct {
	UseTurbopack  bool
	UseTypescript bool
}

func NewNextJSStrategy() *NextJSStrategy {
	return &NextJSStrategy{
		UseTurbopack:  false, // Disabled by default - only enable if Next.js version supports it
		UseTypescript: true,
	}
}

func (s *NextJSStrategy) Name() string {
	return "nextjs"
}

func (s *NextJSStrategy) GetBaseImage() string {
	return "node:20-alpine"
}

func (s *NextJSStrategy) GetInstallCommand() []string {
	return []string{"sh", "-c", "if [ -f package-lock.json ]; then npm ci; else npm install; fi"}
}

func (s *NextJSStrategy) GetDevCommand() []string {
	cmd := []string{"npx", "next", "dev", "-H", "0.0.0.0"}
	if s.UseTurbopack {
		cmd = append(cmd, "--turbopack")
	}
	return cmd
}

func (s *NextJSStrategy) GetDefaultPort() int {
	return 3000
}

func (s *NextJSStrategy) GetEnvVars() map[string]string {
	return map[string]string{
		"NODE_ENV":                "development",
		"NEXT_TELEMETRY_DISABLED": "1",
		"WATCHPACK_POLLING":       "true",
		"NEXT_PRIVATE_STANDALONE": "false",
	}
}

func (s *NextJSStrategy) GetHealthCheckPath() string {
	return "/"
}

func (s *NextJSStrategy) NeedsPolling() bool {
	return true
}

func (s *NextJSStrategy) GetWorkdir() string {
	return "/app"
}

func (s *NextJSStrategy) GetReadyLogPattern() string {
	return `(?i)(ready in \d+|ready - started server|Local:\s+http)`
}

// DetectMainFile returns empty string as Next.js uses npm scripts and doesn't require a main file.
func (s *NextJSStrategy) DetectMainFile(ctx context.Context, projectPath string) string {
	return ""
}
