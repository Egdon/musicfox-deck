#!/usr/bin/env bash
# Builds an RPM of this checkout into dist/, for Fedora and other RPM
# distributions. Needs rpmbuild (rpm-build) and what `make build` needs:
# Go, gcc, flac-devel and alsa-lib-devel.
set -euo pipefail
cd "$(dirname "$0")/.."

# Conda and similar environments set their own compiler and flags, which
# miss the system's FLAC and ALSA headers; build with the system's.
cc=/usr/bin/gcc cxx=/usr/bin/g++
[ -x "$cc" ] || cc=gcc
[ -x "$cxx" ] || cxx=g++
env -u CFLAGS -u CPPFLAGS -u LDFLAGS -u CXXFLAGS CC="$cc" CXX="$cxx" make build

# RPM versions have no hyphens: a snapshot of 5.1.0 is 5.1.0^<date>git<commit>.
base=$(git describe --tags --abbrev=0 2>/dev/null || echo v0)
version="${base#v}^$(git log -1 --date=format:%Y%m%d --format=%cd)git$(git rev-parse --short HEAD)"

top=$(mktemp -d)
trap 'rm -rf "$top"' EXIT
mkdir -p "$top/SOURCES" dist
cp bin/musicfox LICENSE README.md "$top/SOURCES/"
rpmbuild -bb --quiet --define "_topdir $top" --define "deck_version $version" deploy/rpm/musicfox-deck.spec
find "$top/RPMS" -name '*.rpm' -exec cp {} dist/ \; -exec basename {} \;
