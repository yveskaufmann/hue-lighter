.SHELL = /usr/bin/env bash

.PHONY: all build clean test fmt lint

all: build

build:
	go build -o bin/hue-lighter ./cmd/hue-lighter

clean:
	rm -rf bin/

test:
	go test ./...

fmt:
	go fmt ./...

install: build

	if [ ! -f configs/certs/cacert_bundle.pem ]; then \
		echo "Hue Bridge Root CA bundle not found. Please ensure configs/certs/cacert_bundle.pem exists."; \
		echo "Take the CA bundle from https://developers.meethue.com/develop/application-design-guidance/using-https/ and write to configs/certs/cacert_bundle.pem"; \
		exit 1; \
	fi

	sudo systemctl stop hue-lighter || true

	sudo cp bin/hue-lighter /usr/bin/hue-lighter
	sudo cp build/linux/etc/systemd/system/hue-lighter.service /etc/systemd/system/hue-lighter.service

	# Create necessary directories
	sudo mkdir -p /var/lib/hue-lighter

	sudo mkdir -p /etc/hue-lighter
	sudo cp configs/config.yaml /etc/hue-lighter/config.yaml
	sudo cp configs/certs/cacert_bundle.pem /etc/hue-lighter/cacert_bundle.pem

	# Create system user and set ownerships
	sudo useradd --system --no-create-home --shell /usr/sbin/nologin hue-lighter || true
	sudo chown -R hue-lighter:hue-lighter /var/lib/hue-lighter
	sudo chown -R hue-lighter:hue-lighter /etc/hue-lighter

	# Set ownership of installed binary
	sudo systemctl daemon-reload
	sudo systemctl enable hue-lighter
	sudo systemctl start hue-lighter

uninstall:
	sudo systemctl stop hue-lighter || true
	sudo systemctl disable hue-lighter
	sudo rm -f /usr/bin/hue-lighter
	sudo rm -f /etc/systemd/system/hue-lighter.service
	sudo rm -rf /etc/hue-lighter
	sudo systemctl daemon-reload
