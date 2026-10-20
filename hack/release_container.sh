#!/usr/bin/env bash
# Builds the Linux and Windows releases with GoReleaser (.goreleaser.yaml)
# inside an Ubuntu 22.04 container:
#
#   docker run --rm -v "$PWD:/src" -w /src ubuntu:22.04 hack/release_container.sh [goreleaser flags]
#
# .github/workflows/release.yaml and `make release-snapshot` run it; it only
# works in that container. Ubuntu 22.04 keeps the binaries on glibc 2.34,
# libFLAC and libogg are linked statically (hack/pkg-config-static), so
# libasound.so.2 is the only other library they need. Windows builds use
# llvm-mingw.

set -euo pipefail

GO_VERSION=1.26.10
GO_SHA256=ccfe7ae822d2c2f41f4fe7c956199967a4dac22d0ee0d2b2bfdb440aab3b20c5
GORELEASER_VERSION=2.18.3
GORELEASER_SHA256=b8dee4728bca1aba0d71a832a17cd0a3799f75372878e7ffa36aef2343bce0bc
LLVM_MINGW_VERSION=20261006
LLVM_MINGW_SHA256=5f9c6ed95b2d4bdb2869a488c5fd5857fbdabcf288a0aa3eb1da43f6a08d8ab4

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

# fetch URL SHA256 FILE
fetch() {
	curl -fsSL --retry 3 -o "$3" "$1"
	echo "$2  $3" | sha256sum -c --quiet -
}

export DEBIAN_FRONTEND=noninteractive
# arm64 libraries come from ports.ubuntu.com
sed -i 's/^deb http/deb [arch=amd64] http/' /etc/apt/sources.list
for suite in jammy jammy-updates jammy-security; do
	echo "deb [arch=arm64] http://ports.ubuntu.com/ubuntu-ports $suite main universe"
done >/etc/apt/sources.list.d/arm64.list
dpkg --add-architecture arm64
apt-get update -qq
apt-get install -y -qq --no-install-recommends \
	ca-certificates curl git xz-utils pkg-config gcc libc6-dev \
	gcc-aarch64-linux-gnu libc6-dev-arm64-cross \
	libflac-dev libogg-dev libasound2-dev \
	libflac-dev:arm64 libogg-dev:arm64 libasound2-dev:arm64 >/dev/null

cd /tmp
fetch "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" "$GO_SHA256" go.tar.gz
tar -C /usr/local -xzf go.tar.gz
fetch "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz" \
	"$GORELEASER_SHA256" goreleaser.tar.gz
tar -C /usr/local/bin -xzf goreleaser.tar.gz goreleaser
llvm=llvm-mingw-${LLVM_MINGW_VERSION}-ucrt-ubuntu-22.04-x86_64
fetch "https://github.com/mstorsjo/llvm-mingw/releases/download/${LLVM_MINGW_VERSION}/${llvm}.tar.xz" \
	"$LLVM_MINGW_SHA256" llvm-mingw.tar.xz
tar -C /opt -xJf llvm-mingw.tar.xz

export PATH=/usr/local/go/bin:/opt/$llvm/bin:$PATH
export PKG_CONFIG=$ROOT/hack/pkg-config-static

cd "$ROOT"
git config --global --add safe.directory "$ROOT"
status=0
goreleaser release --clean "$@" || status=$?
# A container that runs as root hands dist back to whoever started it.
if [[ -n ${HOST_UID:-} && -d dist ]]; then
	chown -R "$HOST_UID:${HOST_GID:-$HOST_UID}" dist
fi
exit "$status"
