#!/usr/bin/env bash
# Uninstalls hue-lighter.
#
#   uninstall.sh [Linux|Darwin]           uninstall the app (config and API keys are kept)
#   uninstall.sh Darwin --legacy-only     only remove the former macOS pkg installation
set -euo pipefail

APP_ID="com.github.yveskaufmann.hue-lighter"
OS="${1:-$(uname -s)}"
MODE="${2:-}"

# Removes the former pkg installation: a system LaunchDaemon running as hidden
# "hue-lighter" user plus /Applications/hue-lighter.app. Its config and API key
# file are migrated to the per-user config dir, unless that one already exists.
remove_legacy_pkg_install() {
    local legacy_dir="/Library/Application Support/hue-lighter"
    local legacy_daemon="/Library/LaunchDaemons/${APP_ID}.plist"
    local user_dir="${HOME}/Library/Application Support/hue-lighter"

    if [ ! -e "${legacy_daemon}" ] && [ ! -e "${legacy_dir}" ] && [ ! -e "/Applications/hue-lighter.app" ] \
        && ! dscl . -read /Users/hue-lighter >/dev/null 2>&1 \
        && ! dscl . -read /Groups/hue-lighter >/dev/null 2>&1 \
        && ! pkgutil --pkg-info "${APP_ID}" >/dev/null 2>&1; then
        return 0
    fi

    echo "Removing former hue-lighter pkg installation (requires sudo)..."

    if [ -e "${legacy_daemon}" ]; then
        sudo launchctl bootout "system/${APP_ID}" 2>/dev/null || true
        sudo rm -f "${legacy_daemon}"
    fi

    if [ -d "${legacy_dir}" ]; then
        mkdir -p "${user_dir}"
        for file in config.yaml api-keys.json; do
            if [ -f "${legacy_dir}/${file}" ] && [ ! -e "${user_dir}/${file}" ]; then
                echo "Migrating ${legacy_dir}/${file} to ${user_dir}/${file}"
                (umask 077 && sudo cat "${legacy_dir}/${file}" > "${user_dir}/${file}")
            fi
        done
        sudo rm -rf "${legacy_dir}"
    fi

    sudo rm -rf "/Applications/hue-lighter.app"

    if dscl . -read /Users/hue-lighter >/dev/null 2>&1; then
        echo "Removing system user hue-lighter..."
        sudo dscl . -delete /Users/hue-lighter
    fi
    if dscl . -read /Groups/hue-lighter >/dev/null 2>&1; then
        echo "Removing system group hue-lighter..."
        sudo dscl . -delete /Groups/hue-lighter
    fi
    if pkgutil --pkg-info "${APP_ID}" >/dev/null 2>&1; then
        sudo pkgutil --forget "${APP_ID}" >/dev/null
    fi
}

case "${OS}" in
    Linux)
        flatpak uninstall io.github.yveskaufmann.HueLighter
        exit 0
        ;;

    Darwin)
        remove_legacy_pkg_install
        if [ "${MODE}" = "--legacy-only" ]; then
            exit 0
        fi

        echo "Uninstalling hue-lighter..."
        LAUNCH_AGENT="${HOME}/Library/LaunchAgents/${APP_ID}.plist"

        # Stops the app if it was started by launchd (on login) ...
        launchctl bootout "gui/$(id -u)/${APP_ID}" 2>/dev/null || true
        rm -f "${LAUNCH_AGENT}"
        # ... or manually. SIGTERM turns the lights off gracefully.
        pkill -TERM -x hue-lighter 2>/dev/null || true

        rm -rf "/Applications/Hue Lighter.app"

        echo "hue-lighter uninstalled."
        echo "Config is kept in ~/Library/Application Support/hue-lighter and the API key in the Keychain."
        ;;

    *)
        echo "ERROR: Unsupported OS: ${OS}"
        exit 1
        ;;
esac
