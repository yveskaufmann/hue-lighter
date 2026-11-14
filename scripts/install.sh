#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [ ! -f "${REPO_ROOT}/configs/certs/cacert_bundle.pem" ]; then
    echo "ERROR: Hue Bridge Root CA bundle not found."
    echo "Please ensure configs/certs/cacert_bundle.pem exists."
    echo "See: https://developers.meethue.com/develop/application-design-guidance/using-https/"
    exit 1
fi

OS="${1:-$(uname -s)}"

case "${OS}" in
    Linux)
        make -C "${REPO_ROOT}" flatpak
        exit 0
        ;;

    Darwin)
        APP_NAME="Hue Lighter"
        APP_BUILD="${REPO_ROOT}/bin/macos/${APP_NAME}.app"
        APP_INSTALL="/Applications/${APP_NAME}.app"

        # Remove the former pkg installation (LaunchDaemon + system user), otherwise
        # hue-lighter would run twice.
        bash "${SCRIPT_DIR}/uninstall.sh" Darwin --legacy-only

        bash "${SCRIPT_DIR}/package-macos.sh"

        if pgrep -x hue-lighter >/dev/null; then
            echo "Stopping running hue-lighter..."
            pkill -TERM -x hue-lighter || true
            sleep 2
        fi

        echo "Installing ${APP_INSTALL}..."
        rm -rf "${APP_INSTALL}"
        ditto "${APP_BUILD}" "${APP_INSTALL}"

        # Starts the app; on first start it registers the LaunchAgent for login.
        open "${APP_INSTALL}"
        echo "hue-lighter installed. Config: ~/Library/Application Support/hue-lighter, logs: ~/Library/Logs/hue-lighter.log"
        ;;

    *)
        echo "ERROR: Unsupported OS: ${OS}"
        echo "Supported platforms: Linux, Darwin (macOS)"
        exit 1
        ;;
esac
