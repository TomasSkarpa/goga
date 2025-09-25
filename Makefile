.PHONY: build run dev test clean install-deps

# Build the application
build:
	go build -o bin/goga cmd/goga/main.go

# Run the application
run:
	go run cmd/goga/main.go

# Run with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ goga.db uploads/

# Install development dependencies
install-deps:
	go mod tidy
	go install github.com/air-verse/air@latest

# Install and run
setup: install-deps
	@echo "Setup complete! Run 'make dev' to start development server"

# Production build
build-prod:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/goga cmd/goga/main.go