package strategies

import (
	"context"
	"path/filepath"

	"github.com/raghavyuva/nixopus-api/internal/live/sftp"
)

type PythonStrategy struct {
	Framework  string
	UsesPoetry bool
	UsesPipenv bool
	UsesUv     bool
	MainFile   string
	CustomPort int
}

func NewPythonStrategy() *PythonStrategy {
	return &PythonStrategy{
		Framework:  "python",
		UsesPoetry: false,
		UsesPipenv: false,
		UsesUv:     false,
		MainFile:   "main.py",
		CustomPort: 0,
	}
}

func NewFlaskStrategy() *PythonStrategy {
	return &PythonStrategy{
		Framework:  "flask",
		UsesPoetry: false,
		UsesPipenv: false,
		UsesUv:     false,
		MainFile:   "app.py",
		CustomPort: 5000,
	}
}

func NewDjangoStrategy() *PythonStrategy {
	return &PythonStrategy{
		Framework:  "django",
		UsesPoetry: false,
		UsesPipenv: false,
		UsesUv:     false,
		MainFile:   "manage.py",
		CustomPort: 8000,
	}
}

func NewFastAPIStrategy() *PythonStrategy {
	return &PythonStrategy{
		Framework:  "fastapi",
		UsesPoetry: false,
		UsesPipenv: false,
		UsesUv:     false,
		MainFile:   "main.py",
		CustomPort: 8000,
	}
}

func (s *PythonStrategy) Name() string {
	return s.Framework
}

func (s *PythonStrategy) GetBaseImage() string {
	return "python:3.12-slim"
}

func (s *PythonStrategy) GetInstallCommand() []string {
	if s.UsesUv {
		return []string{"sh", "-c", "pip install uv && uv sync"}
	}
	if s.UsesPoetry {
		return []string{"sh", "-c", "pip install poetry && poetry config virtualenvs.create false && poetry install --no-interaction"}
	}
	if s.UsesPipenv {
		return []string{"sh", "-c", "pip install pipenv && pipenv install --deploy --system"}
	}
	return []string{"sh", "-c", "if [ -f requirements.txt ]; then pip install -r requirements.txt; elif [ -f pyproject.toml ]; then pip install -e .; fi"}
}

func (s *PythonStrategy) GetDevCommand() []string {
	switch s.Framework {
	case "django":
		return []string{"python", "manage.py", "runserver", "0.0.0.0:8000"}
	case "flask":
		return []string{"sh", "-c", "flask run --host=0.0.0.0 --port=5000 --reload"}
	case "fastapi":
		appModule := s.getAppModule()
		return []string{"sh", "-c", "uvicorn " + appModule + ":app --host 0.0.0.0 --port 8000 --reload"}
	default:
		return []string{"python", s.MainFile}
	}
}

func (s *PythonStrategy) getAppModule() string {
	mainFile := s.MainFile
	if mainFile == "" {
		return "main"
	}
	if len(mainFile) > 3 && mainFile[len(mainFile)-3:] == ".py" {
		mainFile = mainFile[:len(mainFile)-3]
	}
	result := ""
	for _, c := range mainFile {
		if c == '/' || c == '\\' {
			result += "."
		} else {
			result += string(c)
		}
	}
	return result
}

func (s *PythonStrategy) GetDefaultPort() int {
	if s.CustomPort > 0 {
		return s.CustomPort
	}
	switch s.Framework {
	case "django":
		return 8000
	case "flask":
		return 5000
	case "fastapi":
		return 8000
	default:
		return 8000
	}
}

func (s *PythonStrategy) GetEnvVars() map[string]string {
	envVars := map[string]string{
		"PYTHONDONTWRITEBYTECODE": "1",
		"PYTHONUNBUFFERED":        "1",
		"HOST":                    "0.0.0.0",
	}

	switch s.Framework {
	case "flask":
		envVars["FLASK_ENV"] = "development"
		envVars["FLASK_DEBUG"] = "1"
		envVars["FLASK_APP"] = s.MainFile
	case "django":
		envVars["DJANGO_DEBUG"] = "True"
		envVars["DJANGO_SETTINGS_MODULE"] = "config.settings"
	case "fastapi":
		envVars["DEBUG"] = "1"
	}

	return envVars
}

func (s *PythonStrategy) GetHealthCheckPath() string {
	switch s.Framework {
	case "django":
		return "/admin/"
	default:
		return "/"
	}
}

func (s *PythonStrategy) NeedsPolling() bool {
	return true
}

func (s *PythonStrategy) GetWorkdir() string {
	return "/app"
}

func (s *PythonStrategy) GetReadyLogPattern() string {
	switch s.Framework {
	case "django":
		return `(?i)(Starting development server|Quit the server with|Watching for file changes)`
	case "flask":
		return `(?i)(Running on|Debugger is active|Restarting with)`
	case "fastapi":
		return `(?i)(Uvicorn running on|Started reloader process|Application startup complete)`
	default:
		return `(?i)(running|started|listening|ready|serving)`
	}
}

// DetectMainFile finds the main Python file in the project.
func (s *PythonStrategy) DetectMainFile(ctx context.Context, projectPath string) string {
	candidates := []string{
		"main.py", "app.py", "run.py", "server.py", "wsgi.py", "asgi.py",
		"src/main.py", "src/app.py", "app/main.py", "app/__init__.py",
	}

	for _, candidate := range candidates {
		if fileExists(ctx, filepath.Join(projectPath, candidate)) {
			return candidate
		}
	}
	return s.MainFile // Return the default from the strategy
}

// fileExists checks if a file exists at the given path via SFTP.
func fileExists(ctx context.Context, path string) bool {
	return sftp.FileExists(ctx, path)
}
