#!/usr/bin/env bash
# Called by semantic-release (verifyRelease step, also runs on --dry-run) with
# the next release version.
#
# - Writes the version to .release-version, so the "Determine version" job of
#   .github/workflows/release.yaml can pass it on to the build jobs.
# - If EXPECTED_VERSION is set (in the release job), fails unless the version
#   matches, so the release never gets binaries built for another version.
set -euo pipefail

VERSION="${1:?usage: release-version.sh <version>}"
echo "${VERSION}" > .release-version

if [ -n "${EXPECTED_VERSION:-}" ] && [ "${VERSION}" != "${EXPECTED_VERSION}" ]; then
    echo "ERROR: next release version ${VERSION} differs from built version ${EXPECTED_VERSION}" >&2
    exit 1
fi
