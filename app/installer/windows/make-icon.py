#!/usr/bin/env python3
"""把品牌 SVG 渲染成 Windows 图标 app/installer/windows/gostudy.ico。

安装包（GoStudy.iss 的 SetupIconFile）和 GoStudy.exe 的图标资源都由这一份 .ico 生成，
所以品牌图形只维护 app/frontend/public/favicon.svg 一个源头。

用法：
    python3 app/installer/windows/make-icon.py            # 自动寻找 Chrome/Chromium
    python3 app/installer/windows/make-icon.py --chrome /path/to/chrome --keep-png

依赖：Chrome 或 Chromium（无头渲染 SVG）、Pillow（打包 ICO）。

渲染方式是「一次大图 + 分档缩放」：Chrome 只截一张 1024px 的母图，再用 LANCZOS 缩到各档。
之所以不按档位各截一次，是因为无头 Chrome 在窗口小于约 100px 时会渲染出残缺图（试过 16~64
全部只画出一小块），而反复起 Chrome 既慢又容易互相抢 profile。
256 用 PNG 压缩（Vista 起支持，也是现在图标文件的通行做法），更小的档位用 DIB 位图，
保证 Windows 资源管理器和 Inno Setup 都能读。
"""

from __future__ import annotations

import argparse
import json
import os
import shutil
import struct
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
APP_DIR = HERE.parent.parent
SVG = APP_DIR / "frontend" / "public" / "favicon.svg"
OUT_ICO = HERE / "gostudy.ico"

# 常用的几档；Windows 任务栏、开始菜单、桌面和资源管理器大图标都会各取所需。
SIZES = (16, 24, 32, 48, 64, 128, 256)
# 母图尺寸：比最大档位大得多，缩放时才有足够细节。
MASTER = 1024
# 这一档及以上用 PNG 存放，其余用 DIB。
PNG_FROM = 256

CHROME_CANDIDATES = (
    os.environ.get("CHROME", ""),
    "google-chrome",
    "google-chrome-stable",
    "chromium",
    "chromium-browser",
    "microsoft-edge",
    # 本机常见位置（Playwright / Puppeteer 下载的浏览器）
    "/opt/ai-vision/browsers/chromium-1187/chrome-linux/chrome",
)


def find_chrome(explicit: str | None) -> str:
    for candidate in (explicit or "", *CHROME_CANDIDATES):
        if not candidate:
            continue
        found = shutil.which(candidate) or (candidate if Path(candidate).is_file() else "")
        if found:
            return found
    # Puppeteer 的缓存目录按版本号命名，扫一层即可。
    for root in (Path.home() / ".cache" / "puppeteer" / "chrome", Path.home() / ".cache" / "ms-playwright"):
        if root.is_dir():
            for path in sorted(root.glob("*/chrome-linux*/chrome")):
                return str(path)
    sys.exit("没有找到 Chrome/Chromium，请用 --chrome 指定路径。")


def render_master(chrome: str, size: int, out: Path, profile: Path, page: Path) -> None:
    """把 SVG 渲染成一张 size×size 的透明底 PNG。"""
    subprocess.run(
        [
            chrome,
            "--headless=new",
            "--no-sandbox",
            "--disable-gpu",
            "--hide-scrollbars",
            "--force-device-scale-factor=1",
            # 透明背景：图标在浅色/深色桌面上都要好看
            "--default-background-color=00000000",
            # 独立配置目录，避免和用户自己的 Chrome 抢 profile
            f"--user-data-dir={profile}",
            # 等到渲染稳定再截图，避免截到半成品
            "--virtual-time-budget=3000",
            f"--window-size={size},{size}",
            f"--screenshot={out}",
            page.as_uri(),
        ],
        check=True,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        timeout=120,
    )
    if not out.is_file() or out.stat().st_size == 0:
        sys.exit(f"渲染失败：{out} 没有内容")


def dib_bytes(image) -> bytes:
    """按 ICO 里的 DIB 布局写位图：BITMAPINFOHEADER 的高字段翻倍 + XOR 位图 + AND 掩码。"""
    from io import BytesIO

    buffer = BytesIO()
    image.save(buffer, "dib", mode="RGBA")
    data = buffer.getvalue()
    # ICONDIRENTRY 里 DIB 高度要写成图像高度的两倍（含 AND 掩码）
    width, height = image.size
    return data[:8] + struct.pack("<I", height * 2) + data[12:] + _and_mask(width, height)


