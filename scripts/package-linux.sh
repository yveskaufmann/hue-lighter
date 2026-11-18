#!/usr/bin/env bash
# Builds static Linux binaries (amd64, arm64) and packs each into a tarball.
#
# Output:
#   dist/hue-lighter_<version>_linux_<arch>.tar.gz
#
# VERSION defaults to the latest git tag (without leading "v").
set -euo pipefail

#  Version of the app to build
VERSION="${VERSION:-}"

##
# Parses the command-line arguments for the build script
#
# Usage: parse_args [arguments]
#
parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help) usage; exit 0 ;;
      -v|--version)
        VERSION="$2"; shift 2 ;;
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
# Displays the usage information for the build script
#
# Usage: usage
#
usage() {
  echo "Usage: $0 [options]"
  echo "Build and package the Linux binaries into tarballs"
  echo ""
  echo "Options:"
  echo "  -h, --help    Show this help message"
  echo "  -v, --version <version> Specify the version to use for the build"
  echo ""
}


##
# Builds the Linux binaries for the specified architectures
# and places them in the appropriate directories.
#
# Usage: build_linux_binaries
#
build_linux_binaries() {
  mkdir -p dist
  for arch in amd64 arm64; do
    archive_name="hue-lighter_${VERSION}_linux_${arch}"

    work_dir="$(mktemp -d)"
    trap 'rm -rf "$work_dir"' EXIT

    bash scripts/build.sh --os linux --arch "$arch" --version "$VERSION"
    cp LICENSE README.md configs/config.example.yaml "${work_dir}/"
    cp "bin/linux/hue-lighter" "${work_dir}/"

    tar -C "$work_dir" -czf "dist/${archive_name}.tar.gz" .

    echo "Created dist/${archive_name}.tar.gz"
  done
}


parse_args "$@"
build_linux_binaries
