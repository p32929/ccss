#!/usr/bin/env bash
# Build ccss, install it globally, then run it.
set -euo pipefail

# ccss picks which project to open from the directory it starts in, so build
# from the repo but come back here before launching -- otherwise you would
# always land on this repo's own sessions.
start_dir="$PWD"
repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing ccss…"
cd "$repo_dir"
go install .

bin_dir="$(go env GOBIN)"
[ -n "$bin_dir" ] || bin_dir="$(go env GOPATH)/bin"
ccss_bin="$bin_dir/ccss"

if [ ! -x "$ccss_bin" ]; then
  echo "error: go install did not produce $ccss_bin" >&2
  exit 1
fi

echo "Installed: $ccss_bin"

if ! command -v ccss >/dev/null 2>&1; then
  cat <<MSG

Note: $bin_dir is not on your PATH, so "ccss" alone won't work yet.
To fix that, run:

  echo 'export PATH="\$PATH:$bin_dir"' >> ~/.zshrc && source ~/.zshrc

MSG
fi

echo "Starting…"
cd "$start_dir"
exec "$ccss_bin" "$@"
