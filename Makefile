.SHELL = /usr/bin/env bash

APP_NAME="hue-lighter"
APP_ID="com.github.yveskaufmann.hue-lighter"
VERSION="1.0.0"

BINARY=bin/hue-lighter

.PHONY: all build clean test test-coverage test-race test-verbose fmt lint install uninstall mac_pkg mac_install mac_uninstall

all: build

build:
	go build -o ${BINARY} ./cmd/hue-lighter

clean:
	rm -rf bin/

test:
	go test ./...

test-coverage:
	go test -cover ./...

test-race:
	go test -race ./...

test-verbose:
	go test -v ./...

fmt:
	go fmt ./...

install:
	$(MAKE) build && bash scripts/install.sh "$$(uname -s)"

uninstall:
	bash scripts/uninstall.sh

mac_pkg: build
	HUE_LIGHTER_MAC_PACKAGE_ONLY=1 bash scripts/install.sh Darwin

mac_install: build
	bash scripts/install.sh Darwin

mac_uninstall:
	scripts/uninstall.sh
