#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"
MOBILE_VERSION='v0.0.0-20250813145510-f12310a0cfd9'
TOOLS_BIN="$SCRIPT_DIR/.tools/bin"

fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }
require_command() { command -v "$1" >/dev/null 2>&1 || fail "Required tool not found: $1"; }
build_frontend() {
    require_command npm
    [[ -f "$APP_DIR/frontend/package-lock.json" ]] || fail 'frontend/package-lock.json is missing.'
    # 清理上一次的产物，但保留被版本库跟踪的 README.md——
    # //go:embed dist 需要该目录至少有一个可嵌入文件，否则全新克隆无法编译。
    if [[ -d "$APP_DIR/web/dist" ]]; then
        find "$APP_DIR/web/dist" -mindepth 1 ! -name README.md -exec rm -rf -- {} +
    fi
    (cd -- "$APP_DIR/frontend"; npm ci --no-audit --no-fund; npm run build)
    [[ -s "$APP_DIR/web/dist/index.html" ]] || fail 'Frontend build must produce web/dist/index.html.'
}
# release_version：优先取当前 tag，其次退回 0.0.0。
# 版本号只在这里推导，然后用 -ldflags 注入到 backend.Version，
# 这样"程序自报的版本"和 tag 永远不会对不上。
release_version() {
    local tag
    tag="$(git -C "$APP_DIR" describe --tags --abbrev=0 2>/dev/null || true)"
    tag="${tag#v}"
    if [[ "$tag" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        printf '%s' "$tag"
    else
        printf '0.0.0'
    fi
}

# version_ldflags：把版本号注入 Go 构建。
version_ldflags() {
    local extra="${1:-}"
    printf -- '-X github.com/xianshujun/awsome-go-study/app/backend.Version=%s%s' "$(release_version)" "$extra"
}

require_go() {
    require_command go
    local version major minor
    version="$(go env GOVERSION)"
    [[ "$version" =~ ^go([0-9]+)\.([0-9]+) ]] || fail "Cannot identify Go version: $version"
    major="${BASH_REMATCH[1]}"; minor="${BASH_REMATCH[2]}"
    (( major > 1 || minor >= 25 )) || fail 'Packaging requires Go 1.25 or newer.'
}
