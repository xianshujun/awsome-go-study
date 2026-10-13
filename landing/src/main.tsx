import { useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  AnimatePresence,
  MotionConfig,
  motion,
  useReducedMotion,
} from "motion/react";
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
    caption: "左边讲义和练习，右边编辑器和运行结果。跑通之后每行输出带一个对勾。",
  },
  {
    id: "home",
    file: "screenshots/01-home.png",
    tab: "首页",
    caption: "首页会循环演示一段课程里的代码：怎么写、跑出什么结果、有没有通过。",
  },
  {
    id: "assistant",
    file: "screenshots/04-assistant.png",
    tab: "AI 导师",
    caption: "接自己的模型，问不懂的地方。只有你点发送时才会把这一课和代码发出去。",
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

const downloads = [
  {
    os: "Windows",
    file: "GoStudy.exe",
    note: "双击就能用。界面用 Edge 或 Chrome 的应用窗口打开。",
    hint: "约 12 MB，免安装",
  },
  {
    os: "Android",
    file: "GoStudy-android-release.apk",
    note: "Android 8.0 以上。自签名，装的时候需要允许未知来源。",
    hint: "约 65 MB，含四种 CPU 架构",
  },
  {
    os: "Linux",
    file: "GoStudy-linux-amd64 / arm64",
    note: "静态单文件，chmod +x 之后直接运行。",
    hint: "约 12 MB，无动态依赖",
  },
];

const stack = [
  { term: "引擎", detail: "Go 1.22+ 起一个只监听回环地址的本机服务，练习用 yaegi 受限解释执行" },
  { term: "界面", detail: "React 19 + TypeScript，编辑器是 CodeMirror 6，交互和动效用 Radix UI 与 Motion" },
  { term: "打包", detail: "前端资源在编译期嵌进可执行文件，产物是单个二进制；安卓通过 gomobile 打成 AAR 再套一层 WebView" },
  { term: "课程", detail: "内容来自仓库里的 8 篇 Go 笔记，共 43 节，每节都标注了对应的原文章节" },
];

/* ---------------- 动效：先写起点、再动到终点，关闭动效时元素直接是最终状态 ---------------- */

// 缓动写成四元组，避免被推断成 number[] 而无法赋给 Motion 的 Easing
const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

function useMotionOn() {
  return !useReducedMotion();
}

/** 入场：淡入并轻微上浮。关闭动效时返回空对象，元素首帧即最终状态。 */
function rise(on: boolean, index = 0): MotionProps {
  if (!on) return {};
  return {
    initial: { opacity: 0, y: 16 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.42, ease: EASE, delay: Math.min(index, 6) * 0.05 },
  };
}

/** 滚动进入时浮起，只在支持的内核上生效，且只播一次。 */
function reveal(on: boolean, index = 0): MotionProps {
  if (!on) return {};
  return {
    initial: { opacity: 0, y: 18 },
    whileInView: { opacity: 1, y: 0 },
    viewport: { once: true, amount: 0.25 },
    transition: { duration: 0.45, ease: EASE, delay: Math.min(index, 5) * 0.06 },
  };
}

/* ---------------- 首屏演示台 ---------------- */

/** 代码逐行写出，然后终端出结果、打上通过 —— 数据取自课程里真实的第一课。 */
const heroCode = [
  <><span className="k">package</span> main</>,
  null,
  <><span className="k">import</span> <span className="s">"fmt"</span></>,
  null,
  <><span className="k">func</span> <span className="f">main</span>() {"{"}</>,
  <>  fmt.<span className="f">Println</span>(<span className="s">"Hello, Go!"</span>)</>,
  <>{"}"}</>,
];

function HeroStage() {
  const on = useMotionOn();
  return (
    <motion.div className="stage" {...rise(on, 2)} aria-hidden="true">
      <div className="stage-bar">
        <span className="stage-file">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="m18 16 4-4-4-4M6 8l-4 4 4 4m8.5-12-5 16" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          main.go
        </span>
        <span className="stage-lesson">go-basic · 第一行 Go 代码</span>
      </div>
      <pre className="stage-code"><code>
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
      </code></pre>
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
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="m4 17 6-6-6-6M12 19h8" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          运行结果
          <motion.span
            className="stage-chip"
            initial={on ? { opacity: 0, scale: 0.8 } : false}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ type: "spring", stiffness: 520, damping: 26, delay: 1.6 }}
          >
            ✓ 挑战通过
          </motion.span>
        </div>
        <p className="stage-out"><span className="prompt">$ </span>Hello, Go!</p>
      </motion.div>
    </motion.div>
  );
}

