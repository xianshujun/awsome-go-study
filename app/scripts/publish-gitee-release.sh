#!/usr/bin/env bash
# 把一次发布的产物镜像到 Gitee：建发行版（Release）+ 上传附件。
#
# 为什么要镜像：GitHub 的产物文件走 objects.githubusercontent.com，国内经常只有几十 KB/s。
# Gitee 的下载入口和 GitHub 同形，所以落地页把同一个文件名换个域名就能用：
#   https://gitee.com/<owner>/<repo>/releases/download/<tag>/<文件名>
# <tag> 也可以写成 latest（自动指向最新发行版），落地页用的就是这个稳定地址。
#
# 用法（需要 Gitee 私人令牌，作用域勾 projects）：
#   GITEE_TOKEN=xxx bash app/scripts/publish-gitee-release.sh \
#       --repo xianshujun/gostudy-release --tag v0.5.0 --dir app/release
#
# 幂等：重复跑同一个 tag 会复用已有发行版，同名附件先删后传，所以失败可以直接重试。
set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

API='https://gitee.com/api/v5'
REPO=''
TAG=''
NAME=''
BODY=''
BRANCH='master'
FILES=()
DIR=''
DRY_RUN='no'

usage() {
    printf '%s\n' \
        'Usage: publish-gitee-release.sh --repo owner/repo --tag vX.Y.Z [--dir 目录] [--file 文件]...' \
        'Options:' \
        '  --repo   owner/repo   Gitee 上的镜像仓库（必填）' \
        '  --tag    vX.Y.Z       要发布的 tag（必填）' \
        '  --name   标题         默认用 tag' \
        '  --body   发布说明     默认写明与 GitHub Releases 一致' \
        '  --branch 分支         发行版指向的分支，默认 master' \
        '  --dir    目录         把目录里的文件全部上传' \
        '  --file   文件         单独指定一个文件，可重复' \
        '  --dry-run             只打印将要发的请求，不真的调接口' \
        '环境变量：GITEE_TOKEN（Gitee 私人令牌，勾 projects 权限）'
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) [[ $# -ge 2 ]] || fail '--repo 需要一个值'; REPO="$2"; shift 2 ;;
        --tag) [[ $# -ge 2 ]] || fail '--tag 需要一个值'; TAG="$2"; shift 2 ;;
        --name) [[ $# -ge 2 ]] || fail '--name 需要一个值'; NAME="$2"; shift 2 ;;
        --body) [[ $# -ge 2 ]] || fail '--body 需要一个值'; BODY="$2"; shift 2 ;;
        --branch) [[ $# -ge 2 ]] || fail '--branch 需要一个值'; BRANCH="$2"; shift 2 ;;
        --dir) [[ $# -ge 2 ]] || fail '--dir 需要一个值'; DIR="$2"; shift 2 ;;
        --file) [[ $# -ge 2 ]] || fail '--file 需要一个值'; FILES+=("$2"); shift 2 ;;
        --dry-run) DRY_RUN='yes'; shift ;;
        -h|--help) usage; exit 0 ;;
        *) usage; fail "未知参数：$1" ;;
    esac
done

[[ "$REPO" == */* ]] || fail '--repo 要写成 owner/repo'
[[ -n "$TAG" ]] || fail '缺少 --tag'
[[ -n "$NAME" ]] || NAME="$TAG"
[[ -n "$BODY" ]] || BODY="GoStudy $TAG 的安装包与可执行文件，与 GitHub Releases 上的产物一致。"
if [[ -n "$DIR" ]]; then
    [[ -d "$DIR" ]] || fail "找不到目录：$DIR"
    while IFS= read -r file; do FILES+=("$file"); done < <(find "$DIR" -maxdepth 1 -type f | sort)
fi
[[ ${#FILES[@]} -gt 0 ]] || fail '没有要上传的文件，用 --dir 或 --file 指定'
for file in "${FILES[@]}"; do
    [[ -f "$file" ]] || fail "找不到文件：$file"
done

TOKEN="${GITEE_TOKEN:-}"
if [[ "$DRY_RUN" == 'yes' ]]; then
    TOKEN='dry-run'
else
    [[ -n "$TOKEN" ]] || fail '缺少 GITEE_TOKEN，先在 Gitee 生成私人令牌（勾 projects 权限）'
fi
require_command curl
require_command python3

# 请求统一走这里：令牌只放在请求头里，不拼进 URL，免得被写进日志。
request() {
    local method="$1" path="$2"
    shift 2
    curl -sS --fail-with-body --retry 3 --retry-delay 5 \
        -X "$method" "$API$path" -H "Authorization: token $TOKEN" "$@"
}

# 从 JSON 里取 release 的 id
release_id_of() {
    python3 -c 'import json,sys; d=json.load(sys.stdin); print(d["id"] if isinstance(d, dict) and "id" in d else "")'
}

# 从附件列表里取指定文件名的附件 id（Gitee 不同接口字段名不一致，这里都认）
attach_id_of() {
    python3 -c '
import json, os, sys
name = sys.argv[1]
data = json.load(sys.stdin) or []
def title(item):
    for key in ("name", "title"):
        value = item.get(key)
        if isinstance(value, str) and value:
            return value
    url = item.get("download_url") or item.get("browser_download_url") or ""
    return os.path.basename(url)
print(next((str(item.get("id")) for item in data if title(item) == name), ""))
' "$1"
}

printf 'Gitee 镜像：%s @ %s（%d 个文件）\n' "$REPO" "$TAG" "${#FILES[@]}"

# 1. 找到或创建发行版
release_id=''
if [[ "$DRY_RUN" == 'yes' ]]; then
    printf '[dry-run] GET  %s/repos/%s/releases/tags/%s\n' "$API" "$REPO" "$TAG"
    printf '[dry-run] POST %s/repos/%s/releases  （不存在时创建：tag_name=%s branch=%s）\n' "$API" "$REPO" "$TAG" "$BRANCH"
    release_id='<dry-run>'
else
    existing=''
    if existing="$(request GET "/repos/$REPO/releases/tags/$TAG" 2>/dev/null)"; then
        release_id="$(printf '%s' "$existing" | release_id_of)"
    fi
    if [[ -n "$release_id" ]]; then
        printf '发行版 %s 已存在（id=%s），复用它并覆盖同名附件。\n' "$TAG" "$release_id"
    else
        printf '创建发行版 %s …\n' "$TAG"
        created="$(request POST "/repos/$REPO/releases" \
            --data-urlencode "tag_name=$TAG" \
            --data-urlencode "name=$NAME" \
            --data-urlencode "body=$BODY" \
            --data-urlencode "target_commitish=$BRANCH")"
        release_id="$(printf '%s' "$created" | release_id_of)"
        [[ -n "$release_id" ]] || fail "创建发行版失败：$created"
        printf '发行版 id=%s\n' "$release_id"
    fi
fi

# 2. 逐个上传：同名附件先删掉，保证脚本可以重跑
uploaded=0
for file in "${FILES[@]}"; do
    name="$(basename -- "$file")"
    size="$(du -h -- "$file" | cut -f1)"
    if [[ "$DRY_RUN" == 'yes' ]]; then
        printf '[dry-run] POST %s/repos/%s/releases/%s/attach_files  -F file=@%s（%s）\n' \
            "$API" "$REPO" "$release_id" "$file" "$size"
        continue
    fi
    list="$(request GET "/repos/$REPO/releases/$release_id/attach_files?per_page=100" 2>/dev/null || true)"
    old_id="$(printf '%s' "$list" | attach_id_of "$name" 2>/dev/null || true)"
    if [[ -n "$old_id" ]]; then
        printf '删除同名旧附件 %s（id=%s）\n' "$name" "$old_id"
        request DELETE "/repos/$REPO/releases/$release_id/attach_files/$old_id" >/dev/null
    fi
    printf '上传 %s（%s）…\n' "$name" "$size"
    request POST "/repos/$REPO/releases/$release_id/attach_files" -F "file=@$file" >/dev/null
    uploaded=$((uploaded + 1))
done

if [[ "$DRY_RUN" == 'yes' ]]; then
    printf 'dry-run 结束：没有调用任何写接口。\n'
    exit 0
fi

printf '\n镜像完成（%d 个文件），下载地址：\n' "$uploaded"
for file in "${FILES[@]}"; do
    printf '  https://gitee.com/%s/releases/download/%s/%s\n' "$REPO" "$TAG" "$(basename -- "$file")"
done
printf '把地址里的 %s 换成 latest 同样有效——落地页用的就是 latest。\n' "$TAG"
