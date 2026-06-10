BINARY      := a2acli
REPO_ROOT   := $(shell pwd)
DOCKERFILE  := Dockerfile
PKG         := ./...
COVERAGE_DIR := test
COVERAGE_OUT := $(COVERAGE_DIR)/profile.cov
DIST_DIR     := dist
GOLANGCI_LINT_VERSION := v1.62.2
IMAGE       ?= quay.io/kynoproj/a2acli
VERSION     ?= latest

ifndef GOPATH
GOPATH=$(shell go env GOPATH)
endif

HOST_ARCH=$(shell uname -m)
# Github actions instances are x86_64
ifeq ($(HOST_ARCH),x86_64)
	HOST_ARCH=amd64
endif
ifeq ($(HOST_ARCH),aarch64)
	HOST_ARCH=arm64
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

build-binaries: clean $(DIST_DIR)/$(BINARY)-darwin-amd64.gz $(DIST_DIR)/$(BINARY)-darwin-arm64.gz $(DIST_DIR)/$(BINARY)-linux-amd64.gz $(DIST_DIR)/$(BINARY)-linux-arm64.gz $(DIST_DIR)/$(BINARY)-linux-arm.gz $(DIST_DIR)/$(BINARY)-linux-ppc64le.gz $(DIST_DIR)/$(BINARY)-linux-s390x.gz

$(DIST_DIR)/$(BINARY)-%.gz: $(DIST_DIR)/$(BINARY)-%
	@[[ -e $(DIST_DIR)/$(BINARY)-$*.gz ]] || gzip -k $(DIST_DIR)/$(BINARY)-$*

$(DIST_DIR)/$(BINARY): GOARGS = GOOS= GOARCH=
$(DIST_DIR)/$(BINARY)-darwin-amd64: GOARGS = GOOS=darwin GOARCH=amd64
$(DIST_DIR)/$(BINARY)-darwin-arm64: GOARGS = GOOS=darwin GOARCH=arm64
$(DIST_DIR)/$(BINARY)-linux-amd64: GOARGS = GOOS=linux GOARCH=amd64
$(DIST_DIR)/$(BINARY)-linux-arm64: GOARGS = GOOS=linux GOARCH=arm64
$(DIST_DIR)/$(BINARY)-linux-arm: GOARGS = GOOS=linux GOARCH=arm
$(DIST_DIR)/$(BINARY)-linux-ppc64le: GOARGS = GOOS=linux GOARCH=ppc64le
$(DIST_DIR)/$(BINARY)-linux-s390x: GOARGS = GOOS=linux GOARCH=s390x

$(DIST_DIR)/$(BINARY)-%:
	CGO_ENABLED=0 $(GOARGS) go build -v -ldflags '${LDFLAGS}' -o $(DIST_DIR)/$(BINARY)-$* .

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
	rm -rf $(DIST_DIR)

docker-build: clean dist/$(BINARY)-linux-$(HOST_ARCH)
	docker build \
		-t $(IMAGE):$(VERSION) .

docker-push:
	docker push $(IMAGE):$(VERSION)

buildx-push: clean dist/$(BINARY)-linux-amd64 dist/$(BINARY)-linux-arm64
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-f $(DOCKERFILE) \
		-t $(IMAGE):$(VERSION) \
		--push \
		$(REPO_ROOT)


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
