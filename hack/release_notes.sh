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

v=$version
macos_x86=musicfox-deck_${v}_darwin_amd64.tar.gz
macos_arm=musicfox-deck_${v}_darwin_arm64.tar.gz
windows_x86=musicfox-deck_${v}_windows_amd64.zip
windows_arm=musicfox-deck_${v}_windows_arm64.zip
deb_x86=musicfox-deck_${v}_amd64.deb
deb_arm=musicfox-deck_${v}_arm64.deb
rpm_x86=musicfox-deck-${v}-1.x86_64.rpm
rpm_arm=musicfox-deck-${v}-1.aarch64.rpm
arch_x86=musicfox-deck-${v}-1-x86_64.pkg.tar.zst
arch_arm=musicfox-deck-${v}-1-aarch64.pkg.tar.zst
linux_x86=musicfox-deck_${v}_linux_amd64.tar.gz
linux_arm=musicfox-deck_${v}_linux_arm64.tar.gz

cat <<EOF
$zh

## 下载

| 平台 | x86_64 | ARM64 |
|---|---|---|
| macOS | \`$macos_x86\`（Intel） | \`$macos_arm\`（Apple 芯片） |
| Windows | \`$windows_x86\` | \`$windows_arm\` |
| Debian / Ubuntu | \`$deb_x86\` | \`$deb_arm\` |
| Fedora / openSUSE | \`$rpm_x86\` | \`$rpm_arm\` |
| Arch Linux | \`$arch_x86\` | \`$arch_arm\` |
| 其他 Linux 发行版 | \`$linux_x86\` | \`$linux_arm\` |

- 文件名中的 \`amd64\` 与 \`x86_64\` 均指 x86_64 架构，\`arm64\` 与 \`aarch64\` 均指 ARM64 架构。
- macOS 也可通过 Homebrew 安装：\`brew install egdon/tap/musicfox-deck\`。\`*_macapp.zip\` 为 .app 版本，系统通知中显示应用图标。
- Linux 版需要 glibc 2.34 及以上与 ALSA（\`libasound.so.2\`）。
- macOS 与 Windows 版未签名，首次运行前的设置参见 [README](https://github.com/Egdon/musicfox-deck#安装)。
- 各文件的 SHA-256 校验值见 \`SHA256SUMS\`。

---
EOF
if [[ -n $en ]]; then
	printf '\n%s\n' "$en"
fi
cat <<EOF

## Downloads

| Platform | x86_64 | ARM64 |
|---|---|---|
| macOS | \`$macos_x86\` (Intel) | \`$macos_arm\` (Apple silicon) |
| Windows | \`$windows_x86\` | \`$windows_arm\` |
| Debian / Ubuntu | \`$deb_x86\` | \`$deb_arm\` |
| Fedora / openSUSE | \`$rpm_x86\` | \`$rpm_arm\` |
| Arch Linux | \`$arch_x86\` | \`$arch_arm\` |
| Other Linux distributions | \`$linux_x86\` | \`$linux_arm\` |

- \`amd64\` and \`x86_64\` in file names both mean x86_64; \`arm64\` and \`aarch64\` both mean ARM64.
- macOS can also install with Homebrew: \`brew install egdon/tap/musicfox-deck\`. \`*_macapp.zip\` is the .app version, whose notifications show the app icon.
- The Linux builds need glibc 2.34 or later and ALSA (\`libasound.so.2\`).
- The macOS and Windows builds are not signed; see the [README](https://github.com/Egdon/musicfox-deck/blob/master/README.en.md#installation) for the first run.
- \`SHA256SUMS\` lists the SHA-256 checksum of each file.
EOF
