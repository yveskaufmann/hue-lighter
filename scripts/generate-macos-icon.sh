#!/usr/bin/env bash
# Regenerates build/macos/AppIcon.icns from internal/tray/assets/icon.svg.
# Requires rsvg-convert (brew install librsvg) and iconutil (macOS).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SVG="${REPO_ROOT}/internal/tray/assets/icon.svg"
ICNS="${REPO_ROOT}/build/macos/AppIcon.icns"

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT
ICONSET="${WORK_DIR}/AppIcon.iconset"
mkdir -p "${ICONSET}"

for size in 16 32 128 256 512; do
    rsvg-convert -w "${size}" -h "${size}" "${SVG}" -o "${ICONSET}/icon_${size}x${size}.png"
    rsvg-convert -w "$((size * 2))" -h "$((size * 2))" "${SVG}" -o "${ICONSET}/icon_${size}x${size}@2x.png"
done

iconutil --convert icns --output "${ICNS}" "${ICONSET}"
echo "Generated ${ICNS}"
