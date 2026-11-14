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

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

if [ "$(uname -s)" != "Darwin" ]; then
    echo "ERROR: macOS packaging requires macOS (lipo, codesign, ditto)." >&2
    exit 1
fi

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)}"
VERSION="${VERSION#v}"
# CFBundleVersion must be numeric (e.g. 1.2.0); use 0.0.0 for untagged dev builds.
BUNDLE_VERSION="$(printf '%s' "${VERSION}" | grep -Eo '^[0-9]+\.[0-9]+(\.[0-9]+)?' || echo 0.0.0)"

CA_BUNDLE="internal/hue_client/certs/cacert_bundle.pem"
if [ ! -f "${CA_BUNDLE}" ]; then
    echo "ERROR: ${CA_BUNDLE} not found, run 'make embed-ca-bundle' first." >&2
    exit 1
fi

APP_NAME="Hue Lighter"
EXECUTABLE="hue-lighter"
OUT_DIR="bin/macos"
APP="${OUT_DIR}/${APP_NAME}.app"
ZIP="dist/hue-lighter_${VERSION}_macos_universal.zip"

rm -rf "${OUT_DIR}" "${ZIP}"
mkdir -p "${APP}/Contents/MacOS" "${APP}/Contents/Resources" dist

echo "Building ${APP_NAME} ${VERSION} (bundle version ${BUNDLE_VERSION})..."
for arch in arm64 amd64; do
    CGO_ENABLED=0 GOOS=darwin GOARCH="${arch}" \
        go build -trimpath -ldflags "-s -w" -o "${OUT_DIR}/${EXECUTABLE}-${arch}" ./cmd/hue-lighter
done
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
