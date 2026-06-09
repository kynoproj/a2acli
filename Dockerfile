# syntax=docker/dockerfile:1.7

# ---------- Build stage ----------
FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT
ARG GIT_TAG
ARG GIT_TREE_STATE=clean
ARG BUILD_DATE
ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    : "${GIT_COMMIT:=$(git rev-parse HEAD 2>/dev/null || echo unknown)}"; \
    : "${BUILD_DATE:=$(date -u +'%Y-%m-%dT%H:%M:%SZ')}"; \
    LDFLAGS="-s -w \
        -X main.version=${VERSION} \
        -X main.buildDate=${BUILD_DATE} \
        -X main.gitCommit=${GIT_COMMIT} \
        -X main.gitTreeState=${GIT_TREE_STATE}"; \
    if [ -n "${GIT_TAG}" ]; then LDFLAGS="${LDFLAGS} -X main.gitTag=${GIT_TAG}"; fi; \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
        go build -trimpath -ldflags="${LDFLAGS}" -o /out/a2acli .

# ---------- Runtime stage (debug-friendly) ----------
FROM alpine:3.23

RUN apk add --no-cache \
        bash \
        ca-certificates \
        curl \
        wget \
        bind-tools \
        busybox-extras \
        netcat-openbsd \
        iputils \
        tcpdump \
        jq \
        openssl \
        tzdata

COPY --from=builder /out/a2acli /usr/local/bin/a2acli

# Non-root user is preferable, but in-cluster debug sessions sometimes need
# privileged tools (tcpdump). Default to root so kubectl exec / debug works
# without extra securityContext gymnastics; override at runtime if desired.
WORKDIR /work

ENTRYPOINT ["a2acli"]
CMD ["--help"]
