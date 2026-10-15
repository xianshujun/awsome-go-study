import { useEffect, useRef, useState } from "react";
import gsap from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { Button, Theme } from "@radix-ui/themes";
import "@radix-ui/themes/styles.css";
import { createRoot } from "react-dom/client";
import { AnimatePresence, MotionConfig, motion } from "motion/react";
import type { MotionProps } from "motion/react";
import { Accordion, Dialog, Tabs, Tooltip } from "radix-ui";
import "./styles.css";

type Screenshot = {
  id: string;
  file: string;
  tab: string;
  caption: string;
  tall?: boolean;
};

/** 截图直接取自应用仓库的 app/docs/screenshots，由部署流程复制进站点。 */
const screenshots: Screenshot[] = [
  {
    id: "lesson",
    file: "screenshots/03-lesson.png",
    tab: "课程页",
    caption:
      "左边讲义和练习，右边编辑器和运行结果。跑通之后每行输出带一个对勾。",
  },
  {
    id: "home",
    file: "screenshots/01-home.png",
    tab: "首页",
    caption:
      "首页会循环演示一段课程里的代码：怎么写、跑出什么结果、有没有通过。",
  },
  {
    id: "assistant",
    file: "screenshots/04-assistant.png",
    tab: "AI 导师",
    caption:
      "接自己的模型，问不懂的地方。只有你点发送时才会把这一课和代码发出去。",
  },
  {
    id: "mobile",
    file: "screenshots/06-mobile.png",
    tab: "手机端",
    caption: "手机上界面会竖过来排，代码和输出都在。",
    tall: true,
  },
];

const features = [
  {
    title: "代码是真的在跑",
    body: "练习里的代码交给本机的解释器执行，输出和报错原样返回，不是预先写好的答案比对。整个运行环境不开放文件和网络访问。",
  },
  {
    title: "跑完会对一遍答案",
    body: "每次运行分编译、执行、比对三步。结果跟期望输出逐行对，对上了才打勾；再答对一道理解题，这一课算过。",
  },
  {
    title: "AI 是可选的",
    body: "可以接 OpenAI 兼容的接口，也可以指向自己电脑上跑的模型。不配也完全能用，只是少了提问这个功能。",
  },
  {
    title: "进度在你自己电脑上",
    body: "没有账号，不用登录。进度、代码草稿和笔记都写在本地文件里，换设备不会同步，也不经过任何服务器。",
  },
];

const faq = [
  {
    q: "需要联网吗？",
    a: "不用。课程内容编译在程序里，练习在本机执行。只有用 AI 导师提问时需要联网。",
  },
  {
    q: "用 AI 要花钱吗？",
    a: "看你接的是什么。接 OpenAI 就按它的价格走，接本机模型就不用钱。不配置也完全能用。",
  },
  {
    q: "进度和笔记存在哪？",
    a: "存在你自己的电脑上。Windows 一般放在用户配置目录，安卓放在应用的私有目录里。换设备不会同步。",
  },
  {
    q: "支持哪些系统？",
    a: "Windows、Android 8.0 以上、Linux（amd64 和 arm64）。安卓包是自签名的，安装时需要在系统里允许「未知来源」。",
  },
  {
    q: "Windows 版为什么还要浏览器？",
    a: "界面是用 Edge 或 Chrome 的应用窗口打开的，所以看起来像个独立窗口，没有地址栏。系统里找不到这两个浏览器时会退回默认浏览器。",
  },
  {
    q: "代码具体在哪运行？",
    a: "程序启动时会在本机起一个只监听 127.0.0.1 的服务，练习交给 yaegi 解释器执行，并限制了超时、文件访问和网络访问。AI 的请求也由这个服务发起，浏览器不直接持有你的密钥。",
  },
];

/** 下载区数据。asset 是 Release 里的资源名，落地页按它拼直链；改资源名要同步改这里。 */
type Download = {
  os: string;
  /** 卡片的主动作：点整张卡片下载这个资源 */
  asset: string;
  note: string;
  hint: string;
  /** 同一平台的其他产物，做成卡片下方的次链接 */
  others?: { asset: string; label: string }[];
};

