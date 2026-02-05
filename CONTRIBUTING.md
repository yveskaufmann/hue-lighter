# Contributing to Hue Lighter

This document covers a minimal development setup, build and test instructions, and guidelines for commits and PRs.

## Prerequisites

- Go 1.24 or newer installed and available on `PATH`.
- GNU Make (for provided targets).
- A Linux environment is recommended for full testing (systemd behavior, service install).

## Getting the repository

```sh
git clone https://github.com/yveskaufmann/hue-lighter.git
cd hue-lighter
```

## Prepare configuration for local development

The repository contains an example config at `configs/config.example.yaml`. Do NOT commit your personal `configs/config.yaml` — it is listed in `.gitignore`.

Create a local config for development:

```sh
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with local coordinates and dummy light IDs for testing
```

If you prefer to keep the config outside the repo, set the `CONFIG_PATH` environment variable to point to your config file:

```sh
export CONFIG_PATH=/path/to/config.yaml
```

## Build

Build the binary with the Makefile or `go build`:

```sh
make build
# or
go build -o bin/hue-lighter ./cmd/hue-lighter
```

## Run (development)

Run the binary directly for testing (recommended) rather than installing as a systemd service:

```sh
./bin/hue-lighter
# or
go run ./cmd/hue-lighter
```

If you need to pass an alternate CA bundle for TLS, set `HUE_CA_CERTS_PATH`:

```sh
export HUE_CA_CERTS_PATH=configs/certs/cacert_bundle.pem
```

## Tests

Run unit tests with:

```sh
make test
# or
go test ./...
```

## Formatting and linting

Keep code formatted with `gofmt` / `go fmt`:

```sh
make fmt
# or
go fmt ./...
```

Run `go vet` or other linters as needed.

## Debugging in VS Code

A `.vscode/launch.json` is included for debugging. Open the project in VS Code and use the Go extension to start the debug session.

## Committing and Pull Requests

- Write small, focused commits with clear messages.
- Use conventional commits where helpful (e.g., `feat:`, `fix:`, `chore:`).
- Run `make test` and `make fmt` before creating a PR.
- Explain the purpose of the change in the PR description and link to any related issues.

## Sensitive Data

- Never commit `configs/config.yaml`, API keys, or any private certificates.
- If a secret is accidentally committed, rotate the secret immediately and open an issue describing the remediation.

## Code of Conduct

Be respectful and collaborative. If you'd like, we can add a `CODE_OF_CONDUCT.md` to the repository (suggested for public projects).

