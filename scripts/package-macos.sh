#!/usr/bin/env bash

# Builds "Hue Lighter.app" (universal arm64 + amd64, ad-hoc signed) and zips it
# for distribution via GitHub releases and the Homebrew cask.
#
# Output:
#   bin/macos/Hue Lighter.app
#   dist/hue-lighter_<version>_macos_universal.zip
#
# VERSION defaults to the latest git tag (without leading "v").
set -euo pipefail

#  Version of the macOS app to build
VERSION="${VERSION:-}"

#  Name of the macOS app bundle and related paths
APP_NAME="Hue Lighter"

#  Executable name, output directory, and app bundle path
EXECUTABLE="hue-lighter"

#  Output directory and app bundle path
OUT_DIR="bin/macos"

#  Path to the macOS app bundle
APP="${OUT_DIR}/${APP_NAME}.app"

##
# Parses the command-line arguments for the macOS packaging script
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
  echo "Build and package the macOS app into a zip archive"
  echo ""
  echo "Options:"
  echo "  -h, --help    Show this help message"
  echo "  -v, --version <version> Specify the version to use for the build"
  echo ""
}

build_universal_binary() {
  local ZIP="dist/hue-lighter_${VERSION}_macos_universal.zip"
  # CFBundleVersion must be numeric (e.g. 1.2.0); use 0.0.0 for untagged dev builds.
  local BUNDLE_VERSION="$(printf '%s' "${VERSION}" | grep -Eo '^[0-9]+\.[0-9]+(\.[0-9]+)?' || echo 0.0.0)"

  rm -rf "${OUT_DIR}" "${ZIP}"
  mkdir -p "${APP}/Contents/MacOS" "${APP}/Contents/Resources" dist

  # Build the universal binary for each architecture (arm64 and amd64)
  for arch in arm64 amd64; do
      bash scripts/build.sh --os darwin --arch "${arch}" -a --version "${VERSION}"
  done
  echo "Built universal binary for ${APP_NAME} ${VERSION}"

  lipo -create -output "${APP}/Contents/MacOS/${EXECUTABLE}" \
      "${OUT_DIR}/${EXECUTABLE}-arm64" "${OUT_DIR}/${EXECUTABLE}-amd64"
  rm "${OUT_DIR}/${EXECUTABLE}-arm64" "${OUT_DIR}/${EXECUTABLE}-amd64"

  sed "s/@VERSION@/${BUNDLE_VERSION}/g" build/macos/Info.plist > "${APP}/Contents/Info.plist"
  plutil -lint "${APP}/Contents/Info.plist" >/dev/null
  cp build/macos/AppIcon.icns "${APP}/Contents/Resources/AppIcon.icns"
  printf 'APPL????' > "${APP}/Contents/PkgInfo"

  # Ad-hoc signature (no Apple Developer ID): required to run on Apple Silicon and
  # seals the bundle, Gatekeeper still treats the app as unidentified developer.
  xattr -cr "${APP}"
  codesign --force --sign - --timestamp=none "${APP}"
  codesign --verify --strict "${APP}"

  # No extended attributes / AppleDouble files (__MACOSX), the signature lives in the bundle.
  ditto -c -k --norsrc --noextattr --keepParent "${APP}" "${ZIP}"

  echo "Created ${APP}"
  echo "Created ${ZIP}"
}

if [ "$(uname -s)" != "Darwin" ]; then
    echo "ERROR: macOS packaging requires macOS (lipo, codesign, ditto)." >&2
    exit 1
fi

parse_args "$@"
build_universal_binary










