#!/usr/bin/env bash
# Build then run the ccss TUI. Always builds first.
set -euo pipefail

cd "$(dirname "$0")"

echo "Building ccss…"
go build -o ccss .

echo "Starting…"
exec ./ccss "$@"
