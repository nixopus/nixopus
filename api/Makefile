APP_NAME=nixopus-api
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

format:
	@echo "Formatting Go files..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: fixtures-load fixtures-recreate fixtures-clean fixtures-help format

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
