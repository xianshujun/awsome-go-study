#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

if [[ "${1:-}" == '--help' ]]; then
    printf '%s\n' 'Usage: build-linux.sh' \
        'Requires Go 1.25+, Node.js 20.19+ or 22.12+, and npm.' \
        'Builds frontend assets, then static Linux binaries for amd64 and arm64:' \
        'release/GoStudy-linux-amd64, release/GoStudy-linux-arm64.' \
        'The binary starts a local engine on 127.0.0.1 and opens the system browser' \
        '(xdg-open). Run with -no-browser to just print the URL.'
    exit 0
fi
[[ $# -eq 0 ]] || fail 'Use --help for usage.'
require_go
[[ -d "$APP_DIR/cmd/gostudy" ]] || fail 'cmd/gostudy is missing.'
build_frontend
printf '版本 %s\n' "$(release_version)"
mkdir -p -- "$APP_DIR/release"
# 纯 Go、无 CGO：产物可直接在常见发行版上运行。
for arch in amd64 arm64; do
    (cd -- "$APP_DIR"; GOOS=linux GOARCH="$arch" CGO_ENABLED=0 \
        go build -trimpath -ldflags "-s -w $(version_ldflags)" -o "release/GoStudy-linux-$arch" ./cmd/gostudy)
    chmod 0755 "$APP_DIR/release/GoStudy-linux-$arch"
    printf 'Linux %s: %s\n' "$arch" "$APP_DIR/release/GoStudy-linux-$arch"
done
