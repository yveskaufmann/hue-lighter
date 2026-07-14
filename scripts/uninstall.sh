#!/usr/bin/env bash
set -euo pipefail

APP_ID="com.github.yveskaufmann.hue-lighter"
OS="${1:-$(uname -s)}"

case "${OS}" in
    Linux)
        echo "Uninstalling hue-lighter (Linux/systemd)..."

        sudo systemctl stop hue-lighter 2>/dev/null || true
        sudo systemctl disable hue-lighter 2>/dev/null || true
        sudo rm -f /usr/bin/hue-lighter
        sudo rm -f /etc/systemd/system/hue-lighter.service
        sudo rm -rf /etc/hue-lighter
        sudo systemctl daemon-reload

        echo "hue-lighter uninstalled."
        ;;

    Darwin)
        echo "Uninstalling hue-lighter (macOS/launchd)..."
        LAUNCH_DAEMON="/Library/LaunchDaemons/${APP_ID}.plist"

        if launchctl list | grep -q "${APP_ID}"; then
            sudo launchctl bootout system "${LAUNCH_DAEMON}" 2> /dev/null || true
        fi
        sudo rm -rf "/Applications/hue-lighter.app"
	    sudo rm -rf "/Library/Application Support/hue-lighter"
        sudo rm -f "${LAUNCH_DAEMON}"

        if dscl . -read /Users/hue-lighter >/dev/null 2>&1; then 
            echo "Removing System-User hue-lighter..." 
            sudo dscl . -delete /Users/hue-lighter 
	    fi

	    if dscl . -read /Groups/hue-lighter >/dev/null 2>&1; then 
            echo "Removing system-group hue-lighter..." 
            sudo dscl . -delete /Groups/hue-lighter 
	    fi

        if pkgutil --packages | grep -q "${APP_ID}"; then
            sudo pkgutil --forget "${APP_ID}"
        fi

        echo "hue-lighter uninstalled."
        ;;

    *)
        echo "ERROR: Unsupported OS: ${OS}"
        exit 1
        ;;
esac
