package devrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/live/devrunner/strategies"
	"github.com/raghavyuva/nixopus-api/internal/live/sftp"
)

// DetectFramework analyzes the project at the given path and returns
// the appropriate FrameworkStrategy for running its dev server.
func DetectFramework(ctx context.Context, projectPath string) (FrameworkStrategy, error) {
	for _, detector := range projectDetectors {
		if strategy, err := detector(ctx, projectPath); err == nil && strategy != nil {
			return strategy, nil
		}
	}
	return nil, fmt.Errorf("could not detect framework for project at %s", projectPath)
}

// GetStrategyByName returns a framework strategy by its name.
// For Python frameworks, it will use default settings without project-specific detection.
func GetStrategyByName(name string) (FrameworkStrategy, error) {
	factory, exists := strategyRegistry[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("unknown framework: %s", name)
	}
	return factory(), nil
}

// GetStrategyByNameWithPath returns a framework strategy by its name and configures it
// based on the actual project at the given path. This is preferred over GetStrategyByName
// when you have access to the project directory.
func GetStrategyByNameWithPath(ctx context.Context, name, projectPath string) (FrameworkStrategy, error) {
	factory, exists := strategyRegistry[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("unknown framework: %s", name)
	}

	strategy := factory()

	// For Python frameworks, detect the actual main file and dependency manager
	if pythonStrategy, ok := strategy.(*strategies.PythonStrategy); ok {
		pythonStrategy.MainFile = strategy.DetectMainFile(ctx, projectPath)
		depManager := detectPythonDepManager(ctx, projectPath)
		applyPythonDepManager(strategy, depManager)
	}

	// For Node.js frameworks that need main file detection (includes Express, Fastify, etc.)
	if nodeStrategy, ok := strategy.(*strategies.NodeJSStrategy); ok {
		pkg, err := readPackageJSON(ctx, projectPath)
		if err == nil {
			nodeStrategy.MainFile = strategy.DetectMainFile(ctx, projectPath)
			nodeStrategy.UseTypescript = pkg.hasPackage("typescript")
			nodeStrategy.HasNodemon = pkg.hasPackage("nodemon")
		}
	}

	return strategy, nil
}

// AnalyzeProject returns detailed information about the project.
func AnalyzeProject(ctx context.Context, projectPath string) (*FrameworkInfo, error) {
	for _, analyzer := range projectAnalyzers {
		if info, err := analyzer(ctx, projectPath); err == nil && info != nil {
			return info, nil
		}
	}
	return nil, fmt.Errorf("could not analyze project at %s", projectPath)
}

// FrameworkInfo contains metadata about a detected framework.
type FrameworkInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	HasLockFile bool              `json:"has_lock_file"`
	Scripts     map[string]string `json:"scripts,omitempty"`
}

// packageJSON represents the structure of a package.json file.
type packageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// hasPackage checks if a package is in dependencies or devDependencies.
func (p packageJSON) hasPackage(name string) bool {
	_, inDeps := p.Dependencies[name]
	_, inDevDeps := p.DevDependencies[name]
	return inDeps || inDevDeps
}

// getVersion returns the version of a package from deps or devDeps.
func (p packageJSON) getVersion(name string) string {
	if v, ok := p.Dependencies[name]; ok {
		return v
	}
	if v, ok := p.DevDependencies[name]; ok {
		return v
	}
	return ""
}

type strategyFactory func() FrameworkStrategy

