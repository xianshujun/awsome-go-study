#!/usr/bin/env bash
# 生成发布清单与课程内容包，供应用内更新使用：
#
#   update.json   程序的版本、发布日期、更新说明与每个产物的大小/sha256
#   content.zip   课程内容包（backend/courses/*.json）
#   content.json  内容包清单：版本（内容哈希）、大小、sha256
#
# 用法（在仓库里执行）：
#   bash app/scripts/make-update-manifest.sh --version 0.7.0 --dir app/release \
#       --notes "本版修好了…" --out-dir app/release
#
# 说明：
#   - 只把目录里实际存在的产物写进清单，缺哪个就少哪个，不会写出指向不存在文件的条目。
#   - 内容包版本取内容本身的哈希，所以同一份内容重复发布会得到同一个版本号，
#     用户不会因为"重新发布了一次一模一样的课程"而被提示更新。
set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

VERSION=''
DIR=''
OUT_DIR=''
CONTENT_ONLY='no'
NOTES=''
NOTES_FILE=''
RELEASED="$(date -u +%F)"
CONTENT_SOURCE="$APP_DIR/backend/courses"

usage() {
    printf '%s\n' \
        'Usage: make-update-manifest.sh --version X.Y.Z --dir 产物目录 --out-dir 输出目录' \
        'Options:' \
        '  --version     X.Y.Z        这次发布的版本号（必填）' \
        '  --dir 目录                 产物所在目录（必填）' \
        '  --out-dir 目录             清单与内容包输出目录，默认与 --dir 相同' \
        '  --notes 文本               更新说明，直接写进清单' \
        '  --notes-file 文件          从文件读更新说明（与 --notes 二选一）' \
        '  --released YYYY-MM-DD      发布日期，默认今天' \
        '  --content-dir 目录         课程目录，默认 app/backend/courses' \
        '  --content-only             只生成课程内容包，不生成 update.json（只改课程内容时用）'
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version) [[ $# -ge 2 ]] || fail '--version 需要一个值'; VERSION="$2"; shift 2 ;;
        --dir) [[ $# -ge 2 ]] || fail '--dir 需要一个值'; DIR="$2"; shift 2 ;;
        --out-dir) [[ $# -ge 2 ]] || fail '--out-dir 需要一个值'; OUT_DIR="$2"; shift 2 ;;
        --notes) [[ $# -ge 2 ]] || fail '--notes 需要一个值'; NOTES="$2"; shift 2 ;;
        --notes-file) [[ $# -ge 2 ]] || fail '--notes-file 需要一个值'; NOTES_FILE="$2"; shift 2 ;;
        --released) [[ $# -ge 2 ]] || fail '--released 需要一个值'; RELEASED="$2"; shift 2 ;;
        --content-dir) [[ $# -ge 2 ]] || fail '--content-dir 需要一个值'; CONTENT_SOURCE="$2"; shift 2 ;;
        --content-only) CONTENT_ONLY='yes'; shift ;;
        -h|--help) usage; exit 0 ;;
        *) usage; fail "未知参数：$1" ;;
    esac
done

if [[ "$CONTENT_ONLY" == 'no' ]]; then
    [[ -n "$VERSION" ]] || fail '缺少 --version'
    [[ -d "$DIR" ]] || fail "找不到产物目录：$DIR"
fi
[[ -z "$OUT_DIR" ]] && OUT_DIR="${DIR:-.}"
mkdir -p -- "$OUT_DIR"
require_command python3

if [[ -n "$NOTES_FILE" ]]; then
    [[ -f "$NOTES_FILE" ]] || fail "找不到说明文件：$NOTES_FILE"
    NOTES="$(cat -- "$NOTES_FILE")"
fi

# 清单里的产物名与更新逻辑（backend/update.go 的 assetName）保持一致。
ASSETS=(GoStudy.exe GoStudy-Setup.exe GoStudy-linux-amd64 GoStudy-linux-arm64 GoStudy-android-release.apk)

if [[ "$CONTENT_ONLY" == 'no' ]]; then
VERSION="$VERSION" RELEASED="$RELEASED" NOTES="$NOTES" DIR="$DIR" OUT_DIR="$OUT_DIR" \
    ASSETS="${ASSETS[*]}" python3 - <<'PY'
import hashlib, json, os, pathlib

version = os.environ["VERSION"]
released = os.environ["RELEASED"]
notes = os.environ["NOTES"]
source = pathlib.Path(os.environ["DIR"])
out_dir = pathlib.Path(os.environ["OUT_DIR"])

def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as handle:
        for chunk in iter(lambda: handle.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()

files = {}
for name in os.environ["ASSETS"].split():
    path = source / name
    if path.is_file():
        files[name] = {"size": path.stat().st_size, "sha256": digest(path)}

if not files:
    raise SystemExit(f"产物目录里没有已知的发布文件：{source}")

manifest = {"version": version, "released": released, "notes": notes, "files": files}
(out_dir / "update.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"update.json：{version}，{len(files)} 个文件")
for name, info in files.items():
    print(f"  {name}  {info['size']} 字节  sha256 {info['sha256'][:12]}…")
PY
fi

# 课程内容包：把 courses/*.json 打成一个 zip，版本号取内容的哈希。
CONTENT_DIR="$CONTENT_SOURCE" OUT_DIR="$OUT_DIR" python3 - <<'PY'
import hashlib, json, os, pathlib, zipfile

content_dir = pathlib.Path(os.environ["CONTENT_DIR"])
out_dir = pathlib.Path(os.environ["OUT_DIR"])
names = sorted(path for path in content_dir.glob("*.json"))
if not names:
    raise SystemExit(f"课程目录里没有 JSON：{content_dir}")

signature = hashlib.sha256()
for path in names:
    signature.update(path.name.encode("utf-8") + b"\0")
    signature.update(path.read_bytes() + b"\0")
version = signature.hexdigest()[:10]

pack = out_dir / "content.zip"
# 固定时间戳：同样的课程内容在 CI 与本机打出的包逐字节一致，便于核对与缓存。
FIXED_TIME = (2026, 1, 1, 0, 0, 0)
with zipfile.ZipFile(pack, "w", zipfile.ZIP_DEFLATED) as archive:
    for path in names:
        # 只放文件名，压缩包里不带层级，与服务端的结构校验一致。
        entry = zipfile.ZipInfo(path.name, date_time=FIXED_TIME)
        entry.compress_type = zipfile.ZIP_DEFLATED
        entry.external_attr = 0o600 << 16
        archive.writestr(entry, path.read_bytes())

data = pack.read_bytes()
info = {
    "version": version,
    "file": "content.zip",
    "count": len(names),
    "pack": {"size": len(data), "sha256": hashlib.sha256(data).hexdigest()},
    "notes": f"{len(names)} 门课程的内容包",
}
(out_dir / "content.json").write_text(json.dumps(info, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"content.json：版本 {version}，{len(names)} 个课程文件，{len(data)} 字节")
PY