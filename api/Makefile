APP_NAME=nixopus-api
CLI_NAME=nixopus

build: 
	@go build -o bin/$(APP_NAME)

run: build
	@./bin/$(APP_NAME)

test:
	@go test -p 1 ./internal/features/... -v -count=1

test-all:
	@go test -p 1 ./... -v -count=1

test-routes:
	@go test -p 1 ./internal/tests/routes/... -v -count=1

cover-auth:
	@go test -count=1 -cover ./internal/features/auth/...

format:
	@echo "Formatting Go files..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: format build run test test-all test-routes cover-auth