const downloads: Download[] = [
  {
    os: "Windows",
    asset: "GoStudy-Setup.exe",
    note: "安装版。可以自己选装在哪个目录，装完自动建桌面和开始菜单快捷方式，卸载走系统「应用和功能」。",
    hint: "安装包，可自选目录",
    others: [{ asset: "GoStudy.exe", label: "免安装版：单个 exe，双击即用" }],
  },
  {
    os: "Android",
    asset: "GoStudy-android-release.apk",
    note: "Android 8.0 以上。自签名，装的时候需要允许未知来源。",
    hint: "约 65 MB，含四种 CPU 架构",
  },
  {
    os: "Linux",
    asset: "GoStudy-linux-amd64",
    note: "静态单文件，chmod +x 之后直接运行。",
    hint: "约 12 MB，无动态依赖",
    others: [
      { asset: "GoStudy-linux-arm64", label: "arm64 版：树莓派、ARM 服务器" },
    ],
  },
];

const stack = [
  {
    term: "引擎",
    detail:
      "Go 1.22+ 起一个只监听回环地址的本机服务，练习用 yaegi 受限解释执行",
  },
  {
    term: "界面",
    detail:
      "React 19 + TypeScript，编辑器是 CodeMirror 6，交互和动效用 Radix UI 与 Motion",
  },
  {
    term: "打包",
    detail:
      "前端资源在编译期嵌进可执行文件，产物是单个二进制；安卓通过 gomobile 打成 AAR 再套一层 WebView",
  },
  {
    term: "课程",
    detail:
      "内容来自仓库里的 8 篇 Go 笔记，共 43 节，每节都标注了对应的原文章节",
  },
];

/* ---------------- 动效：先写起点、再动到终点，关闭动效时元素直接是最终状态 ---------------- */

// 缓动写成四元组，避免被推断成 number[] 而无法赋给 Motion 的 Easing
const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

function useMotionOn() {
  const [on, setOn] = useState(
    () => !window.matchMedia("(prefers-reduced-motion: reduce)").matches,
  );
  useEffect(() => {
    const preference = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setOn(!preference.matches);
    preference.addEventListener("change", update);
    update();
    return () => preference.removeEventListener("change", update);
  }, []);
  return on;
}

/** 入场：淡入并轻微上浮。关闭动效时返回空对象，元素首帧即最终状态。 */
function rise(on: boolean, index = 0): MotionProps {
  if (!on)
    return {
      initial: false,
      animate: { opacity: 1, y: 0 },
      transition: { duration: 0 },
    };
  return {
    initial: { opacity: 0, y: 16 },
    animate: { opacity: 1, y: 0 },
    transition: {
      duration: 0.42,
      ease: EASE,
      delay: Math.min(index, 6) * 0.05,
    },
  };
}

/** 滚动进入时浮起，只在支持的内核上生效，且只播一次。 */
function reveal(on: boolean, index = 0): MotionProps {
  if (!on)
    return {
      initial: false,
      animate: { opacity: 1, y: 0 },
      transition: { duration: 0 },
    };
  return {
    initial: { opacity: 0, y: 18 },
    whileInView: { opacity: 1, y: 0 },
    viewport: { once: true, amount: 0.25 },
    transition: {
      duration: 0.45,
      ease: EASE,
      delay: Math.min(index, 5) * 0.06,
    },
  };
}

/* ---------------- 首屏演示台 ---------------- */

/** 代码逐行写出，然后终端出结果、打上通过 —— 数据取自课程里真实的第一课。 */
const heroCode = [
  <>
    <span className="k">package</span> main
  </>,
  null,
  <>
    <span className="k">import</span> <span className="s">"fmt"</span>
  </>,
  null,
  <>
    <span className="k">func</span> <span className="f">main</span>() {"{"}
  </>,
  <>
    {" "}
    fmt.<span className="f">Println</span>(
    <span className="s">"Hello, Go!"</span>)
  </>,
  <>{"}"}</>,
];

