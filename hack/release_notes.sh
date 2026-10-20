#!/usr/bin/env bash
# Prints the release notes of a version, its section of CHANGELOG.md in
# Chinese and then in English, each followed by the files to download, for
# the draft release that .github/workflows/release.yaml creates:
#
#   hack/release_notes.sh v1.0.0

set -euo pipefail

tag=$1
version=${tag#v}
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

# section VERSION prints the CHANGELOG.md section of VERSION. A section
# starts with its version above a line of '=' and ends before the next such
# heading or at a line of '---'. The date under the heading and the blank
# lines around the section are left out.
section() {
	awk -v v="$1" '
		/^=+$/ {
			if (on) { n--; exit }
			if (prev == v) { on = 1; next }
		}
		on && /^---$/ { exit }
		on { line[++n] = $0 }
		{ prev = $0 }
		END {
			if (!on) exit 1
			first = 1
			if (line[1] ~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]$/) first = 2
			while (first <= n && line[first] ~ /^[ \t]*$/) first++
			while (n >= first && line[n] ~ /^[ \t]*$/) n--
			for (i = first; i <= n; i++) print line[i]
		}
	' "$ROOT/CHANGELOG.md"
}

# A prerelease such as 1.0.0-rc.1 has the notes of 1.0.0.
notes=$(section "$version" || section "${version%%-*}") ||
	{ echo "CHANGELOG.md has no section for $version" >&2; exit 1; }

# The Chinese notes come first; a line "<!-- en -->" starts the English ones.
# Each language gets its own table of downloads.
zh=$(printf '%s\n' "$notes" | sed '/^<!-- en -->$/,$d')
en=$(printf '%s\n' "$notes" | sed -n '/^<!-- en -->$/,$p' | sed '1d;/./,$!d')

# f SUFFIX_A SUFFIX_B prints the file names of both architectures.
f() { printf '`musicfox-deck%s` / `musicfox-deck%s`' "$1" "$2"; }
v=$version
macos=$(f "_${v}_darwin_arm64.tar.gz" "_${v}_darwin_amd64.tar.gz")
windows=$(f "_${v}_windows_amd64.zip" "_${v}_windows_arm64.zip")
deb=$(f "_${v}_amd64.deb" "_${v}_arm64.deb")
rpm=$(f "-${v}-1.x86_64.rpm" "-${v}-1.aarch64.rpm")
arch=$(f "-${v}-1-x86_64.pkg.tar.zst" "-${v}-1-aarch64.pkg.tar.zst")
linux=$(f "_${v}_linux_amd64.tar.gz" "_${v}_linux_arm64.tar.gz")

cat <<EOF
$zh

## 下载

| 平台 | 文件 |
|---|---|
| macOS（Apple 芯片 / Intel） | $macos；或 \`brew install egdon/tap/musicfox-deck\` |
| Windows（x64 / ARM64） | $windows |
| Debian / Ubuntu | $deb |
| Fedora / openSUSE | $rpm |
| Arch Linux | $arch |
| 其他 Linux 发行版 | $linux |

- Linux 版需要 glibc 2.34 及以上与 ALSA（\`libasound.so.2\`）。
- macOS 与 Windows 版未签名，首次运行前的设置参见 [README](https://github.com/Egdon/musicfox-deck#安装)。\`*_macapp.zip\` 为 macOS 的 .app 版本，系统通知中显示应用图标。
- 各文件的 SHA-256 校验值见 \`SHA256SUMS\`。

---
EOF
if [[ -n $en ]]; then
	printf '\n%s\n' "$en"
fi
cat <<EOF

## Downloads

| Platform | Files |
|---|---|
| macOS (Apple silicon / Intel) | $macos; or \`brew install egdon/tap/musicfox-deck\` |
| Windows (x64 / ARM64) | $windows |
| Debian / Ubuntu | $deb |
| Fedora / openSUSE | $rpm |
| Arch Linux | $arch |
| Other Linux distributions | $linux |

- The Linux builds need glibc 2.34 or later and ALSA (\`libasound.so.2\`).
- The macOS and Windows builds are not signed; see the [README](https://github.com/Egdon/musicfox-deck/blob/master/README.en.md#installation) for the first run. \`*_macapp.zip\` is the macOS .app version, whose notifications show the app icon.
- \`SHA256SUMS\` lists the SHA-256 checksum of each file.
EOF
