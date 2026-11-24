#!/usr/bin/env bash

set -euo pipefail

# Version of the new flatpak bundle.
VERSION=1.2.0

# Location of the temp repository
TEMP_REPO=$(mktemp -d)

# Build the flatpak
scripts/package-flatpak.sh -v $VERSION

# Preparse the temp repository
ostree --repo="$TEMP_REPO" init --mode=bare-user
flatpak build-import-bundle "$TEMP_REPO" dist/hue-lighter-x86_64.flatpak

ostree --repo="$TEMP_REPO" cat  app/io.github.yveskaufmann.HueLighter/x86_64/master /files/share/metainfo/io.github.yveskaufmann.HueLighter.metainfo.xml
ostree --repo="$TEMP_REPO" cat  app/io.github.yveskaufmann.HueLighter/x86_64/master /files/bin/hue-lighter > $TEMP_REPO/bin

chmod +x $TEMP_REPO/bin
$TEMP_REPO/bin --version

rm -rf TEMP_REPO


