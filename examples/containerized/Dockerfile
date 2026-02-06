FROM golang:1.24 AS builder
WORKDIR /src

#  Leverage Docker layer caching for dependencies before copying the full source code 
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w" -o /out/hue-lighter ./cmd/hue-lighter

FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ENV USER=huelighter \
    CONFIG_PATH=/etc/hue-lighter/config.yaml \
    HUE_CA_CERTS_PATH=/etc/hue-lighter/cacert_bundle.pem

WORKDIR /etc/hue-lighter
RUN groupadd -g 1001 -r ${USER} && \
    useradd -u 1001 -r -g ${USER} -M -s /usr/sbin/nologin ${USER}

COPY --from=builder --chown=${USER}:${USER} /out/hue-lighter /usr/local/bin/hue-lighter

VOLUME ["/etc/hue-lighter", "/var/lib/hue-lighter"]
USER ${USER}

ENTRYPOINT ["/usr/local/bin/hue-lighter"]
