package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raghavyuva/octopack/packer"
)

// TestMain_StdinStdout tests reading from stdin and writing to stdout
func TestMain_StdinStdout(t *testing.T) {
	simpleSpec := `{
		"stages": [
			{
				"from": {
					"image": "alpine",
					"tag": "latest"
				},
				"instructions": [
					{
						"type": "run",
						"command": "echo hello"
					}
				]
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-validate=false")
	cmd.Stdin = strings.NewReader(simpleSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "FROM alpine:latest") {
		t.Errorf("Expected Dockerfile to contain 'FROM alpine:latest', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "RUN") || !strings.Contains(outputStr, "echo hello") {
		t.Errorf("Expected Dockerfile to contain RUN instruction with 'echo hello', got: %s", outputStr)
	}
}

// TestMain_InputFile tests reading from input file
func TestMain_InputFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.json")

	simpleSpec := `{
		"stages": [
			{
				"from": {
					"image": "ubuntu",
					"tag": "22.04"
				},
				"instructions": [
					{
						"type": "workdir",
						"path": "/app"
					}
				]
			}
		]
	}`

	if err := os.WriteFile(inputFile, []byte(simpleSpec), 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	cmd := exec.Command("go", "run", ".", "-input", inputFile, "-validate=false")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "FROM ubuntu:22.04") {
		t.Errorf("Expected Dockerfile to contain 'FROM ubuntu:22.04', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "WORKDIR /app") {
		t.Errorf("Expected Dockerfile to contain 'WORKDIR /app', got: %s", outputStr)
	}
}

// TestMain_OutputFile tests writing to output file
func TestMain_OutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.json")
	outputFile := filepath.Join(tmpDir, "Dockerfile")

	simpleSpec := `{
		"stages": [
			{
				"from": {
					"image": "node",
					"tag": "20"
				},
				"instructions": [
					{
						"type": "cmd",
						"command": ["node", "index.js"]
					}
				]
			}
		]
	}`

	if err := os.WriteFile(inputFile, []byte(simpleSpec), 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	cmd := exec.Command("go", "run", ".", "-input", inputFile, "-output", outputFile, "-validate=false")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "FROM node:20") {
		t.Errorf("Expected Dockerfile to contain 'FROM node:20', got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "CMD [\"node\", \"index.js\"]") {
		t.Errorf("Expected Dockerfile to contain CMD instruction, got: %s", contentStr)
	}
}

// TestMain_ValidationEnabled_ValidSpec tests validation with valid spec
func TestMain_ValidationEnabled_ValidSpec(t *testing.T) {
	validSpec := `{
		"stages": [
			{
				"from": {
					"image": "alpine",
					"tag": "latest"
				},
				"instructions": [
					{
						"type": "run",
						"command": "echo test"
					}
				]
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-validate=true")
	cmd.Stdin = strings.NewReader(validSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "FROM alpine:latest") {
		t.Errorf("Expected Dockerfile to contain 'FROM alpine:latest', got: %s", outputStr)
	}
}

// TestMain_ValidationEnabled_InvalidSpec tests validation with invalid spec (should fail)
func TestMain_ValidationEnabled_InvalidSpec(t *testing.T) {
	invalidSpec := `{
		"stages": []
	}`

	cmd := exec.Command("go", "run", ".", "-validate=true")
	cmd.Stdin = strings.NewReader(invalidSpec)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with invalid spec, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Validation errors") {
		t.Errorf("Expected validation error message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "at least one stage must be defined") {
		t.Errorf("Expected specific validation error, got: %s", outputStr)
	}
}

// TestMain_ValidationEnabled_Warnings tests validation with warnings
func TestMain_ValidationEnabled_Warnings(t *testing.T) {
	// Spec with deprecated MAINTAINER instruction (should generate warning)
	specWithWarning := `{
		"stages": [
			{
				"from": {
					"image": "alpine",
					"tag": "latest"
				},
				"instructions": [
					{
						"type": "maintainer",
						"name": "Test Maintainer"
					}
				]
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-validate=true")
	cmd.Stdin = strings.NewReader(specWithWarning)
	output, err := cmd.CombinedOutput()
	// Command should succeed even with warnings
	if err != nil {
		t.Fatalf("Command should succeed with warnings, but failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should still generate Dockerfile
	if !strings.Contains(outputStr, "FROM alpine:latest") {
		t.Errorf("Expected Dockerfile to contain 'FROM alpine:latest', got: %s", outputStr)
	}
}

// TestMain_ValidationDisabled tests with validation disabled
func TestMain_ValidationDisabled(t *testing.T) {
	invalidSpec := `{
		"stages": []
	}`

	cmd := exec.Command("go", "run", ".", "-validate=false")
	cmd.Stdin = strings.NewReader(invalidSpec)
	output, err := cmd.CombinedOutput()
	// Should succeed without validation
	if err != nil {
		t.Fatalf("Command should succeed without validation, but failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should generate empty Dockerfile (just FROM without stages)
	if outputStr == "" {
		t.Error("Expected some output, got empty string")
	}
}

// TestMain_InvalidJSON tests handling of invalid JSON
func TestMain_InvalidJSON(t *testing.T) {
	invalidJSON := `{
		"stages": [
			{
				"from": {
					"image": "alpine"
					// missing comma
					"tag": "latest"
				}
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-validate=false")
	cmd.Stdin = strings.NewReader(invalidJSON)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with invalid JSON, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error parsing JSON") {
		t.Errorf("Expected JSON parsing error, got: %s", outputStr)
	}
}

// TestMain_InputFileNotFound tests handling of non-existent input file
func TestMain_InputFileNotFound(t *testing.T) {
	nonExistentFile := "/tmp/nonexistent-file-12345.json"

	cmd := exec.Command("go", "run", ".", "-input", nonExistentFile, "-validate=false")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with non-existent file, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error opening input file") {
		t.Errorf("Expected file error message, got: %s", outputStr)
	}
}

// TestMain_ComplexDockerfile tests a very complex Dockerfile with all features
func TestMain_ComplexDockerfile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "complex.json")
	outputFile := filepath.Join(tmpDir, "Dockerfile")

	// Use JSON string directly to avoid marshaling issues with OnbuildInstruction
	complexSpecJSON := `{
		"directives": {
			"syntax": "docker/dockerfile:1",
			"escape": "\\",
			"check": {
				"skip": ["JSONArgsRecommended", "EmptyContinuationLine"],
				"error": false
			}
		},
		"global_args": [
			{
				"name": "VERSION",
				"default_value": "latest"
			},
			{
				"name": "BUILDPLATFORM",
				"default_value": "linux/amd64"
			}
		],
		"stages": [
			{
				"from": {
					"image": "golang",
					"tag": "1.21-alpine",
					"as": "builder"
				},
				"instructions": [
					{
						"type": "workdir",
						"path": "/app"
					},
					{
						"type": "arg",
						"name": "GO_VERSION",
						"default_value": "1.21"
					},
					{
						"type": "env",
						"variables": {
							"CGO_ENABLED": "0",
							"GOOS": "linux",
							"GOARCH": "amd64"
						}
					},
					{
						"type": "copy",
						"sources": ["go.mod", "go.sum"],
						"destination": "./"
					},
					{
						"type": "run",
						"command": "go mod download",
						"mount": [
							{
								"type": "cache",
								"target": "/go/pkg/mod",
								"id": "gomod",
								"sharing": "shared"
							}
						]
					},
					{
						"type": "copy",
						"sources": ["."],
						"destination": "./",
						"exclude": [".git", "node_modules"]
					},
					{
						"type": "run",
						"command": ["go", "build", "-o", "/app/main", "."],
						"mount": [
							{
								"type": "cache",
								"target": "/root/.cache/go-build",
								"id": "gobuild",
								"sharing": "shared"
							}
						],
						"network": "default"
					}
				]
			},
			{
				"from": {
					"image": "alpine",
					"tag": "3.19",
					"platform": "linux/amd64"
				},
				"instructions": [
					{
						"type": "workdir",
						"path": "/app"
					},
					{
						"type": "run",
						"command": "apk add --no-cache ca-certificates tzdata",
						"network": "default"
					},
					{
						"type": "copy",
						"from": "builder",
						"sources": ["/app/main"],
						"destination": "./",
						"chmod": "0755"
					},
					{
						"type": "user",
						"user": "1000",
						"group": "1000"
					},
					{
						"type": "expose",
						"ports": [
							{
								"port": 8080,
								"protocol": "tcp"
							},
							{
								"port": 9090,
								"protocol": "tcp"
							}
						]
					},
					{
						"type": "volume",
						"paths": ["/app/data", "/app/logs"]
					},
					{
						"type": "label",
						"labels": {
							"maintainer": "octopack@example.com",
							"version": "1.0.0",
							"description": "Complex test application"
						}
					},
					{
						"type": "healthcheck",
						"command": ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"],
						"interval": "30s",
						"timeout": "10s",
						"start_period": "40s",
						"retries": 3
					},
					{
						"type": "stopsignal",
						"signal": "SIGTERM"
					},
					{
						"type": "entrypoint",
						"command": ["./main"]
					},
					{
						"type": "cmd",
						"command": ["--config", "/app/config.yaml"],
						"is_default_args": true
					}
				]
			},
			{
				"from": {
					"image": "ubuntu",
					"tag": "22.04"
				},
				"instructions": [
					{
						"type": "shell",
						"shell": ["/bin/bash", "-c"]
					},
					{
						"type": "run",
						"command": "apt-get update && apt-get install -y curl",
						"security": "sandbox"
					},
					{
						"type": "run",
						"command": ["sh", "-c", "echo 'Using exec form with shell'"],
						"mount": [
							{
								"type": "bind",
								"target": "/mnt/data",
								"source": "/host/data",
								"from": "builder"
							},
							{
								"type": "secret",
								"target": "/run/secrets/api-key",
								"id": "api-key",
								"mode": "0600",
								"required": true
							},
							{
								"type": "ssh",
								"target": "/root/.ssh",
								"id": "default",
								"mode": "0600",
								"required": false
							},
							{
								"type": "cache",
								"target": "/var/cache",
								"id": "cache-id",
								"uid": 1000,
								"gid": 1000,
								"sharing": "private"
							}
						]
					},
					{
						"type": "add",
						"sources": ["https://example.com/file.tar.gz"],
						"destination": "/tmp/",
						"checksum": "sha256:abc123...",
						"unpack": true,
						"chown": "root:root"
					},
					{
						"type": "maintainer",
						"name": "Example Maintainer <maintainer@example.com>"
					},
					{
						"type": "onbuild",
						"instruction": {
							"type": "copy",
							"sources": ["config.json"],
							"destination": "/app/config.json"
						}
					}
				]
			}
		]
	}`

	if err := os.WriteFile(inputFile, []byte(complexSpecJSON), 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	cmd := exec.Command("go", "run", ".", "-input", inputFile, "-output", outputFile, "-validate=true")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify output file was created
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	
	// Verify all major components are present
	checks := []string{
		"# syntax=",
		"# escape=",
		"# check=",
		"ARG VERSION",
		"FROM golang",
		"WORKDIR",
		"ENV",
		"COPY",
		"RUN",
		"ARG",
		"USER",
		"EXPOSE",
		"VOLUME",
		"LABEL",
		"HEALTHCHECK",
		"STOPSIGNAL",
		"ENTRYPOINT",
		"CMD",
		"ADD",
		"SHELL",
		"ONBUILD",
		"MAINTAINER",
	}

	for _, check := range checks {
		if !strings.Contains(contentStr, check) {
			t.Errorf("Expected Dockerfile to contain '%s', but it was missing", check)
		}
	}

	// Verify multi-stage build
	fromCount := strings.Count(contentStr, "FROM")
	if fromCount < 3 {
		t.Errorf("Expected at least 3 FROM instructions (multi-stage build), got %d", fromCount)
	}

	// Verify mount options
	if !strings.Contains(contentStr, "--mount=") {
		t.Error("Expected Dockerfile to contain mount options")
	}

	// Verify platform specification
	if !strings.Contains(contentStr, "--platform=") {
		t.Error("Expected Dockerfile to contain platform specification")
	}
}

// TestMain_ComplexDockerfile_Validation tests complex Dockerfile with validation
func TestMain_ComplexDockerfile_Validation(t *testing.T) {
	// Use the same complex spec JSON as TestMain_ComplexDockerfile
	complexSpecJSON := `{
		"directives": {
			"syntax": "docker/dockerfile:1",
			"escape": "\\"
		},
		"global_args": [
			{
				"name": "VERSION",
				"default_value": "latest"
			}
		],
		"stages": [
			{
				"from": {
					"image": "golang",
					"tag": "1.21-alpine",
					"as": "builder"
				},
				"instructions": [
					{
						"type": "workdir",
						"path": "/app"
					},
					{
						"type": "run",
						"command": "go mod download"
					}
				]
			},
			{
				"from": {
					"image": "alpine",
					"tag": "latest"
				},
				"instructions": [
					{
						"type": "copy",
						"from": "builder",
						"sources": ["/app"],
						"destination": "/app"
					}
				]
			}
		]
	}`

	// Validate the spec programmatically first
	var spec packer.DockerfileSpec
	if err := json.Unmarshal([]byte(complexSpecJSON), &spec); err != nil {
		t.Fatalf("Failed to unmarshal spec: %v", err)
	}

	result := packer.Validate(&spec)
	if result.HasErrors() {
		t.Fatalf("Complex spec has validation errors: %v", result.Errors)
	}

	// Test via command line
	cmd := exec.Command("go", "run", ".", "-validate=true")
	cmd.Stdin = strings.NewReader(complexSpecJSON)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Fatal("Expected non-empty Dockerfile output")
	}

	// Verify it's a valid Dockerfile structure
	lines := strings.Split(outputStr, "\n")
	hasFrom := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "FROM") {
			hasFrom = true
			break
		}
	}
	if !hasFrom {
		t.Error("Generated Dockerfile does not contain FROM instruction")
	}
}

