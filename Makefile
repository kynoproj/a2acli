BINARY      := a2acli
PKG         := ./...
COVERAGE_DIR := test
COVERAGE_OUT := $(COVERAGE_DIR)/profile.cov
GOLANGCI_LINT_VERSION := v1.62.2
IMAGE       ?= quay.io/kynoproj/a2acli
VERSION     ?= dev

ifndef GOPATH
GOPATH=$(shell go env GOPATH)
endif

BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git rev-parse --symbolic-full-name --verify --quiet --abbrev-ref HEAD)
GIT_TAG=$(shell if [[ -z "`git status --porcelain`" ]]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [[ -z "`git status --porcelain`" ]]; then echo "clean" ; else echo "dirty"; fi)

override LDFLAGS += \
  -X main.version=${VERSION} \
  -X main.buildDate=${BUILD_DATE} \
  -X main.gitCommit=${GIT_COMMIT} \
  -X main.gitTreeState=${GIT_TREE_STATE}

ifneq (${GIT_TAG},)
VERSION=$(GIT_TAG)
override LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif

.PHONY: all build test test-coverage lint lint-install fmt vet tidy clean docker-build docker-push help

all: lint test

build:
	go build -v -ldflags '${LDFLAGS}' -o $(BINARY) .

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

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GIT_TAG=$(GIT_TAG) \
		--build-arg GIT_TREE_STATE=$(GIT_TREE_STATE) \
		-t $(IMAGE):$(VERSION) .

docker-push:
	docker push $(IMAGE):$(VERSION)

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
	@echo "  docker-build   Build debug image $(IMAGE):$(VERSION)"
	@echo "  docker-push    Push image $(IMAGE):$(VERSION)"
