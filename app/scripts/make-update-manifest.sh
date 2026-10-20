#!/usr/bin/env bash
# 生成发布清单与课程内容包，供应用内更新使用：
#
#   update.json   程序的版本、发布日期、更新说明与每个产物的大小/sha256
#   content.zip   课程内容包（backend/courses/*.json）
#   content.json  内容包清单：版本（内容哈希）、大小、sha256
#
# 用法（在仓库里执行）：
#   bash app/scripts/make-update-manifest.sh --version 0.9.0 --dir app/release \
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
LABS_SOURCE="$APP_DIR/backend/labs/labs.json"
MIN_APP_VERSION=''
# 事故演练场景改变了内容包结构（多了 labs.json），必须告诉旧程序"先升级"。
# 版本号写在文件里，改包结构时同步改这一处即可。
MIN_APP_FILE="$APP_DIR/backend/labs/min-app-version"

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
        '  --labs-file 文件           事故演练场景，默认 app/backend/labs/labs.json' \
        '  --min-app-version X.Y.Z    课程内容要求的最低程序版本（可选）' \
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
        --labs-file) [[ $# -ge 2 ]] || fail '--labs-file 需要一个值'; LABS_SOURCE="$2"; shift 2 ;;
        --min-app-version) [[ $# -ge 2 ]] || fail '--min-app-version 需要一个值'; MIN_APP_VERSION="$2"; shift 2 ;;
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

# 没显式指定最低版本时，读 labs 旁边的版本文件：带 labs.json 的内容包
# 需要能识别它的程序，否则 v0.9.0 及更早版本会把 labs.json 当成坏课程文件。
if [[ -z "$MIN_APP_VERSION" && -f "$MIN_APP_FILE" ]]; then
    MIN_APP_VERSION="$(tr -d '[:space:]' < "$MIN_APP_FILE")"
fi

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

# 内容包：courses/*.json + labs.json。版本号取所有文件的内容哈希，
# 算法必须与 backend/content.go 的 contentSignature 完全一致。
CONTENT_DIR="$CONTENT_SOURCE" LABS_FILE="$LABS_SOURCE" OUT_DIR="$OUT_DIR" MIN_APP_VERSION="$MIN_APP_VERSION" python3 - <<'PY'
import hashlib, json, os, pathlib, zipfile

content_dir = pathlib.Path(os.environ["CONTENT_DIR"])
labs_file = pathlib.Path(os.environ["LABS_FILE"])
out_dir = pathlib.Path(os.environ["OUT_DIR"])
names = sorted(path for path in content_dir.glob("*.json"))
if not names:
    raise SystemExit(f"课程目录里没有 JSON：{content_dir}")
if not labs_file.is_file():
    raise SystemExit(f"找不到事故演练场景：{labs_file}")

entries = [(path.name, path.read_bytes()) for path in names]
entries.append(("labs.json", labs_file.read_bytes()))
entries.sort(key=lambda item: item[0])

signature = hashlib.sha256()
for name, data in entries:
    signature.update(name.encode("utf-8") + b"\0")
    signature.update(data + b"\0")
version = signature.hexdigest()[:10]

pack = out_dir / "content.zip"
# 固定时间戳：同样的内容在 CI 与本机打出的包逐字节一致，便于核对与缓存。
FIXED_TIME = (2026, 1, 1, 0, 0, 0)
with zipfile.ZipFile(pack, "w", zipfile.ZIP_DEFLATED) as archive:
    for name, data in entries:
        # 只放文件名，压缩包里不带层级，与服务端的结构校验一致。
        entry = zipfile.ZipInfo(name, date_time=FIXED_TIME)
        entry.compress_type = zipfile.ZIP_DEFLATED
        entry.external_attr = 0o600 << 16
        archive.writestr(entry, data)

labs = json.loads(labs_file.read_text(encoding="utf-8"))
scenario_count = len(labs.get("scenarios", []))
data = pack.read_bytes()
info = {
    "version": version,
    "file": "content.zip",
    "count": len(names),
    "scenarios": scenario_count,
    "pack": {"size": len(data), "sha256": hashlib.sha256(data).hexdigest()},
    "notes": f"{len(names)} 门课程、{scenario_count} 个事故场景的内容包",
}
min_app = os.environ.get("MIN_APP_VERSION", "").strip()
if min_app:
    info["minAppVersion"] = min_app
(out_dir / "content.json").write_text(json.dumps(info, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"content.json：版本 {version}，{len(names)} 个课程文件、{scenario_count} 个事故场景，{len(data)} 字节")
PY