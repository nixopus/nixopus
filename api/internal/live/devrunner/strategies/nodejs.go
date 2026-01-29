package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/raghavyuva/nixopus-api/internal/live/sftp"
)

type NodeJSStrategy struct {
	UseTypescript bool
	Framework     string
	HasNodemon    bool
	MainFile      string
	CustomPort    int
}

func NewNodeJSStrategy() *NodeJSStrategy {
	return &NodeJSStrategy{
		UseTypescript: false,
		Framework:     "node",
		HasNodemon:    false,
		MainFile:      "index.js",
		CustomPort:    0,
	}
}

func NewExpressStrategy() *NodeJSStrategy {
	return &NodeJSStrategy{
		UseTypescript: false,
		Framework:     "express",
		HasNodemon:    false,
		MainFile:      "index.js",
		CustomPort:    0,
	}
}

func NewFastifyStrategy() *NodeJSStrategy {
	return &NodeJSStrategy{
		UseTypescript: false,
		Framework:     "fastify",
		HasNodemon:    false,
		MainFile:      "index.js",
		CustomPort:    0,
	}
}

func (s *NodeJSStrategy) Name() string {
	return s.Framework
}

func (s *NodeJSStrategy) GetBaseImage() string {
	return "node:20-alpine"
}

func (s *NodeJSStrategy) GetInstallCommand() []string {
	return []string{"sh", "-c", "if [ -f package-lock.json ]; then npm ci; else npm install; fi"}
}

func (s *NodeJSStrategy) GetDevCommand() []string {
	if s.HasNodemon {
		if s.UseTypescript {
			return []string{"npx", "nodemon", "--exec", "ts-node", s.MainFile}
		}
		return []string{"npx", "nodemon", s.MainFile}
	}

	cmd := "if npm run dev --if-present 2>/dev/null; then exit 0; " +
		"elif npm start --if-present 2>/dev/null; then exit 0; " +
		"else node " + s.MainFile + "; fi"

	return []string{"sh", "-c", cmd}
}

func (s *NodeJSStrategy) GetDefaultPort() int {
	if s.CustomPort > 0 {
		return s.CustomPort
	}
	return 3000
}

func (s *NodeJSStrategy) GetEnvVars() map[string]string {
	return map[string]string{
		"NODE_ENV":            "development",
		"CHOKIDAR_USEPOLLING": "true",
		"HOST":                "0.0.0.0",
	}
}

func (s *NodeJSStrategy) GetHealthCheckPath() string {
	return "/"
}

func (s *NodeJSStrategy) NeedsPolling() bool {
	return true
}

func (s *NodeJSStrategy) GetWorkdir() string {
	return "/app"
}

func (s *NodeJSStrategy) GetReadyLogPattern() string {
	return `(?i)(listening on|server (is )?(running|started|listening)|port \d+|ready|started)`
}

// DetectMainFile determines the main entry file from package.json.
func (s *NodeJSStrategy) DetectMainFile(ctx context.Context, projectPath string) string {
	pkg, err := readPackageJSON(ctx, projectPath)
	if err != nil {
		// Fallback to default based on TypeScript usage
		if s.UseTypescript {
			return "index.ts"
		}
		return "index.js"
	}

	if pkg.Main != "" {
		return pkg.Main
	}

	// TypeScript vs JavaScript defaults
	if s.UseTypescript || hasPackage(pkg, "typescript") {
		return "index.ts"
	}
	return "index.js"
}

// packageJSON represents the structure of a package.json file.
type packageJSON struct {
	Main            string            `json:"main"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// readPackageJSON reads and parses package.json from the project path via SFTP.
func readPackageJSON(ctx context.Context, projectPath string) (packageJSON, error) {
	var pkg packageJSON
	filePath := filepath.Join(projectPath, "package.json")
	data, err := sftp.ReadFile(ctx, filePath)
	if err != nil {
		return pkg, fmt.Errorf("failed to read package.json via SFTP: %w", err)
	}
	if err := json.Unmarshal([]byte(data), &pkg); err != nil {
		return pkg, fmt.Errorf("failed to parse package.json: %w", err)
	}
	return pkg, nil
}

// hasPackage checks if a package is in dependencies or devDependencies.
func hasPackage(pkg packageJSON, name string) bool {
	_, inDeps := pkg.Dependencies[name]
	_, inDevDeps := pkg.DevDependencies[name]
	return inDeps || inDevDeps
}