var strategyRegistry = map[string]strategyFactory{
	// Node.js frameworks
	"nextjs":           func() FrameworkStrategy { return strategies.NewNextJSStrategy() },
	"next":             func() FrameworkStrategy { return strategies.NewNextJSStrategy() },
	"react":            func() FrameworkStrategy { return strategies.NewReactStrategy() },
	"cra":              func() FrameworkStrategy { return strategies.NewReactStrategy() },
	"create-react-app": func() FrameworkStrategy { return strategies.NewReactStrategy() },
	"vite-react":       func() FrameworkStrategy { return strategies.NewViteReactStrategy() },
	"vite":             func() FrameworkStrategy { return strategies.NewViteReactStrategy() },
	"express":          func() FrameworkStrategy { return strategies.NewExpressStrategy() },
	"fastify":          func() FrameworkStrategy { return strategies.NewFastifyStrategy() },
	"node":             func() FrameworkStrategy { return strategies.NewNodeJSStrategy() },
	"nodejs":           func() FrameworkStrategy { return strategies.NewNodeJSStrategy() },

	// Python frameworks
	"python":  func() FrameworkStrategy { return strategies.NewPythonStrategy() },
	"flask":   func() FrameworkStrategy { return strategies.NewFlaskStrategy() },
	"django":  func() FrameworkStrategy { return strategies.NewDjangoStrategy() },
	"fastapi": func() FrameworkStrategy { return strategies.NewFastAPIStrategy() },
}

type projectDetector func(ctx context.Context, projectPath string) (FrameworkStrategy, error)

var projectDetectors = []projectDetector{
	detectNodeProject,
	detectPythonProject,
	// TODO: Add more detectors here:
	// detectGoProject,
	// detectRustProject,
}

type projectAnalyzer func(ctx context.Context, projectPath string) (*FrameworkInfo, error)

var projectAnalyzers = []projectAnalyzer{
	analyzeNodeProject,
	analyzePythonProject,
}

// nodeFrameworkDetector defines how to detect and create a Node.js framework strategy.
type nodeFrameworkDetector struct {
	packages []string                                // packages that indicate this framework
	create   func(pkg packageJSON) FrameworkStrategy // factory function
	priority int                                     // lower = higher priority
}

// nodeFrameworkDetectors defines detection rules in priority order.
// More specific frameworks should come before generic ones.
var nodeFrameworkDetectors = []nodeFrameworkDetector{
	{
		packages: []string{"next"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewNextJSStrategy()
			s.UseTypescript = pkg.hasPackage("typescript")
			return s
		},
	},
	{
		packages: []string{"vite", "react"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewViteReactStrategy()
			s.UseTypescript = pkg.hasPackage("typescript")
			return s
		},
	},
	{
		packages: []string{"react-scripts"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewReactStrategy()
			s.UseTypescript = pkg.hasPackage("typescript")
			return s
		},
	},
	{
		packages: []string{"react", "react-dom"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewReactStrategy()
			s.UseTypescript = pkg.hasPackage("typescript")
			return s
		},
	},
	{
		packages: []string{"express"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewExpressStrategy()
			s.UseTypescript = pkg.hasPackage("typescript")
			s.HasNodemon = pkg.hasPackage("nodemon")
			return s
		},
	},
	{
		packages: []string{"fastify"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewFastifyStrategy()
			s.UseTypescript = pkg.hasPackage("typescript")
			s.HasNodemon = pkg.hasPackage("nodemon")
			return s
		},
	},
	{
		packages: []string{"koa"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewNodeJSStrategy()
			s.Framework = "koa"
			s.UseTypescript = pkg.hasPackage("typescript")
			s.HasNodemon = pkg.hasPackage("nodemon")
			return s
		},
	},
	{
		packages: []string{"@hapi/hapi"},
		create: func(pkg packageJSON) FrameworkStrategy {
			s := strategies.NewNodeJSStrategy()
			s.Framework = "hapi"
			s.UseTypescript = pkg.hasPackage("typescript")
			s.HasNodemon = pkg.hasPackage("nodemon")
			return s
		},
	},
}

func detectNodeProject(ctx context.Context, projectPath string) (FrameworkStrategy, error) {
	pkg, err := readPackageJSON(ctx, projectPath)
	if err != nil {
		return nil, err
	}
	strategy, err := detectNodeFramework(pkg)
	if err != nil {
		return nil, err
	}
	// Set MainFile using the strategy's DetectMainFile method
	if nodeStrategy, ok := strategy.(*strategies.NodeJSStrategy); ok {
		nodeStrategy.MainFile = strategy.DetectMainFile(ctx, projectPath)
	}
	return strategy, nil
}

