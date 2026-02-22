APP_NAME=nixopus-api
CLI_NAME=nixopus

build: 
	@go build -o bin/$(APP_NAME)

run: build
	@./bin/$(APP_NAME)

# CLI build requires build-time config. Set vars in env or: make cli-build-dev
# Example: make cli-build CLI_API_URL=https://api.nixopus.com ...
CLI_PKG := github.com/raghavyuva/nixopus-api/pkg/cli/cliconfig
CLI_LDFLAGS := -X $(CLI_PKG).APIURL=$(CLI_API_URL) -X $(CLI_PKG).BetterAuthURL=$(CLI_BETTER_AUTH_URL) -X $(CLI_PKG).FrontendURL=$(CLI_FRONTEND_URL) -X $(CLI_PKG).OAuthClientID=$(CLI_OAUTH_CLIENT_ID) -X $(CLI_PKG).AgentEndpoint=$(CLI_AGENT_ENDPOINT) -X $(CLI_PKG).InitTimeout=$(CLI_INIT_TIMEOUT) -X $(CLI_PKG).WorkflowTimeout=$(CLI_WORKFLOW_TIMEOUT) -X $(CLI_PKG).WorkflowID=$(CLI_WORKFLOW_ID) -X $(CLI_PKG).DebugStream=$(or $(CLI_DEBUG_STREAM),) -X $(CLI_PKG).ConfigFileName=$(or $(CLI_CONFIG_FILE),) -X $(CLI_PKG).AuthFileName=$(or $(CLI_AUTH_FILE),) -X $(CLI_PKG).SyncStateFileName=$(or $(CLI_SYNC_STATE_FILE),) \
	-X $(CLI_PKG).MoverSendBuffer=$(or $(CLI_MOVER_SEND_BUFFER),) -X $(CLI_PKG).MoverReceiveBuffer=$(or $(CLI_MOVER_RECEIVE_BUFFER),) -X $(CLI_PKG).MoverWriteWait=$(or $(CLI_MOVER_WRITE_WAIT),) -X $(CLI_PKG).MoverPongWait=$(or $(CLI_MOVER_PONG_WAIT),) -X $(CLI_PKG).MoverPingPeriod=$(or $(CLI_MOVER_PING_PERIOD),) -X $(CLI_PKG).MoverMaxMessageSize=$(or $(CLI_MOVER_MAX_MESSAGE_SIZE),) -X $(CLI_PKG).MoverHandshakeTimeout=$(or $(CLI_MOVER_HANDSHAKE_TIMEOUT),) -X $(CLI_PKG).MoverInitialReconnectDelay=$(or $(CLI_MOVER_INITIAL_RECONNECT_DELAY),) -X $(CLI_PKG).MoverMaxReconnectDelay=$(or $(CLI_MOVER_MAX_RECONNECT_DELAY),) -X $(CLI_PKG).MoverReconnectBackoffRate=$(or $(CLI_MOVER_RECONNECT_BACKOFF_RATE),) -X $(CLI_PKG).MoverMaxReconnectAttempts=$(or $(CLI_MOVER_MAX_RECONNECT_ATTEMPTS),) -X $(CLI_PKG).MoverCloseFlushDelay=$(or $(CLI_MOVER_CLOSE_FLUSH_DELAY),) -X $(CLI_PKG).MoverDebounceMs=$(or $(CLI_MOVER_DEBOUNCE_MS),) -X $(CLI_PKG).MoverLargeSyncThreshold=$(or $(CLI_MOVER_LARGE_SYNC_THRESHOLD),) -X $(CLI_PKG).MoverChunkSize=$(or $(CLI_MOVER_CHUNK_SIZE),) -X $(CLI_PKG).MoverManifestWaitTimeout=$(or $(CLI_MOVER_MANIFEST_WAIT_TIMEOUT),) -X $(CLI_PKG).MoverSyncWorkers=$(or $(CLI_MOVER_SYNC_WORKERS),) -X $(CLI_PKG).MoverSyncConcurrency=$(or $(CLI_MOVER_SYNC_CONCURRENCY),) -X $(CLI_PKG).MoverEventsBufferSize=$(or $(CLI_MOVER_EVENTS_BUFFER_SIZE),) -X $(CLI_PKG).MoverWatcherDebounceMs=$(or $(CLI_MOVER_WATCHER_DEBOUNCE_MS),) -X $(CLI_PKG).MoverGitCheckTimeout=$(or $(CLI_MOVER_GIT_CHECK_TIMEOUT),) -X $(CLI_PKG).MoverSyncStateDebounceMs=$(or $(CLI_MOVER_SYNC_STATE_DEBOUNCE_MS),)
