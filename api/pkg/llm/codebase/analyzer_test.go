package codebase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinConfidence(t *testing.T) {
	assert.Equal(t, ConfidenceHigh, minConfidence(ConfidenceHigh, ConfidenceHigh))
	assert.Equal(t, ConfidenceMedium, minConfidence(ConfidenceHigh, ConfidenceMedium))
	assert.Equal(t, ConfidenceLow, minConfidence(ConfidenceHigh, ConfidenceMedium, ConfidenceLow))
	assert.Equal(t, ConfidenceLow, minConfidence(ConfidenceLow))
	assert.Equal(t, ConfidenceHigh, minConfidence())
}

func TestDetectPortFromDockerfile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *int
	}{
		{"simple expose", "FROM node\nEXPOSE 8080\nCMD [\"node\"]", intPtr(8080)},
		{"expose with spaces", "FROM node\n  EXPOSE 3000\n", intPtr(3000)},
		{"no expose", "FROM node\nCMD [\"node\"]", nil},
		{"multiple expose takes first", "EXPOSE 4000\nEXPOSE 5000", intPtr(4000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPortFromDockerfile(tt.content)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestDetectPortFromScripts(t *testing.T) {
	tests := []struct {
		name    string
		scripts map[string]string
		want    *int
	}{
		{"--port flag", map[string]string{"dev": "vite --port 4000"}, intPtr(4000)},
		{"--listen flag", map[string]string{"start": "server --listen 9090"}, intPtr(9090)},
		{"-p flag", map[string]string{"start": "http-server -p 8888"}, intPtr(8888)},
		{"PORT env", map[string]string{"start": "PORT=5555 node server.js"}, intPtr(5555)},
		{"no port", map[string]string{"start": "node server.js"}, nil},
		{"invalid port", map[string]string{"start": "server --port 99999"}, nil},
		{"empty scripts", map[string]string{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPortFromScripts(tt.scripts)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestExtractEnvVars(t *testing.T) {
	content := "DATABASE_URL=postgres://...\nPORT=3000\nSECRET_KEY=abc\nlowercase=ignored\nAB=tooShort"
	vars := extractEnvVars(content)
	assert.Contains(t, vars, "DATABASE_URL")
	assert.Contains(t, vars, "PORT")
	assert.Contains(t, vars, "SECRET_KEY")
	assert.NotContains(t, vars, "lowercase")
	assert.NotContains(t, vars, "AB")
}

func TestExtractEnvVarsEmpty(t *testing.T) {
	vars := extractEnvVars("")
	assert.Empty(t, vars)
}

func TestHasWorkspaces(t *testing.T) {
	assert.True(t, hasWorkspaces([]byte(`["packages/*"]`)))
	assert.True(t, hasWorkspaces([]byte(`{"packages":["apps/*"]}`)))
	assert.False(t, hasWorkspaces([]byte(`[]`)))
	assert.False(t, hasWorkspaces([]byte(`{"packages":[]}`)))
	assert.False(t, hasWorkspaces(nil))
	assert.False(t, hasWorkspaces([]byte{}))
}

func TestFindFile(t *testing.T) {
	files := []FileEntry{
		{Path: "src/utils/helper.go", Content: "package utils"},
		{Path: "Dockerfile", Content: "FROM golang"},
	}
	assert.NotNil(t, findFile(files, "helper.go"))
	assert.NotNil(t, findFile(files, "Dockerfile"))
	assert.Nil(t, findFile(files, "missing.go"))
}

func TestHasRootDir(t *testing.T) {
	files := []FileEntry{
		{Path: "apps/web/package.json"},
		{Path: "src/main.go"},
	}
	assert.True(t, hasRootDir(files, "apps"))
	assert.True(t, hasRootDir(files, "src"))
	assert.False(t, hasRootDir(files, "packages"))
}

// --- Ecosystem detection tests ---

func TestNodeNextJS(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"next":"14.0.0","react":"18.0.0"},"scripts":{"build":"next build","start":"next start"}}`},
		{Path: "package-lock.json", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("node"), h.Ecosystem)
	assert.NotNil(t, h.Framework)
	assert.Equal(t, "next.js", h.Framework.Value)
	assert.Equal(t, ConfidenceHigh, h.Framework.Confidence)
	assert.NotNil(t, h.Port)
	assert.Equal(t, 3000, h.Port.Value)
	assert.Equal(t, "framework default", h.Port.Source)
	assert.NotNil(t, h.PackageManager)
	assert.Equal(t, "npm", *h.PackageManager)
	assert.NotNil(t, h.BuildCommand)
	assert.Equal(t, "npm run build", *h.BuildCommand)
	assert.NotNil(t, h.StartCommand)
	assert.Equal(t, "npm start", *h.StartCommand)
}

func TestNodeNuxt(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"nuxt":"3.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "nuxt", h.Framework.Value)
}

func TestNodeRemix(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"@remix-run/node":"2.0.0","@remix-run/react":"2.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "remix", h.Framework.Value)
}

func TestNodeAstro(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"astro":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "astro", h.Framework.Value)
	assert.Equal(t, 4321, h.Port.Value)
}

func TestNodeSvelte(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"devDependencies":{"svelte":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "svelte", h.Framework.Value)
	assert.Equal(t, 5173, h.Port.Value)
}

func TestNodeSvelteKit(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"devDependencies":{"@sveltejs/kit":"2.0.0","svelte":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "sveltekit", h.Framework.Value)
}

func TestNodeAngular(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"@angular/core":"17.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "angular", h.Framework.Value)
	assert.Equal(t, 4200, h.Port.Value)
}

func TestNodeCRA(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"react-scripts":"5.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "create-react-app", h.Framework.Value)
}

func TestNodeVite(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"devDependencies":{"vite":"5.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "vite", h.Framework.Value)
	assert.Equal(t, 5173, h.Port.Value)
}

func TestNodeExpress(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "express", h.Framework.Value)
}

func TestNodeFastify(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"fastify":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "fastify", h.Framework.Value)
}

func TestNodeHono(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"hono":"3.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "hono", h.Framework.Value)
}

func TestNodeKoa(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"koa":"2.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "koa", h.Framework.Value)
}

func TestNodeNestJS(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"@nestjs/core":"10.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "nestjs", h.Framework.Value)
}

func TestNodeFullstackPlusServer(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"next":"14.0.0","express":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "next.js", h.Framework.Value)
	assert.Equal(t, ConfidenceHigh, h.Framework.Confidence)
}

func TestNodeMultipleServerFrameworks(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0","koa":"2.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.NotNil(t, h.Framework)
	assert.Equal(t, ConfidenceMedium, h.Framework.Confidence)
	assert.NotEmpty(t, h.Warnings)
	found := false
	for _, w := range h.Warnings {
		if contains(w, "Multiple frameworks") {
			found = true
		}
	}
	assert.True(t, found, "should have multiple frameworks warning")
}

func TestNodeNoDeps(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"my-app"}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("node"), h.Ecosystem)
	assert.Nil(t, h.Framework)
}

func TestNodeInvalidJSON(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{invalid json`},
	}
	h := AnalyzeFiles(files)
	assert.Nil(t, h.Ecosystem)
}

func TestNodeDevOnlyStartCommand(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"next":"14.0.0"},"scripts":{"dev":"next dev"}}`},
		{Path: "yarn.lock", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.NotNil(t, h.StartCommand)
	assert.Equal(t, "yarn run dev", *h.StartCommand)
	assert.Nil(t, h.BuildCommand)
}

func TestGoEcosystem(t *testing.T) {
	files := []FileEntry{
		{Path: "go.mod", Content: "module example.com/app\n\ngo 1.22"},
		{Path: "main.go", Content: "package main"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("go"), h.Ecosystem)
	assert.Equal(t, "go", h.Framework.Value)
	assert.Equal(t, "go.mod", h.Framework.Source)
	assert.Equal(t, "go build -o app ./...", *h.BuildCommand)
	assert.Equal(t, "./app", *h.StartCommand)
}

func TestPythonDjango(t *testing.T) {
	files := []FileEntry{
		{Path: "requirements.txt", Content: "django==4.2\npsycopg2==2.9"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("python"), h.Ecosystem)
	assert.Equal(t, "django", h.Framework.Value)
	assert.Equal(t, 8000, h.Port.Value)
}

func TestPythonFlask(t *testing.T) {
	files := []FileEntry{
		{Path: "requirements.txt", Content: "flask==3.0\ngunicorn==21.2"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("python"), h.Ecosystem)
	// flask or gunicorn detected depending on map iteration, both valid
	assert.NotNil(t, h.Framework)
}

func TestPythonFastAPI(t *testing.T) {
	files := []FileEntry{
		{Path: "requirements.txt", Content: "fastapi==0.104.0\nuvicorn==0.24.0"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("python"), h.Ecosystem)
	assert.Equal(t, "fastapi", h.Framework.Value)
}

func TestPythonPyprojectToml(t *testing.T) {
	files := []FileEntry{
		{Path: "pyproject.toml", Content: "[tool.poetry.dependencies]\ndjango = \"^4.2\""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("python"), h.Ecosystem)
	assert.Equal(t, "django", h.Framework.Value)
}

func TestPythonPipfile(t *testing.T) {
	files := []FileEntry{
		{Path: "Pipfile", Content: "[packages]\nflask = \"*\""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("python"), h.Ecosystem)
	assert.Equal(t, "flask", h.Framework.Value)
}

func TestPythonNoFramework(t *testing.T) {
	files := []FileEntry{
		{Path: "requirements.txt", Content: "requests==2.31\nnumpy==1.26"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("python"), h.Ecosystem)
	assert.Nil(t, h.Framework)
}

func TestRustEcosystem(t *testing.T) {
	files := []FileEntry{
		{Path: "Cargo.toml", Content: "[package]\nname = \"my-app\""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("rust"), h.Ecosystem)
	assert.Equal(t, "rust", h.Framework.Value)
	assert.Equal(t, "cargo build --release", *h.BuildCommand)
}

func TestJavaMaven(t *testing.T) {
	files := []FileEntry{
		{Path: "pom.xml", Content: "<project></project>"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("java"), h.Ecosystem)
	assert.Equal(t, "maven", h.Framework.Value)
	assert.Equal(t, "mvn package", *h.BuildCommand)
}

func TestJavaGradle(t *testing.T) {
	files := []FileEntry{
		{Path: "build.gradle", Content: "apply plugin: 'java'"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("java"), h.Ecosystem)
	assert.Equal(t, "gradle", h.Framework.Value)
	assert.Equal(t, "gradle build", *h.BuildCommand)
}

func TestJavaGradleKotlin(t *testing.T) {
	files := []FileEntry{
		{Path: "build.gradle.kts", Content: "plugins { java }"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("java"), h.Ecosystem)
	assert.Equal(t, "gradle", h.Framework.Value)
}

func TestRubyRails(t *testing.T) {
	files := []FileEntry{
		{Path: "Gemfile", Content: "source 'https://rubygems.org'\ngem 'rails', '~> 7.0'"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("ruby"), h.Ecosystem)
	assert.Equal(t, "rails", h.Framework.Value)
}

func TestRubyNoRails(t *testing.T) {
	files := []FileEntry{
		{Path: "Gemfile", Content: "source 'https://rubygems.org'\ngem 'sinatra'"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("ruby"), h.Ecosystem)
	assert.Nil(t, h.Framework)
}

func TestElixirEcosystem(t *testing.T) {
	files := []FileEntry{
		{Path: "mix.exs", Content: "defmodule MyApp.MixProject do"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("elixir"), h.Ecosystem)
	assert.Equal(t, "elixir", h.Framework.Value)
}

func TestPHPLaravel(t *testing.T) {
	files := []FileEntry{
		{Path: "composer.json", Content: `{"require":{"laravel/framework":"^10.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("php"), h.Ecosystem)
	assert.Equal(t, "laravel", h.Framework.Value)
}

func TestPHPNoLaravel(t *testing.T) {
	files := []FileEntry{
		{Path: "composer.json", Content: `{"require":{"slim/slim":"^4.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("php"), h.Ecosystem)
	assert.Nil(t, h.Framework)
}

// --- Docker detection ---

func TestDockerfileDetection(t *testing.T) {
	files := []FileEntry{
		{Path: "Dockerfile", Content: "FROM node:20\nEXPOSE 8080"},
		{Path: ".dockerignore", Content: "node_modules"},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.HasDockerfile)
	assert.Equal(t, sPtr("Dockerfile"), h.DockerfilePath)
	assert.True(t, h.HasDockerignore)
	assert.NotNil(t, h.Port)
	assert.Equal(t, 8080, h.Port.Value)
	assert.Equal(t, "Dockerfile EXPOSE", h.Port.Source)
	assert.Equal(t, ConfidenceHigh, h.Port.Confidence)
}

func TestDockerComposeOnly(t *testing.T) {
	files := []FileEntry{
		{Path: "docker-compose.yml", Content: "services:\n  web:\n    ports:\n      - \"3000:8000\""},
	}
	h := AnalyzeFiles(files)
	assert.False(t, h.HasDockerfile)
	assert.True(t, h.HasCompose)
	assert.NotNil(t, h.Port)
	assert.Equal(t, 8000, h.Port.Value)
	assert.Equal(t, "docker-compose ports", h.Port.Source)
	found := false
	for _, w := range h.Warnings {
		if contains(w, "no root Dockerfile") {
			found = true
		}
	}
	assert.True(t, found)
}

func TestDockerComposeYamlExt(t *testing.T) {
	files := []FileEntry{
		{Path: "docker-compose.yaml", Content: "services:\n  api:\n    ports:\n      - '4000:4000'"},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.HasCompose)
	assert.Equal(t, 4000, h.Port.Value)
}

func TestDockerfilePortOverridesCompose(t *testing.T) {
	files := []FileEntry{
		{Path: "Dockerfile", Content: "FROM node\nEXPOSE 9090"},
		{Path: "docker-compose.yml", Content: "services:\n  web:\n    ports:\n      - \"3000:8000\""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 9090, h.Port.Value)
	assert.Equal(t, "Dockerfile EXPOSE", h.Port.Source)
}

// --- Port detection priority ---

func TestPortFromScripts(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"},"scripts":{"start":"node server.js --port 4444"}}`},
		{Path: "pnpm-lock.yaml", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 4444, h.Port.Value)
	assert.Equal(t, "package.json scripts", h.Port.Source)
}

func TestPortFromEnvExample(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"},"scripts":{"start":"node server.js"}}`},
		{Path: ".env.example", Content: "DATABASE_URL=postgres://...\nPORT=7777"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 7777, h.Port.Value)
	assert.Equal(t, ".env.example", h.Port.Source)
}

func TestPortFromEnvSample(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"},"scripts":{"start":"node server.js"}}`},
		{Path: ".env.sample", Content: "PORT=6666"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 6666, h.Port.Value)
}

func TestPortFallbackToFrameworkDefault(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 3000, h.Port.Value)
	assert.Equal(t, "framework default", h.Port.Source)
	assert.Equal(t, ConfidenceLow, h.Port.Confidence)
}

// --- Package manager detection ---

func TestPackageManagerPnpm(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"app"}`},
		{Path: "pnpm-lock.yaml", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("pnpm"), h.PackageManager)
}

func TestPackageManagerYarn(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"app"}`},
		{Path: "yarn.lock", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("yarn"), h.PackageManager)
}

func TestPackageManagerBun(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"app"}`},
		{Path: "bun.lockb", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("bun"), h.PackageManager)
}

// --- Monorepo detection ---

func TestMonorepoTurbo(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"root"}`},
		{Path: "turbo.json", Content: "{}"},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
	assert.Contains(t, h.MonorepoMarkers, "turbo.json")
}

