package codebase

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type ConfidenceValue[T any] struct {
	Value      T          `json:"value"`
	Source     string     `json:"source"`
	Confidence Confidence `json:"confidence"`
}

type RepoHints struct {
	Confidence      Confidence               `json:"confidence"`
	Ecosystem       *string                  `json:"ecosystem"`
	Framework       *ConfidenceValue[string] `json:"framework"`
	Port            *ConfidenceValue[int]    `json:"port"`
	HasDockerfile   bool                     `json:"has_dockerfile"`
	DockerfilePath  *string                  `json:"dockerfile_path"`
	HasCompose      bool                     `json:"has_docker_compose"`
	HasDockerignore bool                     `json:"has_dockerignore"`
	PackageManager  *string                  `json:"package_manager"`
	IsMonorepo      bool                     `json:"is_monorepo"`
	MonorepoMarkers []string                 `json:"monorepo_markers"`
	BuildCommand    *string                  `json:"build_command"`
	StartCommand    *string                  `json:"start_command"`
	EnvFiles        []string                 `json:"env_files"`
	RequiredEnvVars []string                 `json:"required_env_vars"`
	Warnings        []string                 `json:"warnings"`
}

type FileEntry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type frameworkDetector struct {
	dep         string
	framework   string
	ecosystem   string
	defaultPort int
}

type pythonFrameworkInfo struct {
	framework string
	port      int
}

