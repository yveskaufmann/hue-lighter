#!/usr/bin/env bash
# Builds static Linux binaries (amd64, arm64) and packs each into a tarball.
#
# Output:
#   dist/hue-lighter_<version>_linux_<arch>.tar.gz
#
# VERSION defaults to the latest git tag (without leading "v").
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)}"
VERSION="${VERSION#v}"

CA_BUNDLE="internal/hue_client/certs/cacert_bundle.pem"
if [ ! -f "${CA_BUNDLE}" ]; then
    echo "ERROR: ${CA_BUNDLE} not found, run 'make embed-ca-bundle' first." >&2
    exit 1
fi

mkdir -p dist
for arch in amd64 arm64; do
    name="hue-lighter_${VERSION}_linux_${arch}"
    work_dir="bin/linux/${name}"
    rm -rf "${work_dir}" "dist/${name}.tar.gz"
    mkdir -p "${work_dir}"

    echo "Building ${name}..."
    CGO_ENABLED=0 GOOS=linux GOARCH="${arch}" \
        go build -trimpath -ldflags "-s -w" -o "${work_dir}/hue-lighter" ./cmd/hue-lighter
    cp LICENSE README.md configs/config.example.yaml "${work_dir}/"

    tar -C "bin/linux" -czf "dist/${name}.tar.gz" "${name}"
    echo "Created dist/${name}.tar.gz"
done