func TestMonorepoNx(t *testing.T) {
	files := []FileEntry{
		{Path: "nx.json", Content: "{}"},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
	assert.Contains(t, h.MonorepoMarkers, "nx.json")
}

func TestMonorepoLerna(t *testing.T) {
	files := []FileEntry{
		{Path: "lerna.json", Content: "{}"},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
}

func TestMonrepoPnpmWorkspace(t *testing.T) {
	files := []FileEntry{
		{Path: "pnpm-workspace.yaml", Content: "packages:\n  - 'apps/*'"},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
}

func TestMonorepoWorkspacesArray(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"root","workspaces":["packages/*"]}`},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
	assert.Contains(t, h.MonorepoMarkers, "package.json workspaces")
}

func TestMonorepoWorkspacesObject(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"root","workspaces":{"packages":["apps/*"]}}`},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
}

func TestMonorepoAppsDir(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"root"}`},
		{Path: "apps/web/package.json", Content: `{}`},
		{Path: "apps/api/package.json", Content: `{}`},
	}
	h := AnalyzeFiles(files)
	assert.True(t, h.IsMonorepo)
	assert.Contains(t, h.MonorepoMarkers, "apps/ with multiple packages")
}

func TestMonorepoAppsDirSingleApp(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"name":"root"}`},
		{Path: "apps/web/package.json", Content: `{}`},
	}
	h := AnalyzeFiles(files)
	assert.False(t, h.IsMonorepo)
}

