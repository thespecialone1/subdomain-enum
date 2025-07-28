# Advanced Subdomain Enumeration Tool v2.2.0 - Makefile
# Professional build and deployment automation

# Build variables
VERSION ?= 2.2.0
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Go build variables
GO_VERSION := 1.21
BINARY_NAME := subdomain-enum
MAIN_PATH := cmd/server/main.go
DIST_DIR := dist
DOCKER_REGISTRY := ghcr.io
DOCKER_REPOSITORY := thespecialone1/subdomain-enum

# Build flags
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.buildTime=$(BUILD_TIME) \
	-X main.gitCommit=$(GIT_COMMIT)

BUILD_FLAGS := -ldflags="$(LDFLAGS)" -a -installsuffix cgo

# Platform targets
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	freebsd/amd64

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
NC := \033[0m # No Color

.PHONY: help build build-all clean test lint fmt vet deps docker docker-build docker-run docker-push docker-compose-up docker-compose-down install uninstall release check-tools

# Default target
all: clean fmt lint test build

# Help target
help: ## Show this help message
	@echo "$(CYAN)Advanced Subdomain Enumeration Tool v$(VERSION)$(NC)"
	@echo "$(CYAN)==========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  $(GREEN)make build$(NC)          # Build for current platform"
	@echo "  $(GREEN)make build-all$(NC)      # Build for all platforms"
	@echo "  $(GREEN)make docker-build$(NC)   # Build Docker image"
	@echo "  $(GREEN)make docker-run$(NC)     # Run with Docker"
	@echo "  $(GREEN)make test$(NC)           # Run tests"
	@echo "  $(GREEN)make release$(NC)        # Create release build"

# Build targets
build: deps ## Build binary for current platform
	@echo "$(BLUE)Building $(BINARY_NAME) v$(VERSION) for $(shell go env GOOS)/$(shell go env GOARCH)...$(NC)"
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Build completed: $(DIST_DIR)/$(BINARY_NAME)$(NC)"

build-race: deps ## Build with race detection (for testing)
	@echo "$(BLUE)Building $(BINARY_NAME) with race detection...$(NC)"
	@mkdir -p $(DIST_DIR)
	@go build -race $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-race $(MAIN_PATH)
	@echo "$(GREEN)✅ Race build completed: $(DIST_DIR)/$(BINARY_NAME)-race$(NC)"

build-debug: deps ## Build with debug symbols
	@echo "$(BLUE)Building $(BINARY_NAME) with debug symbols...$(NC)"
	@mkdir -p $(DIST_DIR)
	@go build -gcflags="all=-N -l" -o $(DIST_DIR)/$(BINARY_NAME)-debug $(MAIN_PATH)
	@echo "$(GREEN)✅ Debug build completed: $(DIST_DIR)/$(BINARY_NAME)-debug$(NC)"

build-all: deps ## Build for all supported platforms
	@echo "$(BLUE)Building $(BINARY_NAME) v$(VERSION) for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		SUFFIX=$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then SUFFIX=$$SUFFIX.exe; fi; \
		echo "$(YELLOW)Building for $$OS/$$ARCH...$(NC)"; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build $(BUILD_FLAGS) \
			-o $(DIST_DIR)/$(BINARY_NAME)-$$SUFFIX $(MAIN_PATH); \
		if [ $$? -eq 0 ]; then \
			echo "$(GREEN)✅ $$OS/$$ARCH build completed$(NC)"; \
		else \
			echo "$(RED)❌ $$OS/$$ARCH build failed$(NC)"; \
		fi; \
	done

# Development targets
run: build ## Build and run the application
	@echo "$(BLUE)Starting $(BINARY_NAME)...$(NC)"
	@./$(DIST_DIR)/$(BINARY_NAME)

run-dev: ## Run in development mode with hot reloading
	@echo "$(BLUE)Starting development server with hot reloading...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)Installing air for hot reloading...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Testing and quality targets
test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(NC)"

test-race: ## Run tests with race detection
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	@go test -race -v ./...

bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

lint: ## Run linting tools
	@echo "$(BLUE)Running linting tools...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)Installing goimports...$(NC)"; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...

# Dependency management
deps: ## Download and verify dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify

deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy

deps-clean: ## Clean module cache
	@echo "$(BLUE)Cleaning module cache...$(NC)"
	@go clean -modcache

# Docker targets
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_REPOSITORY):$(VERSION) \
		-t $(DOCKER_REPOSITORY):latest \
		.
	@echo "$(GREEN)✅ Docker image built: $(DOCKER_REPOSITORY):$(VERSION)$(NC)"