function HeroStage() {
  const on = useMotionOn();
  return (
    <motion.div className="stage" {...rise(on, 2)} aria-hidden="true">
      <div className="stage-bar">
        <span className="stage-file">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path
              d="m18 16 4-4-4-4M6 8l-4 4 4 4m8.5-12-5 16"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          main.go
        </span>
        <span className="stage-lesson">go-basic · 第一行 Go 代码</span>
      </div>
      <pre className="stage-code">
        <code>
          {heroCode.map((line, i) => (
            <motion.span
              className="stage-line"
              key={i}
              initial={on ? { opacity: 0, x: -6 } : false}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.24, delay: 0.35 + i * 0.09 }}
            >
              {line ?? "\u00a0"}
            </motion.span>
          ))}
        </code>
      </pre>
      <motion.div
        className="stage-flow"
        initial={on ? { opacity: 0 } : false}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.3, delay: 1.05 }}
      >
        <span>本地解释执行</span>
        <i />
      </motion.div>
      <motion.div
        className="stage-term"
        initial={on ? { opacity: 0 } : false}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.34, delay: 1.3 }}
      >
        <div className="stage-term-head">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path
              d="m4 17 6-6-6-6M12 19h8"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          运行结果
          <motion.span
            className="stage-chip"
            initial={on ? { opacity: 0, scale: 0.8 } : false}
            animate={{ opacity: 1, scale: 1 }}
            transition={{
              type: "spring",
              stiffness: 520,
              damping: 26,
              delay: 1.6,
            }}
          >
            ✓ 挑战通过
          </motion.span>
        </div>
        <p className="stage-out">
          <span className="prompt">$ </span>Hello, Go!
        </p>
      </motion.div>
    </motion.div>
  );
}

/* ---------------- 截图区：标签切换 + 点开放大 ---------------- */

function Gallery({ on }: { on: boolean }) {
  const [active, setActive] = useState(screenshots[0].id);
  const [zoom, setZoom] = useState<Screenshot>();
  const zoomTrigger = useRef<HTMLButtonElement | null>(null);
  const current =
    screenshots.find((item) => item.id === active) ?? screenshots[0];
  return (
    <>
      <Tabs.Root value={active} onValueChange={setActive}>
        <Tabs.List className="shots-tabs" aria-label="界面截图">
          {screenshots.map((item) => (
            <Tabs.Trigger
              key={item.id}
              value={item.id}
              id={`shot-tab-${item.id}`}
              aria-controls="shot-panel"
              className={active === item.id ? "selected" : ""}
            >
              {item.tab}
              {active === item.id && (
                <motion.span
                  className="tabs-underline"
                  layoutId="shots-underline"
                  transition={{ type: "spring", stiffness: 520, damping: 34 }}
                />
              )}
            </Tabs.Trigger>
          ))}
        </Tabs.List>
      </Tabs.Root>
      <div
        className="shots-frame"
        id="shot-panel"
        role="tabpanel"
        aria-labelledby={`shot-tab-${active}`}
      >
        <AnimatePresence mode="wait" initial={false}>
          <motion.button
            key={current.id}
            className="shot"
            ref={zoomTrigger}
            onClick={() => setZoom(current)}
            aria-label={`放大查看：${current.tab}`}
            initial={on ? { opacity: 0, y: 8 } : false}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -6 }}
            transition={{ duration: 0.24, ease: [0.22, 1, 0.36, 1] }}
          >
            <img
              src={current.file}
              alt={current.caption}
              width={current.tall ? 780 : 1440}
              height={current.tall ? 1688 : 940}
              className={current.tall ? "tall" : undefined}
            />
          </motion.button>
        </AnimatePresence>
      </div>
      <p className="shots-caption">
        {current.caption}
        <span className="shots-hint">点图片可以放大</span>
      </p>

      <Dialog.Root
        open={Boolean(zoom)}
        onOpenChange={(open) => !open && setZoom(undefined)}
      >
        <AnimatePresence>
          {zoom && (
            <Dialog.Portal forceMount>
              <Dialog.Overlay asChild forceMount>
                <motion.div
                  className="zoom-scrim"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.18 }}
                />
              </Dialog.Overlay>
              <Dialog.Content
                asChild
                forceMount
                aria-describedby={undefined}
                onCloseAutoFocus={(event) => {
                  event.preventDefault();
                  zoomTrigger.current?.focus();
                }}
              >
                <motion.figure
                  className="zoom"
                  initial={on ? { opacity: 0, scale: 0.97 } : false}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.98 }}
                  transition={{ type: "spring", stiffness: 300, damping: 32 }}
                >
                  <Dialog.Title className="zoom-title">{zoom.tab}</Dialog.Title>
                  <img src={zoom.file} alt={zoom.caption} />
                  <figcaption>{zoom.caption}</figcaption>
                  <Dialog.Close asChild>
                    <button className="zoom-close" aria-label="关闭">
                      关闭
                    </button>
                  </Dialog.Close>
                </motion.figure>
              </Dialog.Content>
            </Dialog.Portal>
          )}
        </AnimatePresence>
      </Dialog.Root>
    </>
  );
}

