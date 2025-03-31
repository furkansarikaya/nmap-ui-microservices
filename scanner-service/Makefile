.PHONY: build run test clean docker docker-run docker-stop lint format

# Variables
APP_NAME=scanner-service
MAIN_PATH=./cmd/main
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_COMPOSE_FILE=./deployments/docker/docker-compose.yml

# Build
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(MAIN_PATH)

# Run
run:
	@echo "Running $(APP_NAME)..."
	go run $(MAIN_PATH)

# Test
test:
	@echo "Running tests..."
	go test ./... -v

# Clean
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	go clean

# Docker
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) -f deployments/docker/Dockerfile .

# Docker run
docker-run:
	@echo "Running with Docker Compose..."
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

# Docker stop
docker-stop:
	@echo "Stopping Docker Compose..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down

# Lint
lint:
	@echo "Linting..."
	golangci-lint run ./...

# Format
format:
	@echo "Formatting..."
	gofmt -s -w .

# Generate
generate:
	@echo "Generating code..."
	go generate ./...

# Build CLI
build-cli:
	@echo "Building CLI tool..."
	go build -o scan-cli ./tools/scripts/scan-cli.go

# Install
install: build
	@echo "Installing $(APP_NAME)..."
	cp $(APP_NAME) $(GOPATH)/bin/

# Help
help:
	@echo "Make targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker       - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose"
	@echo "  lint         - Run linter"
	@echo "  format       - Format code"
	@echo "  generate     - Generate code"
	@echo "  build-cli    - Build CLI tool"
	@echo "  install      - Install to GOPATH/bin"
	@echo "  help         - Show this help"