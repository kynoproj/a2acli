FROM alpine:3.23

ARG TARGETOS
ARG TARGETARCH

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

COPY dist/a2acli-${TARGETOS}-${TARGETARCH} /usr/local/bin/a2acli

WORKDIR /work

ENTRYPOINT ["a2acli"]
CMD ["--help"]
