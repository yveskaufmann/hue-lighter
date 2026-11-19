#!/usr/bin/env bash

#
# Adds a release entry to the Flatpak manifest file.
#

set -euo pipefail

MANIFEST_FILE="build/flatpak/io.github.yveskaufmann.HueLighter.metainfo.xml"
VERSION=""
DATE=""
DESCRIPTION=""
VERSION_SET=false
OVERWRITE=false

##
# Displays the usage information for the release script
#
# Usage: usage
#
usage() {
  echo "Usage: $0 <version> [options]"
  echo "Adds a release entry to the Flatpak manifest file"
  echo ""
  echo "Options:"
  echo "  -h, --help               Show this help message"
  echo "  -d, --date <date>        Release date (or 'now'; default: today)"
  echo "      --description <text> Release description"
  echo "  -o, --overwrite          Update an existing release version"
  echo ""
}

##
# Parses the command-line arguments for the release script
#
# Usage: parse_args [arguments]
#
parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        usage
        exit 0
        ;;
      -d|--date)
        if [[ $# -lt 2 ]]; then
          echo "Missing value for $1" >&2
          exit 1
        fi
        DATE="$2"
        shift 2
        ;;
      --description)
        if [[ $# -lt 2 ]]; then
          echo "Missing value for $1" >&2
          exit 1
        fi
        DESCRIPTION="$2"
        shift 2
        ;;
      -o|--overwrite)
        OVERWRITE=true
        shift
        ;;
      --)
        shift
        if [[ $# -ne 1 || "$VERSION_SET" == true ]]; then
          echo "Exactly one version argument is required" >&2
          exit 1
        fi
        VERSION="$1"
        VERSION_SET=true
        shift
        ;;
      -*)
        echo "Unknown argument: $1" >&2
        exit 1
        ;;
      *)
        if [[ "$VERSION_SET" == true ]]; then
          echo "Only the version may be given as a positional argument" >&2
          exit 1
        fi
        VERSION="$1"
        VERSION_SET=true
        shift
        ;;
    esac
  done

  if [[ "$VERSION_SET" != true || -z "$VERSION" ]]; then
    echo "Version is required" >&2
    usage >&2
    exit 1
  fi
}

##
# Adds release information to the Flatpak manifest file
#
# Usage: add_release_info
#
add_release_info() {
  if [[ ! -f "$MANIFEST_FILE" ]]; then
    echo "Manifest file not found: $MANIFEST_FILE" >&2
    exit 1
  fi

  if [[ -z "$DATE" || "$DATE" == "now" ]]; then
    DATE="$(date -u +%Y-%m-%d)"
  else
    DATE="$(date -u -d "$DATE" +%Y-%m-%d)"
  fi

  local release_versions
  release_versions="$(yq -p xml -o yml '.component.releases.release[].+@version' "$MANIFEST_FILE")"

  if grep -Fxq -- "$VERSION" <<< "$release_versions"; then
    if [[ "$OVERWRITE" != true ]]; then
      echo "Version $VERSION is already in use" >&2
      exit 1
    fi

    local update_expression
    update_expression='(.component.releases.release[] | select(.+@version == strenv(VERSION)) | .+@date) = strenv(DATE)'
    if [[ -n "$DESCRIPTION" ]]; then
      update_expression+=' | (.component.releases.release[] | select(.+@version == strenv(VERSION)) | .description) = strenv(DESCRIPTION)'
    fi

    VERSION="$VERSION" DATE="$DATE" DESCRIPTION="$DESCRIPTION" \
      yq -p xml -o xml "$update_expression" "$MANIFEST_FILE" > "$MANIFEST_FILE.tmp"
    mv "$MANIFEST_FILE.tmp" "$MANIFEST_FILE"
    return
  fi

  local version_patch
  version_patch="$(jq -n --arg version "$VERSION" --arg date "$DATE" \
    '{"+@version": $version, "+@date": $date}')"

  if [[ -n "$DESCRIPTION" ]]; then
    # TODO: convert description to the appropriate format for the manifest if needed
    version_patch="$(jq --arg description "$DESCRIPTION" '. + {description: $description}' <<< "$version_patch")"
  fi

  yq -p xml -o xml '.component.releases.release +=['"$version_patch"']' "$MANIFEST_FILE" > "$MANIFEST_FILE.tmp"
  mv "$MANIFEST_FILE.tmp" "$MANIFEST_FILE"

  # TODO: Format with same xml formatter as used by the rest of the manifest file
}

parse_args "$@"
add_release_info