cli-build:
	@test -n "$(CLI_API_URL)" || (echo "CLI build requires CLI_API_URL. Set at build time. See api/.env.cli"; exit 1)
	@test -n "$(CLI_BETTER_AUTH_URL)" || (echo "CLI build requires CLI_BETTER_AUTH_URL"; exit 1)
	@test -n "$(CLI_FRONTEND_URL)" || (echo "CLI build requires CLI_FRONTEND_URL"; exit 1)
	@test -n "$(CLI_OAUTH_CLIENT_ID)" || (echo "CLI build requires CLI_OAUTH_CLIENT_ID"; exit 1)
	@test -n "$(CLI_AGENT_ENDPOINT)" || (echo "CLI build requires CLI_AGENT_ENDPOINT"; exit 1)
	@test -n "$(CLI_INIT_TIMEOUT)" || (echo "CLI build requires CLI_INIT_TIMEOUT"; exit 1)
	@test -n "$(CLI_WORKFLOW_TIMEOUT)" || (echo "CLI build requires CLI_WORKFLOW_TIMEOUT"; exit 1)
	@test -n "$(CLI_WORKFLOW_ID)" || (echo "CLI build requires CLI_WORKFLOW_ID"; exit 1)
	@go build -ldflags "$(CLI_LDFLAGS)" -o bin/$(CLI_NAME) ./cmd/nixopus

cli-run: cli-build
	@./bin/$(CLI_NAME)

# Local development: build with localhost defaults. Source .env.cli first to override.
cli-build-dev:
	@[ -f .env.cli ] && set -a && . ./.env.cli && set +a || true; \
	$(MAKE) cli-build \
		CLI_API_URL="$${CLI_API_URL:-http://localhost:8080}" \
		CLI_BETTER_AUTH_URL="$${CLI_BETTER_AUTH_URL:-http://localhost:9090}" \
		CLI_FRONTEND_URL="$${CLI_FRONTEND_URL:-http://localhost:3000}" \
		CLI_OAUTH_CLIENT_ID="$${CLI_OAUTH_CLIENT_ID:-nixopus-cli}" \
		CLI_AGENT_ENDPOINT="$${CLI_AGENT_ENDPOINT:-http://localhost:4117}" \
		CLI_INIT_TIMEOUT="$${CLI_INIT_TIMEOUT:-30s}" \
		CLI_WORKFLOW_TIMEOUT="$${CLI_WORKFLOW_TIMEOUT:-30m}" \
		CLI_WORKFLOW_ID="$${CLI_WORKFLOW_ID:-deploymentWorkflow}"

cli-install: cli-build
	@cp bin/$(CLI_NAME) $(shell go env GOPATH)/bin/$(CLI_NAME)

test:
	@go test -p 1 ./internal/features/... -v -count=1

test-all:
	@go test -p 1 ./... -v -count=1

test-routes:
	@go test -p 1 ./internal/tests/routes/... -v -count=1

format:
	@echo "Formatting Go files..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: fixtures-load fixtures-recreate fixtures-clean fixtures-help format cli-build cli-build-dev cli-run cli-install

fixtures-help:
	@echo "Nixopus API Fixtures Commands:"
	@echo ""
	@echo "  fixtures-load     - Load fixtures without affecting existing data"
	@echo "  fixtures-recreate - Drop and recreate all tables, then load fixtures"
	@echo "  fixtures-clean    - Truncate all tables, then load fixtures"
	@echo ""
	@echo "Usage Examples:"
	@echo "  make fixtures-load     # Load fixtures normally"
	@echo "  make fixtures-recreate # Clean slate - drops and recreates tables"
	@echo "  make fixtures-clean    # Truncates tables and loads fresh data"
	@echo ""
	@echo "Available Fixture Files:"
	@echo "  fixtures/development/complete.yml      # Loads all fixtures (uses imports)"
	@echo "  fixtures/development/users.yml         # User data only"
	@echo "  fixtures/development/organizations.yml # Organization data only"
	@echo "  fixtures/development/roles.yml         # Role data only"
	@echo "  fixtures/development/permissions.yml   # Permission data only"
	@echo "  fixtures/development/role_permissions.yml # Role-permission mappings"
	@echo "  fixtures/development/feature_flags.yml # Feature flags"
	@echo "  fixtures/development/organization_users.yml # User-organization relationships"
	@echo ""
	@echo "Note: The complete.yml file uses import statements to load all individual files."

fixtures-load:
	go run cmd/fixtures/main.go -fixture=fixtures/development/complete.yml

fixtures-recreate:
	go run cmd/fixtures/main.go -fixture=fixtures/development/complete.yml -recreate

fixtures-clean:
	go run cmd/fixtures/main.go -fixture=fixtures/development/complete.yml -truncate
