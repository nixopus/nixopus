APP_NAME=nixopus-api

build:
	@go build -o bin/$(APP_NAME)

run: build
	@./bin/$(APP_NAME)

vet:
	@go vet ./...

test: vet
	@go test -race -p 1 ./internal/features/... -v -count=1

test-integration:
	@go test -race -p 1 -tags integration ./internal/tests/... -v -count=1

test-all: vet
	@go test -race -p 1 ./... -v -count=1

test-routes:
	@go test -race -p 1 ./internal/tests/routes/... -v -count=1

cover:
	@go test -race -coverprofile=coverage.out -covermode=atomic ./internal/features/... -count=1 -timeout 300s
	@go tool cover -func=coverage.out | tail -1
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Full report: coverage.html"

cover-check: cover
	@TOTAL=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $$TOTAL%"; \
	if awk "BEGIN {exit !($$TOTAL + 0 < 95)}"; then \
		echo "FAIL: Coverage $$TOTAL% is below 95% threshold"; \
		exit 1; \
	fi; \
	echo "PASS: Coverage $$TOTAL% meets the 95% threshold"

cover-domain:
	@if [ -z "$(DOMAIN)" ]; then echo "Usage: make cover-domain DOMAIN=deploy"; exit 1; fi
	@go test -race -coverprofile=$(DOMAIN)_coverage.out -covermode=atomic ./internal/features/$(DOMAIN)/... -count=1 -timeout 120s
	@go tool cover -func=$(DOMAIN)_coverage.out | grep -v "_test.go"
	@echo "---"
	@go tool cover -func=$(DOMAIN)_coverage.out | tail -1

cover-auth:
	@go test -race -count=1 -cover ./internal/features/auth/...

format:
	@echo "Formatting Go files..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: vet build run test test-integration test-all test-routes cover cover-check cover-domain cover-auth format
