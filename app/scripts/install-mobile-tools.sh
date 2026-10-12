#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

if [[ "${1:-}" == '--help' ]]; then
    printf '%s\n' 'Usage: install-mobile-tools.sh' \
        'Installs pinned gomobile and gobind into scripts/.tools/bin.' \
        "Version: $MOBILE_VERSION; requires Go 1.25+ and network access." \
        'Does not install Android SDK/NDK or modify the application go.mod.'
    exit 0
fi
[[ $# -eq 0 ]] || fail 'Use --help for usage.'
require_go
mkdir -p -- "$TOOLS_BIN"
# gomobile init 内部使用 @latest，故分别安装同一固定版本，避免隐式升级。
GOBIN="$TOOLS_BIN" GOTOOLCHAIN=local go install "golang.org/x/mobile/cmd/gomobile@$MOBILE_VERSION"
GOBIN="$TOOLS_BIN" GOTOOLCHAIN=local go install "golang.org/x/mobile/cmd/gobind@$MOBILE_VERSION"
printf 'Pinned mobile tools installed in %s\n' "$TOOLS_BIN"