def _and_mask(width: int, height: int) -> bytes:
    """1 位掩码，行按 4 字节对齐；现代 Windows 只看 alpha，全 0 即可。"""
    stride = ((width + 31) // 32) * 4
    return bytes(stride * height)


def pack_ico(frames: list[tuple[int, bytes]], out: Path, *, png_from: int) -> None:
    header = struct.pack("<HHH", 0, 1, len(frames))  # reserved, type=icon, count
    offset = len(header) + 16 * len(frames)
    entries, blobs = [], []
    for size, blob in frames:
        entries.append(
            struct.pack(
                "<BBBBHHII",
                size if size < 256 else 0,
                size if size < 256 else 0,
                0,  # 调色板颜色数：真彩色为 0
                0,  # reserved
                1,  # planes
                32,  # bpp
                len(blob),
                offset,
            )
        )
        blobs.append(blob)
        offset += len(blob)
    out.write_bytes(header + b"".join(entries) + b"".join(blobs))
    del png_from  # 具体格式由调用方决定，这里只负责拼装


def find_winres(explicit: str | None) -> str | None:
    """go-winres 用来把图标写进 GoStudy.exe 的资源段；没装就只产出 .ico。"""
    for candidate in (explicit or "", os.environ.get("GO_WINRES", ""), "go-winres"):
        if candidate:
            found = shutil.which(candidate)
            if found:
                return found
    gobin = subprocess.run(["go", "env", "GOBIN"], capture_output=True, text=True)
    if gobin.returncode == 0 and gobin.stdout.strip():
        candidate = Path(gobin.stdout.strip()) / "go-winres"
        if candidate.is_file():
            return str(candidate)
    gopath = subprocess.run(["go", "env", "GOPATH"], capture_output=True, text=True)
    if gopath.returncode == 0 and gopath.stdout.strip():
        candidate = Path(gopath.stdout.strip()) / "bin" / "go-winres"
        if candidate.is_file():
            return str(candidate)
    return None


def write_syso(winres: str, pngs: dict[int, Path], work: Path) -> Path:
    """生成 app/cmd/gostudy/rsrc_windows_amd64.syso，Go 构建 Windows 版时自动链接进去。"""
    (work / "winres.json").write_text(
        json.dumps(
            {"RT_GROUP_ICON": {"APP": {"0409": [p.name for p in pngs.values()]}}},
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    subprocess.run(
        [winres, "make", "--in", "winres.json", "--out", "rsrc", "--arch", "amd64"],
        cwd=work,
        check=True,
        stdout=subprocess.DEVNULL,
    )
    target = APP_DIR / "cmd" / "gostudy" / "rsrc_windows_amd64.syso"
    shutil.copy(work / "rsrc_windows_amd64.syso", target)
    return target


def main() -> None:
    parser = argparse.ArgumentParser(description="生成 Windows 图标 gostudy.ico 与 GoStudy.exe 图标资源")
    parser.add_argument("--chrome", help="Chrome/Chromium 可执行文件路径")
    parser.add_argument("--winres", help="go-winres 可执行文件路径")
    parser.add_argument("--keep-png", action="store_true", help="保留各尺寸 PNG 到 <out>.png/ 便于核对")
    parser.add_argument("--all-dib", action="store_true", help="全部用 DIB（不推荐，仅排查 Inno Setup 兼容性）")
    args = parser.parse_args()

    if not SVG.is_file():
        sys.exit(f"缺少品牌 SVG：{SVG}")

    from io import BytesIO

    from PIL import Image

    chrome = find_chrome(args.chrome)
    svg = SVG.read_text(encoding="utf-8")
    png_from = 1 << 30 if args.all_dib else PNG_FROM

    with tempfile.TemporaryDirectory() as tmp:
        tmp = Path(tmp)
        if args.keep_png:
            Path(str(OUT_ICO) + ".png").mkdir(parents=True, exist_ok=True)
        page = tmp / "icon.html"
        page.write_text(
            '<!doctype html><meta charset="utf-8"><style>'
            "html,body{margin:0;padding:0;background:transparent;overflow:hidden}"
            f"svg{{display:block;width:{MASTER}px;height:{MASTER}px}}"
            "</style>" + svg,
            encoding="utf-8",
        )
        master = tmp / "master.png"
        render_master(chrome, MASTER, master, tmp / "profile", page)
        source = Image.open(master).convert("RGBA")
        if source.size != (MASTER, MASTER):
            sys.exit(f"母图尺寸不符：{source.size}")
        if source.getextrema()[3][1] == 0:
            sys.exit(f"母图是空白图，检查 Chrome 能否无头打开 {SVG}")

        frames: list[tuple[int, bytes]] = []
        pngs: dict[int, Path] = {}
        for size in SIZES:
            image = source.resize((size, size), Image.LANCZOS)
            png_path = tmp / f"icon-{size}.png"
            image.save(png_path)
            pngs[size] = png_path
            if args.keep_png:
                shutil.copy(png_path, Path(str(OUT_ICO) + ".png") / f"{size}.png")
            # 边框细，缩到 16px 也必须有实心像素，否则说明图形丢了
            if sum(1 for pixel in image.getdata() if pixel[3] > 128) < size:
                sys.exit(f"{size}x{size} 缩放后几乎没有内容，检查品牌 SVG 是否有效")
            if size >= png_from:
                buffer = BytesIO()
                image.save(buffer, "png")
                frames.append((size, buffer.getvalue()))
            else:
                frames.append((size, dib_bytes(image)))
        pack_ico(frames, OUT_ICO, png_from=png_from)

        winres = find_winres(args.winres)
        if winres:
            syso = write_syso(winres, pngs, tmp)
            print(f"已生成 {syso.relative_to(APP_DIR)}（{syso.stat().st_size / 1024:.1f} KB）")
        else:
            print(
                "没有找到 go-winres，跳过 GoStudy.exe 图标资源（.ico 已生成）。"
                "补装：go install github.com/tc-hib/go-winres@latest"
            )

    print(f"已生成 {OUT_ICO.relative_to(APP_DIR)}（{OUT_ICO.stat().st_size / 1024:.1f} KB，"
          f"{len(SIZES)} 档尺寸：{', '.join(str(s) for s in SIZES)}）")


if __name__ == "__main__":
    main()