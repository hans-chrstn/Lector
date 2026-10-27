.PHONY: build dev test test-backend test-frontend test-e2e lint format format-check check
build:
	mkdir -p .local
	go build -o .local/lector ./cmd/lector
	npm --prefix frontend run build
dev:
	go run ./cmd/lector
test: test-backend test-frontend
test-backend:
	go test -race ./cmd/... ./internal/...
test-frontend:
	npm --prefix frontend test
test-e2e: build
	npm --prefix frontend run test:e2e
lint:
	golangci-lint run ./cmd/... ./internal/...
	npm --prefix frontend run lint
format:
	gofmt -w cmd internal
	npm --prefix frontend run format
format-check:
	@test -z "$$(gofmt -l cmd internal)"
	npm --prefix frontend run format:check
check: test lint format-check build
