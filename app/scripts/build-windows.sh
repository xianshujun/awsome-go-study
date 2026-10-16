#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

if [[ "${1:-}" == '--help' ]]; then
    printf '%s\n' 'Usage: build-windows.sh' \
        'Requires Go 1.25+, Node.js 20.19+ or 22.12+, and npm.' \
        'Builds frontend assets, then release/GoStudy.exe for Windows amd64.' \
        'GoStudy.exe uses the Windows GUI subsystem: double-click opens the app window' \
        'without a console. Startup errors appear in a system dialog.' \
        'release/GoStudy-console.exe is the same app with a console for troubleshooting.'
    exit 0
fi
[[ $# -eq 0 ]] || fail 'Use --help for usage.'
require_go
[[ -d "$APP_DIR/cmd/gostudy" ]] || fail 'cmd/gostudy is missing.'
build_frontend
mkdir -p -- "$APP_DIR/release"
# GUI 版本：-H=windowsgui 去掉控制台窗口，双击后直接显示应用界面；
# 启动失败由 cmd/gostudy 的系统对话框兜底，不会静默消失。
(cd -- "$APP_DIR"; GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w -H=windowsgui $(version_ldflags)" -o release/GoStudy.exe ./cmd/gostudy)
# 控制台版本用于排查问题，保留标准输出。
(cd -- "$APP_DIR"; GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w $(version_ldflags)" -o release/GoStudy-console.exe ./cmd/gostudy)
printf '版本 %s\n' "$(release_version)"
printf 'Windows app: %s\n' "$APP_DIR/release/GoStudy.exe"
printf 'Windows console helper: %s\n' "$APP_DIR/release/GoStudy-console.exe"
