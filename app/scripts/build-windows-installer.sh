#!/usr/bin/env bash
# 用 Inno Setup 把 app/release/GoStudy.exe 打成 Windows 安装包 GoStudy-Setup.exe。
#
# 只在 Windows 上跑（本地 Git Bash、CI 的 windows-latest 都行）：ISCC.exe 是 Windows 程序。
# 安装包脚本见 ../installer/windows/GoStudy.iss，图标由 installer/windows/make-icon.py 生成。
#
# 用法：
#   bash app/scripts/build-windows-installer.sh                 # 版本取 git tag，取不到就用 0.0.0
#   bash app/scripts/build-windows-installer.sh --version 0.5.0 # 指定版本
#   bash app/scripts/build-windows-installer.sh --app <exe>     # 指定要打包的可执行文件
#   ISCC="/c/Program Files (x86)/Inno Setup 6/ISCC.exe" bash app/scripts/build-windows-installer.sh
set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

INSTALLER_DIR="$APP_DIR/installer/windows"
EXE="$APP_DIR/release/GoStudy.exe"
VERSION=''

usage() {
    printf '%s\n' \
        'Usage: build-windows-installer.sh [--version X.Y.Z] [--app path/to/GoStudy.exe]' \
        'Requires: Inno Setup 6.3+（提供 ISCC.exe，未安装时可用 choco install innosetup），' \
        '以及已经构建好的 release/GoStudy.exe（先跑 build-windows.sh）。' \
        'Produces: release/GoStudy-Setup.exe'
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version) [[ $# -ge 2 ]] || fail '--version 需要一个值'; VERSION="$2"; shift 2 ;;
        --app) [[ $# -ge 2 ]] || fail '--app 需要一个路径'; EXE="$2"; shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *) usage; fail "未知参数：$1" ;;
    esac
done

# 版本号：优先命令行，其次当前 tag（v0.5.0 → 0.5.0），最后退回 0.0.0。
# Inno Setup 的 VersionInfoVersion 只接受数字点分，所以这里顺手做个校验。
version_from_git() {
    git -C "$APP_DIR" describe --tags --exact-match 2>/dev/null \
        || git -C "$APP_DIR" describe --tags --abbrev=0 2>/dev/null \
        || true
}
if [[ -z "$VERSION" ]]; then
    tag="$(version_from_git)"
    VERSION="${tag#v}"
fi
[[ "$VERSION" =~ ^[0-9]+(\.[0-9]+)*$ ]] || fail "版本号必须是数字点分形式，当前是：$VERSION"

[[ -f "$EXE" ]] || fail "找不到可执行文件 $EXE，先运行 bash app/scripts/build-windows.sh"

# 找 ISCC.exe：环境变量优先，其次是 Inno Setup 的默认安装位置。
find_iscc() {
    local candidate
    for candidate in "${ISCC:-}" \
        "/c/Program Files (x86)/Inno Setup 6/ISCC.exe" \
        "/c/Program Files/Inno Setup 6/ISCC.exe" \
        "$(command -v iscc 2>/dev/null || true)"; do
        if [[ -n "$candidate" && -f "$candidate" ]]; then
            printf '%s' "$candidate"
            return 0
        fi
    done
    return 1
}
ISCC_BIN="$(find_iscc)" || fail '没有找到 ISCC.exe，请安装 Inno Setup 6.3+ 或设置 ISCC 环境变量'

# Git Bash 会把 /Dxxx 这种参数当成路径改写，MSYS2_ARG_CONV_EXCL 关掉这个转换。
export MSYS2_ARG_CONV_EXCL='*'

printf '用 %s 编译安装包，版本 %s\n' "$ISCC_BIN" "$VERSION"
"$ISCC_BIN" \
    "/DAppVersion=$VERSION" \
    "/DSourceExe=$(cygpath -w "$EXE" 2>/dev/null || printf '%s' "$EXE")" \
    "$(cygpath -w "$INSTALLER_DIR/GoStudy.iss" 2>/dev/null || printf '%s' "$INSTALLER_DIR/GoStudy.iss")"

OUT="$APP_DIR/release/GoStudy-Setup.exe"
[[ -s "$OUT" ]] || fail "ISCC 没有产出 $OUT"
printf 'Windows 安装包：%s（%s）\n' "$OUT" "$(du -h -- "$OUT" | cut -f1)"