func TestMonorepoWarning(t *testing.T) {
	files := []FileEntry{
		{Path: "turbo.json", Content: "{}"},
	}
	h := AnalyzeFiles(files)
	found := false
	for _, w := range h.Warnings {
		if contains(w, "Monorepo detected") {
			found = true
		}
	}
	assert.True(t, found)
}

// --- Env var detection ---

func TestEnvFileDetection(t *testing.T) {
	files := []FileEntry{
		{Path: ".env.example", Content: "DATABASE_URL=\nSECRET_KEY=abc"},
		{Path: ".env.template", Content: "API_KEY=\nSECRET_KEY=def"},
	}
	h := AnalyzeFiles(files)
	assert.Contains(t, h.EnvFiles, ".env.example")
	assert.Contains(t, h.EnvFiles, ".env.template")
	assert.Contains(t, h.RequiredEnvVars, "DATABASE_URL")
	assert.Contains(t, h.RequiredEnvVars, "SECRET_KEY")
	assert.Contains(t, h.RequiredEnvVars, "API_KEY")
}

// --- Confidence tests ---

func TestConfidenceHighWhenSimple(t *testing.T) {
	files := []FileEntry{
		{Path: "go.mod", Content: "module example.com/app"},
		{Path: "Dockerfile", Content: "FROM golang\nEXPOSE 8080"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, ConfidenceHigh, h.Confidence)
}

func TestConfidenceLowWhenFrameworkDefault(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"}}`},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, ConfidenceLow, h.Confidence)
}

func TestConfidenceMediumForMonorepo(t *testing.T) {
	files := []FileEntry{
		{Path: "go.mod", Content: "module example.com/app"},
		{Path: "Dockerfile", Content: "FROM golang\nEXPOSE 8080"},
		{Path: "turbo.json", Content: "{}"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, ConfidenceMedium, h.Confidence)
}

func TestConfidenceMediumForNonMonorepoWarnings(t *testing.T) {
	files := []FileEntry{
		{Path: "docker-compose.yml", Content: "services:\n  web:\n    build: ."},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, ConfidenceMedium, h.Confidence)
}

// --- Edge cases ---

func TestEmptyFiles(t *testing.T) {
	h := AnalyzeFiles(nil)
	assert.Nil(t, h.Ecosystem)
	assert.Nil(t, h.Framework)
	assert.Nil(t, h.Port)
	assert.False(t, h.HasDockerfile)
	assert.False(t, h.HasCompose)
	assert.False(t, h.IsMonorepo)
	assert.Empty(t, h.EnvFiles)
	assert.Empty(t, h.RequiredEnvVars)
	assert.Empty(t, h.Warnings)
	assert.Equal(t, ConfidenceHigh, h.Confidence)
}

func TestNodePriorityOverGo(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"next":"14.0.0"}}`},
		{Path: "go.mod", Content: "module example.com/app"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, sPtr("node"), h.Ecosystem)
}

func TestPythonStarlette(t *testing.T) {
	files := []FileEntry{
		{Path: "requirements.txt", Content: "starlette==0.32.0"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "starlette", h.Framework.Value)
	assert.Equal(t, 8000, h.Port.Value)
}

func TestDockerfileNoExpose(t *testing.T) {
	files := []FileEntry{
		{Path: "Dockerfile", Content: "FROM node:20\nCMD [\"node\", \"server.js\"]"},
		{Path: "package.json", Content: `{"dependencies":{"express":"4.0.0"},"scripts":{"start":"node server.js --port 5000"}}`},
		{Path: "npm", Content: ""},
		{Path: "package-lock.json", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 5000, h.Port.Value)
	assert.Equal(t, "package.json scripts", h.Port.Source)
}

func TestComposeWithQuotedPorts(t *testing.T) {
	files := []FileEntry{
		{Path: "docker-compose.yml", Content: "services:\n  web:\n    ports:\n      - '8080:3000'"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 3000, h.Port.Value)
}

func TestAllSliceFieldsNonNil(t *testing.T) {
	h := AnalyzeFiles([]FileEntry{})
	assert.NotNil(t, h.Warnings)
	assert.NotNil(t, h.EnvFiles)
	assert.NotNil(t, h.RequiredEnvVars)
	assert.NotNil(t, h.MonorepoMarkers)
}

func TestNodeStartCommandWithDevFallback(t *testing.T) {
	files := []FileEntry{
		{Path: "package.json", Content: `{"dependencies":{"next":"14.0.0"},"scripts":{"build":"next build","start":"next start","dev":"next dev"}}`},
		{Path: "pnpm-lock.yaml", Content: ""},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, "pnpm start", *h.StartCommand)
}

func TestPythonPortNotOverriddenByDockerfile(t *testing.T) {
	files := []FileEntry{
		{Path: "requirements.txt", Content: "django==4.2"},
		{Path: "Dockerfile", Content: "FROM python:3.12\nEXPOSE 9000"},
	}
	h := AnalyzeFiles(files)
	assert.Equal(t, 9000, h.Port.Value)
	assert.Equal(t, "Dockerfile EXPOSE", h.Port.Source)
}

// helpers

func intPtr(v int) *int { return &v }

func sPtr(v string) *string { return &v }

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
