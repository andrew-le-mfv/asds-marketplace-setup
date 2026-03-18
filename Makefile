.PHONY: build run test test-verbose clean fmt vet lint install

# Binary output directory
BIN_DIR := bin
BINARY := $(BIN_DIR)/asds

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOVET := $(GOCMD) vet

# Main package
MAIN_PACKAGE := ./cmd/asds/

all: clean build

build: ## Build the binary
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BINARY) $(MAIN_PACKAGE)

run: ## Run the application
	$(GOCMD) run $(MAIN_PACKAGE)

test: ## Run all tests
	$(GOTEST) ./...

test-verbose: ## Run all tests with verbose output
	$(GOTEST) ./... -v

test-package: ## Run tests for a specific package (make test-package PKG=./internal/config/)
	$(GOTEST) $(PKG) -v

clean: ## Remove build artifacts
	@rm -rf $(BIN_DIR)
	$(GOCLEAN)

fmt: ## Format code
	$(GOFMT) -s -w .

vet: ## Run go vet
	$(GOVET) ./...

lint: vet ## Run linting (requires golangci-lint)
	@golangci-lint run ./...

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

install: build ## Build and install to GOPATH/bin
	@cp $(BINARY) $(GOPATH)/bin/asds

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