docker-build-dev: ## Build Docker development image
	@echo "$(BLUE)Building Docker development image...$(NC)"
	@docker build \
		--target development \
		-t $(DOCKER_REPOSITORY):dev \
		.
	@echo "$(GREEN)✅ Docker development image built: $(DOCKER_REPOSITORY):dev$(NC)"

docker-run: docker-build ## Build and run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	@docker run -p 8080:8080 -p 9090:9090 --rm $(DOCKER_REPOSITORY):$(VERSION)

docker-run-dev: docker-build-dev ## Build and run development Docker container
	@echo "$(BLUE)Running development Docker container...$(NC)"
	@docker run -p 8081:8080 -p 9091:9090 -v $(PWD):/app --rm $(DOCKER_REPOSITORY):dev

docker-push: docker-build ## Build and push Docker image to registry
	@echo "$(BLUE)Pushing Docker image to $(DOCKER_REGISTRY)...$(NC)"
	@docker tag $(DOCKER_REPOSITORY):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):$(VERSION)
	@docker tag $(DOCKER_REPOSITORY):latest $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):latest
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):latest
	@echo "$(GREEN)✅ Docker images pushed to registry$(NC)"

docker-scan: docker-build ## Scan Docker image for vulnerabilities
	@echo "$(BLUE)Scanning Docker image for vulnerabilities...$(NC)"
	@if command -v trivy > /dev/null; then \
		trivy image $(DOCKER_REPOSITORY):$(VERSION); \
	else \
		echo "$(YELLOW)Trivy not found. Install from https://github.com/aquasecurity/trivy$(NC)"; \
	fi

# Docker Compose targets
docker-compose-up: ## Start services with Docker Compose
	@echo "$(BLUE)Starting services with Docker Compose...$(NC)"
	@docker-compose up -d

docker-compose-up-dev: ## Start development services with Docker Compose
	@echo "$(BLUE)Starting development services with Docker Compose...$(NC)"
	@docker-compose --profile dev up -d

docker-compose-up-full: ## Start all services including monitoring
	@echo "$(BLUE)Starting all services with monitoring...$(NC)"
	@docker-compose --profile dev --profile monitoring --profile cache up -d

docker-compose-down: ## Stop Docker Compose services
	@echo "$(BLUE)Stopping Docker Compose services...$(NC)"
	@docker-compose down

docker-compose-logs: ## View Docker Compose logs
	@docker-compose logs -f

# Installation targets
install: build ## Install binary to system
	@echo "$(BLUE)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@sudo cp $(DIST_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ $(BINARY_NAME) installed successfully$(NC)"

uninstall: ## Uninstall binary from system
	@echo "$(BLUE)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ $(BINARY_NAME) uninstalled successfully$(NC)"

