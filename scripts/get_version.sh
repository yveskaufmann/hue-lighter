#!/usr/bin/env bash

##
# Retrieves the version of the application to be used in the build output
#
# Usage: get_version

set -euo pipefail


# Version global variable defaulting to empty if not set externally
VERSION="${VERSION:-}"

##
# Determines the version of the application to be used in the build output
#
# Usage: get_version
#
get_version() {
  if [[ -n "$VERSION" ]]; then
    echo "$VERSION"
    return 0
  fi

  local version="$(git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)"
  version="${version#v}"

  echo "$version"
  return 0
}


get_version
