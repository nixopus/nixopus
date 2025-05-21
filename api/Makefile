APP_NAME=nixopus-api
build: 
	@go build -o bin/$(APP_NAME)

run: build
	@./bin/$(APP_NAME)

test:
	@go test -p 1 ./... -v -count=1