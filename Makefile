run-api-dev:
	@air

run-api-with-docker:
	@docker-compose up

build-api:
	@go build -v -o ./tmp/bin ./cmd/api/main.go