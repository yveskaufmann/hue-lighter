.SHELL = /usr/bin/env bash

APP_NAME="hue-lighter"
APP_ID="com.github.yveskaufmann.hue-lighter"
# Latest git tag without leading "v", e.g. 1.2.0 (or 1.2.0-3-gabc123 between tags).
VERSION ?= $(patsubst v%,%,$(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0))

BINARY=bin/hue-lighter

.PHONY: all build embed-ca-bundle clean test test-coverage test-race test-verbose fmt lint install uninstall mac_app linux_dist mac_icon mac_install mac_uninstall flatpak flatpak-run

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
	rm -rf bin/ dist/

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

# Builds bin/macos/Hue Lighter.app and dist/hue-lighter_$(VERSION)_macos_universal.zip (as used by the Homebrew cask).
mac_app: embed-ca-bundle
	VERSION="$(VERSION)" bash scripts/package-macos.sh

# Builds dist/hue-lighter_$(VERSION)_linux_{amd64,arm64}.tar.gz.
linux_dist: embed-ca-bundle
	VERSION="$(VERSION)" bash scripts/package-linux.sh

# Regenerates build/macos/AppIcon.icns from the tray icon SVG.
mac_icon:
	bash scripts/generate-macos-icon.sh

# Builds and installs the app to /Applications for local testing.
mac_install: embed-ca-bundle
	VERSION="$(VERSION)" bash scripts/install.sh Darwin

mac_uninstall:
	bash scripts/uninstall.sh Darwin
