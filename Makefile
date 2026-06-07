BINARY      := a2acli
PKG         := ./...
COVERAGE_DIR := test
COVERAGE_OUT := $(COVERAGE_DIR)/profile.cov
GOLANGCI_LINT_VERSION := v1.62.2

.PHONY: all build test test-coverage lint lint-install fmt vet tidy clean help

all: lint test

build:
	go build -o $(BINARY) .

test:
	go test -race -v $(PKG)

test-coverage:
	@mkdir -p $(COVERAGE_DIR)
	go test -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic $(PKG)
	go tool cover -func=$(COVERAGE_OUT) | tail -n 1

lint: lint-install
	$(shell go env GOPATH)/bin/golangci-lint run ./...

lint-install:
	@if ! command -v $(shell go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
	rm -rf $(COVERAGE_DIR)

help:
	@echo "Targets:"
	@echo "  build          Build the $(BINARY) binary"
	@echo "  test           Run unit tests with race detector"
	@echo "  test-coverage  Run tests and emit $(COVERAGE_OUT)"
	@echo "  lint           Run golangci-lint (installs if missing)"
	@echo "  fmt            Run go fmt"
	@echo "  vet            Run go vet"
	@echo "  tidy           Run go mod tidy"
	@echo "  clean          Remove build/coverage artifacts"
