.PHONY: run test lint build docker-up docker-down

run:
	@echo "Starting API server..."
	go run ./cmd/api

test:
	go test -v -race -cover  ./...

lint:
	golangci-lint run

build:
	go build -o bin/api ./cmd/api

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v
