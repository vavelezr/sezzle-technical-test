COVERAGE_PKGS := ./internal/...

.PHONY: build run test coverage lint fmt tidy clean

build: ## Build the server binary into bin/
	CGO_ENABLED=0 go build -trimpath -o bin/server ./cmd/server

run: ## Run the server locally
	go run ./cmd/server

test: ## Run all tests with the race detector
	go test -race ./...

coverage: ## Generate coverage.out and coverage.html for internal packages
	go test -covermode=atomic -coverprofile=coverage.out $(COVERAGE_PKGS)
	go tool cover -func=coverage.out | tail -n 1
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	golangci-lint fmt ./...

tidy: ## Tidy module dependencies
	go mod tidy

clean: ## Remove build and coverage artifacts
	rm -rf bin coverage.out coverage.html
