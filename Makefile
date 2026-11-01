.PHONY: run build test test-integration lint fmt up down clean

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test -v -race ./internal/...

test-integration:
	go test -v -tags=integration ./test/integration/...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

up:
	docker compose up --build -d

down:
	docker compose down

clean:
	rm -rf bin/

