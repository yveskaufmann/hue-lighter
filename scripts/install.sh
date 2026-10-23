#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
APP_NAME="hue-lighter"
APP_ID="com.github.yveskaufmann.hue-lighter"
VERSION="${HUE_LIGHTER_VERSION:-1.0.0}"

if [ ! -f "${REPO_ROOT}/configs/certs/cacert_bundle.pem" ]; then
    echo "ERROR: Hue Bridge Root CA bundle not found."
    echo "Please ensure configs/certs/cacert_bundle.pem exists."
    echo "See: https://developers.meethue.com/develop/application-design-guidance/using-https/"
    exit 1
fi

OS="${1:-$(uname -s)}"

case "${OS}" in
    Linux)
        make flatpak 
        echo "On Linux hue-lighter is distributed as Flatpak, run: make flatpak"
        exit 0
        ;;

    Darwin)
        echo "Installing hue-lighter for macOS (pkg/app)..."

        APP_EXEC="${REPO_ROOT}/build/macos/Applications/hue-lighter.app/Contents/MacOS/hue-lighter"
        DAEMON_EXEC="${REPO_ROOT}/build/macos/Library/Application Support/hue-lighter/hue-lighter-daemon"
        APP_INFO="${REPO_ROOT}/build/macos/Applications/hue-lighter.app/Contents/Info.plist"
        CA_CERT_DST="${REPO_ROOT}/build/macos/Library/Application Support/hue-lighter/cacert_bundle.pem"
        PKG_PATH="${REPO_ROOT}/bin/${APP_NAME}-${VERSION}.pkg"

        if [ ! -f "${APP_INFO}" ]; then
            echo "ERROR: macOS app bundle template not found at ${APP_INFO}."
            exit 1
        fi

        find "${REPO_ROOT}/build/macos" -name ".DS_Store" -depth -exec rm {} \;
        mkdir -p "${REPO_ROOT}/bin"

        plutil "${APP_INFO}"
        cp "${REPO_ROOT}/bin/hue-lighter" "${APP_EXEC}"
        cp "${REPO_ROOT}/bin/hue-lighter" "${DAEMON_EXEC}"
        cp "${REPO_ROOT}/configs/certs/cacert_bundle.pem" "${CA_CERT_DST}"

        chmod +x "${APP_EXEC}"
        chmod +x "${DAEMON_EXEC}"

        pkgbuild --root "${REPO_ROOT}/build/macos" \
            --scripts "${REPO_ROOT}/build/macos/scripts" \
            --filter "/scripts" \
            --filter "\\.DS_Store" \
            --identifier "${APP_ID}" \
            --version "${VERSION}" \
            --install-location "/" \
            "${PKG_PATH}"

        if [ "${HUE_LIGHTER_MAC_PACKAGE_ONLY:-0}" = "1" ]; then
            echo "hue-lighter macOS package created at ${PKG_PATH}."
            exit 0
        fi

        sudo installer -pkg "${PKG_PATH}" -target /
        echo "hue-lighter installed via macOS pkg."
        ;;

    *)
        echo "ERROR: Unsupported OS: ${OS}"
        echo "Supported platforms: Linux, Darwin (macOS)"
        exit 1
        ;;
esac