type packageJSON struct {
	Name            string            `json:"name"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
	Workspaces      json.RawMessage   `json:"workspaces"`
}

var frameworkDetectors = []frameworkDetector{
	{dep: "next", framework: "next.js", ecosystem: "node", defaultPort: 3000},
	{dep: "nuxt", framework: "nuxt", ecosystem: "node", defaultPort: 3000},
	{dep: "@remix-run/node", framework: "remix", ecosystem: "node", defaultPort: 3000},
	{dep: "@remix-run/react", framework: "remix", ecosystem: "node", defaultPort: 3000},
	{dep: "astro", framework: "astro", ecosystem: "node", defaultPort: 4321},
	{dep: "svelte", framework: "svelte", ecosystem: "node", defaultPort: 5173},
	{dep: "@sveltejs/kit", framework: "sveltekit", ecosystem: "node", defaultPort: 3000},
	{dep: "@angular/core", framework: "angular", ecosystem: "node", defaultPort: 4200},
	{dep: "react-scripts", framework: "create-react-app", ecosystem: "node", defaultPort: 3000},
	{dep: "vite", framework: "vite", ecosystem: "node", defaultPort: 5173},
	{dep: "express", framework: "express", ecosystem: "node", defaultPort: 3000},
	{dep: "fastify", framework: "fastify", ecosystem: "node", defaultPort: 3000},
	{dep: "hono", framework: "hono", ecosystem: "node", defaultPort: 3000},
	{dep: "koa", framework: "koa", ecosystem: "node", defaultPort: 3000},
	{dep: "nest", framework: "nestjs", ecosystem: "node", defaultPort: 3000},
	{dep: "@nestjs/core", framework: "nestjs", ecosystem: "node", defaultPort: 3000},
}

var serverFrameworks = map[string]bool{
	"express": true, "fastify": true, "hono": true, "koa": true, "nestjs": true,
}

var fullstackFrameworks = map[string]bool{
	"next.js": true, "nuxt": true, "remix": true, "sveltekit": true,
}

var pythonFrameworks = map[string]pythonFrameworkInfo{
	"django":    {framework: "django", port: 8000},
	"flask":     {framework: "flask", port: 5000},
	"fastapi":   {framework: "fastapi", port: 8000},
	"uvicorn":   {framework: "fastapi", port: 8000},
	"gunicorn":  {framework: "gunicorn", port: 8000},
	"starlette": {framework: "starlette", port: 8000},
}

var monorepoMarkers = []string{"turbo.json", "nx.json", "lerna.json", "pnpm-workspace.yaml"}

// Ordered so first match wins — pnpm before npm since pnpm-lock.yaml is more specific.
var lockfileToPM = []struct {
	file string
	pm   string
}{
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "yarn"},
	{"package-lock.json", "npm"},
	{"bun.lockb", "bun"},
}

var (
	exposeRe       = regexp.MustCompile(`(?m)^\s*EXPOSE\s+(\d+)`)
	envVarRe       = regexp.MustCompile(`(?m)^([A-Z][A-Z0-9_]{2,})=`)
	portInScriptRe = regexp.MustCompile(`(?:--port|--listen|-p)\s+(\d+)|PORT[=:\s]+(\d+)`)
)

func findFile(files []FileEntry, name string) *FileEntry {
	for i := range files {
		if files[i].Path == name || strings.HasSuffix(files[i].Path, "/"+name) {
			return &files[i]
		}
	}
	return nil
}

func findRootFile(files []FileEntry, name string) *FileEntry {
	for i := range files {
		if files[i].Path == name {
			return &files[i]
		}
	}
	return nil
}

func hasRootFile(files []FileEntry, name string) bool {
	return findRootFile(files, name) != nil
}

func hasRootDir(files []FileEntry, dir string) bool {
	prefix := dir
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	for _, f := range files {
		if strings.HasPrefix(f.Path, prefix) {
			return true
		}
	}
	return false
}

func minConfidence(values ...Confidence) Confidence {
	for _, v := range values {
		if v == ConfidenceLow {
			return ConfidenceLow
		}
	}
	for _, v := range values {
		if v == ConfidenceMedium {
			return ConfidenceMedium
		}
	}
	return ConfidenceHigh
}

func detectPortFromDockerfile(content string) *int {
	m := exposeRe.FindStringSubmatch(content)
	if m == nil {
		return nil
	}
	p, err := strconv.Atoi(m[1])
	if err != nil {
		return nil
	}
	return &p
}

func detectPortFromScripts(scripts map[string]string) *int {
	for _, cmd := range scripts {
		m := portInScriptRe.FindStringSubmatch(cmd)
		if m == nil {
			continue
		}
		raw := m[1]
		if raw == "" {
			raw = m[2]
		}
		p, err := strconv.Atoi(raw)
		if err != nil {
			continue
		}
		if p > 0 && p < 65536 {
			return &p
		}
	}
	return nil
}

func extractEnvVars(content string) []string {
	matches := envVarRe.FindAllStringSubmatch(content, -1)
	vars := make([]string, 0, len(matches))
	for _, m := range matches {
		vars = append(vars, m[1])
	}
	return vars
}

func detectPythonFramework(files []FileEntry) *pythonFrameworkInfo {
	var content string
	for _, name := range []string{"requirements.txt", "pyproject.toml", "Pipfile"} {
		if f := findRootFile(files, name); f != nil {
			content = f.Content
			break
		}
	}
	if content == "" {
		return nil
	}
	for pkg, info := range pythonFrameworks {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(pkg) + `\b`)
		if re.MatchString(content) {
			result := info
			return &result
		}
	}
	return nil
}

func hasWorkspaces(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var arr []string
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		return true
	}
	var obj struct {
		Packages []string `json:"packages"`
	}
	if json.Unmarshal(raw, &obj) == nil && len(obj.Packages) > 0 {
		return true
	}
	return false
}

func strPtr(s string) *string { return &s }

func AnalyzeFiles(files []FileEntry) RepoHints {
	warnings := make([]string, 0)
	var ecosystem *string
	var framework *ConfidenceValue[string]
	var port *ConfidenceValue[int]
	var buildCommand *string
	var startCommand *string

	hasDF := hasRootFile(files, "Dockerfile")
	var dockerfilePath *string
	if hasDF {
		dockerfilePath = strPtr("Dockerfile")
	}
	hasDC := hasRootFile(files, "docker-compose.yml") || hasRootFile(files, "docker-compose.yaml")
	hasDI := hasRootFile(files, ".dockerignore")

	var packageManager *string
	for _, lf := range lockfileToPM {
		if hasRootFile(files, lf.file) {
			packageManager = strPtr(lf.pm)
			break
		}
	}

	foundMonorepoMarkers := make([]string, 0)
	for _, marker := range monorepoMarkers {
		if hasRootFile(files, marker) {
			foundMonorepoMarkers = append(foundMonorepoMarkers, marker)
		}
	}
	hasApps := hasRootDir(files, "apps")
	isMonorepo := len(foundMonorepoMarkers) > 0

	pkgFile := findRootFile(files, "package.json")
	var pkg *packageJSON
	if pkgFile != nil {
		var p packageJSON
		if json.Unmarshal([]byte(pkgFile.Content), &p) == nil {
			pkg = &p
		}
	}

	if pkg != nil && hasWorkspaces(pkg.Workspaces) {
		isMonorepo = true
		foundMonorepoMarkers = append(foundMonorepoMarkers, "package.json workspaces")
	}

	if !isMonorepo && hasApps {
		count := 0
		for _, f := range files {
			if strings.HasPrefix(f.Path, "apps/") &&
				strings.HasSuffix(f.Path, "/package.json") &&
				len(strings.Split(f.Path, "/")) == 3 {
				count++
			}
		}
		if count >= 2 {
			isMonorepo = true
			foundMonorepoMarkers = append(foundMonorepoMarkers, "apps/ with multiple packages")
		}
	}

	if pkg != nil {
		ecosystem = strPtr("node")
		allDeps := make(map[string]string)
		for k, v := range pkg.Dependencies {
			allDeps[k] = v
		}
		for k, v := range pkg.DevDependencies {
			allDeps[k] = v
		}

		var matched []frameworkDetector
		for _, d := range frameworkDetectors {
			if _, ok := allDeps[d.dep]; ok {
				matched = append(matched, d)
			}
		}

		var fullstackMatch, serverMatch *frameworkDetector
		for i := range matched {
			if fullstackFrameworks[matched[i].framework] && fullstackMatch == nil {
				fullstackMatch = &matched[i]
			}
			if serverFrameworks[matched[i].framework] && serverMatch == nil {
				serverMatch = &matched[i]
			}
		}

		switch {
		case len(matched) == 1:
			framework = &ConfidenceValue[string]{
				Value: matched[0].framework, Source: "package.json dep: " + matched[0].dep, Confidence: ConfidenceHigh,
			}
		case fullstackMatch != nil:
			framework = &ConfidenceValue[string]{
				Value: fullstackMatch.framework, Source: "package.json dep: " + fullstackMatch.dep, Confidence: ConfidenceHigh,
			}
		case len(matched) > 1:
			framework = &ConfidenceValue[string]{
				Value: matched[0].framework, Source: "package.json dep: " + matched[0].dep, Confidence: ConfidenceMedium,
			}
			names := make([]string, len(matched))
			for i, m := range matched {
				names[i] = m.framework
			}
			warnings = append(warnings, "Multiple frameworks detected ("+strings.Join(names, ", ")+") — verify primary")
		}

		if packageManager != nil && pkg.Scripts != nil {
			if _, ok := pkg.Scripts["build"]; ok {
				buildCommand = strPtr(*packageManager + " run build")
			}
			if _, ok := pkg.Scripts["start"]; ok {
				startCommand = strPtr(*packageManager + " start")
			} else if _, ok := pkg.Scripts["dev"]; ok {
				startCommand = strPtr(*packageManager + " run dev")
			}
		}
	}

	if ecosystem == nil && hasRootFile(files, "go.mod") {
		ecosystem = strPtr("go")
		framework = &ConfidenceValue[string]{Value: "go", Source: "go.mod", Confidence: ConfidenceHigh}
		buildCommand = strPtr("go build -o app ./...")
		startCommand = strPtr("./app")
	}

	if ecosystem == nil {
		pyResult := detectPythonFramework(files)
		if pyResult != nil {
			ecosystem = strPtr("python")
			framework = &ConfidenceValue[string]{Value: pyResult.framework, Source: "requirements", Confidence: ConfidenceHigh}
			if port == nil {
				port = &ConfidenceValue[int]{Value: pyResult.port, Source: "framework default", Confidence: ConfidenceLow}
			}
		} else if hasRootFile(files, "requirements.txt") || hasRootFile(files, "pyproject.toml") {
			ecosystem = strPtr("python")
		}
	}

	if ecosystem == nil && hasRootFile(files, "Cargo.toml") {
		ecosystem = strPtr("rust")
		framework = &ConfidenceValue[string]{Value: "rust", Source: "Cargo.toml", Confidence: ConfidenceHigh}
		buildCommand = strPtr("cargo build --release")
	}

	if ecosystem == nil {
		if hasRootFile(files, "pom.xml") {
			ecosystem = strPtr("java")
			framework = &ConfidenceValue[string]{Value: "maven", Source: "pom.xml", Confidence: ConfidenceHigh}
			buildCommand = strPtr("mvn package")
		} else if hasRootFile(files, "build.gradle") || hasRootFile(files, "build.gradle.kts") {
			ecosystem = strPtr("java")
			framework = &ConfidenceValue[string]{Value: "gradle", Source: "build.gradle", Confidence: ConfidenceHigh}
			buildCommand = strPtr("gradle build")
		}
	}

	if ecosystem == nil && hasRootFile(files, "Gemfile") {
		ecosystem = strPtr("ruby")
		gemfile := findRootFile(files, "Gemfile")
		if gemfile != nil && regexp.MustCompile(`(?i)\brails\b`).MatchString(gemfile.Content) {
			framework = &ConfidenceValue[string]{Value: "rails", Source: "Gemfile", Confidence: ConfidenceHigh}
		}
	}

	if ecosystem == nil && hasRootFile(files, "mix.exs") {
		ecosystem = strPtr("elixir")
		framework = &ConfidenceValue[string]{Value: "elixir", Source: "mix.exs", Confidence: ConfidenceHigh}
	}

	if ecosystem == nil && hasRootFile(files, "composer.json") {
		ecosystem = strPtr("php")
		composer := findRootFile(files, "composer.json")
		if composer != nil && regexp.MustCompile(`(?i)laravel`).MatchString(composer.Content) {
			framework = &ConfidenceValue[string]{Value: "laravel", Source: "composer.json", Confidence: ConfidenceHigh}
		}
	}

	// Port detection priority: Dockerfile EXPOSE > docker-compose ports > scripts > .env.example > framework default
	if hasDF {
		df := findRootFile(files, "Dockerfile")
		if df != nil {
			if ep := detectPortFromDockerfile(df.Content); ep != nil {
				port = &ConfidenceValue[int]{Value: *ep, Source: "Dockerfile EXPOSE", Confidence: ConfidenceHigh}
			}
		}
	}

	if port == nil && hasDC {
		composeFile := findRootFile(files, "docker-compose.yml")
		if composeFile == nil {
			composeFile = findRootFile(files, "docker-compose.yaml")
		}
		if composeFile != nil {
			re := regexp.MustCompile(`ports:\s*\n\s*-\s*["']?(\d+):(\d+)`)
			m := re.FindStringSubmatch(composeFile.Content)
			if m != nil {
				if p, err := strconv.Atoi(m[2]); err == nil {
					port = &ConfidenceValue[int]{Value: p, Source: "docker-compose ports", Confidence: ConfidenceHigh}
				}
			}
		}
	}

	if port == nil && pkg != nil && pkg.Scripts != nil {
		if sp := detectPortFromScripts(pkg.Scripts); sp != nil {
			port = &ConfidenceValue[int]{Value: *sp, Source: "package.json scripts", Confidence: ConfidenceMedium}
		}
	}

	if port == nil {
		for _, name := range []string{".env.example", ".env.sample"} {
			envFile := findRootFile(files, name)
			if envFile == nil {
				continue
			}
			re := regexp.MustCompile(`(?m)^PORT=(\d+)`)
			m := re.FindStringSubmatch(envFile.Content)
			if m != nil {
				if p, err := strconv.Atoi(m[1]); err == nil {
					port = &ConfidenceValue[int]{Value: p, Source: ".env.example", Confidence: ConfidenceMedium}
					break
				}
			}
		}
	}

	if port == nil && framework != nil {
		for _, d := range frameworkDetectors {
			if d.framework == framework.Value {
				port = &ConfidenceValue[int]{Value: d.defaultPort, Source: "framework default", Confidence: ConfidenceLow}
				break
			}
		}
	}

	envFiles := make([]string, 0)
	envVarSet := make(map[string]bool)
	for _, name := range []string{".env.example", ".env.sample", ".env.template"} {
		ef := findRootFile(files, name)
		if ef == nil {
			continue
		}
		envFiles = append(envFiles, name)
		for _, v := range extractEnvVars(ef.Content) {
			envVarSet[v] = true
		}
	}
	requiredEnvVars := make([]string, 0, len(envVarSet))
	for v := range envVarSet {
		requiredEnvVars = append(requiredEnvVars, v)
	}

	if isMonorepo {
		warnings = append(warnings, "Monorepo detected ("+strings.Join(foundMonorepoMarkers, ", ")+") — may need per-service deploy")
	}
	if hasDC && !hasDF {
		warnings = append(warnings, "docker-compose.yml found but no root Dockerfile — likely multi-service setup")
	}

	fieldConfidences := []Confidence{ConfidenceHigh}
	if framework != nil {
		fieldConfidences = append(fieldConfidences, framework.Confidence)
	}
	if port != nil {
		fieldConfidences = append(fieldConfidences, port.Confidence)
	}
	if isMonorepo {
		fieldConfidences = append(fieldConfidences, ConfidenceMedium)
	}
	if len(warnings) > 0 {
		allMonorepo := true
		for _, w := range warnings {
			if !strings.Contains(w, "Monorepo") {
				allMonorepo = false
				break
			}
		}
		if !allMonorepo {
			fieldConfidences = append(fieldConfidences, ConfidenceMedium)
		}
	}

	return RepoHints{
		Confidence:      minConfidence(fieldConfidences...),
		Ecosystem:       ecosystem,
		Framework:       framework,
		Port:            port,
		HasDockerfile:   hasDF,
		DockerfilePath:  dockerfilePath,
		HasCompose:      hasDC,
		HasDockerignore: hasDI,
		PackageManager:  packageManager,
		IsMonorepo:      isMonorepo,
		MonorepoMarkers: foundMonorepoMarkers,
		BuildCommand:    buildCommand,
		StartCommand:    startCommand,
		EnvFiles:        envFiles,
		RequiredEnvVars: requiredEnvVars,
		Warnings:        warnings,
	}
}
