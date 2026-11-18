.SHELL = /usr/bin/env bash

APP_NAME := hue-lighter
APP_ID := com.github.yveskaufmann.hue-lighter
FLATPAK_ID := io.github.yveskaufmann.HueLighter
VERSION ?= $(patsubst v%,%,$(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0))
BINARY := bin/hue-lighter

.PHONY: all
all: build

.PHONY: build
build: embed-ca-bundle
	go build -o ${BINARY} ./cmd/hue-lighter

# Copies the Philips Hue CA bundle next to the hue_client package so it gets embedded into the binary.
CA_BUNDLE_SRC := configs/certs/cacert_bundle.pem
CA_BUNDLE_EMBED := internal/hue_client/certs/cacert_bundle.pem
.PHONY: embed-ca-bundle
embed-ca-bundle:
	@if [ -f "${CA_BUNDLE_SRC}" ]; then \
		cp "${CA_BUNDLE_SRC}" "${CA_BUNDLE_EMBED}"; \
	else \
		echo "WARNING: ${CA_BUNDLE_SRC} not found, building without embedded CA bundle (HUE_CA_CERTS_PATH required at runtime)"; \
	fi

.PHONY: clean
clean:
	rm -rf bin/ dist/

.PHONY: test
test:
	go test ./...

.PHONY: test-coverage
test-coverage:
	go test -cover ./...

.PHONY: test-race
test-race:
	go test -race ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: install
install:
	$(MAKE) build && bash scripts/install.sh "$$(uname -s)"

# Builds the Flatpak (offline, from vendored Go modules) and installs it for the current user.
# vendor/ is removed afterwards, otherwise regular Go builds would use it as well.
.PHONY: flatpak
flatpak: embed-ca-bundle
	bash scripts/package-flatpak.sh --version "$(VERSION)"

.PHONY: flatpak-run
flatpak-run:
	flatpak run ${FLATPAK_ID}

.PHONY: uninstall
uninstall:
	bash scripts/uninstall.sh

# Builds dist/hue-lighter_$(VERSION)_linux_{amd64,arm64}.tar.gz.
.PHONY: linux_app
linux_app: embed-ca-bundle
	bash scripts/package-linux.sh --version "$(VERSION)"

# Regenerates build/macos/AppIcon.icns from the tray icon SVG.
.PHONY: mac_icon
mac_icon:
	bash scripts/generate-macos-icon.sh

# Builds bin/macos/Hue Lighter.app and dist/hue-lighter_$(VERSION)_macos_universal.zip (as used by the Homebrew cask).
.PHONY: mac_app
mac_app: embed-ca-bundle
	bash scripts/package-macos.sh --version "$(VERSION)"

# Builds and installs the app to /Applications for local testing.
.PHONY: mac_install
mac_install: embed-ca-bundle
	VERSION="$(VERSION)" bash scripts/install.sh Darwin

.PHONY: mac_uninstall
mac_uninstall:
	bash scripts/uninstall.sh Darwin
