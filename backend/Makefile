COVERAGE_PKGS := ./internal/...
COVERAGE_MIN  := 90

.PHONY: build run test coverage lint fmt tidy clean

build: ## Build the server and healthcheck binaries into bin/
	CGO_ENABLED=0 go build -trimpath -o bin/ ./cmd/...

run: ## Run the server locally
	go run ./cmd/server

test: ## Run all tests with the race detector
	go test -race ./...

coverage: ## Generate coverage.out and coverage.html; fail below COVERAGE_MIN%
	go test -covermode=atomic -coverprofile=coverage.out $(COVERAGE_PKGS)
	go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | awk -v min=$(COVERAGE_MIN) '/^total:/ { \
		sub("%", "", $$3); printf "total coverage: %s%% (minimum %s%%)\n", $$3, min; \
		if ($$3 + 0 < min) { print "coverage below threshold"; exit 1 } }'

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	golangci-lint fmt ./...

tidy: ## Tidy module dependencies
	go mod tidy

clean: ## Remove build and coverage artifacts
	rm -rf bin coverage.out coverage.html
