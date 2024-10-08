run-api-dev:
	@air

run-api-with-docker:
	@docker-compose up

build-api:
	@go build -o ./tmp/main cmd/api/main.go