# Release targets
release: clean build-all ## Create release archives
	@echo "$(BLUE)Creating release archives...$(NC)"
	@mkdir -p $(DIST_DIR)/release
	@for file in $(DIST_DIR)/$(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			basename=$$(basename $$file); \
			if [[ "$$basename" == *".exe" ]]; then \
				cd $(DIST_DIR) && zip release/$$basename.zip $$basename; \
			else \
				cd $(DIST_DIR) && tar -czf release/$$basename.tar.gz $$basename; \
			fi; \
		fi; \
	done
	@echo "$(GREEN)✅ Release archives created in $(DIST_DIR)/release/$(NC)"

release-github: ## Create GitHub release (requires gh CLI)
	@echo "$(BLUE)Creating GitHub release...$(NC)"
	@if command -v gh > /dev/null; then \
		gh release create v$(VERSION) \
			--title "Release v$(VERSION)" \
			--notes "See CHANGELOG.md for details" \
			$(DIST_DIR)/release/*; \
	else \
		echo "$(RED)GitHub CLI (gh) not found. Install from https://cli.github.com/$(NC)"; \
	fi

# Utility targets
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✅ Clean completed$(NC)"

clean-all: clean ## Clean everything including caches
	@echo "$(BLUE)Cleaning everything...$(NC)"
	@go clean -cache -testcache -modcache
	@docker system prune -f
	@echo "$(GREEN)✅ Deep clean completed$(NC)"

version: ## Show version information
	@echo "$(CYAN)Version Information:$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Git Branch: $(GIT_BRANCH)"
	@echo "  Go Version: $(shell go version)"

check-tools: ## Check required tools
	@echo "$(BLUE)Checking required tools...$(NC)"
	@echo -n "Go: "; go version 2>/dev/null && echo "$(GREEN)✅$(NC)" || echo "$(RED)❌$(NC)"
	@echo -n "Docker: "; docker --version 2>/dev/null && echo "$(GREEN)✅$(NC)" || echo "$(RED)❌$(NC)"
	@echo -n "Docker Compose: "; docker-compose --version 2>/dev/null && echo "$(GREEN)✅$(NC)" || echo "$(RED)❌$(NC)"
	@echo -n "Git: "; git --version 2>/dev/null && echo "$(GREEN)✅$(NC)" || echo "$(RED)❌$(NC)"

# Development workflows
dev-setup: ## Set up development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@go mod download
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)✅ Development environment ready$(NC)"

# Quick development workflow
quick: fmt lint test build ## Quick development workflow (format, lint, test, build)

# CI/CD simulation
ci: clean fmt lint vet test build-all docker-build ## Simulate CI/CD pipeline

# Security checks
security: ## Run security checks
	@echo "$(BLUE)Running security checks...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)Installing gosec...$(NC)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Performance profiling
profile-cpu: ## Generate CPU profile
	@echo "$(BLUE)Generating CPU profile...$(NC)"
	@go build -o $(DIST_DIR)/$(BINARY_NAME)-profile $(MAIN_PATH)
	@echo "Run your workload then access http://localhost:8080/debug/pprof/profile"

profile-mem: ## Generate memory profile
	@echo "$(BLUE)Generating memory profile...$(NC)"
	@go build -o $(DIST_DIR)/$(BINARY_NAME)-profile $(MAIN_PATH)
	@echo "Run your workload then access http://localhost:8080/debug/pprof/heap"

# Database and migration targets (for future use)
db-migrate: ## Run database migrations (placeholder)
	@echo "$(YELLOW)Database migrations not implemented yet$(NC)"

# Monitoring and logging
logs: ## View application logs
	@echo "$(BLUE)Viewing application logs...$(NC)"
	@tail -f /var/log/subdomain-enum.log 2>/dev/null || echo "$(YELLOW)Log file not found$(NC)"

monitor: ## Start monitoring dashboard
	@echo "$(BLUE)Starting monitoring dashboard...$(NC)"
	@echo "Visit http://localhost:3000 for Grafana (admin/admin123)"
	@echo "Visit http://localhost:9092 for Prometheus"

# Print build information
info:
	@echo "$(CYAN)Build Information:$(NC)"
	@echo "  Binary Name:      $(BINARY_NAME)"
	@echo "  Version:          $(VERSION)"
	@echo "  Build Time:       $(BUILD_TIME)"
	@echo "  Git Commit:       $(GIT_COMMIT)"
	@echo "  Git Branch:       $(GIT_BRANCH)"
	@echo "  Main Path:        $(MAIN_PATH)"
	@echo "  Distribution Dir: $(DIST_DIR)"
	@echo "  Docker Registry:  $(DOCKER_REGISTRY)"
	@echo "  Docker Repo:      $(DOCKER_REPOSITORY)"
	@echo ""
	@echo "$(CYAN)Platform Targets:$(NC)"
	@for platform in $(PLATFORMS); do echo "  $$platform"; done