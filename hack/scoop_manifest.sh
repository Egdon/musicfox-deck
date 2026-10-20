#!/usr/bin/env bash
# Prints the Scoop manifest of a release for the Egdon/scoop-bucket
# repository (.github/workflows/package-managers.yaml):
#
#   hack/scoop_manifest.sh v1.0.0 SHA256SUMS [URL]
#
# URL is where the Windows archives are, by default the release's downloads.
# The data stays in %AppData%\go-musicfox, shared with the other ways of
# installing tapefox and with go-musicfox.

set -euo pipefail

tag=$1
version=${tag#v}
sums=$2
repo=https://github.com/Egdon/tapefox
base=${3:-$repo/releases/download/$tag}

# checksum ARCH
checksum() {
	local file=tapefox_${version}_windows_$1.zip
	awk -v f="$file" '$2 == f || $2 == "*" f { print $1; found = 1 } END { exit !found }' "$sums" ||
		{ echo "$sums has no checksum for $file" >&2; return 1; }
}

amd64=$(checksum amd64)
arm64=$(checksum arm64)

cat <<EOF
{
    "version": "${version}",
    "description": "磁带卡座风格的命令行网易云音乐客户端 · A terminal client for NetEase Cloud Music with a cassette-deck interface",
    "homepage": "${repo}",
    "license": "GPL-3.0-only",
    "notes": "推荐在 Windows Terminal 中运行 tapefox。Run tapefox in Windows Terminal.",
    "architecture": {
        "64bit": {
            "url": "${base}/tapefox_${version}_windows_amd64.zip",
            "hash": "${amd64}",
            "extract_dir": "tapefox_${version}_windows_amd64"
        },
        "arm64": {
            "url": "${base}/tapefox_${version}_windows_arm64.zip",
            "hash": "${arm64}",
            "extract_dir": "tapefox_${version}_windows_arm64"
        }
    },
    "bin": "tapefox.exe",
    "checkver": "github",
    "autoupdate": {
        "architecture": {
            "64bit": {
                "url": "${repo}/releases/download/v\$version/tapefox_\$version_windows_amd64.zip",
                "extract_dir": "tapefox_\$version_windows_amd64"
            },
            "arm64": {
                "url": "${repo}/releases/download/v\$version/tapefox_\$version_windows_arm64.zip",
                "extract_dir": "tapefox_\$version_windows_arm64"
            }
        },
        "hash": {
            "url": "\$baseurl/SHA256SUMS"
        }
    }
}
EOF
