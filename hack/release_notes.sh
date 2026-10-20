#!/usr/bin/env bash
# Prints the release notes of a version, its section of CHANGELOG.md followed
# by the files to download, for the draft release that
# .github/workflows/release.yaml creates:
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
section "$version" || section "${version%%-*}" ||
	{ echo "CHANGELOG.md has no section for $version" >&2; exit 1; }

cat <<EOF

## 下载

| 系统 | 文件 |
|---|---|
| macOS（Apple 芯片 / Intel） | \`musicfox-deck_${version}_darwin_arm64.tar.gz\` / \`darwin_amd64.tar.gz\`，或用 Homebrew：\`brew install egdon/tap/musicfox-deck\` |
| Windows（x64 / ARM64） | \`musicfox-deck_${version}_windows_amd64.zip\` / \`windows_arm64.zip\` |
| Debian、Ubuntu | \`musicfox-deck_${version}_amd64.deb\` / \`arm64.deb\` |
| Fedora、openSUSE | \`musicfox-deck-${version}-1.x86_64.rpm\` / \`aarch64.rpm\` |
| Arch Linux | \`musicfox-deck-${version}-1-x86_64.pkg.tar.zst\` / \`aarch64.pkg.tar.zst\` |
| 其他 Linux | \`musicfox-deck_${version}_linux_amd64.tar.gz\` / \`linux_arm64.tar.gz\` |

- Linux 版需要 glibc 2.34 以上（2021 年以后的发行版）和 ALSA。
- macOS 版和 Windows 版没有签名，第一次运行前的放行方法见 [README](https://github.com/Egdon/musicfox-deck#安装)。macOS 的 \`*_macapp.zip\` 是 .app 版，系统通知会带上应用图标。
- 各文件的 SHA-256 在 \`SHA256SUMS\` 里。
EOF
