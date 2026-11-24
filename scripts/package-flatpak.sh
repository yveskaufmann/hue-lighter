#!/usr/bin/env bash
# Builds and bundle the app in flatpak bundle
#
# Usage: scripts/package-flatpak.sh [options]
#
set -euo pipefail

#  Flatpak bundle name
FLATPAK_NAME="hue-lighter"

#  Flatpak APP ID
FLATPAK_ID="io.github.yveskaufmann.HueLighter"

#  Flatpak manifest file
FLATPAK_MANIFEST="build/flatpak/${FLATPAK_ID}.yml"

#  Version of the app to build
VERSION="${VERSION:-}"

# Commit hash to be embedded into the binary
COMMIT_HASH="${COMMIT_HASH:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"

#  Build directory for flatpak
BUILD_DIR="build/flatpak/build-dir"

#  Temporary repository for flatpak
REPO="build/flatpak/.tmp-repo"

# List of architectures to build for
ARCHES_TO_BUILD=("x86_64")

# Remember if architectures were specified explicitly
ARCHES_SPECIFIED=false

##
# Parses the command-line arguments for the package script
#
# Usage: parse_args [arguments]
#
parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help) usage; exit 0 ;;
      -v|--version)
        VERSION="$2"; shift 2 ;;
      -c|--commit)
        COMMIT_HASH="$2"; shift 2 ;;
      -a|--arch)
        if [[ "$ARCHES_SPECIFIED" == false ]]; then
          ARCHES_TO_BUILD=()
          ARCHES_SPECIFIED=true
        fi
        ARCHES_TO_BUILD+=("$2"); shift 2 ;;
      *)
        echo "Unknown argument: $1"
        exit 1
        ;;
    esac
  done

  if [[ -z "$VERSION" ]]; then
    VERSION="$(bash scripts/get_version.sh)"
  fi
}

##
# Displays the usage information for the package script
#
# Usage: usage
#
usage() {
  echo "Usage: $0 [options]"
  echo "Build and package the Linux binaries into flatpak bundle"
  echo ""
  echo "Options:"
  echo "  -h, --help    Show this help message"
  echo "  -v, --version <version> Specify the version to use for the build"
  echo "  -c, --commit <commit> Specify the commit hash to embed into the binary"
  echo "  -a, --arch <arch> Specify the architecture to build for can be specified multiple times (default: x86_64)"
  echo ""
}

build_flatpak_bundle() {
  mkdir -p dist
  go mod vendor

  # Build and extract flatpak bundles for each architecture in ARCHES_TO_BUILD
  for ARCH in "${ARCHES_TO_BUILD[@]}"; do
    local build_args=()
    build_args+=("--user")
    build_args+=("--install-deps-from=flathub")
    build_args+=("--arch=$ARCH")
    build_args+=("--force-clean")
    build_args+=("--repo" "$REPO")
    build_args+=("$BUILD_DIR/$ARCH")
    build_args+=("${FLATPAK_MANIFEST}")

    APP_VERSION="$VERSION" APP_COMMIT="$COMMIT_HASH" flatpak-builder "${build_args[@]}"

    local bundle_args=()
    bundle_args+=("--arch=$ARCH")
    bundle_args+=("$REPO")
    bundle_args+=("dist/${FLATPAK_NAME}-$ARCH.flatpak")
    bundle_args+=("${FLATPAK_ID}")
    flatpak build-bundle "${bundle_args[@]}"
  done

  # Clean up temporary files and directories
  rm -rf vendor;
  rm -rf build/flatpak/.tmp-repo
  rm -rf build/flatpak/build-dir
}

parse_args "$@"
build_flatpak_bundle
