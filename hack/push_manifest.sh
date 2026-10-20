#!/usr/bin/env bash
# Commits FILE as PATH to the GitHub repository REPO over SSH, with the
# private deploy key in $DEPLOY_KEY (.github/workflows/package-managers.yaml):
#
#   hack/push_manifest.sh Egdon/homebrew-tap Formula/tapefox.rb FILE "tapefox 1.0.0"

set -euo pipefail

repo=$1
path=$2
file=$3
message=$4

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
printf '%s\n' "$DEPLOY_KEY" >"$work/key"
chmod 600 "$work/key"
# github.com's published host key (api.github.com/meta)
echo "github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl" >"$work/known_hosts"
export GIT_SSH_COMMAND="ssh -i $work/key -o IdentitiesOnly=yes -o UserKnownHostsFile=$work/known_hosts"

git clone -q "git@github.com:$repo.git" "$work/repo"
mkdir -p "$work/repo/$(dirname "$path")"
cp "$file" "$work/repo/$path"
cd "$work/repo"
git add "$path"
if git diff --cached --quiet; then
	echo "$repo already has this $path"
	exit 0
fi
git -c user.name="github-actions[bot]" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
	commit -q -m "$message"
git push -q
echo "pushed $path to $repo"
