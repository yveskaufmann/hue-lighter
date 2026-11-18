#!/usr/bin/env bash

#
# Builds the golang application against different operating systems and architectures.
#

set -euo pipefail

# Target operating system
TARGE_OS="$(go env GOOS)"

# Target architecture
TARGET_ARCH="$(go env GOARCH)"

# Application name
APP_NAME="hue-lighter"

# Application version
VERSION="${VERSION:-}"

# Where to place the build output
OUTPUT_DIR="bin"

# When enable the version will be included in the output binary name
INCLUDE_VERSION=0

# When enabled the architecture will be included in the output binary name
INCLUDE_ARCH=0

# Package path for build information (version, commit, build date) injection
BUILD_INFO_PKG="com.github.yveskaufmann/hue-lighter/internal/buildinfo"

# Use go mod vendor for dependency management
USE_GO_MOD_VENDOR=0

##
# Parses the command-line arguments for the build script
#
# Usage: parse_args [arguments]
#
parse_args() {

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        usage
        exit 0
        ;;
      -V|--apply-version-suffix)
        INCLUDE_VERSION=1
        shift
        ;;
      -a|--include-arch)
        INCLUDE_ARCH=1
        shift
        ;;
      --arch)
        TARGET_ARCH="$2"
        shift 2
        ;;
      -v|--version)
        VERSION="$2"
        shift 2
        ;;
      --os)
        TARGET_OS="$2"
        shift 2
        ;;
      --use-go-mod-vendor)
        USE_GO_MOD_VENDOR=1
        shift
        ;;
      *)
        echo "Unknown argument: $1"
        exit 1
        ;;
    esac
  done

  ## Verify and normalize the architectures
  case "$TARGET_ARCH" in
    x86_64|amd64)  TARGET_ARCH=amd64 ;;
    aarch64|arm64) TARGET_ARCH=arm64 ;;
    *) echo "Unsupported architecture: $TARGET_ARCH"; exit 1 ;;
  esac

  case "$TARGET_OS" in
    linux|darwin) ;;
    *) echo "Unsupported operating system: $TARGET_OS"; exit 1 ;;
  esac

  if [[ -z "$VERSION" ]]; then
    VERSION="$(bash scripts/get_version.sh)"
  fi
}

##
# Displays the usage information for the build script
#
# Usage: usage
#
usage() {
  echo "Usage: $0 [options]"
  echo "Builds the application for the specified TARGET_OS and TARGET_ARCH"
  echo ""
  echo "Options:"
  echo "  -h, --help    Show this help message"
  echo "  -V, --apply-version-suffix When set, appends the version suffix to the output binary"
  echo "  -a, --include-arch When set, appends the architecture to the output binary"
  echo "  -v, --version <version> Specify the version to use for the build"
  echo "      --os <os> Specify the target operating system (e.g., linux, darwin, windows)"
  echo "      --arch <arch> Specify the target architecture (e.g., amd64/x86_64, arm64/aarch64)"
  echo "      --use-go-mod-vendor  Use go mod vendor for dependency management"
  echo ""
  echo "Environment Variables:"
  echo "  VERSION      Version to use for the build (overrides --version)"
  echo ""
  echo "Example:"
  echo "Build for the current os and architecture"
  echo "  $0"
  echo ""
  echo "#Build the application for Linux amd64 with version 1.0.0, including version and architecture in the output binary"
  echo "  $0 --os linux --arch amd64 --version 1.0.0 -V -a"
  echo ""
}

##
# Builds the application
#
# Usage: build
#
build() {

  CA_BUNDLE="internal/hue_client/certs/cacert_bundle.pem"
  if [ ! -f "${CA_BUNDLE}" ]; then
    make embed-ca-bundle
  fi

  local binary_name="$APP_NAME"
  local binary_sub_dir="$TARGET_OS"


  if [[ $INCLUDE_VERSION -eq 1 ]]; then
    binary_name+="-${VERSION}"
  fi

  if [[ $INCLUDE_ARCH -eq 1 ]]; then
    binary_name+="-${TARGET_ARCH}"
  fi

  if [[ "$TARGET_OS" == "windows" ]]; then
    binary_name+=".exe"
  fi

  if [[ $TARGET_OS == "darwin" ]]; then
    binary_sub_dir="macos"
  fi

  local binary_path="$OUTPUT_DIR/$binary_sub_dir/$binary_name"
  local ldflags="-s -w"
  ldflags+=" -X ${BUILD_INFO_PKG}.version=${VERSION}"
  ldflags+=" -X ${BUILD_INFO_PKG}.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)"
  ldflags+=" -X ${BUILD_INFO_PKG}.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  local build_args=()

  if [[ $USE_GO_MOD_VENDOR -eq 1 ]]; then
    build_args+=("-mod=vendor")
  fi

  build_args+=("-trimpath")
  build_args+=("-ldflags" "$ldflags")
  build_args+=("-o" "$binary_path")
  build_args+=("./cmd/hue-lighter")

  mkdir -p "$(dirname "$binary_path")"
  CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build "${build_args[@]}"

  return 0
}

main() {
  parse_args "$@"
  build
}

main "$@"