// TestMain_AllInstructionTypes tests all instruction types are handled
func TestMain_AllInstructionTypes(t *testing.T) {
	allTypesSpec := `{
		"directives": {
			"syntax": "docker/dockerfile:1",
			"escape": "\\",
			"check": {
				"skip": ["JSONArgsRecommended"],
				"error": false
			}
		},
		"global_args": [
			{
				"name": "BUILD_VERSION",
				"default_value": "1.0.0"
			}
		],
		"stages": [
			{
				"from": {
					"image": "alpine",
					"tag": "latest",
					"as": "builder"
				},
				"instructions": [
					{
						"type": "arg",
						"name": "APP_NAME",
						"default_value": "myapp"
					},
					{
						"type": "env",
						"variables": {
							"PATH": "/usr/local/bin:$PATH",
							"NODE_ENV": "production"
						}
					},
					{
						"type": "workdir",
						"path": "/app"
					},
					{
						"type": "run",
						"command": "apk add --no-cache curl",
						"network": "default"
					},
					{
						"type": "run",
						"command": ["sh", "-c", "echo 'Exec form'"],
						"mount": [
							{
								"type": "cache",
								"target": "/var/cache/apk",
								"id": "apk-cache",
								"sharing": "shared"
							}
						]
					},
					{
						"type": "copy",
						"sources": ["package.json"],
						"destination": "./",
						"chmod": "0644"
					},
					{
						"type": "copy",
						"sources": ["src"],
						"destination": "./src",
						"chown": "1000:1000",
						"exclude": [".git"]
					},
					{
						"type": "add",
						"sources": ["https://example.com/file.tar.gz"],
						"destination": "/tmp/",
						"checksum": "sha256:abc123",
						"unpack": true
					},
					{
						"type": "user",
						"user": "appuser",
						"group": "appgroup"
					},
					{
						"type": "expose",
						"ports": [
							{
								"port": 8080,
								"protocol": "tcp"
							},
							{
								"port": 9090,
								"protocol": "udp"
							}
						]
					},
					{
						"type": "volume",
						"paths": ["/app/data", "/app/logs"]
					},
					{
						"type": "label",
						"labels": {
							"version": "1.0.0",
							"maintainer": "test@example.com"
						}
					},
					{
						"type": "healthcheck",
						"command": ["CMD", "curl", "-f", "http://localhost:8080/health"],
						"interval": "30s",
						"timeout": "10s",
						"start_period": "40s",
						"retries": 3
					},
					{
						"type": "stopsignal",
						"signal": "SIGTERM"
					},
					{
						"type": "shell",
						"shell": ["/bin/sh", "-c"]
					},
					{
						"type": "entrypoint",
						"command": ["./app"]
					},
					{
						"type": "cmd",
						"command": ["--config", "/app/config.yaml"],
						"is_default_args": true
					},
					{
						"type": "onbuild",
						"instruction": {
							"type": "copy",
							"sources": ["config.json"],
							"destination": "/app/"
						}
					},
					{
						"type": "maintainer",
						"name": "Test Maintainer <test@example.com>"
					}
				]
			},
			{
				"from": {
					"image": "alpine",
					"tag": "latest",
					"platform": "linux/amd64"
				},
				"instructions": [
					{
						"type": "copy",
						"from": "builder",
						"sources": ["/app"],
						"destination": "/app",
						"link": true,
						"parents": true
					}
				]
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-validate=true")
	cmd.Stdin = strings.NewReader(allTypesSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	
	// Verify all instruction types are present
	instructionChecks := map[string]string{
		"ARG":        "ARG BUILD_VERSION",
		"FROM":       "FROM alpine",
		"ENV":        "ENV",
		"WORKDIR":    "WORKDIR",
		"RUN":        "RUN",
		"COPY":       "COPY",
		"ADD":        "ADD",
		"USER":       "USER",
		"EXPOSE":     "EXPOSE",
		"VOLUME":     "VOLUME",
		"LABEL":      "LABEL",
		"HEALTHCHECK": "HEALTHCHECK",
		"STOPSIGNAL": "STOPSIGNAL",
		"SHELL":      "SHELL",
		"ENTRYPOINT": "ENTRYPOINT",
		"CMD":        "CMD",
		"ONBUILD":    "ONBUILD",
		"MAINTAINER": "MAINTAINER",
	}

	for instruction, check := range instructionChecks {
		if !strings.Contains(outputStr, check) {
			t.Errorf("Expected Dockerfile to contain %s instruction (%s), but it was missing", instruction, check)
		}
	}
}

// TestMain_DefaultFlags tests default flag values
func TestMain_DefaultFlags(t *testing.T) {
	// Test that validation is enabled by default
	invalidSpec := `{
		"stages": []
	}`

	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader(invalidSpec)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with validation enabled by default, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Validation errors") {
		t.Errorf("Expected validation error (validation should be enabled by default), got: %s", outputStr)
	}
}

// TestMain_ComplexExampleFile tests using the complex-example.json file
func TestMain_ComplexExampleFile(t *testing.T) {
	// Get the path to the testdata directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Go up from cmd/ to octopack/ then to testdata/
	projectRoot := filepath.Dir(wd)
	complexFile := filepath.Join(projectRoot, "testdata", "complex-example.json")
	
	// Check if file exists
	if _, err := os.Stat(complexFile); os.IsNotExist(err) {
		t.Skipf("Complex example file not found at %s", complexFile)
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "Dockerfile")

	cmd := exec.Command("go", "run", ".", "-input", complexFile, "-output", outputFile, "-validate=true")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify output file was created
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Fatal("Expected non-empty Dockerfile output")
	}

	// Verify it contains multiple stages
	fromCount := strings.Count(contentStr, "FROM")
	if fromCount < 5 {
		t.Errorf("Expected at least 5 FROM instructions (multi-stage build), got %d", fromCount)
	}

	// Verify complex features
	checks := []string{
		"# syntax=",
		"# escape=",
		"# check=",
		"ARG VERSION",
		"golang",
		"node",
		"alpine",
		"nginx",
		"ubuntu",
		"python",
		"--mount=",
		"--platform=",
		"HEALTHCHECK",
		"ONBUILD",
	}

	for _, check := range checks {
		if !strings.Contains(contentStr, check) {
			t.Errorf("Expected Dockerfile to contain '%s', but it was missing. Content length: %d", check, len(contentStr))
		}
	}
}

// TestMain_OutputFileWriteError tests handling of write errors
func TestMain_OutputFileWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a directory path instead of a file path to cause write error
	outputFile := filepath.Join(tmpDir, "nonexistent", "subdir", "Dockerfile")

	simpleSpec := `{
		"stages": [
			{
				"from": {
					"image": "alpine",
					"tag": "latest"
				}
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-output", outputFile, "-validate=false")
	cmd.Stdin = strings.NewReader(simpleSpec)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with invalid output path, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error writing output file") {
		t.Errorf("Expected write error message, got: %s", outputStr)
	}
}

// TestMain_ComplexMountOptions tests complex mount options
func TestMain_ComplexMountOptions(t *testing.T) {
	complexMountSpec := `{
		"stages": [
			{
				"from": {
					"image": "alpine",
					"tag": "latest"
				},
				"instructions": [
					{
						"type": "run",
						"command": "echo test",
						"mount": [
							{
								"type": "cache",
								"target": "/cache",
								"id": "test-cache",
								"sharing": "private",
								"uid": 1000,
								"gid": 1000
							},
							{
								"type": "secret",
								"target": "/run/secrets/key",
								"id": "secret-key",
								"mode": "0600",
								"required": true
							},
							{
								"type": "ssh",
								"target": "/root/.ssh",
								"id": "default",
								"mode": "0600",
								"required": false
							},
							{
								"type": "bind",
								"target": "/mnt/data",
								"source": "/host/data",
								"from": "builder",
								"readonly": true
							},
							{
								"type": "tmpfs",
								"target": "/tmp",
								"uid": 1000,
								"gid": 1000
							}
						]
					}
				]
			}
		]
	}`

	cmd := exec.Command("go", "run", ".", "-validate=true")
	cmd.Stdin = strings.NewReader(complexMountSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "--mount=") {
		t.Error("Expected Dockerfile to contain mount options")
	}
}

// createComplexSpec creates a very complex DockerfileSpec for testing
// NOTE: This function is kept for reference but complex tests now use JSON strings
// to avoid marshaling issues with OnbuildInstruction
func createComplexSpec() *packer.DockerfileSpec {
	uid := 1000
	gid := 1000
	retries := 3

	return &packer.DockerfileSpec{
		Directives: &packer.ParserDirectives{
			Syntax: "docker/dockerfile:1",
			Escape: "\\",
			Check: &packer.CheckDirective{
				Skip:  []string{"JSONArgsRecommended", "EmptyContinuationLine"},
				Error: false,
			},
		},
		GlobalArgs: []packer.ArgInstruction{
			{Name: "VERSION", DefaultValue: "latest"},
			{Name: "BUILDPLATFORM", DefaultValue: "linux/amd64"},
		},
		Stages: []packer.Stage{
			{
				From: packer.FromInstruction{
					Image: "golang",
					Tag:   "1.21-alpine",
					As:    "builder",
				},
				Instructions: []packer.Instruction{
					packer.WorkdirInstruction{Path: "/app"},
					packer.ArgInstruction{Name: "GO_VERSION", DefaultValue: "1.21"},
					packer.EnvInstruction{
						Variables: map[string]string{
							"CGO_ENABLED": "0",
							"GOOS":        "linux",
							"GOARCH":      "amd64",
						},
					},
					packer.CopyInstruction{
						Sources:     []string{"go.mod", "go.sum"},
						Destination: "./",
					},
					packer.RunInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: "go mod download",
						},
						Mount: []packer.MountOption{
							{
								Type:    "cache",
								Target:  "/go/pkg/mod",
								ID:      "gomod",
								Sharing: "shared",
							},
						},
					},
					packer.CopyInstruction{
						Sources:     []string{"."},
						Destination: "./",
						Exclude:     []string{".git", "node_modules"},
					},
					packer.RunInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: []string{"go", "build", "-o", "/app/main", "."},
						},
						Mount: []packer.MountOption{
							{
								Type:    "cache",
								Target:  "/root/.cache/go-build",
								ID:      "gobuild",
								Sharing: "shared",
							},
						},
						Network: "default",
					},
				},
			},
			{
				From: packer.FromInstruction{
					Image:    "alpine",
					Tag:      "3.19",
					Platform: "linux/amd64",
				},
				Instructions: []packer.Instruction{
					packer.WorkdirInstruction{Path: "/app"},
					packer.RunInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: "apk add --no-cache ca-certificates tzdata",
						},
						Network: "default",
					},
					packer.CopyInstruction{
						From:        "builder",
						Sources:     []string{"/app/main"},
						Destination: "./",
						Chmod:       "0755",
					},
					packer.UserInstruction{
						User:  "1000",
						Group: "1000",
					},
					packer.ExposeInstruction{
						Ports: []packer.PortDefinition{
							{Port: 8080, Protocol: "tcp"},
							{Port: 9090, Protocol: "tcp"},
						},
					},
					packer.VolumeInstruction{
						Paths: []string{"/app/data", "/app/logs"},
					},
					packer.LabelInstruction{
						Labels: map[string]string{
							"maintainer": "octopack@example.com",
							"version":    "1.0.0",
							"description": "Complex test application",
						},
					},
					packer.HealthcheckInstruction{
						Command:     []interface{}{"CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"},
						Interval:    "30s",
						Timeout:     "10s",
						StartPeriod: "40s",
						Retries:     &retries,
					},
					packer.StopsignalInstruction{
						Signal: "SIGTERM",
					},
					packer.EntrypointInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: []string{"./main"},
						},
					},
					packer.CmdInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: []string{"--config", "/app/config.yaml"},
						},
						IsDefaultArgs: true,
					},
				},
			},
			{
				From: packer.FromInstruction{
					Image: "ubuntu",
					Tag:   "22.04",
				},
				Instructions: []packer.Instruction{
					packer.ShellInstruction{
						Shell: []string{"/bin/bash", "-c"},
					},
					packer.RunInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: "apt-get update && apt-get install -y curl",
						},
						Security: "sandbox",
					},
					packer.RunInstruction{
						BaseInstruction: packer.BaseInstruction{
							Command: []string{"sh", "-c", "echo 'Using exec form with shell'"},
						},
						Mount: []packer.MountOption{
							{
								Type:   "bind",
								Target: "/mnt/data",
								Source: "/host/data",
								From:   "builder",
							},
							{
								Type:     "secret",
								Target:   "/run/secrets/api-key",
								ID:       "api-key",
								Mode:     "0600",
								Required: true,
							},
							{
								Type:     "ssh",
								Target:   "/root/.ssh",
								ID:       "default",
								Mode:     "0600",
								Required: false,
							},
							{
								Type:  "cache",
								Target: "/var/cache",
								ID:    "cache-id",
								UID:   &uid,
								GID:   &gid,
								Sharing: "private",
							},
						},
					},
					packer.AddInstruction{
						CopyInstruction: packer.CopyInstruction{
							Sources:     []string{"https://example.com/file.tar.gz"},
							Destination: "/tmp/",
							Chown:       "root:root",
						},
						Checksum: "sha256:abc123...",
						Unpack:   true,
					},
					packer.MaintainerInstruction{
						Name: "Example Maintainer <maintainer@example.com>",
					},
					packer.OnbuildInstruction{
						Instruction: packer.CopyInstruction{
							Sources:     []string{"config.json"},
							Destination: "/app/config.json",
						},
					},
				},
			},
		},
	}
}

