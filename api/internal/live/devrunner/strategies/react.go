package strategies

import "context"

type ReactStrategy struct {
	UseTypescript bool

	UsesVite bool
}

func NewReactStrategy() *ReactStrategy {
	return &ReactStrategy{
		UseTypescript: false,
		UsesVite:      false,
	}
}

func NewViteReactStrategy() *ReactStrategy {
	return &ReactStrategy{
		UseTypescript: true,
		UsesVite:      true,
	}
}

func (s *ReactStrategy) Name() string {
	if s.UsesVite {
		return "vite-react"
	}
	return "react"
}

func (s *ReactStrategy) GetBaseImage() string {
	return "node:20-alpine"
}

func (s *ReactStrategy) GetInstallCommand() []string {
	return []string{"sh", "-c", "if [ -f package-lock.json ]; then npm ci; else npm install; fi"}
}

func (s *ReactStrategy) GetDevCommand() []string {
	if s.UsesVite {
		return []string{"npx", "vite", "--host", "0.0.0.0"}
	}
	return []string{"npm", "start"}
}

func (s *ReactStrategy) GetDefaultPort() int {
	if s.UsesVite {
		return 5173
	}
	return 3000
}

func (s *ReactStrategy) GetEnvVars() map[string]string {
	if s.UsesVite {
		return map[string]string{
			"NODE_ENV":            "development",
			"CHOKIDAR_USEPOLLING": "true",
		}
	}

	return map[string]string{
		"NODE_ENV":            "development",
		"HOST":                "0.0.0.0",
		"BROWSER":             "none",
		"CHOKIDAR_USEPOLLING": "true",
		"WATCHPACK_POLLING":   "true",
		"GENERATE_SOURCEMAP":  "false",
	}
}

func (s *ReactStrategy) GetHealthCheckPath() string {
	return "/"
}

func (s *ReactStrategy) NeedsPolling() bool {
	return true
}

func (s *ReactStrategy) GetWorkdir() string {
	return "/app"
}

func (s *ReactStrategy) GetReadyLogPattern() string {
	if s.UsesVite {
		return `(?i)(Local:\s+http|ready in \d+)`
	}
	return `(?i)(Compiled successfully|You can now view|Local:\s+http|Starting the development server)`
}

// DetectMainFile returns empty string as React uses npm scripts and doesn't require a main file.
func (s *ReactStrategy) DetectMainFile(ctx context.Context, projectPath string) string {
	return ""
}