func detectNodeFramework(pkg packageJSON) (FrameworkStrategy, error) {
	// Check each framework detector in priority order
	for _, detector := range nodeFrameworkDetectors {
		if hasAllPackages(pkg, detector.packages) {
			return detector.create(pkg), nil
		}
	}

	// Fallback: Generic Node.js project with start/dev script
	if pkg.Scripts["start"] != "" || pkg.Scripts["dev"] != "" {
		s := strategies.NewNodeJSStrategy()
		s.UseTypescript = pkg.hasPackage("typescript")
		s.HasNodemon = pkg.hasPackage("nodemon")
		// MainFile will be set via DetectMainFile when projectPath is available
		return s, nil
	}

	return nil, fmt.Errorf("no supported Node.js framework detected")
}

func analyzeNodeProject(ctx context.Context, projectPath string) (*FrameworkInfo, error) {
	pkg, err := readPackageJSON(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	info := &FrameworkInfo{
		Scripts:     pkg.Scripts,
		HasLockFile: hasNodeLockFile(ctx, projectPath),
	}

	// Detect framework name and version
	info.Name, info.Version = detectNodeFrameworkInfo(pkg)

	return info, nil
}

// detectNodeFrameworkInfo returns framework name and version from package.json.
func detectNodeFrameworkInfo(pkg packageJSON) (name, version string) {
	// Priority-ordered detection rules
	rules := []struct {
		packages []string
		name     string
	}{
		{[]string{"next"}, "nextjs"},
		{[]string{"vite", "react"}, "vite-react"},
		{[]string{"react-scripts"}, "react"},
		{[]string{"react"}, "react"},
		{[]string{"express"}, "express"},
		{[]string{"fastify"}, "fastify"},
		{[]string{"koa"}, "koa"},
		{[]string{"@hapi/hapi"}, "hapi"},
	}

	for _, rule := range rules {
		if hasAllPackages(pkg, rule.packages) {
			return rule.name, pkg.getVersion(rule.packages[0])
		}
	}

	// Fallback to generic node
	if pkg.Scripts["start"] != "" || pkg.Scripts["dev"] != "" {
		return "node", ""
	}

	return "", ""
}

var pythonIndicators = []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile", "uv.lock"}
var pythonLockFiles = []string{"poetry.lock", "Pipfile.lock", "uv.lock", "requirements.lock"}
var pythonFrameworks = []string{"django", "fastapi", "flask"}

func detectPythonProject(ctx context.Context, projectPath string) (FrameworkStrategy, error) {
	if !hasAnyFile(ctx, projectPath, pythonIndicators) {
		return nil, fmt.Errorf("not a Python project")
	}

	deps := collectPythonDeps(ctx, projectPath)
	depManager := detectPythonDepManager(ctx, projectPath)

	// Check frameworks in order of specificity
	if containsDep(deps, "django") {
		s := strategies.NewDjangoStrategy()
		applyPythonDepManager(s, depManager)
		return s, nil
	}

	if containsDep(deps, "fastapi") {
		s := strategies.NewFastAPIStrategy()
		applyPythonDepManager(s, depManager)
		s.MainFile = s.DetectMainFile(ctx, projectPath)
		return s, nil
	}

	if containsDep(deps, "flask") {
		s := strategies.NewFlaskStrategy()
		applyPythonDepManager(s, depManager)
		s.MainFile = s.DetectMainFile(ctx, projectPath)
		return s, nil
	}

	// Generic Python
	s := strategies.NewPythonStrategy()
	applyPythonDepManager(s, depManager)
	s.MainFile = s.DetectMainFile(ctx, projectPath)
	return s, nil
}

func analyzePythonProject(ctx context.Context, projectPath string) (*FrameworkInfo, error) {
	if !hasAnyFile(ctx, projectPath, pythonIndicators) {
		return nil, fmt.Errorf("not a Python project")
	}

	deps := collectPythonDeps(ctx, projectPath)

	info := &FrameworkInfo{
		HasLockFile: hasAnyFile(ctx, projectPath, pythonLockFiles),
	}

	// Detect framework
	for _, framework := range pythonFrameworks {
		if containsDep(deps, framework) {
			info.Name = framework
			return info, nil
		}
	}

	info.Name = "python"
	return info, nil
}

// pythonDepManager holds dependency manager flags.
type pythonDepManager struct {
	usesPoetry bool
	usesPipenv bool
	usesUv     bool
}

func detectPythonDepManager(ctx context.Context, projectPath string) pythonDepManager {
	return pythonDepManager{
		usesPoetry: fileExists(ctx, filepath.Join(projectPath, "poetry.lock")),
		usesPipenv: fileExists(ctx, filepath.Join(projectPath, "Pipfile.lock")),
		usesUv:     fileExists(ctx, filepath.Join(projectPath, "uv.lock")),
	}
}

// applyPythonDepManager sets dependency manager flags on a Python strategy.
func applyPythonDepManager(s FrameworkStrategy, dm pythonDepManager) {
	if strategy, ok := s.(*strategies.PythonStrategy); ok {
		strategy.UsesPoetry = dm.usesPoetry
		strategy.UsesPipenv = dm.usesPipenv
		strategy.UsesUv = dm.usesUv
	}
}

// collectPythonDeps gathers dependencies from all Python config files.
func collectPythonDeps(ctx context.Context, projectPath string) []string {
	var deps []string

	// Parse requirements.txt
	deps = append(deps, parseRequirementsTxt(ctx, projectPath)...)

	// Check pyproject.toml, Pipfile for framework keywords
	for _, file := range []string{"pyproject.toml", "Pipfile"} {
		deps = append(deps, extractFrameworksFromFile(ctx, projectPath, file)...)
	}

	return deps
}

func parseRequirementsTxt(ctx context.Context, projectPath string) []string {
	var deps []string
	filePath := filepath.Join(projectPath, "requirements.txt")
	data, err := sftp.ReadFile(ctx, filePath)
	if err != nil {
		return deps
	}

	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if pkg := extractPackageName(line); pkg != "" {
			deps = append(deps, strings.ToLower(pkg))
		}
	}
	return deps
}

