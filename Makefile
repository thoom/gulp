# Makefile for GULP v1.0+

# Variables
BINARY_NAME=gulp
# Generate timestamp in the format: YYYYMMDD.HHMMAM/PM.TZ-SNAPSHOT
SNAPSHOT_VERSION=$(shell date '+%Y%m%d.%I%M%p.%Z')-SNAPSHOT
VERSION?=$(SNAPSHOT_VERSION)
LDFLAGS=-X github.com/thoom/gulp/client.buildVersion=$(VERSION)

# Docker variables
IMAGE_NAME?=gulp
DOCKER_USER?=$(shell echo $$DOCKER_USER)
DOCKER_PASS?=$(shell echo $$DOCKER_PASS)
GH_USER?=$(shell echo $$GH_USER)
GH_PASS?=$(shell echo $$GH_PASS)

# Build output directory
BUILD_DIR=./build

# Default target
.PHONY: all
all: frontend build

# Frontend build
.PHONY: frontend
frontend: frontend-deps frontend-build

.PHONY: frontend-deps
frontend-deps:
	@echo "Installing frontend dependencies..."
	cd ui/frontend && npm install

.PHONY: frontend-build
frontend-build:
	@echo "Building React frontend..."
	cd ui/frontend && npm run build
	@echo "Copying frontend build to static directory..."
	@mkdir -p ui/static
	@cp -r ui/frontend/build/* ui/static/

.PHONY: frontend-dev
frontend-dev:
	@echo "Starting frontend development server..."
	cd ui/frontend && npm start

# Build for current platform
.PHONY: build
build: frontend
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build snapshot version (explicit snapshot with timestamp)
.PHONY: snapshot
snapshot: frontend
	go build -ldflags="-X github.com/thoom/gulp/client.buildVersion=$(SNAPSHOT_VERSION)" -o $(BINARY_NAME)

# Build with custom version
.PHONY: build-version
build-version:
	@if [ -z "$(V)" ]; then echo "Usage: make build-version V=1.2.3"; exit 1; fi
	go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$(V)" -o $(BINARY_NAME)

# Build for all platforms (replaces scripts/build.sh)
.PHONY: build-all
build-all:
	@if [ -z "$(RELEASE_VERSION)" ]; then echo "Usage: make build-all RELEASE_VERSION=1.2.3"; exit 1; fi
	@echo "Building for all platforms with version: $(RELEASE_VERSION)"
	@mkdir -p $(BUILD_DIR)
	
	# Linux 386
	@echo "Building for Linux 386..."
	@env GOOS=linux GOARCH=386 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$(RELEASE_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)
	@cd $(BUILD_DIR) && tar czf $(BINARY_NAME).linux-386.tar.gz $(BINARY_NAME) && rm $(BINARY_NAME)
	
	# Linux AMD64
	@echo "Building for Linux AMD64..."
	@env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$(RELEASE_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)
	@cd $(BUILD_DIR) && tar czf $(BINARY_NAME).linux-amd64.tar.gz $(BINARY_NAME) && rm $(BINARY_NAME)
	
	# Darwin AMD64
	@echo "Building for Darwin AMD64..."
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$(RELEASE_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)
	@cd $(BUILD_DIR) && tar czf $(BINARY_NAME).darwin-amd64.tar.gz $(BINARY_NAME) && rm $(BINARY_NAME)
	
	# Darwin ARM64
	@echo "Building for Darwin ARM64..."
	@env GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$(RELEASE_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)
	@cd $(BUILD_DIR) && tar czf $(BINARY_NAME).darwin-arm64.tar.gz $(BINARY_NAME) && rm $(BINARY_NAME)
	
	# Windows AMD64
	@echo "Building for Windows AMD64..."
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$(RELEASE_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME).exe
	@cd $(BUILD_DIR) && zip $(BINARY_NAME).windows.zip $(BINARY_NAME).exe && rm $(BINARY_NAME).exe
	
	@echo "All platform builds complete in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Docker build
.PHONY: docker-build
docker-build:
	@if [ -z "$(RELEASE_VERSION)" ]; then echo "Usage: make docker-build RELEASE_VERSION=1.2.3"; exit 1; fi
	@echo "Building Docker image with version: $(RELEASE_VERSION)"
	docker build -t $(BINARY_NAME) -f Dockerfile --no-cache --build-arg BUILD_VERSION=$(RELEASE_VERSION) .

# Docker deploy (replaces scripts/deploy.sh)
.PHONY: docker-deploy
docker-deploy: docker-build
	@if [ -z "$(RELEASE_VERSION)" ]; then echo "Usage: make docker-deploy RELEASE_VERSION=1.2.3"; exit 1; fi
	@if [ -z "$(DOCKER_USER)" ] || [ -z "$(DOCKER_PASS)" ]; then echo "Error: DOCKER_USER and DOCKER_PASS environment variables required"; exit 1; fi
	@if [ -z "$(GH_USER)" ] || [ -z "$(GH_PASS)" ]; then echo "Error: GH_USER and GH_PASS environment variables required"; exit 1; fi
	@if [ -z "$(IMAGE_NAME)" ]; then echo "Error: IMAGE_NAME variable required"; exit 1; fi
	
	@echo "Deploying to Docker Hub..."
	@echo "$(DOCKER_PASS)" | docker login -u "$(DOCKER_USER)" --password-stdin
	@docker tag $(BINARY_NAME) $(IMAGE_NAME):latest
	@docker tag $(BINARY_NAME) $(IMAGE_NAME):$(RELEASE_VERSION)
	@docker push $(IMAGE_NAME):latest
	@docker push $(IMAGE_NAME):$(RELEASE_VERSION)
	
	@echo "Deploying to GitHub Container Registry..."
	@echo "$(GH_PASS)" | docker login ghcr.io -u "$(GH_USER)" --password-stdin
	@docker tag $(BINARY_NAME) ghcr.io/$(IMAGE_NAME):latest
	@docker tag $(BINARY_NAME) ghcr.io/$(IMAGE_NAME):$(RELEASE_VERSION)
	@docker push ghcr.io/$(IMAGE_NAME):latest
	@docker push ghcr.io/$(IMAGE_NAME):$(RELEASE_VERSION)
	
	@echo "Docker deployment complete!"

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f *.tar.gz *.zip

# Clean Docker images
.PHONY: docker-clean
docker-clean:
	-docker rmi $(BINARY_NAME) 2>/dev/null || true
	-docker rmi $(IMAGE_NAME):latest 2>/dev/null || true
	-docker rmi ghcr.io/$(IMAGE_NAME):latest 2>/dev/null || true

# Test
.PHONY: test
test:
	go test ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run
.PHONY: run
run: build
	./$(BINARY_NAME)

# Show version after building
.PHONY: version
version: build
	./$(BINARY_NAME) --version

# Show what snapshot version would be generated
.PHONY: show-snapshot-version
show-snapshot-version:
	@echo "Snapshot version: $(SNAPSHOT_VERSION)"

# Install dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Full release pipeline
.PHONY: release
release: clean test build-all docker-deploy
	@echo "Release $(RELEASE_VERSION) complete!"

.PHONY: help
help:
	@echo "GULP Makefile - v1.0+ Standard Build System"
	@echo ""
	@echo "Frontend Targets:"
	@echo "  frontend        - Build React frontend and embed in Go binary"
	@echo "  frontend-deps   - Install frontend dependencies"
	@echo "  frontend-build  - Build React frontend production bundle"
	@echo "  frontend-dev    - Start frontend development server"
	@echo ""
	@echo "Build Targets:"
	@echo "  build           - Build binary with embedded frontend"
	@echo "  snapshot        - Build with explicit timestamp snapshot version"
	@echo "  build-version   - Build with custom version (usage: make build-version V=1.2.3)"
	@echo "  build-all       - Build for all platforms (usage: make build-all RELEASE_VERSION=1.2.3)"
	@echo ""
	@echo "Docker Targets:"
	@echo "  docker-build    - Build Docker image (usage: make docker-build RELEASE_VERSION=1.2.3)"
	@echo "  docker-deploy   - Build and deploy Docker image (usage: make docker-deploy RELEASE_VERSION=1.2.3)"
	@echo "  docker-clean    - Remove Docker images"
	@echo ""
	@echo "Development Targets:"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  run             - Build and run the application"
	@echo "  version         - Build and show version"
	@echo "  deps            - Install/update dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter (requires golangci-lint)"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean           - Remove build artifacts"
	@echo "  show-snapshot-version - Show what snapshot version would be generated"
	@echo ""
	@echo "Release Pipeline:"
	@echo "  release         - Full release pipeline: clean, test, build-all, docker-deploy"
	@echo ""
	@echo "Environment Variables:"
	@echo "  RELEASE_VERSION - Version for releases (required for build-all, docker-*)"
	@echo "  IMAGE_NAME      - Docker image name (default: gulp)"
	@echo "  DOCKER_USER     - Docker Hub username"
	@echo "  DOCKER_PASS     - Docker Hub password/token" 
	@echo "  GH_USER         - GitHub username"
	@echo "  GH_PASS         - GitHub token" 