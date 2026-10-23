.SHELL = /usr/bin/env bash

APP_NAME="hue-lighter"
APP_ID="com.github.yveskaufmann.hue-lighter"
VERSION="1.0.0"

BINARY=bin/hue-lighter

.PHONY: all build embed-ca-bundle clean test test-coverage test-race test-verbose fmt lint install uninstall mac_pkg mac_install mac_uninstall flatpak flatpak-run

all: build

CA_BUNDLE_SRC=configs/certs/cacert_bundle.pem
CA_BUNDLE_EMBED=internal/hue_client/certs/cacert_bundle.pem

build: embed-ca-bundle
	go build -o ${BINARY} ./cmd/hue-lighter

# Copies the Philips Hue CA bundle next to the hue_client package so it gets embedded into the binary.
embed-ca-bundle:
	@if [ -f "${CA_BUNDLE_SRC}" ]; then \
		cp "${CA_BUNDLE_SRC}" "${CA_BUNDLE_EMBED}"; \
	else \
		echo "WARNING: ${CA_BUNDLE_SRC} not found, building without embedded CA bundle (HUE_CA_CERTS_PATH required at runtime)"; \
	fi

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

FLATPAK_ID=io.github.yveskaufmann.HueLighter
FLATPAK_MANIFEST=build/flatpak/${FLATPAK_ID}.yml

# Builds the Flatpak (offline, from vendored Go modules) and installs it for the current user.
# vendor/ is removed afterwards, otherwise regular Go builds would use it as well.
flatpak: embed-ca-bundle
	go mod vendor
	flatpak-builder --user --install --force-clean \
		--state-dir=.flatpak-builder build/flatpak/build-dir ${FLATPAK_MANIFEST}; \
		status=$$?; rm -rf vendor; exit $$status

flatpak-run:
	flatpak run ${FLATPAK_ID}

uninstall:
	bash scripts/uninstall.sh

mac_pkg: build
	HUE_LIGHTER_MAC_PACKAGE_ONLY=1 bash scripts/install.sh Darwin

mac_install: build
	bash scripts/install.sh Darwin

mac_uninstall:
	scripts/uninstall.sh
