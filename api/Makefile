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