gsap.registerPlugin(ScrollTrigger);

function Arrow({ diagonal = false }: { diagonal?: boolean }) {
  return (
    <svg
      className="arrow"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      aria-hidden="true"
    >
      <path
        d={diagonal ? "M6 18 18 6M6 6h12v12" : "M4 12h15m-6-6 6 6-6 6"}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/** 仓库坐标与 Star 数据。拉不到数字时只显示链接，绝不显示编造的数值。 */
const REPO = "xianshujun/awsome-go-study";
/** 下载直链：latest/download 会跳到最新一次 Release 的对应资源，发新版本不用改页面。 */
const RELEASES_PAGE = `https://github.com/${REPO}/releases`;
const latestAssetUrl = (name: string) =>
  `${RELEASES_PAGE}/latest/download/${name}`;
const STAR_CACHE_KEY = "gostudy-stars";
const STAR_CACHE_TTL = 10 * 60 * 1000;

let starRequest: Promise<number | null> | undefined;

function loadStars(): Promise<number | null> {
  starRequest ??= (async () => {
    // 未认证的 GitHub API 限 60 次/小时/IP：同一会话先读缓存，过期再回源。
    try {
      const cached = sessionStorage.getItem(STAR_CACHE_KEY);
      if (cached) {
        const parsed = JSON.parse(cached) as { count?: number; at?: number };
        if (
          typeof parsed.count === "number" &&
          Date.now() - (parsed.at ?? 0) < STAR_CACHE_TTL
        ) {
          return parsed.count;
        }
      }
    } catch {
      // 隐私模式下 sessionStorage 不可用，直接回源。
    }
    // 网络抖动时重试一次；被限流（非 2xx）不重试，避免继续消耗配额。
    for (let attempt = 0; attempt < 2; attempt += 1) {
      try {
        // 加超时：接口不通时尽快放弃，不留一个挂着的请求。
        const response = await fetch(`https://api.github.com/repos/${REPO}`, {
          signal: AbortSignal.timeout(5000),
        });
        if (!response.ok) return null;
        const data = (await response.json()) as { stargazers_count?: number };
        if (typeof data.stargazers_count !== "number") return null;
        try {
          sessionStorage.setItem(
            STAR_CACHE_KEY,
            JSON.stringify({ count: data.stargazers_count, at: Date.now() }),
          );
        } catch {
          // 存不下只影响下次访问。
        }
        return data.stargazers_count;
      } catch {
        if (attempt === 1) return null;
        await new Promise((resolve) => setTimeout(resolve, 800));
      }
    }
    return null;
  })();
  return starRequest;
}

function formatStars(count: number) {
  return count < 1000 ? String(count) : `${(count / 1000).toFixed(1)}k`;
}

/**
 * 最新 Release 里有哪些产物。
 *
 * 下载按钮的地址本身写成 releases/latest/download/<资源名>，新版本一发就自动指过去，
 * 不需要改这个文件；这里查接口只是为了把版本号、发布时间和文件大小显示出来，
 * 顺便确认某个平台的包在这个版本里确实存在。拿不到数据时按钮照旧能用。
 */
type ReleaseInfo = {
  tag: string;
  /** 发布日期，YYYY-MM-DD */
  date: string;
  /** 资源名 → 字节数 */
  sizes: Record<string, number>;
};

const RELEASE_CACHE_KEY = "gostudy-release";
const RELEASE_CACHE_TTL = 30 * 60 * 1000;

let releaseRequest: Promise<ReleaseInfo | null> | undefined;

function loadRelease(): Promise<ReleaseInfo | null> {
  releaseRequest ??= (async () => {
    // 未认证的 GitHub API 限 60 次/小时/IP：同一会话先读缓存，过期再回源。
    try {
      const cached = sessionStorage.getItem(RELEASE_CACHE_KEY);
      if (cached) {
        const parsed = JSON.parse(cached) as { value?: ReleaseInfo; at?: number };
        if (parsed.value && Date.now() - (parsed.at ?? 0) < RELEASE_CACHE_TTL) {
          return parsed.value;
        }
      }
    } catch {
      // 隐私模式下 sessionStorage 不可用，直接回源。
    }
    for (let attempt = 0; attempt < 2; attempt += 1) {
      try {
        const response = await fetch(
          `https://api.github.com/repos/${REPO}/releases/latest`,
          { signal: AbortSignal.timeout(5000) },
        );
        // 还没有 Release 时接口返回 404，这时保持静态文案即可。
        if (!response.ok) return null;
        const data = (await response.json()) as {
          tag_name?: string;
          published_at?: string;
          assets?: { name?: string; size?: number }[];
        };
        if (!data.tag_name) return null;
        const sizes: Record<string, number> = {};
        for (const asset of data.assets ?? []) {
          if (asset.name && typeof asset.size === "number") {
            sizes[asset.name] = asset.size;
          }
        }
        const value: ReleaseInfo = {
          tag: data.tag_name,
          date: (data.published_at ?? "").slice(0, 10),
          sizes,
        };
        try {
          sessionStorage.setItem(
            RELEASE_CACHE_KEY,
            JSON.stringify({ value, at: Date.now() }),
          );
        } catch {
          // 存不下只影响下次访问。
        }
        return value;
      } catch {
        if (attempt === 1) return null;
        await new Promise((resolve) => setTimeout(resolve, 800));
      }
    }
    return null;
  })();
  return releaseRequest;
}

function formatSize(bytes: number) {
  const megabytes = bytes / 1024 / 1024;
  return megabytes >= 1
    ? `${megabytes.toFixed(1)} MB`
    : `${Math.round(bytes / 1024)} KB`;
}

/** 判断访客自己的系统，用来在对应卡片上标一句「你的系统」。 */
function detectPlatform(agent: string): string | null {
  if (/Android/i.test(agent)) return "Android";
  if (/Windows NT/i.test(agent)) return "Windows";
  if (/Linux/i.test(agent)) return "Linux";
  return null;
}

function StarMark() {
  return (
    <svg className="star-mark" viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M12 2.6l2.86 5.8 6.4.93-4.63 4.51 1.09 6.37L12 17.2l-5.72 3.01 1.09-6.37L2.74 9.33l6.4-.93z"
        fill="currentColor"
      />
    </svg>
  );
}

/** 实时 Star 数：未拿到数据时只显示图标与文字，拿到后补上数字。 */
function StarCount({ className }: { className: string }) {
  const [count, setCount] = useState<number | null>(null);
  useEffect(() => {
    let alive = true;
    void loadStars().then((value) => {
      if (alive) setCount(value);
    });
    return () => {
      alive = false;
    };
  }, []);
  return (
    <a
      className={className}
      href={`https://github.com/${REPO}/stargazers`}
      title="在 GitHub 上给 GoStudy 点个 Star"
      aria-label={
        count === null
          ? "去 GitHub 给 GoStudy 加 Star"
          : `当前 ${count} 个 Star，去 GitHub 加 Star`
      }
    >
      <StarMark />
      <span>Star</span>
      {count !== null && <strong>{formatStars(count)}</strong>}
    </a>
  );
}

function MobileMenu() {
  const [open, setOpen] = useState(false);
  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <Button
          variant="ghost"
          color="gray"
          className="menu-toggle"
          aria-label="打开导航"
        >
          菜单 <span aria-hidden="true">＋</span>
        </Button>
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="menu-scrim" />
        <Dialog.Content className="mobile-menu" aria-describedby={undefined}>
          <Dialog.Title>GoStudy</Dialog.Title>
          <nav aria-label="移动端导航">
            {[
              ["features", "功能"],
              ["shots", "截图"],
              ["faq", "常见问题"],
              ["get", "下载"],
            ].map(([id, title], i) => (
              <a href={`#${id}`} onClick={() => setOpen(false)} key={id}>
                <span>0{i + 1}</span>
                {title}
                <Arrow />
              </a>
            ))}
          </nav>
          <Dialog.Close asChild>
            <Button variant="soft" color="gray">
              关闭导航
            </Button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

/* ---------------- 页面 ---------------- */

function Page() {
  const on = useMotionOn();
  const page = useRef<HTMLDivElement>(null);
  // 下载区展示的版本信息；接口不通就退回卡片上的静态文案，链接照样可用。
  const [release, setRelease] = useState<ReleaseInfo | null>(null);
  const [platform] = useState(() => detectPlatform(navigator.userAgent));
  useEffect(() => {
    let alive = true;
    void loadRelease().then((value) => {
      if (alive) setRelease(value);
    });
    return () => {
      alive = false;
    };
  }, []);
  useEffect(() => {
    const context = gsap.context(() => {
      if (!on) return;
      gsap.fromTo(
        ".hero-rule",
        { scaleX: 0 },
        { scaleX: 1, duration: 1.2, ease: "power3.out", delay: 0.15 },
      );
      gsap.to(".reading-progress", {
        scaleX: 1,
        ease: "none",
        scrollTrigger: {
          trigger: ".site",
          start: "top top",
          end: "bottom bottom",
          scrub: 0.3,
        },
      });
      gsap.to(".stage-surround", {
        y: -24,
        ease: "none",
        scrollTrigger: {
          trigger: ".hero",
          start: "top top",
          end: "bottom top",
          scrub: 0.8,
        },
      });
      gsap.fromTo(
        ".shots-frame",
        { scale: 0.97 },
        {
          scale: 1,
          ease: "none",
          scrollTrigger: {
            trigger: "#shots",
            start: "top 90%",
            end: "top 25%",
            scrub: 0.6,
          },
        },
      );
      document.querySelectorAll(".section").forEach((section) => {
        gsap.fromTo(
          section.querySelector(".eyebrow"),
          { x: -10 },
          {
            x: 0,
            duration: 0.8,
            ease: "power3.out",
            scrollTrigger: { trigger: section, start: "top 85%", once: true },
          },
        );
      });
    }, page);
    return () => context.revert();
  }, [on]);
  return (
    <Theme accentColor="grass" grayColor="sand" radius="small" scaling="100%">
      <div className="site" ref={page}>
        <a className="skip-link" href="#features">
          跳到正文
        </a>
        <MotionConfig reducedMotion="user">
          <Tooltip.Provider delayDuration={300}>
            <header className="topbar">
              <div className="reading-progress" aria-hidden="true" />
              <div className="topbar-inner">
                <a className="brand" href="#top">
                  <svg
                    className="mark"
                    viewBox="0 0 32 32"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M16 2.6 27.4 9.1v13L16 28.6 4.6 22.1v-13z"
                      stroke="currentColor"
                      strokeWidth="2.4"
                      strokeLinejoin="round"
                    />
                    <path
                      d="M20.7 12.9a5.3 5.3 0 1 0 .9 4.5h-4.9"
                      stroke="currentColor"
                      strokeWidth="2.3"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <span>GoStudy</span>
                </a>
                <nav className="nav" aria-label="页面导航">
                  <a href="#features">功能</a>
                  <a href="#shots">截图</a>
                  <a href="#faq">常见问题</a>
                  <a href="#get">下载</a>
                </nav>
                <StarCount className="star-link" />
                <MobileMenu />
                <Button asChild className="button primary">
                  <a href="#get">
                    开始学习 <Arrow />
                  </a>
                </Button>
              </div>
            </header>

            <main id="top">
              <section className="hero">
                <div className="hero-index">
                  <span>GOSTUDY / 本地 Go 练习</span>
                  <span>写下代码，理解代码。</span>
                </div>
                <motion.div className="hero-text">
                  <motion.span className="pill" {...rise(on, 0)}>
                    <span className="pill-dot" />
                    yaegi · 本地解释执行
                  </motion.span>
                  <motion.h1 {...rise(on, 1)}>
                    <span className="title-line">
                      <span>学 Go，</span>
                    </span>
                    <span className="title-line">
                      <span>
                        亲手<span className="title-accent">写一遍。</span>
                      </span>
                    </span>
                  </motion.h1>
                  <motion.p className="lead" {...rise(on, 2)}>
                    43 道 Go 练习，从语法写到并发。装完就能用，不用配 Go
                    环境；课程、进度和笔记都留在本机，
                    断网也能学。卡住了可以接自己的模型问一问。
                  </motion.p>
                  <motion.div className="actions" {...rise(on, 3)}>
                    <Button asChild size="3" className="button primary large">
                      <a href="#get">
                        下载 GoStudy <Arrow />
                      </a>
                    </Button>
                    <Button
                      asChild
                      size="3"
                      variant="ghost"
                      color="gray"
                      className="button text-button large"
                    >
                      <a href="https://github.com/xianshujun/awsome-go-study">
                        查看源码 <Arrow diagonal />
                      </a>
                    </Button>
                  </motion.div>
                  <motion.p className="footnote" {...rise(on, 4)}>
                    Windows <span>·</span> Android 8.0+ <span>·</span> Linux{" "}
                    <span>·</span> 不用装 Go
                  </motion.p>
                </motion.div>
                <div className="stage-surround">
                  <div className="stage-label">
                    <span>练习 01 / HELLO, GO</span>
                    <span>
                      本机运行 <i />
                    </span>
                  </div>
                  <HeroStage />
                  <div className="stage-annotation">
                    <span className="annotation-number">01—43</span>
                    <span>
                      改一行。跑一下。
                      <br />
                      把知识变成自己的。
                    </span>
                    <Arrow />
                  </div>
                </div>
                <div className="hero-rule" aria-hidden="true" />
              </section>
              <motion.div className="facts" {...reveal(on)}>
                <div>
                  <strong>
                    43<span>节</span>
                  </strong>
                  <p>从语法到并发</p>
                </div>
                <div>
                  <strong>
                    3<span>个平台</span>
                  </strong>
                  <p>桌面和手机，随手练习</p>
                </div>
                <div>
                  <strong>
                    100<span>% 本地</span>
                  </strong>
                  <p>课程、代码与学习进度</p>
                </div>
                <a href="#shots">
                  先看一眼界面 <Arrow diagonal />
                </a>
              </motion.div>

              <section id="features" className="section">
                <motion.p className="eyebrow" {...reveal(on)}>
                  01 / 为学习留出空间
                </motion.p>
                <motion.h2 {...reveal(on, 1)}>
                  少一点准备，
                  <br />
                  多一点动手。
                </motion.h2>
                <div className="cards">
                  {features.map((item, i) => (
                    <motion.article
                      className="card"
                      key={item.title}
                      {...reveal(on, i)}
                    >
                      <span className="feature-number">0{i + 1}</span>
                      <div>
                        <h3>{item.title}</h3>
                        <p>{item.body}</p>
                      </div>
                    </motion.article>
                  ))}
                </div>
              </section>

              <section id="shots" className="section">
                <motion.p className="eyebrow" {...reveal(on)}>
                  02 / 学习现场
                </motion.p>
                <motion.h2 {...reveal(on, 1)}>
                  每一次理解，
                  <br />
                  都有迹可循。
                </motion.h2>
                <Gallery on={on} />
              </section>

              <section id="faq" className="section">
                <motion.p className="eyebrow" {...reveal(on)}>
                  03 / 开始之前
                </motion.p>
                <motion.h2 {...reveal(on, 1)}>你可能还想知道。</motion.h2>
                <Accordion.Root type="single" collapsible className="faq">
                  {faq.map((item) => (
                    <Accordion.Item
                      key={item.q}
                      value={item.q}
                      className="faq-item"
                    >
                      <Accordion.Header>
                        <Accordion.Trigger asChild>
                          <button className="faq-q">
                            {item.q}
                            <svg
                              viewBox="0 0 24 24"
                              fill="none"
                              stroke="currentColor"
                              strokeWidth="2"
                              aria-hidden="true"
                            >
                              <path
                                d="m6 9 6 6 6-6"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                              />
                            </svg>
                          </button>
                        </Accordion.Trigger>
                      </Accordion.Header>
                      <Accordion.Content className="faq-a">
                        <p>{item.a}</p>
                      </Accordion.Content>
                    </Accordion.Item>
                  ))}
                </Accordion.Root>
              </section>

              <section id="get" className="section">
                <motion.p className="eyebrow" {...reveal(on)}>
                  04 / 从这里开始
                </motion.p>
                <motion.h2 {...reveal(on, 1)}>
                  下一行 Go，
                  <br />
                  现在就写。
                </motion.h2>
                {release && (
                  <motion.p className="downloads-version" {...reveal(on, 2)}>
                    当前最新版本 {release.tag}
                    {release.date && ` · ${release.date} 发布`}
                  </motion.p>
                )}
                <div className="downloads">
                  {downloads.map((item, i) => {
                    const size = release?.sizes[item.asset];
                    // 只有「确实读到了最新版本、但里面没有这个产物」才改跳 Releases；
                    // 读不到接口时按老办法给直链，latest/download 本来就是猜得中的。
                    const missing = release !== null && size === undefined;
                    return (
                      <motion.div key={item.os} {...reveal(on, i)}>
                        <Tooltip.Root>
                          <Tooltip.Trigger asChild>
                            <a
                              className="download"
                              href={
                                missing
                                  ? RELEASES_PAGE
                                  : latestAssetUrl(item.asset)
                              }
                            >
                              <span className="download-os">
                                {item.os}
                                <span className="download-flags">
                                  {item.os === platform && (
                                    <em className="download-you">你的系统</em>
                                  )}
                                  <Arrow diagonal />
                                </span>
                              </span>
                              <strong>{item.asset}</strong>
                              <span className="download-note">{item.note}</span>
                              <span className="download-meta">
                                {missing
                                  ? "这个版本里还没有打包，先去 Releases 看看"
                                  : size === undefined
                                    ? item.hint
                                    : `${formatSize(size)} · 直接下载`}
                              </span>
                            </a>
                          </Tooltip.Trigger>
                          <Tooltip.Portal>
                            <Tooltip.Content className="tooltip" sideOffset={8}>
                              {missing
                                ? "打开全部版本列表，选一个包含该产物的版本"
                                : item.hint}
                              <Tooltip.Arrow className="tooltip-arrow" />
                            </Tooltip.Content>
                          </Tooltip.Portal>
                        </Tooltip.Root>
                        {item.others && (
                          <div className="download-alts">
                            {item.others.map((other) => (
                              <a
                                className="download-alt"
                                key={other.asset}
                                href={latestAssetUrl(other.asset)}
                              >
                                {other.label}
                              </a>
                            ))}
                          </div>
                        )}
                      </motion.div>
                    );
                  })}
                </div>
                <motion.div className="fromsource" {...reveal(on, 3)}>
                  <h3>或者从源码跑</h3>
                  <pre>
                    <code>{`git clone https://github.com/xianshujun/awsome-go-study
cd awsome-go-study/app
go run ./cmd/gostudy -no-browser`}</code>
                  </pre>
                  <p className="muted">
                    需要 Go 1.25+ 和 Node.js
                    22.12+（第一次构建前端要用）。打包各平台产物的命令写在
                    <a href="https://github.com/xianshujun/awsome-go-study/blob/master/app/README.md">
                      应用说明
                    </a>
                    里。
                  </p>
                </motion.div>
              </section>

              <section className="section">
                <motion.p className="eyebrow" {...reveal(on)}>
                  05 / 开源，保持透明
                </motion.p>
                <motion.h2 {...reveal(on, 1)}>
                  简单的工具，
                  <br />
                  清楚的构成。
                </motion.h2>
                <dl className="stack">
                  {stack.map((row, i) => (
                    <motion.div key={row.term} {...reveal(on, i)}>
                      <dt>{row.term}</dt>
                      <dd>{row.detail}</dd>
                    </motion.div>
                  ))}
                </dl>
              </section>
            </main>

            <footer className="footer">
              <div className="footer-inner">
                <span className="footer-brand">
                  <svg
                    className="mark"
                    viewBox="0 0 32 32"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M16 2.6 27.4 9.1v13L16 28.6 4.6 22.1v-13z"
                      stroke="currentColor"
                      strokeWidth="2.4"
                      strokeLinejoin="round"
                    />
                    <path
                      d="M20.7 12.9a5.3 5.3 0 1 0 .9 4.5h-4.9"
                      stroke="currentColor"
                      strokeWidth="2.3"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  GoStudy
                </span>
                <nav aria-label="页脚导航">
                  <a href="https://github.com/xianshujun/awsome-go-study">
                    GitHub 仓库
                  </a>
                  <StarCount className="star-foot" />
                  <a href="https://github.com/xianshujun/awsome-go-study/releases">
                    全部版本
                  </a>
                  <a href="https://xianshujun.github.io/">
                    xianshujun.github.io
                  </a>
                </nav>
                <span className="footer-note">没有账号，没有上传。</span>
              </div>
            </footer>
          </Tooltip.Provider>
        </MotionConfig>
      </div>
    </Theme>
  );
}

const root = document.getElementById("root");
if (root) createRoot(root).render(<Page />);