func extractFrameworksFromFile(ctx context.Context, projectPath, filename string) []string {
	var deps []string
	filePath := filepath.Join(projectPath, filename)
	data, err := sftp.ReadFile(ctx, filePath)
	if err != nil {
		return deps
	}

	content := strings.ToLower(data)
	for _, framework := range pythonFrameworks {
		if strings.Contains(content, framework) {
			deps = append(deps, framework)
		}
	}
	return deps
}

// extractPackageName extracts package name from a requirements.txt line.
func extractPackageName(line string) string {
	// Handle: package, package==1.0, package>=1.0, package[extra], etc.
	separators := "=><[;~ !"
	for i, c := range line {
		if strings.ContainsRune(separators, c) {
			return line[:i]
		}
	}
	return line
}

// readPackageJSON reads and parses package.json from the project path.
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

// hasAllPackages checks if all specified packages are present.
func hasAllPackages(pkg packageJSON, packages []string) bool {
	for _, p := range packages {
		if !pkg.hasPackage(p) {
			return false
		}
	}
	return len(packages) > 0
}

// fileExists checks if a file exists at the given path via SFTP.
func fileExists(ctx context.Context, path string) bool {
	return sftp.FileExists(ctx, path)
}

// hasAnyFile checks if any of the specified files exist in the directory.
func hasAnyFile(ctx context.Context, dir string, files []string) bool {
	for _, f := range files {
		if fileExists(ctx, filepath.Join(dir, f)) {
			return true
		}
	}
	return false
}

// containsDep checks if a dependency is in the list (case-insensitive).
func containsDep(deps []string, name string) bool {
	name = strings.ToLower(name)
	for _, dep := range deps {
		if dep == name {
			return true
		}
	}
	return false
}

// hasNodeLockFile checks for Node.js lock files.
func hasNodeLockFile(ctx context.Context, projectPath string) bool {
	lockFiles := []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "bun.lockb"}
	return hasAnyFile(ctx, projectPath, lockFiles)
}