/* ---------------- 截图区：标签切换 + 点开放大 ---------------- */

function Gallery({ on }: { on: boolean }) {
  const [active, setActive] = useState(screenshots[0].id);
  const [zoom, setZoom] = useState<Screenshot>();
  const current = screenshots.find((item) => item.id === active) ?? screenshots[0];
  return (
    <>
      <Tabs.Root value={active} onValueChange={setActive}>
        <Tabs.List className="shots-tabs" aria-label="界面截图">
          {screenshots.map((item) => (
            <Tabs.Trigger
              key={item.id}
              value={item.id}
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
      <div className="shots-frame">
        <AnimatePresence mode="wait" initial={false}>
          <motion.button
            key={current.id}
            className="shot"
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

      <Dialog.Root open={Boolean(zoom)} onOpenChange={(open) => !open && setZoom(undefined)}>
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
              <Dialog.Content asChild forceMount aria-describedby={undefined}>
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
                    <button className="zoom-close" aria-label="关闭">关闭</button>
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

/* ---------------- 页面 ---------------- */

function Page() {
  const on = useMotionOn();
  const ready = useRef(false);
  useEffect(() => {
    ready.current = true;
  }, []);
  return (
    <MotionConfig reducedMotion={on ? "never" : "always"}>
      <Tooltip.Provider delayDuration={300}>
      <header className="topbar">
        <div className="topbar-inner">
          <a className="brand" href="#top">
            <svg className="mark" viewBox="0 0 32 32" fill="none" aria-hidden="true">
              <path d="M16 2.6 27.4 9.1v13L16 28.6 4.6 22.1v-13z" stroke="currentColor" strokeWidth="2.4" strokeLinejoin="round" />
              <path d="M20.7 12.9a5.3 5.3 0 1 0 .9 4.5h-4.9" stroke="currentColor" strokeWidth="2.3" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <span>GoStudy</span>
          </a>
          <nav className="nav" aria-label="页面导航">
            <a href="#features">功能</a>
            <a href="#shots">截图</a>
            <a href="#faq">常见问题</a>
            <a href="#get">下载</a>
          </nav>
          <a className="button primary" href="https://github.com/xianshujun/awsome-go-study/releases/latest">
            下载
          </a>
        </div>
      </header>

      <main id="top">
        <section className="hero">
          <motion.div className="hero-text">
            <motion.span className="pill" {...rise(on, 0)}>
              <span className="pill-dot" />
              yaegi · 本地解释执行
            </motion.span>
            <motion.h1 {...rise(on, 1)}>
              改代码，跑一下，
              <br />
              看结果。
            </motion.h1>
            <motion.p className="lead" {...rise(on, 2)}>
              43 道 Go 练习，从语法写到并发。装完就能用，不用配 Go 环境；课程、进度和笔记都留在本机，
              断网也能学。卡住了可以接自己的模型问一问。
            </motion.p>
            <motion.div className="actions" {...rise(on, 3)}>
              <a className="button primary large" href="https://github.com/xianshujun/awsome-go-study/releases/latest">
                下载 GoStudy
              </a>
              <a className="button large" href="https://github.com/xianshujun/awsome-go-study">看源码</a>
            </motion.div>
            <motion.p className="footnote" {...rise(on, 4)}>
              Windows <span>·</span> Android 8.0+ <span>·</span> Linux <span>·</span> 不用装 Go
            </motion.p>
          </motion.div>
          <HeroStage />
        </section>

        <section id="features" className="section">
          <motion.p className="eyebrow" {...reveal(on)}>功能</motion.p>
          <motion.h2 {...reveal(on, 1)}>它主要干这几件事</motion.h2>
          <div className="cards">
            {features.map((item, i) => (
              <motion.article className="card" key={item.title} {...reveal(on, i)}>
                <h3>{item.title}</h3>
                <p>{item.body}</p>
              </motion.article>
            ))}
          </div>
        </section>

        <section id="shots" className="section">
          <motion.p className="eyebrow" {...reveal(on)}>截图</motion.p>
          <motion.h2 {...reveal(on, 1)}>界面长这样</motion.h2>
          <Gallery on={on} />
        </section>

        <section id="faq" className="section">
          <motion.p className="eyebrow" {...reveal(on)}>常见问题</motion.p>
          <motion.h2 {...reveal(on, 1)}>装之前可能会想问的</motion.h2>
          <Accordion.Root type="single" collapsible className="faq">
            {faq.map((item) => (
              <Accordion.Item key={item.q} value={item.q} className="faq-item">
                <Accordion.Header>
                  <Accordion.Trigger asChild>
                    <button className="faq-q">
                      {item.q}
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                        <path d="m6 9 6 6 6-6" strokeLinecap="round" strokeLinejoin="round" />
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
          <motion.p className="eyebrow" {...reveal(on)}>下载</motion.p>
          <motion.h2 {...reveal(on, 1)}>三个平台，选一个</motion.h2>
          <div className="downloads">
            {downloads.map((item, i) => (
              <motion.div key={item.os} {...reveal(on, i)}>
                <Tooltip.Root>
                  <Tooltip.Trigger asChild>
                    <a className="download" href="https://github.com/xianshujun/awsome-go-study/releases/latest">
                      <span className="download-os">{item.os}</span>
                      <strong>{item.file}</strong>
                      <span className="download-note">{item.note}</span>
                    </a>
                  </Tooltip.Trigger>
                  <Tooltip.Portal>
                    <Tooltip.Content className="tooltip" sideOffset={8}>
                      {item.hint}
                      <Tooltip.Arrow className="tooltip-arrow" />
                    </Tooltip.Content>
                  </Tooltip.Portal>
                </Tooltip.Root>
              </motion.div>
            ))}
          </div>
          <motion.div className="fromsource" {...reveal(on, 3)}>
            <h3>或者从源码跑</h3>
            <pre><code>{`git clone https://github.com/xianshujun/awsome-go-study
cd awsome-go-study/app
go run ./cmd/gostudy -no-browser`}</code></pre>
            <p className="muted">
              需要 Go 1.25+ 和 Node.js 22.12+（第一次构建前端要用）。打包各平台产物的命令写在
              <a href="https://github.com/xianshujun/awsome-go-study/blob/master/app/README.md">应用说明</a>里。
            </p>
          </motion.div>
        </section>

        <section className="section">
          <motion.p className="eyebrow" {...reveal(on)}>构成</motion.p>
          <motion.h2 {...reveal(on, 1)}>怎么做的</motion.h2>
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
            <svg className="mark" viewBox="0 0 32 32" fill="none" aria-hidden="true">
              <path d="M16 2.6 27.4 9.1v13L16 28.6 4.6 22.1v-13z" stroke="currentColor" strokeWidth="2.4" strokeLinejoin="round" />
              <path d="M20.7 12.9a5.3 5.3 0 1 0 .9 4.5h-4.9" stroke="currentColor" strokeWidth="2.3" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            GoStudy
          </span>
          <nav aria-label="页脚导航">
            <a href="https://github.com/xianshujun/awsome-go-study">GitHub 仓库</a>
            <a href="https://github.com/xianshujun/awsome-go-study/releases">全部版本</a>
            <a href="https://xianshujun.github.io/">xianshujun.github.io</a>
          </nav>
          <span className="footer-note">没有账号，没有上传。</span>
        </div>
      </footer>
      </Tooltip.Provider>
    </MotionConfig>
  );
}

const root = document.getElementById("root");
if (root) createRoot(root).render(<Page />);
