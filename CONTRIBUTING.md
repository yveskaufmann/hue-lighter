# Contributing to Hue Lighter

This document covers a minimal development setup, build and test instructions, and guidelines for commits and PRs.

## Prerequisites

- Go 1.25 or newer installed and available on `PATH`.
- GNU Make (for provided targets).
- A Linux desktop is recommended for full testing (tray icon, keyring, Flatpak build via `make flatpak`).

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

## Build release packages

The packaging scripts default to the version from the latest Git tag (or the
current Git description when there is no tag). Pass `VERSION` to a Make target
to choose a version explicitly.

### Linux

Build tarballs for both `amd64` and `arm64`:

```sh
make linux_app
# or choose a version
make linux_app VERSION=1.2.3
```

The archives are written to `dist/` as
`hue-lighter_<version>_linux_<arch>.tar.gz`. To run the packaging script
directly, use `bash scripts/package-linux.sh --version 1.2.3`.

### Flatpak

Build a Flatpak bundle with `make flatpak`. This requires Flatpak and
`flatpak-builder`; the builder may need network access to install SDK/runtime
dependencies from Flathub. The default target architecture is `x86_64`.

```sh
make flatpak
# or build bundles for multiple Flatpak architectures
bash scripts/package-flatpak.sh --version 1.2.3 --arch x86_64 --arch aarch64
```

Bundles are written to `dist/` as `hue-lighter-<arch>.flatpak`. The package
script also accepts `--version <version>`; `--arch <arch>` can be repeated.

### macOS

Run the macOS packaging script on macOS; it uses Apple tools including `lipo`,
`codesign`, and `ditto`. Build the universal `arm64` + `amd64` app and zip:

```sh
make mac_app
# or choose a version
make mac_app VERSION=1.2.3
# alternatively
bash scripts/package-macos.sh --version 1.2.3
```

The app is created at `bin/macos/Hue Lighter.app`, and the distributable archive
at `dist/hue-lighter_<version>_macos_universal.zip`. The app is ad-hoc signed,
not signed with an Apple Developer ID.

## Run (development)

Run the binary directly for testing (recommended) rather than installing the Flatpak:

```sh
./bin/hue-lighter
# or
go run ./cmd/hue-lighter
```

`make build` embeds `configs/certs/cacert_bundle.pem` into the binary. With `go run`/`go build` (no embedded bundle) or to use an alternate CA bundle for TLS, set `HUE_CA_CERTS_PATH`:

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

## Architecture

See the architecture diagrams and sequence flows in [docs/architecture.md](docs/architecture.md).

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

