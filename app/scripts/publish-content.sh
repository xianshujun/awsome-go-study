#!/usr/bin/env bash
# 只发布课程内容：重新打包 app/backend/courses/*.json 与 app/backend/labs/labs.json，
# 覆盖掉当前最新一次发布里的 content.json 与 content.zip。程序不用重新发版，
# 用户下次检查更新（或第二天自动检查）就能拿到新内容；内容包的版本取内容哈希，
# 内容没变就不会打扰任何人。事故演练场景也走这条通道：新增一场事故不必发版。
#
# 需要：gh 已登录（上传到 GitHub Release）；GitHub 上的 GITEE_TOKEN 之外，
# 本机还要有 Gitee 令牌（默认读环境变量 GITEE_TOKEN，镜像仓库可传 --repo）。
#
# 用法：
#   GITEE_TOKEN=$(cat /root/.gitee-token) bash app/scripts/publish-content.sh
#   GITEE_TOKEN=… bash app/scripts/publish-content.sh --tag v0.7.1   # 指定发布到哪次发布上
#   GITEE_TOKEN=… bash app/scripts/publish-content.sh --skip-gitee   # 只发 GitHub
set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

REPO='pang-xianshu/gostudy-release'
TAG=''
SKIP_GITEE='no'
CONTENT_DIR="$APP_DIR/backend/courses"
LABS_FILE="$APP_DIR/backend/labs/labs.json"
MIN_APP_VERSION=''

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) [[ $# -ge 2 ]] || fail '--repo 需要一个值'; REPO="$2"; shift 2 ;;
        --tag) [[ $# -ge 2 ]] || fail '--tag 需要一个值'; TAG="$2"; shift 2 ;;
        --content-dir) [[ $# -ge 2 ]] || fail '--content-dir 需要一个值'; CONTENT_DIR="$2"; shift 2 ;;
        --labs-file) [[ $# -ge 2 ]] || fail '--labs-file 需要一个值'; LABS_FILE="$2"; shift 2 ;;
        --min-app-version) [[ $# -ge 2 ]] || fail '--min-app-version 需要一个值'; MIN_APP_VERSION="$2"; shift 2 ;;
        --skip-gitee) SKIP_GITEE='yes'; shift ;;
        -h|--help)
            printf '%s\n' 'Usage: publish-content.sh [--tag vX.Y.Z] [--repo owner/repo] [--content-dir 目录] [--labs-file 文件] [--min-app-version X.Y.Z] [--skip-gitee]'
            exit 0 ;;
        *) fail "未知参数：$1" ;;
    esac
done

require_command gh
require_command python3

if [[ -z "$TAG" ]]; then
    TAG="$(gh release view --json tagName --jq .tagName)"
fi
[[ -n "$TAG" ]] || fail '拿不到当前最新发布的 tag，用 --tag 指定一个'

STAGING="$(mktemp -d)"
trap 'rm -rf -- "$STAGING"' EXIT

EXTRA_ARGS=()
if [[ -n "$MIN_APP_VERSION" ]]; then
    EXTRA_ARGS+=(--min-app-version "$MIN_APP_VERSION")
fi

printf '打包课程与事故场景内容并发布到 %s：\n' "$TAG"
bash "$(dirname -- "${BASH_SOURCE[0]}")/make-update-manifest.sh" \
    --content-only --content-dir "$CONTENT_DIR" --labs-file "$LABS_FILE" \
    "${EXTRA_ARGS[@]}" --out-dir "$STAGING"

printf '上传到 GitHub Release %s …\n' "$TAG"
gh release upload "$TAG" "$STAGING/content.json" "$STAGING/content.zip" --clobber

if [[ "$SKIP_GITEE" == 'yes' ]]; then
    printf '按参数跳过 Gitee 镜像。\n'
    exit 0
fi
TOKEN="${GITEE_TOKEN:-}"
[[ -n "$TOKEN" ]] || fail '缺少 GITEE_TOKEN：Gitee 私人令牌（也可加 --skip-gitee 只发 GitHub）'
GITEE_TOKEN="$TOKEN" bash "$(dirname -- "${BASH_SOURCE[0]}")/publish-gitee-release.sh" \
    --repo "$REPO" --tag "$TAG" --file "$STAGING/content.json" --file "$STAGING/content.zip"
printf '课程与事故场景内容已发布。用户下次检查更新即可拿到（也可以等每天一次自动检查）。\n'