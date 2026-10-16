import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties, PointerEvent as ReactPointerEvent, ReactNode } from "react";
import { createRoot } from "react-dom/client";
import { flushSync } from "react-dom";
import CodeMirror from "@uiw/react-codemirror";
import { go } from "@codemirror/lang-go";
import ReactMarkdown from "react-markdown";
import {
  AnimatePresence,
  MotionConfig,
  motion,
  useReducedMotion,
} from "motion/react";
import {
  Accordion,
  Dialog,
  Progress as Meter,
  RadioGroup,
  Slider,
  Switch,
  Tabs,
  Toast,
  Tooltip,
} from "radix-ui";
import {
  ArrowDownToLine,
  ArrowRight,
  BookOpen,
  Check,
  CheckCheck,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleHelp,
  Clock3,
  Code2,
  Compass,
  FileText,
  Flag,
  History,
  Layers3,
  LoaderCircle,
  MessageCircle,
  MessagesSquare,
  Pencil,
  Play,
  Plus,
  Power,
  RotateCcw,
  Send,
  Settings2,
  ShieldCheck,
  Sprout,
  SquarePen,
  Terminal,
  Trash2,
  WifiOff,
  X,
} from "lucide-react";
import { api, reportWindowClosed } from "./api";
import { editorTheme } from "./editor";
import type {
  Bootstrap,
  CloseAction,
  Course,
  Lesson,
  Progress,
  RunResult,
  Settings,
} from "./api";
import {
  dialogVariants,
  fadeIn,
  lift,
  rise,
  runPhases,
  springControl,
  springMeter,
  springPop,
  useMotionEnabled,
  viewEnter,
} from "./motion";
import { EDITOR_SIZE, motionPolicy, useDisplayPrefs } from "./prefs";
import {
  ASSISTANT_WINDOW,
  clampAssistantWindow,
  defaultAssistantWindow,
  dockedAssistantWindow,
  readAssistantWindow,
  saveAssistantWindow,
} from "./prefs";
import type { AssistantWindow } from "./prefs";
import { boundedText, MESSAGE_BYTES, useTutor } from "./tutor";
import { useUpdater } from "./updater";
import type { Updater } from "./updater";
import "./styles.css";

type View = "home" | "courses" | "learn" | "notes" | "settings";
const icons = [Sprout, Layers3, Code2, Flag];
const editorExtensions = [go(), ...editorTheme];
const navigation: { id: View; label: string; icon: typeof Compass }[] = [
  { id: "home", label: "学习空间", icon: Compass },
  { id: "courses", label: "探索课程", icon: BookOpen },
  { id: "notes", label: "我的笔记", icon: FileText },
];

/**
 * 首页演示台用的示例：全部取自仓库课程里的真实片段，输出也真的是解释器的结果。
 * 与其画一张抽象插画，不如直接让产品自己说话。
 */
const stageExamples = [
  {
    lesson: "go-basic · 第一行 Go 代码",
    code:
      'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello, Go!")\n}',
    output: ["Hello, Go!"],
  },
  {
    lesson: "go-basic · goroutine 与 channel",
    code:
      'package main\n\nimport "fmt"\n\nfunc main() {\n\tch := make(chan int)\n\tgo func() { ch <- 42 }()\n\tfmt.Println(<-ch)\n}',
    output: ["42"],
  },
  {
    lesson: "go-basic · defer 的执行顺序",
    code:
      "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfor i := 1; i <= 3; i++ {\n\t\tdefer fmt.Println(i)\n\t}\n}",
    output: ["3", "2", "1"],
  },
];

const goKeywords = new Set([
  "break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough",
  "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range",
  "return", "select", "struct", "switch", "type", "var", "make", "new", "nil", "true", "false",
]);

type Phase = "idle" | "compile" | "execute" | "compare";

const phaseLabel: Record<Exclude<Phase, "idle">, string> = {
  compile: "正在编译…",
  execute: "正在本地执行…",
  compare: "比对期望输出…",
};

type Kind = "kw" | "str" | "num" | "com" | "fn" | "plain";
type Token = { text: string; kind: Kind };

/** 只为首页演示台的固定片段着色，够用就好，不引入语法高亮依赖。 */
function tokenizeGo(source: string): Token[] {
  const pattern = /(\/\/[^\n]*)|("(?:[^"\\]|\\.)*")|(\d+)|([A-Za-z_][A-Za-z0-9_]*)|([\s\S])/g;
  const tokens: Token[] = [];
  for (const match of source.matchAll(pattern)) {
    const [text, comment, quoted, digit, word] = match;
    if (comment) tokens.push({ text, kind: "com" });
    else if (quoted) tokens.push({ text, kind: "str" });
    else if (digit) tokens.push({ text, kind: "num" });
    // Go 里首字母大写即导出标识符，顺带按函数名扱色，读起来更接近真实编辑器。
    else if (word) tokens.push({ text, kind: goKeywords.has(word) ? "kw" : /^[A-Z]/.test(word) ? "fn" : "plain" });
    else tokens.push({ text, kind: "plain" });
  }
  return tokens;
}

const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

/**
 * 共享元素转场：点中的元素带上 view-transition-name，由浏览器把"旧卡片 → 新页面"
 * 连成一帧动画。平台不支持（含 jsdom）或用户要求减少动效时，退化成直接切换。
 */
function withViewTransition(update: () => void, source?: HTMLElement) {
  const doc = document as Document & {
    startViewTransition?: (callback: () => void) => { finished: Promise<void> };
  };
  const clear = () => {
    document
      .querySelectorAll<HTMLElement>("[style*='view-transition-name']")
      .forEach((node) => {
        node.style.viewTransitionName = "";
      });
  };
  clear();
  if (typeof doc.startViewTransition !== "function") {
    update();
    return;
  }
  if (source) source.style.viewTransitionName = "course-hero";
  void doc.startViewTransition(() => flushSync(update)).finished.finally(clear);
}

/** 安卓 WebView 上给成功与失败一点触觉区分；其他平台静默跳过。 */
function haptic(pattern: number | number[]) {
  if (typeof navigator.vibrate === "function" && /android/i.test(navigator.userAgent)) {
    navigator.vibrate(pattern);
  }
}

/** 把 token 展开成逐字符序列，打字动画只推进一个下标，着色不会因为半个字符串而错乱。 */
function expand(tokens: Token[]) {
  return tokens.flatMap((token) =>
    [...token.text].map((character) => ({ character, kind: token.kind })),
  );
}

function groupVisible(chars: { character: string; kind: Kind }[], visible: number) {
  const shown = chars.slice(0, visible);
  const runs: { text: string; kind: Kind }[] = [];
  for (const item of shown) {
    const last = runs.at(-1);
    if (last && last.kind === item.kind) last.text += item.character;
    else runs.push({ text: item.character, kind: item.kind });
  }
  return runs;
}

export function App() {
  const [data, setData] = useState<Bootstrap>();
  const [loadingError, setLoadingError] = useState("");
  const [view, setView] = useState<View>("home");
  const [lessonId, setLessonId] = useState("");
  const [assistantOpen, setAssistantOpen] = useState(false);
  const [assistantCode, setAssistantCode] = useState<string>();
  const [toast, setToast] = useState("");
  const [quit, setQuit] = useState(false);
  const [uninstalling, setUninstalling] = useState(false);
  const tutor = useTutor();
  const hydrateTutor = tutor.hydrate;
  const { prefs, update } = useDisplayPrefs();
  const updater = useUpdater({
    autoUpdate: prefs.autoUpdate,
    version: data?.runtime.version ?? "",
    content: data?.runtime.content ?? "",
  });
  // 这里在 MotionConfig 之外读取，拿到的是系统设置；用户在本机的选择优先。
  const systemReduce = useReducedMotion();
  const reduce = prefs.reduceMotion ?? Boolean(systemReduce);
  const notify = useCallback((text: string) => setToast(text), []);
  const load = useCallback(() => {
    setLoadingError("");
    api<Bootstrap>("bootstrap")
      .then((value) => {
        setData(value);
        setLessonId(value.progress.lastLesson || value.lessons[0]?.id || "");
      })
      .catch((error) => setLoadingError(error.message));
  }, []);
  useEffect(load, [load]);
  // 关闭或刷新页面时告诉本地引擎一声：窗口关掉之后是要顺带结束进程、留在后台，
  // 还是弹一次询问框，由引擎按设置决定（刷新不算关闭，引擎会等页面回来）。
  useEffect(() => {
    window.addEventListener("pagehide", reportWindowClosed);
    return () => window.removeEventListener("pagehide", reportWindowClosed);
  }, []);
  // bootstrap 回来后再恢复上次的对话：面板在没打开时就已经准备好了内容。
  useEffect(() => hydrateTutor(data?.conversations), [data?.conversations, hydrateTutor]);
  const openLesson = (id: string) => {
    setLessonId(id);
    setAssistantCode(undefined);
    setView("learn");
    window.scrollTo(0, 0);
  };
  const updateProgress = useCallback(
    (progress: Progress) =>
      setData((current) => (current ? { ...current, progress } : current)),
    [],
  );
  const updateSettings = (settings: Settings) =>
    setData((current) => (current ? { ...current, settings } : current));
  const quitApp = useCallback(async () => {
    try {
      await api("quit", {});
    } catch {
      // 引擎可能已经停止，仍然进入退出状态。
    }
    setQuit(true);
  }, []);
  const navigate = (id: View) => {
    setView(id);
    window.scrollTo(0, 0);
  };
  const openAssistant = useCallback((open: boolean) => setAssistantOpen(open), []);

  return (
    <MotionConfig reducedMotion={motionPolicy(reduce)}>
      <Tooltip.Provider delayDuration={350} skipDelayDuration={200}>
        {uninstalling ? (
          <BootScreen
            title="正在卸载 GoStudy"
            description="卸载程序已经启动，请在弹出的窗口里完成卸载。学习进度、笔记与对话都留在本机，卸载不会删除它们；如果中途取消，重新打开 GoStudy 即可继续使用。"
          />
        ) : quit ? (
          <BootScreen
            title="GoStudy 已退出"
            description="本地学习引擎已经停止，现在可以关闭这个窗口。下次打开应用会重新启动引擎。"
          />
        ) : !data ? (
          <BootScreen
            busy
            title={loadingError ? "还差一步，连接本地引擎" : "正在打开你的学习空间"}
            description={loadingError || "准备课程、练习和属于你的进度。"}
            error={loadingError ? load : undefined}
          />
        ) : (
          <Shell
            data={data}
            view={view}
            onNavigate={navigate}
            assistantOpen={assistantOpen}
            tutor={tutor}
            updater={updater}
            onAssistant={openAssistant}
            notesCount={Object.values(data.progress.notes || {}).filter(Boolean).length}
              lessonId={lessonId}
              assistantCode={assistantCode}
              onCodeChange={setAssistantCode}
              onOpenLesson={openLesson}
            onProgress={updateProgress}
            onSettings={updateSettings}
            onNotify={notify}
            onQuit={quitApp}
            onUninstalling={() => setUninstalling(true)}
            prefs={prefs}
            onPrefs={update}
          />
        )}
        <Toast.Provider duration={4500} swipeDirection="up">
          <Toast.Root
            className="toast"
            open={Boolean(toast)}
            onOpenChange={(open) => !open && setToast("")}
          >
            <CheckCheck size={17} />
            <Toast.Description>{toast}</Toast.Description>
            <Toast.Close className="icon-button" aria-label="关闭通知">
              <X size={14} />
            </Toast.Close>
          </Toast.Root>
          <Toast.Viewport className="toast-viewport" />
        </Toast.Provider>
      </Tooltip.Provider>
    </MotionConfig>
  );
}

function Shell({
  data,
  view,
  onNavigate,
  assistantOpen,
  tutor,
  updater,
  onAssistant,
  notesCount,
  lessonId,
  assistantCode,
  onOpenLesson,
  onProgress,
  onSettings,
  onNotify,
  onQuit,
  onUninstalling,
  prefs,
  onPrefs,
  onCodeChange,
}: {
  data: Bootstrap;
  view: View;
  onNavigate: (view: View) => void;
  assistantOpen: boolean;
  tutor: Tutor;
  updater: Updater;
  onAssistant: (open: boolean) => void;
  notesCount: number;
  lessonId: string;
  assistantCode?: string;
  onCodeChange: (code: string) => void;
  onOpenLesson: (id: string) => void;
  onProgress: (progress: Progress) => void;
  onSettings: (settings: Settings) => void;
  onNotify: (text: string) => void;
  onQuit: () => void;
  onUninstalling: () => void;
  prefs: { reduceMotion: boolean | null; editorSize: number; autoUpdate: boolean };
  onPrefs: (patch: { reduceMotion?: boolean | null; editorSize?: number; autoUpdate?: boolean }) => void;
}) {
  const motionOn = useMotionEnabled();
  const [tutorSpace, setTutorSpace] = useState(0);
  const completed = data.progress.completed || [];
  const percent = Math.round((completed.length / data.lessons.length) * 100) || 0;
  const next =
    data.lessons.find(
      (item) => item.id === data.progress.lastLesson && !completed.includes(item.id),
    ) ||
    data.lessons.find((item) => !completed.includes(item.id)) ||
    data.lessons[0];
  const currentLesson = data.lessons.find((item) => item.id === lessonId) || data.lessons[0];
  const currentCourse =
    data.courses.find((item) => item.id === currentLesson.courseId) || data.courses[0];
  const masterCourses = data.courses.filter((course) =>
    course.lessonIds.every((id) => completed.includes(id)),
  ).length;
  /** 首页课程卡 → 课程页是低频的跨层级跳转，值得一段共享元素转场。 */
  const openFrom = (id: string, source?: HTMLElement) => {
    if (motionOn && source) withViewTransition(() => onOpenLesson(id), source);
    else onOpenLesson(id);
  };

  return (
    <div
      className="app"
      style={{ "--tutor-space": assistantOpen ? `${tutorSpace}px` : "0px" } as CSSProperties}
    >
      <header className="topbar">
        <span className="topbar-edge" aria-hidden="true" />
        <div className="topbar-inner">
          <button className="brand" aria-label="GoStudy 首页" onClick={() => onNavigate("home")}>
            <Logo />
            <span>GoStudy</span>
          </button>
          <nav className="main-nav" aria-label="主导航">
            {navigation.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                className={`nav-item ${view === id ? "active" : ""}`}
                onClick={() => onNavigate(id)}
              >
                {view === id && (
                  <motion.span className="nav-pill" layoutId="nav-pill" transition={springControl} />
                )}
                <Icon size={16} strokeWidth={2} />
                {label}
                {id === "notes" && <span className="nav-count">{notesCount}</span>}
              </button>
            ))}
          </nav>
          <div className="topbar-actions">
            <span className="offline-badge">
              <WifiOff size={13} />
              离线可用
            </span>
            <button
              className="assistant-trigger"
              onClick={() => onAssistant(true)}
            >
              <MessagesSquare size={15} />
              AI 导师
            </button>
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <button
                  className="icon-button settings-link"
                  aria-label="偏好与 AI 设置"
                  title={updater.hasApp || updater.hasContent ? "有新版本可以更新" : "偏好与 AI 设置"}
                  onClick={() => onNavigate("settings")}
                >
                  <Settings2 size={19} />
                  {(updater.hasApp || updater.hasContent) && (
                    <span className="update-dot" aria-hidden="true" />
                  )}
                </button>
              </Tooltip.Trigger>
              <Tooltip.Portal>
                <Tooltip.Content className="tooltip" sideOffset={8}>
                  偏好与 AI 设置
                  <Tooltip.Arrow className="tooltip-arrow" />
                </Tooltip.Content>
              </Tooltip.Portal>
            </Tooltip.Root>
            <div className="profile-avatar" title="本地学习者">
              G
            </div>
          </div>
        </div>
      </header>
      <main className="page">
        {view === "home" && (
          <motion.div className="home" key="home" {...viewEnter(motionOn)}>
            <section className="hero">
              <div className="hero-text">
                <span className="pill">
                  <span className="pill-dot" />
                  {data.runtime.engine} · 就绪
                </span>
                <h1>
                  {completed.length ? (
                    <>
                      接着上次的进度，<span className="hl">继续跑</span>。
                    </>
                  ) : (
                    <>
                      先把 Go <span className="hl">跑起来</span>，再读懂它。
                    </>
                  )}
                </h1>
                <p className="hero-lead">
                  每课十几分钟：读一段讲解，改几行真代码，在本地解释执行看输出，卡住了再问导师。
                  不需要安装 Go 环境，也不用联网。
                </p>
                <div className="hero-actions">
                  <button className="button primary large" onClick={() => onOpenLesson(next.id)}>
                    {completed.length ? "继续我的学习" : "开始第一课"}
                    <ArrowRight size={18} />
                  </button>
                  <button className="button secondary large" onClick={() => onNavigate("courses")}>
                    探索课程
                  </button>
                </div>
                <span className="hero-footnote">
                  <Clock3 size={14} />
                  每课约 8–16 分钟 <span>·</span> {data.lessons.length} 个练习{" "}
                  <span>·</span> 进度保存在本机
                </span>
              </div>
              <HeroStage />
            </section>
            <section className="stats-row">
              <StatCard label="学习进度" icon={Compass} index={0} motionOn={motionOn}>
                <strong className="stat-value">
                  {percent}
                  <em>%</em>
                </strong>
                <span className="stat-detail">
                  已完成 {completed.length} / {data.lessons.length} 个练习
                </span>
                <Meter.Root className="meter" value={percent} aria-label="学习进度">
                  <motion.div
                    className="meter-fill"
                    initial={motionOn ? { width: 0 } : false}
                    animate={{ width: `${percent}%` }}
                    transition={springMeter}
                  />
                </Meter.Root>
              </StatCard>
              <StatCard label="完成的章节" icon={Layers3} index={1} motionOn={motionOn}>
                <strong className="stat-value">
                  {masterCourses}
                  <em>/ {data.courses.length}</em>
                </strong>
                <span className="stat-detail">
                  共 {data.courses.reduce((sum, course) => sum + course.lessonIds.length, 0)} 节，
                  按自己的顺序学
                </span>
              </StatCard>
              <StatCard label="练习引擎" icon={Terminal} index={2} motionOn={motionOn}>
                <strong className="stat-value stat-phrase">{data.runtime.engine}</strong>
                <span className="stat-detail">
                  {data.runtime.platform} · 不开放文件与网络访问
                </span>
              </StatCard>
            </section>
            <section className="section">
              <div className="section-head">
                <div>
                  <span className="eyebrow">课程路线</span>
                  <h2>四章内容，从语法到工程</h2>
                </div>
                <button className="text-button" onClick={() => onNavigate("courses")}>
                  全部课程 <ArrowRight size={16} />
                </button>
              </div>
              <div className="course-grid">
                {data.courses.map((course, index) => (
                  <CourseCard
                    key={course.id}
                    course={course}
                    index={index}
                    completed={completed}
                    onOpen={openFrom}
                    motionOn={motionOn}
                  />
                ))}
              </div>
            </section>
            <motion.section
              className="cta-note"
              {...(motionOn
                ? {
                    initial: { opacity: 0, y: 18 },
                    whileInView: { opacity: 1, y: 0 },
                    viewport: { once: true, amount: 0.4 },
                    transition: { duration: 0.3, ease: [0.22, 1, 0.36, 1] },
                  }
                : {})}
            >
              <span className="cta-icon">
                <MessagesSquare size={20} />
              </span>
              <div>
                <strong>卡住的时候，先问，再自己跑一遍。</strong>
                <p>
                  AI 导师可以解释概念、给一点线索、陪你看懂报错；模型由你指定，课程本身完全离线。
                </p>
              </div>
              <button
                className="text-button"
                onClick={() => onAssistant(true)}
              >
                打开导师 <ArrowRight size={16} />
              </button>
            </motion.section>
            <Footer />
          </motion.div>
        )}
        {view === "courses" && (
          <motion.div className="courses-page" key="courses" {...viewEnter(motionOn)}>
            <PageHeading
              label="全部课程"
              title="按章节浏览，也可以直接挑一节开跑。"
              description="四章内容按依赖顺序排好了，进度自动保存在本机，随时可以跳着学。"
            />
            <div className="catalog">
              {data.courses.map((course, index) => {
                const Icon = icons[index % icons.length];
                const done = course.lessonIds.filter((id) => completed.includes(id)).length;
                return (
                  <motion.section
                    className="catalog-course"
                    key={course.id}
                    {...rise(motionOn, index)}
                  >
                    <div className="catalog-heading">
                      <div className="course-icon">
                        <Icon size={24} />
                      </div>
                      <div>
                        <span className="eyebrow">
                          {String(index + 1).padStart(2, "0")} · {course.level}
                        </span>
                        <h2>{course.title}</h2>
                        <p>{course.description}</p>
                      </div>
                      <span className="catalog-progress">
                        {done} / {course.lessonIds.length} 已完成
                      </span>
                    </div>
                    <div className="lesson-list">
                      {course.lessonIds.map((id, i) => {
                        const lesson = data.lessons.find((item) => item.id === id)!;
                        return (
                          <motion.button
                            key={id}
                            className="lesson-row"
                            onClick={() => onOpenLesson(id)}
                            {...fadeIn(motionOn, i)}
                          >
                            <span
                              className={`lesson-number ${completed.includes(id) ? "done" : ""}`}
                            >
                              {completed.includes(id) ? (
                                <Check size={15} />
                              ) : (
                                String(i + 1).padStart(2, "0")
                              )}
                            </span>
                            <div>
                              <strong>{lesson.title}</strong>
                              <span>{lesson.description}</span>
                            </div>
                            <span className="lesson-time">
                              <Clock3 size={13} />
                              {lesson.duration} 分钟
                            </span>
                            <ArrowRight size={16} className="row-arrow" />
                          </motion.button>
                        );
                      })}
                    </div>
                  </motion.section>
                );
              })}
            </div>
            <Footer />
          </motion.div>
        )}
        {view === "learn" && (
          <LearningWorkspace
            key={currentLesson.id}
            lesson={currentLesson}
            course={currentCourse}
            data={data}
            onCodeChange={onCodeChange}
            onProgress={onProgress}
            onNext={onOpenLesson}
            onCourses={() => onNavigate("courses")}
            onAI={() => onAssistant(true)}
            onNotify={onNotify}
            motionOn={motionOn}
          />
        )}
        {view === "notes" && (
          <motion.div className="notes-page" key="notes" {...viewEnter(motionOn)}>
            <Notes data={data} onOpen={onOpenLesson} onNotify={onNotify} motionOn={motionOn} />
          </motion.div>
        )}
        {view === "settings" && (
          <motion.div className="settings-page" key="settings" {...viewEnter(motionOn)}>
            <Preferences
              settings={data.settings}
              onSave={onSettings}
              onNotify={onNotify}
              runtime={data.runtime}
              onQuit={onQuit}
              onUninstalling={onUninstalling}
              prefs={prefs}
              onPrefs={onPrefs}
              updater={updater}
              motionOn={motionOn}
            />
          </motion.div>
        )}
      </main>
      <AnimatePresence>
        {assistantOpen && (
          <Assistant
            tutor={tutor}
            lesson={view === "learn" ? currentLesson : next}
            code={
              view === "learn"
                ? assistantCode ?? data.progress.drafts?.[currentLesson.id]
                : data.progress.drafts?.[next.id]
            }
            lessons={data.lessons}
            settings={data.settings}
            onClose={() => onAssistant(false)}
            onSettings={() => {
              onNavigate("settings");
              onAssistant(false);
            }}
            onDock={setTutorSpace}
            motionOn={motionOn}
          />
        )}
      </AnimatePresence>
    </div>
  );
}

function Logo() {
  return (
    <svg className="logo-mark" viewBox="0 0 32 32" fill="none" aria-hidden="true">
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
  );
}

function BootScreen({
  title,
  description,
  error,
  busy,
}: {
  title: string;
  description: string;
  error?: () => void;
  busy?: boolean;
}) {
  const motionOn = useMotionEnabled();
  return (
    <div className="boot-screen">
      <motion.span
        initial={motionOn ? { scale: 0.86, opacity: 0 } : false}
        animate={{ scale: 1, opacity: 1 }}
        transition={springControl}
      >
        <Logo />
      </motion.span>
      <motion.h1 {...rise(motionOn, 0)}>{title}</motion.h1>
      <motion.p {...rise(motionOn, 1)}>{description}</motion.p>
      {error ? (
        <motion.div {...rise(motionOn, 2)}>
          <p className="muted">
            请通过 GoStudy 启动时给出的完整链接打开，链接中包含本次运行的访问凭证。
          </p>
          <button className="button primary" onClick={error}>
            重新连接 <RotateCcw size={16} />
          </button>
        </motion.div>
      ) : (
        busy && <LoaderCircle className="spin" size={24} color="var(--muted)" />
      )}
    </div>
  );
}

function Footer() {
  return (
    <footer>
      <span className="footer-brand">
        <Logo />
        GoStudy
      </span>
      <span className="footer-motto">课程与练习都在本机，离线可用。</span>
      <span className="footer-local">
        <ShieldCheck size={13} />
        没有账号，没有上传
      </span>
    </footer>
  );
}

function PageHeading({
  label,
  title,
  description,
}: {
  label: string;
  title: string;
  description: string;
}) {
  return (
    <div className="page-heading">
      <span className="eyebrow">{label}</span>
      <h1>{title}</h1>
      <p>{description}</p>
    </div>
  );
}

/* ---------------- 首页演示台 ---------------- */

type StagePhase = "typing" | "running" | "output" | "hold";
type Stage = { index: number; phase: StagePhase; typed: number; out: number };

const stageDelay: Record<StagePhase, number> = {
  typing: 14,
  running: 480,
  output: 280,
  hold: 2600,
};

function advanceStage(stage: Stage, length: number): Stage {
  switch (stage.phase) {
    case "typing": {
      const typed = Math.min(length, stage.typed + 4);
      return typed < length ? { ...stage, typed } : { ...stage, typed, phase: "running" };
    }
    case "running":
      return { ...stage, phase: "output", out: 1 };
    case "output":
      return { ...stage, out: stage.out + 1 };
    default:
      return { index: (stage.index + 1) % stageExamples.length, phase: "typing", typed: 0, out: 0 };
  }
}

function HeroStage() {
  const motionOn = useMotionEnabled();
  const [stage, setStage] = useState<Stage>({ index: 0, phase: "typing", typed: 0, out: 0 });
  const example = stageExamples[stage.index];
  const chars = useMemo(() => expand(tokenizeGo(example.code)), [example.code]);
  const fullLines = useMemo(() => Math.max(1, example.code.split("\n").length), [example.code]);
  const finished = !motionOn || stage.phase === "hold";
  const output = finished ? example.output : example.output.slice(0, stage.out);

  useEffect(() => {
    if (!motionOn) return;
    if (stage.phase === "output" && stage.out >= example.output.length) {
      setStage((current) => ({ ...current, phase: "hold" }));
      return;
    }
    if (stage.phase === "output" && stage.out > example.output.length) return;
    const timer = setTimeout(() => setStage((current) => advanceStage(current, chars.length)), stageDelay[stage.phase]);
    return () => clearTimeout(timer);
  }, [stage, motionOn, chars.length, example.output.length]);

  const typed = motionOn ? stage.typed : chars.length;
  const runs = groupVisible(chars, typed);
  const waiting = motionOn && stage.phase === "typing";
  const running = motionOn && stage.phase === "running";
  return (
    <div className="hero-stage" aria-hidden="true">
      <div className="stage-bar">
        <span className="stage-file">
          <Code2 size={13} />
          main.go
        </span>
        <span className="stage-lesson">{example.lesson}</span>
      </div>
      <div className="stage-code">
        <div className="stage-lines">
          {Array.from({ length: fullLines }, (_, i) => (
            <div key={i}>{i + 1}</div>
          ))}
        </div>
        <div className="stage-source">
          <code>
            {runs.map((run, i) => (
              <span key={i} className={run.kind === "plain" ? undefined : `tok-${run.kind}`}>
                {run.text}
              </span>
            ))}
            {waiting && <span className="stage-caret" />}
          </code>
        </div>
      </div>
      <div className="stage-flow">
        <span>{running ? "go run main.go" : "本地解释执行"}</span>
        <i />
      </div>
      <div className="stage-term">
        <div className="stage-term-head">
          <Terminal size={13} />
          运行结果
          {stage.phase === "hold" && (
            <span className="stage-chip">
              <Check size={11} />
              挑战通过
            </span>
          )}
        </div>
        {output.length ? (
          <div className="stage-out">
            {output.map((line, i) => (
              <div key={i}>
                <span className="stage-prompt">$ </span>
                {line}
              </div>
            ))}
          </div>
        ) : (
          <span className="muted">{running ? "正在执行…" : "等待运行"}</span>
        )}
      </div>
    </div>
  );
}

/* ---------------- 首页卡片 ---------------- */

function StatCard({
  label,
  icon: Icon,
  index,
  motionOn,
  children,
}: {
  label: string;
  icon: typeof Compass;
  index: number;
  motionOn: boolean;
  children: ReactNode;
}) {
  return (
    <motion.div className="stat-card" {...rise(motionOn, index)}>
      <span className="stat-label">
        <Icon size={16} />
        {label}
      </span>
      {children}
    </motion.div>
  );
}

function CourseCard({
  course,
  index,
  completed,
  onOpen,
  motionOn,
}: {
  course: Course;
  index: number;
  completed: string[];
  onOpen: (id: string, source?: HTMLElement) => void;
  motionOn: boolean;
}) {
  const Icon = icons[index % icons.length];
  const done = course.lessonIds.filter((id) => completed.includes(id)).length;
  return (
    <motion.button
      className="course-card"
      onClick={(event) =>
        onOpen(
          course.lessonIds.find((id) => !completed.includes(id)) || course.lessonIds[0],
          event.currentTarget,
        )
      }
      {...rise(motionOn, index)}
      {...lift(motionOn)}
    >
      <div className="course-card-top">
        <div className="course-icon">
          <Icon size={22} strokeWidth={1.9} />
        </div>
        <span className="course-index">{String(index + 1).padStart(2, "0")}</span>
      </div>
      <span className="course-meta">
        {course.level} · {course.lessonIds.length} 节练习 · {course.duration}
      </span>
      <h3>{course.title}</h3>
      <p>{course.description}</p>
      <div className="course-card-bottom">
        <span>
          {done ? `已完成 ${done} / ${course.lessonIds.length}` : "开始第一课"}
          <ArrowRight size={15} />
        </span>
      </div>
      {done > 0 && (
        <motion.div
          className="card-progress"
          initial={motionOn ? { width: 0 } : false}
          animate={{ width: `${(done / course.lessonIds.length) * 100}%` }}
          transition={springMeter}
        />
      )}
    </motion.button>
  );
}

/* ---------------- 学习空间 ---------------- */

const lessonTabs = [
  { id: "concept", name: "理解概念" },
  { id: "task", name: "动手挑战" },
  { id: "notes", name: "我的笔记" },
] as const;

function LearningWorkspace({
  lesson,
  course,
  data,
  onCodeChange,
  onProgress,
  onNext,
  onCourses,
  onAI,
  onNotify,
  motionOn,
}: {
  lesson: Lesson;
  course: Course;
  data: Bootstrap;
  onCodeChange: (code: string) => void;
  onProgress: (p: Progress) => void;
  onNext: (id: string) => void;
  onCourses: () => void;
  onAI: () => void;
  onNotify: (s: string) => void;
  motionOn: boolean;
}) {
  const [code, setCode] = useState(data.progress.drafts?.[lesson.id] ?? lesson.starter);
  const [note, setNote] = useState(data.progress.notes?.[lesson.id] ?? "");
  const [tab, setTab] = useState<"concept" | "task" | "notes">("concept");
  const [result, setResult] = useState<RunResult>();
  const [running, setRunning] = useState(false);
  const [error, setError] = useState("");
  const [hint, setHint] = useState(0);
  const [answer, setAnswer] = useState<number>();
  const [completing, setCompleting] = useState(false);
  // idle → compile → execute → compare：把"跑一次代码"拆成看得见的过程
  const [phase, setPhase] = useState<Phase>("idle");
  const [saved, setSaved] = useState("");
  const lastSaved = useRef({ code, note });
  const savePromise = useRef<Promise<Progress> | undefined>(undefined);
  const save = useCallback(
    (nextCode: string, nextNote: string) => {
      // 写入串行化，避免较早的自动保存覆盖完成状态。
      const request = (savePromise.current ?? Promise.resolve())
        .catch(() => undefined)
        .then(() => api<Progress>("progress", { lessonId: lesson.id, code: nextCode, note: nextNote }));
      savePromise.current = request;
      return request;
    },
    [lesson.id],
  );
  useEffect(() => {
    if (code === lastSaved.current.code && note === lastSaved.current.note) return;
    setSaved("保存中…");
    const timer = setTimeout(() => {
      save(code, note)
        .then((progress) => {
          lastSaved.current = { code, note };
          onProgress(progress);
          setSaved("已保存到本机");
        })
        .catch((err) => {
          setSaved("保存失败");
          onNotify(err.message);
        });
    }, 700);
    return () => clearTimeout(timer);
  }, [code, note, save, onProgress, onNotify]);
  useEffect(() => onCodeChange(code), [code, onCodeChange]);
  const latestDraft = useRef({ code, note });
  latestDraft.current = { code, note };
  useEffect(
    () => () => {
      const draft = latestDraft.current;
      if (draft.code !== lastSaved.current.code || draft.note !== lastSaved.current.note) {
        void save(draft.code, draft.note).then(onProgress).catch((err) => onNotify(err.message));
      }
    },
    [save, onProgress, onNotify],
  );
  const run = async () => {
    setRunning(true);
    setError("");
    setResult(undefined);
    setPhase(motionOn ? "compile" : "execute");
    try {
      if (motionOn) await wait(runPhases.compile);
      setPhase("execute");
      const value = await api<RunResult>("run", { lessonId: lesson.id, code });
      if (motionOn && value.passed) {
        setPhase("compare");
        await wait(runPhases.compare);
      }
      setResult(value);
      setTab("task");
      setPhase("idle");
      haptic(value.passed ? [12] : [8, 40, 8]);
    } catch (err) {
      setError((err as Error).message);
      setPhase("idle");
      haptic([8, 40, 8]);
    } finally {
      setRunning(false);
    }
  };
  const complete = async () => {
    setCompleting(true);
    setError("");
    try {
      await savePromise.current;
      const progress = await api<Progress>("progress", {
        lessonId: lesson.id,
        code,
        note,
        complete: true,
        quizAnswer: answer,
      });
      onProgress(progress);
      onNotify("这一课，真正理解了。进度已保存。");
      haptic([10, 30, 14]);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setCompleting(false);
    }
  };
  // 逐行比对：通过时每一行都能对上期望输出，这个反馈带信息量而不只是装饰
  const expectedLines = useMemo(
    () => lesson.expectedOutput.replace(/\n$/, "").split("\n"),
    [lesson.expectedOutput],
  );
  const actualLines = useMemo(
    () => (result?.output ?? "").replace(/\n$/, "").split("\n"),
    [result?.output],
  );
  const index = data.lessons.findIndex((item) => item.id === lesson.id);
  const next = data.lessons[index + 1];
  const done = data.progress.completed?.includes(lesson.id);
  const completedCount = data.progress.completed?.length ?? 0;
  const correct = answer !== undefined && answer === lesson.quiz.answer;
  return (
    <motion.div className="lesson-page" {...viewEnter(motionOn)}>
      <nav className="breadcrumb" aria-label="当前位置">
        <span className="crumb-icon">
          <Compass size={13} />
        </span>
        <button className="crumb-link" onClick={onCourses}>
          {course.title}
        </button>
        <ChevronRight size={13} />
        <span>第 {index + 1} 课</span>
        <ChevronRight size={13} />
        <strong>{lesson.title}</strong>
      </nav>
      <header className="lesson-header">
        <div className="lesson-heading">
          <h1>{lesson.title}</h1>
          <p>{lesson.description}</p>
          <div className="lesson-meta">
            <span>
              <Clock3 size={14} />约 {lesson.duration} 分钟
            </span>
            <span className="meta-divider" />
            <span>
              {done ? (
                <>
                  <Check size={14} />
                  已掌握
                </>
              ) : (
                "进行中"
              )}
            </span>
            <span className="meta-divider" />
            <div className="lesson-progress">
              <motion.span
                className="progress-chip"
                animate={done && motionOn ? { scale: [1, 1.05, 1] } : { scale: 1 }}
                transition={{ duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
              >
                已完成 {completedCount} / {data.lessons.length} 个练习
              </motion.span>
              <Meter.Root
                className="meter slim"
                value={Math.round((completedCount / data.lessons.length) * 100)}
                aria-label="整体进度"
              >
                <motion.div
                  className="meter-fill"
                  initial={motionOn ? { width: 0 } : false}
                  animate={{ width: `${Math.round((completedCount / data.lessons.length) * 100)}%` }}
                  transition={springMeter}
                />
              </Meter.Root>
            </div>
          </div>
          {lesson.source && (
            <p className="lesson-source">
              <BookOpen size={14} />
              内容依据仓库资料：{lesson.source}
            </p>
          )}
        </div>
        <button
          className="button secondary"
          onClick={onAI}
        >
          <MessagesSquare size={16} />
          和导师一起想
        </button>
      </header>
      <div className="lesson-grid">
        <section className="lesson-card">
          <Tabs.Root value={tab} onValueChange={(value) => setTab(value as typeof tab)}>
            <Tabs.List className="tabs" aria-label="课程内容">
              {lessonTabs.map((item) => (
                <Tabs.Trigger
                  key={item.id}
                  value={item.id}
                  id={`lesson-tab-${item.id}`}
                  aria-controls={`lesson-panel-${item.id}`}
                  className={tab === item.id ? "selected" : ""}
                >
                  {item.name}
                  {tab === item.id && (
                    <motion.span
                      className="tab-underline"
                      layoutId="tab-underline"
                      transition={springControl}
                    />
                  )}
                </Tabs.Trigger>
              ))}
            </Tabs.List>
          </Tabs.Root>
          <div
            className="lesson-body panel-in"
            role="tabpanel"
            id={`lesson-panel-${tab}`}
            aria-labelledby={`lesson-tab-${tab}`}
            key={tab}
          >
            {tab === "concept" && (
              <>
                <span className="eyebrow">01 · 先理解，再动手</span>
                <div className="markdown">
                  <ReactMarkdown>{lesson.concept}</ReactMarkdown>
                </div>
                <Accordion.Root type="single" collapsible className="reference-example">
                  <Accordion.Item value="example">
                    <Accordion.Header>
                      <Accordion.Trigger asChild>
                        <button className="reference-trigger">
                          看看一个完整示例 <ChevronDown size={15} />
                        </button>
                      </Accordion.Trigger>
                    </Accordion.Header>
                    <Accordion.Content className="accordion-content">
                      <div className="markdown accordion-inner">
                        <ReactMarkdown>{"```go\n" + lesson.example + "\n```"}</ReactMarkdown>
                      </div>
                    </Accordion.Content>
                  </Accordion.Item>
                </Accordion.Root>
                <div className="takeaways">
                  <span>
                    <Sprout size={16} />
                    记住这几件事
                  </span>
                  <ul>
                    {lesson.takeaways.map((text) => (
                      <li key={text}>{text}</li>
                    ))}
                  </ul>
                </div>
                <button className="button primary" onClick={() => setTab("task")}>
                  我准备好了，开始挑战 <ArrowRight size={16} />
                </button>
              </>
            )}
            {tab === "task" && (
              <>
                <span className="eyebrow">02 · 把理解写成代码</span>
                <h2>你的挑战</h2>
                <div className="markdown">
                  <ReactMarkdown>{lesson.task}</ReactMarkdown>
                </div>
                <div className="expected-output">
                  <span>期望输出</span>
                  <pre>{lesson.expectedOutput}</pre>
                </div>
                <button
                  className="text-button hint-button"
                  disabled={hint >= lesson.hints.length}
                  onClick={() => setHint((value) => value + 1)}
                >
                  <CircleHelp size={16} />
                  {hint >= lesson.hints.length
                    ? "所有提示已展开"
                    : `给我一点提示 ${hint}/${lesson.hints.length}`}
                  <ChevronDown size={14} />
                </button>
                {lesson.hints.slice(0, hint).map((text, i) => (
                  <motion.div className="hint-box" key={i} {...rise(motionOn, i)}>
                    <span>提示 {i + 1}</span>
                    {text}
                  </motion.div>
                ))}
                <div className="quiz">
                  <span className="eyebrow">03 · 确认你真的懂了</span>
                  <h3>{lesson.quiz.question}</h3>
                  <RadioGroup.Root
                    className="quiz-options"
                    aria-label="理解题"
                    value={answer === undefined ? "" : String(answer)}
                    onValueChange={(value) => setAnswer(Number(value))}
                  >
                    {lesson.quiz.options.map((option, i) => {
                      const picked = answer === i;
                      const state = !picked ? "" : i === lesson.quiz.answer ? "correct" : "incorrect";
                      return (
                        <RadioGroup.Item
                          key={option}
                          value={String(i)}
                          className={`quiz-option ${state}`}
                        >
                          <motion.span
                            className="quiz-letter"
                            animate={motionOn && state === "incorrect" ? { x: [0, -5, 4, -2, 0] } : { x: 0 }}
                            transition={{ duration: 0.32 }}
                          >
                            {String.fromCharCode(65 + i)}
                          </motion.span>
                          {option}
                          {picked && i === lesson.quiz.answer && (
                            <motion.span
                              className="quiz-mark"
                              initial={motionOn ? { scale: 0.4, opacity: 0 } : false}
                              animate={{ scale: 1, opacity: 1 }}
                              transition={springPop}
                            >
                              <Check size={16} />
                            </motion.span>
                          )}
                        </RadioGroup.Item>
                      );
                    })}
                  </RadioGroup.Root>
                  {answer !== undefined && (
                    <motion.p className="quiz-explanation" {...fadeIn(motionOn)}>
                      {correct ? "理解正确。" : "再想一想。"}
                      {lesson.quiz.explanation}
                    </motion.p>
                  )}
                </div>
                <button
                  className="button primary complete-button"
                  disabled={!result?.passed || !correct || completing}
                  onClick={complete}
                >
                  {completing ? (
                    <LoaderCircle className="spin" size={16} />
                  ) : (
                    <CheckCheck size={16} />
                  )}
                  {done ? "再次完成练习" : "完成这一课"}
                </button>
                <p className="small muted">
                  代码通过并答对理解题后，就可以记录这次进步。
                </p>
                {done && (
                  <motion.div
                    className="completion-banner"
                    initial={motionOn ? { opacity: 0, y: 10, scale: 0.98 } : false}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    transition={springControl}
                  >
                    <Check size={18} />
                    <div>
                      <strong>已掌握，做得很好。</strong>
                      <p>
                        {next
                          ? "准备好，把这份理解带到下一课。"
                          : "你已走完当前课程。可以回顾笔记，或继续挑战自己的写法。"}
                      </p>
                    </div>
                    {next && (
                      <button
                        className="icon-button"
                        onClick={() => onNext(next.id)}
                        aria-label="开始下一课"
                      >
                        <ArrowRight size={20} />
                      </button>
                    )}
                  </motion.div>
                )}
              </>
            )}
            {tab === "notes" && (
              <>
                <span className="eyebrow">留下一点自己的理解</span>
                <h2>写给未来的自己</h2>
                <p className="muted">用自己的话解释一遍，比记住答案更有用。</p>
                <textarea
                  className="note-input"
                  aria-label="课程笔记"
                  placeholder="我理解了什么？还有哪些地方想再探索？"
                  value={note}
                  onChange={(event) => setNote(boundedText(event.target.value, 8000))}
                  maxLength={2600}
                />
                <span className="small muted">{saved || "笔记自动保存在本机"}</span>
              </>
            )}
          </div>
        </section>
        <section className="editor-panel">
          <div className="editor-header">
            <span>
              <Code2 size={15} />
              main.go
              <span className="language-tag">Go</span>
            </span>
            <Tooltip.Root>
              <Tooltip.Trigger asChild>
                <button
                  className="icon-button"
                  aria-label="还原初始代码"
                  disabled={running}
                  onClick={() => {
                    setCode(lesson.starter);
                    setResult(undefined);
                  }}
                >
                  <RotateCcw size={15} />
                </button>
              </Tooltip.Trigger>
              <Tooltip.Portal>
                <Tooltip.Content className="tooltip" sideOffset={8}>
                  还原为初始代码
                  <Tooltip.Arrow className="tooltip-arrow" />
                </Tooltip.Content>
              </Tooltip.Portal>
            </Tooltip.Root>
          </div>
          <CodeMirror
            value={code}
            extensions={editorExtensions}
            editable={!running}
            onChange={(value) => {
              setCode(value);
              setResult(undefined);
            }}
            height="360px"
            basicSetup={{ foldGutter: false, highlightActiveLine: true }}
            aria-label="Go 代码编辑器"
          />
          <div className="editor-toolbar">
            <span>
              <span className="tiny-dot" />
              {saved || "Go · 本地解释执行"}
            </span>
            <motion.button
              className="button primary run-button"
              disabled={running}
              onClick={run}
              {...lift(motionOn)}
            >
              {running ? (
                <LoaderCircle className="spin" size={15} />
              ) : (
                <Play size={14} fill="currentColor" />
              )}
              {running ? "运行中" : "运行代码"}
            </motion.button>
          </div>
          <div className="console-header">
            <span>
              <Terminal size={15} />
              运行结果
            </span>
            {phase !== "idle" ? (
              <span className="console-phase-chip">
                <LoaderCircle className="spin" size={11} />
                {phaseLabel[phase]}
              </span>
            ) : result ? (
              <span className={result.passed ? "result-success" : "muted"}>
                {result.passed ? "挑战通过" : result.error ? "需要调整" : "继续试试"} ·{" "}
                {result.durationMs} ms
              </span>
            ) : null}
          </div>
          <div
            className={`console ${result?.error || error ? "console-error" : ""}`}
            aria-live="polite"
          >
            {phase !== "idle" ? (
              <div className="console-phase">
                <LoaderCircle className="spin" size={14} />
                {phaseLabel[phase]}
              </div>
            ) : error ? (
              <pre>{error}</pre>
            ) : result ? (
              <motion.div {...fadeIn(motionOn)}>
                {result.error ? (
                  <pre>{result.error}</pre>
                ) : (
                  <div className="console-lines">
                    {actualLines.map((line, i) => {
                      const matched = expectedLines[i] === line;
                      return (
                        <div className="console-line" key={i}>
                          {result.passed ? (
                            <Check size={12} className="console-line-mark" />
                          ) : (
                            <span className={`console-line-dot ${matched ? "" : "off"}`} />
                          )}
                          <span>{line || "\u00a0"}</span>
                        </div>
                      );
                    })}
                  </div>
                )}
                {!result.passed && !result.error && (
                  <p className="console-tip">运行成功。再对照左侧期望输出调整一下。</p>
                )}
              </motion.div>
            ) : (
              <div className="console-empty">
                <Terminal size={24} />
                <p>运行结果会显示在这里。</p>
                <span>改完代码点「运行代码」，输出和报错都会原样贴出来。</span>
              </div>
            )}
          </div>
          <div className="editor-safety">
            <ShieldCheck size={13} />
            受限 Go 运行环境 · 不开放文件与网络访问
          </div>
        </section>
      </div>
    </motion.div>
  );
}

/* ---------------- AI 导师 ---------------- */

function modelReady(settings: Settings) {
  if (!settings.baseUrl || !settings.model) return false;
  try {
    return (
      settings.hasKey ||
      ["127.0.0.1", "localhost", "[::1]"].includes(new URL(settings.baseUrl).hostname)
    );
  } catch {
    return false;
  }
}

function viewport() {
  return { width: window.innerWidth, height: window.innerHeight };
}

/** 停靠位置要紧贴顶栏下方；顶栏在窄屏会换行变高，所以每次都量一遍。 */
function topbarHeight() {
  return document.querySelector<HTMLElement>(".topbar")?.getBoundingClientRect().height ?? 66;
}

/** 只有宽屏才有"停靠"这回事；窄屏下面板是贴底的一整幅。 */
function wideLayout() {
  return window.matchMedia("(min-width: 900px)").matches;
}

/** 正文放下两栏练习所需的最小宽度；不够宽就不让位，面板浮在内容上（仍然可以拖走）。 */
const CONTENT_MIN_WIDTH = 810;

function canReserveSpace(width: number) {
  return wideLayout() && window.innerWidth - (width + ASSISTANT_WINDOW.margin * 2) >= CONTENT_MIN_WIDTH;
}

/** 列表里的时间只到"哪天/多久以前"这个精度，够用又不吵。 */
function relativeTime(seconds: number) {
  if (!seconds) return "";
  const minutes = (Date.now() / 1000 - seconds) / 60;
  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${Math.floor(minutes)} 分钟前`;
  if (minutes < 60 * 24) return `${Math.floor(minutes / 60)} 小时前`;
  if (minutes < 60 * 24 * 7) return `${Math.floor(minutes / (60 * 24))} 天前`;
  return new Date(seconds * 1000).toLocaleDateString("zh-CN");
}

type Tutor = ReturnType<typeof useTutor>;

/**
 * 导师面板默认停靠在右侧：页面内容会给它让出空间，所以打开面板
 * 不会盖住编辑器、运行按钮或测验，左边该干嘛干嘛。拖走之后它就是一扇普通浮窗。
 * 对话由 useTutor 持有，收起面板只是收起视图，问题和回答都还在。
 */
function Assistant({
  tutor,
  lesson,
  code,
  lessons,
  settings,
  onClose,
  onSettings,
  onDock,
  motionOn,
}: {
  tutor: Tutor;
  lesson: Lesson;
  code?: string;
  lessons: Lesson[];
  settings: Settings;
  onClose: () => void;
  onSettings: () => void;
  /** 停靠时把占位宽度报给外层，页面据此让出位置。 */
  onDock: (width: number) => void;
  motionOn: boolean;
}) {
  const [input, setInput] = useState("");
  const [historyOpen, setHistoryOpen] = useState(false);
  const [renaming, setRenaming] = useState("");
  const [renameText, setRenameText] = useState("");
  const [confirming, setConfirming] = useState("");
  const [moving, setMoving] = useState(false);
  const [box, setBox] = useState<AssistantWindow>(() => readAssistantWindow(viewport(), topbarHeight()));
  const bottom = useRef<HTMLDivElement>(null);
  const inputBox = useRef<HTMLTextAreaElement>(null);
  const panel = useRef<HTMLElement>(null);
  const ready = modelReady(settings);
  const lessonTitles = useMemo(
    () => new Map(lessons.map((item) => [item.id, item.title])),
    [lessons],
  );

  useEffect(() => {
    bottom.current?.scrollIntoView({ behavior: "smooth" });
  }, [tutor.messages, tutor.sending, historyOpen]);
  useEffect(() => {
    const fit = () =>
      setBox((current) =>
        current.docked && wideLayout()
          ? dockedAssistantWindow(current.width, viewport(), topbarHeight())
          : clampAssistantWindow(current, viewport()),
      );
    window.addEventListener("resize", fit);
    return () => window.removeEventListener("resize", fit);
  }, []);
  // 停靠时让页面让出右侧空间：面板与内容并排，而不是盖在内容上。
  useEffect(() => {
    onDock(box.docked && canReserveSpace(box.width) ? box.width + ASSISTANT_WINDOW.margin * 2 : 0);
    return () => onDock(0);
  }, [box.docked, box.width, onDock]);

  const context = { lessonId: lesson.id, code };
  const send = (message = input) => {
    const text = message.trim();
    if (!text || tutor.sending || !ready) return;
    setInput("");
    void tutor.send(text, context);
  };
  const startFresh = async () => {
    setHistoryOpen(false);
    await tutor.create(lesson.id);
    inputBox.current?.focus();
  };

  // 拖动与缩放都用指针事件直接改 left/top/width/height：
  // 面板的坐标要落到 localStorage，不能只停在 motion 的 transform 上。
  const drag = (event: ReactPointerEvent<HTMLElement>) => {
    if (event.button !== 0 || !wideLayout()) return;
    if ((event.target as HTMLElement).closest("button")) return;
    const node = event.currentTarget;
    // 从停靠状态拖走时，先把当前的实际位置接过来，面板不会跳一下。
    const rect = panel.current?.getBoundingClientRect();
    const origin = rect
      ? { docked: false, x: rect.x, y: rect.y, width: rect.width, height: rect.height }
      : { ...box, docked: false };
    const start = { x: event.clientX, y: event.clientY, box: origin };
    node.setPointerCapture(event.pointerId);
    setMoving(true);
    setBox(origin);
    const move = (moveEvent: PointerEvent) =>
      setBox(
        clampAssistantWindow(
          { ...start.box, x: start.box.x + moveEvent.clientX - start.x, y: start.box.y + moveEvent.clientY - start.y },
          viewport(),
        ),
      );
    const stop = () => {
      node.removeEventListener("pointermove", move);
      node.removeEventListener("pointerup", stop);
      node.removeEventListener("pointercancel", stop);
      setMoving(false);
      setBox((current) => {
        saveAssistantWindow(current);
        return current;
      });
    };
    node.addEventListener("pointermove", move);
    node.addEventListener("pointerup", stop);
    node.addEventListener("pointercancel", stop);
  };
  const resize = (event: ReactPointerEvent<HTMLElement>) => {
    if (event.button !== 0) return;
    event.stopPropagation();
    const node = event.currentTarget;
    const rect = panel.current?.getBoundingClientRect();
    const origin = rect
      ? { docked: box.docked, x: rect.x, y: rect.y, width: rect.width, height: rect.height }
      : box;
    const start = { x: event.clientX, y: event.clientY, box: origin };
    const right = origin.x + origin.width;
    const bottomEdge = origin.y + origin.height;
    node.setPointerCapture(event.pointerId);
    setMoving(true);
    const move = (moveEvent: PointerEvent) => {
      const width = start.box.width - (moveEvent.clientX - start.x);
      // 停靠时只调宽度：页面占位跟着变，仍然谁也不挡。
      if (start.box.docked) {
        setBox(dockedAssistantWindow(width, viewport(), topbarHeight()));
        return;
      }
      const height = start.box.height - (moveEvent.clientY - start.y);
      // 从左上角拉伸，右下角保持不动，手感才和拖边框一致。
      setBox(
        clampAssistantWindow(
          { docked: false, width, height, x: right - Math.max(width, ASSISTANT_WINDOW.minWidth), y: bottomEdge - Math.max(height, ASSISTANT_WINDOW.minHeight) },
          viewport(),
        ),
      );
    };
    const stop = () => {
      node.removeEventListener("pointermove", move);
      node.removeEventListener("pointerup", stop);
      node.removeEventListener("pointercancel", stop);
      setMoving(false);
      setBox((current) => {
        saveAssistantWindow(current);
        return current;
      });
    };
    node.addEventListener("pointermove", move);
    node.addEventListener("pointerup", stop);
    node.addEventListener("pointercancel", stop);
  };
  /** 双击标题栏：拖乱了、缩过头了，一键回到默认的停靠位置。 */
  const dockBack = () => {
    const next = defaultAssistantWindow(viewport(), topbarHeight());
    const fitted = { ...next, width: box.width };
    saveAssistantWindow(fitted);
    setBox(fitted);
  };

  return (
    <Dialog.Root open modal={false} onOpenChange={(open) => !open && onClose()}>
      <Dialog.Portal forceMount>
        <Dialog.Content
          asChild
          forceMount
          // 非模态：不许点外面就关闭，否则换个焦点、点一下编辑器就把面板关了。
          onInteractOutside={(event) => event.preventDefault()}
          onPointerDownOutside={(event) => event.preventDefault()}
          onOpenAutoFocus={(event) => event.preventDefault()}
        >
          <motion.aside
            ref={panel}
            className={`assistant-panel ${moving ? "moving" : ""} ${box.docked ? "docked" : ""}`}
            style={
              {
                "--panel-x": `${box.x}px`,
                "--panel-y": `${box.y}px`,
                "--panel-w": `${box.width}px`,
                "--panel-h": `${box.height}px`,
              } as CSSProperties
            }
            variants={dialogVariants}
            initial={motionOn ? "hidden" : false}
            animate="shown"
            exit="out"
          >
            <div
              className="assistant-resize"
              onPointerDown={resize}
              title="拖动调整面板大小"
              aria-hidden="true"
            />
            <header
              className="assistant-header"
              onPointerDown={drag}
              onDoubleClick={dockBack}
              title="拖动可以移动面板 · 双击回到右下角"
            >
              <div className="ai-avatar">
                <MessagesSquare size={19} />
              </div>
              <div className="assistant-heading">
                <Dialog.Title>AI 导师</Dialog.Title>
                <Dialog.Description asChild>
                  <span>
                    {tutor.active?.title || (ready ? settings.model : "尚未连接模型")}
                  </span>
                </Dialog.Description>
              </div>
              <div className="assistant-actions">
                <button
                  className={`icon-button ${historyOpen ? "active" : ""}`}
                  aria-label="对话历史"
                  title="对话历史"
                  onClick={() => setHistoryOpen((open) => !open)}
                >
                  <History size={18} />
                </button>
                <button
                  className="icon-button"
                  aria-label="新建对话"
                  title="新建对话"
                  onClick={() => void startFresh()}
                >
                  <SquarePen size={17} />
                </button>
                <Dialog.Close asChild>
                  <button className="icon-button" aria-label="关闭 AI 导师" title="收起面板，对话会留着">
                    <X size={19} />
                  </button>
                </Dialog.Close>
              </div>
            </header>
            {!historyOpen && (
              <div className="assistant-context">
                <BookOpen size={14} />
                当前上下文<span>{lesson.title}</span>
              </div>
            )}
            <div className="assistant-body">
              {historyOpen ? (
                <div className="assistant-history">
                  <div className="history-head">
                    <button
                      className="icon-button"
                      aria-label="返回对话"
                      onClick={() => setHistoryOpen(false)}
                    >
                      <ChevronLeft size={18} />
                    </button>
                    <strong>对话历史</strong>
                    <span>{tutor.items.length ? `${tutor.items.length} 段` : "还没有"}</span>
                  </div>
                  {tutor.items.length === 0 ? (
                    <p className="history-empty">
                      还没有对话记录。回到对话问第一个问题，就会自动建一段，之后随时能回来接着问。
                    </p>
                  ) : (
                    <ul className="history-list">
                      {tutor.items.map((item) => {
                        const label = item.title || "新对话";
                        return (
                          <li
                            key={item.id}
                            className={`history-item ${item.id === tutor.activeId ? "active" : ""}`}
                          >
                            {renaming === item.id ? (
                              <form
                                className="history-rename"
                                onSubmit={(event) => {
                                  event.preventDefault();
                                  void tutor.rename(item.id, renameText);
                                  setRenaming("");
                                }}
                              >
                                <input
                                  aria-label="对话名称"
                                  value={renameText}
                                  maxLength={60}
                                  autoFocus
                                  onChange={(event) => setRenameText(event.target.value)}
                                  onKeyDown={(event) => event.key === "Escape" && setRenaming("")}
                                />
                                <button className="button primary" type="submit">
                                  保存
                                </button>
                                <button
                                  className="text-button"
                                  type="button"
                                  onClick={() => setRenaming("")}
                                >
                                  取消
                                </button>
                              </form>
                            ) : (
                              <>
                                <button
                                  className="history-open"
                                  onClick={() => {
                                    void tutor.open(item.id);
                                    setHistoryOpen(false);
                                  }}
                                >
                                  <span className="history-title">{label}</span>
                                  <span className="history-meta">
                                    {lessonTitles.get(item.lessonId) ?? "已归档的课程"} ·{" "}
                                    {relativeTime(item.updatedAt)} · {item.messageCount} 条
                                  </span>
                                  {item.preview && (
                                    <span className="history-preview">{item.preview}</span>
                                  )}
                                </button>
                                <div className="history-actions">
                                  {confirming === item.id ? (
                                    <>
                                      <button
                                        className="text-button danger"
                                        onClick={() => {
                                          void tutor.remove(item.id);
                                          setConfirming("");
                                        }}
                                      >
                                        确认删除
                                      </button>
                                      <button
                                        className="text-button"
                                        onClick={() => setConfirming("")}
                                      >
                                        取消
                                      </button>
                                    </>
                                  ) : (
                                    <>
                                      <button
                                        className="icon-button"
                                        aria-label={`重命名 ${label}`}
                                        onClick={() => {
                                          setRenaming(item.id);
                                          setRenameText(label);
                                        }}
                                      >
                                        <Pencil size={15} />
                                      </button>
                                      <button
                                        className="icon-button"
                                        aria-label={`删除 ${label}`}
                                        onClick={() => setConfirming(item.id)}
                                      >
                                        <Trash2 size={15} />
                                      </button>
                                    </>
                                  )}
                                </div>
                              </>
                            )}
                          </li>
                        );
                      })}
                    </ul>
                  )}
                </div>
              ) : (
                <>
                  <div className="chat-messages">
                    {tutor.opening && (
                      <div className="thinking">
                        <LoaderCircle size={15} className="spin" />
                        正在打开这段对话…
                      </div>
                    )}
                    {!tutor.opening && tutor.messages.length === 0 && (
                      <div className="assistant-welcome">
                        <span className="welcome-mark">
                          <MessagesSquare size={24} />
                        </span>
                        <h3>先自己跑一遍，再来问为什么。</h3>
                        <p>
                          概念、线索、报错信息都可以拿进来讨论；
                          <br />
                          但它不会替你写完这一课。
                        </p>
                        {!ready ? (
                          <div className="ai-setup">
                            <span>先连接你选择的模型</span>
                            <p>
                              对话需要你自己的 API Key；课程讲解、练习和进度不需要联网，仍然可以正常使用。
                            </p>
                            <button className="button primary" onClick={onSettings}>
                              配置 AI 导师 <ArrowRight size={16} />
                            </button>
                          </div>
                        ) : (
                          <div className="suggestions">
                            {[
                              "用一个生活中的例子解释这一课",
                              "给我一点提示，先不要告诉我答案",
                              "看看我的代码，引导我找出问题",
                            ].map((text, i) => (
                              <motion.button key={text} onClick={() => send(text)} {...rise(motionOn, i)}>
                                {text}
                                <Plus size={14} />
                              </motion.button>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                    {tutor.messages.map((message, i) => (
                      <motion.div
                        key={i}
                        className={`chat-message ${message.role}`}
                        {...rise(motionOn, 0)}
                      >
                        <span className="message-label">
                          {message.role === "user" ? "你" : "AI 导师"}
                        </span>
                        <div className="markdown">
                          <ReactMarkdown>{message.content}</ReactMarkdown>
                        </div>
                      </motion.div>
                    ))}
                    {tutor.sending && (
                      <div className="thinking">
                        <LoaderCircle size={15} className="spin" />
                        正在换个角度想一想…
                      </div>
                    )}
                    {tutor.error && (
                      <div className="error-box" role="alert">
                        {tutor.error}
                        <button className="text-button" onClick={() => void tutor.retry(context)}>
                          重试
                        </button>
                      </div>
                    )}
                    <div ref={bottom} />
                  </div>
                  <form
                    className="chat-compose"
                    onSubmit={(event) => {
                      event.preventDefault();
                      send();
                    }}
                  >
                    <textarea
                      ref={inputBox}
                      aria-label="向 AI 导师提问"
                      placeholder={ready ? "把你的问题说出来…" : "配置模型后，开始你的第一次对话"}
                      value={input}
                      onChange={(event) => setInput(boundedText(event.target.value, MESSAGE_BYTES))}
                      disabled={!ready}
                      maxLength={2000}
                      onKeyDown={(event) => {
                        if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing) {
                          event.preventDefault();
                          send();
                        }
                      }}
                    />
                    <div>
                      <span>Enter 发送 · Shift + Enter 换行</span>
                      <button
                        className="send-button"
                        type="submit"
                        disabled={tutor.sending || !input.trim() || !ready}
                        aria-label="发送消息"
                      >
                        <Send size={17} />
                      </button>
                    </div>
                  </form>
                  <p className="ai-disclaimer">
                    <ShieldCheck size={12} />
                    仅发送当前课题、代码、课程笔记与本轮对话；对话保存在本机。
                  </p>
                </>
              )}
            </div>
          </motion.aside>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}


/* ---------------- 版本与更新 ---------------- */

/** 字节数变成人能读的大小。 */
function readableSize(bytes: number) {
  if (!bytes) return "";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

const RELEASES_PAGE = "https://github.com/xianshujun/awsome-go-study/releases/latest";

function UpdateCard({
  updater,
  runtime,
  prefs,
  onPrefs,
  motionOn,
}: {
  updater: Updater;
  runtime: Bootstrap["runtime"];
  prefs: { autoUpdate: boolean };
  onPrefs: (patch: { autoUpdate?: boolean }) => void;
  motionOn: boolean;
}) {
  const { status, checking, download, error, applying } = updater;
  const current = status?.current || runtime.version || "dev";
  const content = status?.content || runtime.content || "";
  const busy = download.state === "downloading" || download.state === "verifying";
  const ready = download.state === "ready";
  const percent = download.total > 0 ? Math.min(100, Math.round((download.received / download.total) * 100)) : 0;
  const app = status?.app;
  const contentUpdate = status?.contentUpdate;

  return (
    <motion.section className="settings-card" {...rise(motionOn, 2)}>
      <div className="settings-card-heading">
        <div className="course-icon">
          <ArrowDownToLine size={22} />
        </div>
        <div>
          <h2>版本与更新</h2>
          <p>更新包只从 GitHub / Gitee 的发布页下载，落到本机前会校验 sha256；程序不常驻、不开后台服务。</p>
        </div>
        <span className={`status-pill ${app || contentUpdate ? "connected" : ""}`}>
          {app ? `有新版本 ${app.version}` : contentUpdate ? "有课程内容更新" : "已是最新"}
        </span>
      </div>

      <div className="update-summary">
        <div>
          <strong>GoStudy {current}</strong>
          <span>
            {content ? `课程内容 ${content} · ` : ""}
            {status?.checked ? `上次检查：${relativeTime(status.checked)}` : "还没有检查过"}
          </span>
        </div>
        <button className="button secondary" onClick={() => void updater.check(true)} disabled={checking || busy}>
          {checking ? "正在检查…" : "检查更新"}
        </button>
      </div>

      <div className="pref-row">
        <div className="pref-copy">
          <strong>每天自动检查一次</strong>
          <p>只拉一个很小的版本清单，不含任何学习数据；关掉之后仍然可以手动检查。</p>
        </div>
        <div className="pref-control">
          <Switch.Root
            className="switch"
            aria-label="每天自动检查一次"
            checked={prefs.autoUpdate}
            onCheckedChange={(value) => onPrefs({ autoUpdate: value })}
          >
            <Switch.Thumb className="switch-thumb" />
          </Switch.Root>
        </div>
      </div>

      {app && !busy && !ready && (
        <div className="update-box">
          <div className="update-box-head">
            <strong>
              新版本 {app.version}
              {app.released ? ` · ${app.released}` : ""}
            </strong>
            <span>{readableSize(app.size)}</span>
          </div>
          {app.notes && (
            <div className="markdown update-notes">
              <ReactMarkdown>{app.notes}</ReactMarkdown>
            </div>
          )}
          <div className="update-actions">
            <button className="button primary" onClick={() => void updater.downloadUpdate()}>
              <ArrowDownToLine size={16} />
              下载并安装 {app.version}
            </button>
            <a className="text-button" href={RELEASES_PAGE} target="_blank" rel="noreferrer">
              打开发布页
            </a>
          </div>
        </div>
      )}

      {busy && (
        <div className="update-box">
          <div className="update-box-head">
            <strong>{download.state === "verifying" ? "正在校验更新包…" : `正在下载 ${download.asset ?? ""}`}</strong>
            <span>
              {readableSize(download.received)}
              {download.total ? ` / ${readableSize(download.total)}` : ""}
            </span>
          </div>
          <Meter.Root className="meter update-meter" value={percent} max={100} aria-label="下载进度">
            <div className="meter-fill" style={{ width: `${percent}%` }} />
          </Meter.Root>
        </div>
      )}

      {ready && (
        <div className="update-box">
          <div className="update-box-head">
            <strong>更新包已就绪</strong>
            <span>{readableSize(download.received)}</span>
          </div>
          <p className="update-hint">
            {download.mode === "android"
              ? "点下面的按钮后，系统会问一次是否允许安装；装完重新打开就是新版本。"
              : download.mode === "installer"
                ? "安装版会静默升级并保留你的设置，可能需要你在系统弹窗里点一次「是」。"
                : "点下面的按钮后应用会自动替换并重启，学习进度和对话都在，不受影响。"}
          </p>
          <div className="update-actions">
            <button className="button primary" onClick={() => void updater.install()}>
              {download.mode === "android" ? "安装更新" : "立即重启并更新"}
            </button>
          </div>
        </div>
      )}

      {applying === "app" && !ready && !busy && !error && (
        <p className="update-hint">正在准备更新…</p>
      )}

      {contentUpdate && (
        <div className="update-box">
          <div className="update-box-head">
            <strong>课程内容有更新</strong>
            <span>{readableSize(contentUpdate.size)}</span>
          </div>
          <p className="update-hint">
            内容包只含课程讲义与练习，不改程序：更新后页面会自动刷新，
            {content ? `当前内容是 ${content}，新的是 ${contentUpdate.version}。` : `新版是 ${contentUpdate.version}。`}
          </p>
          <div className="update-actions">
            <button className="button secondary" onClick={() => void updater.updateContent()} disabled={applying === "content"}>
              {applying === "content" ? "正在更新内容…" : "更新课程内容"}
            </button>
          </div>
        </div>
      )}

      {!app && !contentUpdate && !busy && !ready && !checking && !status?.error && (
        <p className="update-hint">
          当前已经是最新的版本{content ? `，课程内容是 ${content}` : ""}。
        </p>
      )}

      {(error || status?.error || status?.warning) && (
        <div className="error-box" role="alert">
          {error || status?.error || status?.warning}
          <a className="text-button" href={RELEASES_PAGE} target="_blank" rel="noreferrer">
            手动下载
          </a>
        </div>
      )}
    </motion.section>
  );
}

/* ---------------- 设置 ---------------- */

function PrefsPanel({
  prefs,
  onPrefs,
}: {
  prefs: { reduceMotion: boolean | null; editorSize: number };
  onPrefs: (patch: { reduceMotion?: boolean | null; editorSize?: number }) => void;
}) {
  // 在 MotionConfig 之内读取，拿到的已经是"系统 + 本机选择"叠加后的结果。
  const effectiveReduce = prefs.reduceMotion ?? Boolean(useReducedMotion());
  return (
    <div className="pref-list">
      <div className="pref-row">
        <div className="pref-copy">
          <strong>减少动效</strong>
          <p>
            关闭界面过渡与演示动画。
            {prefs.reduceMotion === null
              ? "当前跟随系统设置。"
              : "已在本机手动指定，不跟随系统。"}
          </p>
        </div>
        <div className="pref-control">
          <Switch.Root
            className="switch"
            aria-label="减少动效"
            checked={effectiveReduce}
            onCheckedChange={(checked) => onPrefs({ reduceMotion: checked })}
          >
            <motion.span
              className="switch-thumb"
              animate={{ x: effectiveReduce ? 19 : 0 }}
              transition={springControl}
            />
          </Switch.Root>
        </div>
      </div>
      <div className="pref-row">
        <div className="pref-copy">
          <strong>代码字号</strong>
          <p>调整练习编辑器的字号，只影响这台设备。</p>
        </div>
        <div className="pref-control">
          <span className="pref-readout">{prefs.editorSize} px</span>
          <Slider.Root
            className="slider"
            min={EDITOR_SIZE.min}
            max={EDITOR_SIZE.max}
            step={EDITOR_SIZE.step}
            value={[prefs.editorSize]}
            onValueChange={([value]) => onPrefs({ editorSize: value })}
          >
            <Slider.Track className="slider-track">
              <Slider.Range className="slider-range" />
            </Slider.Track>
            {/* role="slider" 落在 Thumb 上，可访问名也要加在这里 */}
            <Slider.Thumb className="slider-thumb" aria-label="代码字号" />
          </Slider.Root>
        </div>
      </div>
      {prefs.reduceMotion !== null && (
        <div className="pref-row">
          <div className="pref-copy">
            <strong>恢复跟随系统</strong>
            <p>清除本机指定，重新按系统设置决定是否播放动效。</p>
          </div>
          <div className="pref-control">
            <button className="text-button pref-reset" onClick={() => onPrefs({ reduceMotion: null })}>
              跟随系统
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function Preferences({
  settings,
  onSave,
  onNotify,
  runtime,
  onQuit,
  onUninstalling,
  prefs,
  onPrefs,
  updater,
  motionOn,
}: {
  settings: Settings;
  onSave: (s: Settings) => void;
  onNotify: (s: string) => void;
  runtime: Bootstrap["runtime"];
  onQuit: () => void;
  onUninstalling: () => void;
  prefs: { reduceMotion: boolean | null; editorSize: number; autoUpdate: boolean };
  onPrefs: (patch: { reduceMotion?: boolean | null; editorSize?: number; autoUpdate?: boolean }) => void;
  updater: Updater;
  motionOn: boolean;
}) {
  const [baseUrl, setBaseUrl] = useState(settings.baseUrl || "https://api.openai.com/v1");
  const [model, setModel] = useState(settings.model || "gpt-4o-mini");
  const [key, setKey] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState("");
  const submit = async (clearKey = false) => {
    setSaving(true);
    setError("");
    setTestResult("");
    try {
      const updated = await api<Settings>("settings", {
        baseUrl: baseUrl.trim(),
        model: model.trim(),
        ...(key ? { apiKey: key.trim() } : {}),
        clearKey,
      });
      onSave(updated);
      setKey("");
      onNotify(clearKey ? "本机保存的 API Key 已移除。" : "模型配置已保存到本机。");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSaving(false);
    }
  };
  const test = async () => {
    setTesting(true);
    setError("");
    setTestResult("");
    try {
      const result = await api<{ reply: string }>("chat", {
        message: "请用一句简短的中文确认连接成功。",
      });
      setTestResult(result.reply);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setTesting(false);
    }
  };
  return (
    <>
      <PageHeading
        label="模型与偏好"
        title="课程离线跑，模型由你指定。"
        description="学习数据留在本机；只有 AI 对话会发给你自己配置的服务商。"
      />
      <motion.section className="settings-card" {...rise(motionOn, 0)}>
        <div className="settings-card-heading">
          <div className="course-icon">
            <MessagesSquare size={22} />
          </div>
          <div>
            <h2>连接 AI 导师</h2>
            <p>支持 OpenAI Chat Completions 兼容接口，包括兼容的本地模型服务。</p>
          </div>
          <span className={`status-pill ${modelReady(settings) ? "connected" : ""}`}>
            {modelReady(settings) ? "已配置" : "未配置"}
          </span>
        </div>
        <form
          onSubmit={(event) => {
            event.preventDefault();
            void submit();
          }}
        >
          <label>
            API 地址
            <span>填写基础地址，例如 https://api.openai.com/v1；不要添加 /chat/completions。</span>
            <input
              type="url"
              value={baseUrl}
              onChange={(event) => setBaseUrl(event.target.value)}
              required
              placeholder="https://api.openai.com/v1"
              spellCheck={false}
            />
          </label>
          <div className="form-columns">
            <label>
              模型名称
              <span>使用服务商提供的准确模型 ID。</span>
              <input
                value={model}
                onChange={(event) => setModel(event.target.value)}
                required
                placeholder="gpt-4o-mini"
                spellCheck={false}
              />
            </label>
            <label>
              API Key
              <span>
                {settings.hasKey ? "已保存；留空保留现有 Key。" : "保存在本机文件中，当前版本未加密。"}
              </span>
              <input
                type="password"
                value={key}
                onChange={(event) => setKey(event.target.value)}
                placeholder={settings.hasKey ? "•••••••• 已配置" : "输入你的 API Key"}
                autoComplete="off"
                spellCheck={false}
              />
            </label>
          </div>
          <div className="privacy-note">
            <ShieldCheck size={18} />
            <p>
              课程、代码、当前课的笔记与本轮对话才会发送到配置的服务商。API Key
              保存在本机，支持 POSIX 权限的系统将其设为仅当前用户可读；不要在代码或对话里放入敏感信息。
            </p>
          </div>
          {error && (
            <div className="error-box" role="alert">
              {error}
            </div>
          )}
          {testResult && (
            <motion.div className="test-result" {...fadeIn(motionOn)}>
              <Check size={16} />
              <span>{testResult}</span>
            </motion.div>
          )}
          <div className="form-actions">
            <button className="button primary" type="submit" disabled={saving}>
              {saving ? <LoaderCircle className="spin" size={16} /> : <Check size={16} />}
              保存配置
            </button>
            <button
              className="button secondary"
              type="button"
              disabled={!modelReady(settings) || testing || saving}
              onClick={test}
            >
              {testing ? <LoaderCircle size={16} className="spin" /> : <MessageCircle size={16} />}
              测试已保存的连接
            </button>
            {settings.hasKey && (
              <button
                className="text-button remove-key"
                type="button"
                disabled={saving}
                onClick={() => submit(true)}
              >
                移除 Key
              </button>
            )}
          </div>
          <p className="small muted">
            测试会向已保存的模型发送一条简短消息，可能产生 API
            费用。本地 HTTP 接口仅允许回环地址；本地模型若无需密钥，可填写 local。
          </p>
        </form>
      </motion.section>
      <motion.section className="settings-card" {...rise(motionOn, 1)}>
        <div className="settings-card-heading">
          <div className="course-icon">
            <Settings2 size={22} />
          </div>
          <div>
            <h2>显示与动效</h2>
            <p>这些偏好只保存在这台设备上，不影响课程本身。</p>
          </div>
        </div>
        <PrefsPanel prefs={prefs} onPrefs={onPrefs} />
      </motion.section>
      <UpdateCard
        updater={updater}
        runtime={runtime}
        prefs={prefs}
        onPrefs={onPrefs}
        motionOn={motionOn}
      />
      <motion.section className="settings-card" {...rise(motionOn, 3)}>
        <div className="settings-card-heading">
          <div className="course-icon">
            <ShieldCheck size={22} />
          </div>
          <div>
            <h2>本地优先，安心探索</h2>
            <p>没有账号，学习数据不上传。你的进度属于你。</p>
          </div>
        </div>
        <div className="runtime-grid">
          <div>
            <span>运行平台</span>
            <strong>{runtime.platform}</strong>
          </div>
          <div>
            <span>练习引擎</span>
            <strong>{runtime.engine}</strong>
          </div>
          <div>
            <span>课程与笔记</span>
            <strong>保存在当前设备</strong>
          </div>
          <div>
            <span>程序版本</span>
            <strong>GoStudy {runtime.version || "dev"}</strong>
          </div>
        </div>
      </motion.section>
      <CloseAndUninstallCard
        settings={settings}
        runtime={runtime}
        onSave={onSave}
        onNotify={onNotify}
        onQuit={onQuit}
        onUninstalling={onUninstalling}
        motionOn={motionOn}
      />
      <Footer />
    </>
  );
}

/* ---------------- 关闭与卸载 ---------------- */

const CLOSE_OPTIONS: { value: CloseAction; label: string; hint: string }[] = [
  { value: "ask", label: "每次问我", hint: "每次关窗口都弹一个询问框，可以勾选不再提示。" },
  { value: "background", label: "后台静默运行", hint: "窗口关掉之后进程留着，下次打开 GoStudy 会连回它，不会再起第二个。" },
  { value: "quit", label: "直接退出程序", hint: "窗口关掉就结束进程，任务管理器里不会留下 GoStudy。" },
];

function CloseAndUninstallCard({
  settings,
  runtime,
  onSave,
  onNotify,
  onQuit,
  onUninstalling,
  motionOn,
}: {
  settings: Settings;
  runtime: Bootstrap["runtime"];
  onSave: (settings: Settings) => void;
  onNotify: (text: string) => void;
  onQuit: () => void;
  onUninstalling: () => void;
  motionOn: boolean;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [confirming, setConfirming] = useState(false);
  const desktop = !/android|ios/i.test(runtime.platform);
  const action = settings.closeAction ?? "ask";
  const changeCloseAction = async (value: CloseAction) => {
    setBusy(true);
    setError("");
    try {
      const updated = await api<Settings>("close-behavior", { action: value });
      onSave(updated);
      onNotify("关闭方式已保存。");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  };
  const uninstall = async () => {
    setBusy(true);
    setError("");
    try {
      await api<{ status: string }>("uninstall", {});
      // 引擎随后会关掉窗口、退出进程，让卸载程序能删掉程序文件。
      setConfirming(false);
      onUninstalling();
    } catch (err) {
      setError((err as Error).message);
      setBusy(false);
    }
  };
  if (!desktop) return null;
  return (
    <motion.section className="settings-card" {...rise(motionOn, 4)}>
      <div className="settings-card-heading">
        <div className="course-icon">
          <Power size={22} />
        </div>
        <div>
          <h2>关闭与卸载</h2>
          <p>关掉应用窗口之后怎么收尾，以及在这台设备上退出或卸载 GoStudy。</p>
        </div>
      </div>
      {runtime.closePrompt ? (
        <div className="choice-row">
          <div className="pref-copy">
            <strong>关闭应用窗口时</strong>
            <p>关掉应用窗口不会通知本地引擎，所以先在这里定好。{CLOSE_OPTIONS.find((option) => option.value === action)?.hint}</p>
          </div>
          <RadioGroup.Root
            className="choice-group"
            aria-label="关闭应用窗口时"
            value={action}
            disabled={busy}
            onValueChange={(value) => void changeCloseAction(value as CloseAction)}
          >
            {CLOSE_OPTIONS.map((option) => (
              <RadioGroup.Item key={option.value} value={option.value} className="choice">
                {option.label}
              </RadioGroup.Item>
            ))}
          </RadioGroup.Root>
        </div>
      ) : (
        <p className="small muted">
          这个平台关掉窗口之后会直接结束进程；后台静默运行与关闭提示由 Windows 版提供。
        </p>
      )}
      <div className="quit-row">
        <div>
          <strong>退出 GoStudy</strong>
          <p>立刻停止本机学习引擎并结束进程。学习进度与对话都已经写在磁盘上，重新打开就在。</p>
        </div>
        <button className="button secondary" onClick={() => void onQuit()}>
          <Power size={16} />
          退出 GoStudy
        </button>
      </div>
      {runtime.canUninstall && (
        <div className="quit-row">
          <div>
            <strong>卸载 GoStudy</strong>
            <p>
              删掉程序文件、开始菜单与桌面快捷方式。学习进度、笔记与对话保存在用户配置目录，
              卸载不会删除它们。
            </p>
          </div>
          <button
            className="button secondary uninstall"
            onClick={() => {
              setError("");
              setConfirming(true);
            }}
          >
            <Trash2 size={16} />
            卸载
          </button>
        </div>
      )}
      {error && !confirming && (
        <div className="error-box" role="alert">
          {error}
        </div>
      )}
      <AnimatePresence>
        {confirming && (
          <Dialog.Root open onOpenChange={(open) => !open && setConfirming(false)} modal>
            <Dialog.Portal forceMount>
              <Dialog.Content asChild forceMount aria-label="确认卸载 GoStudy">
                <motion.div
                  className="export-dialog"
                  initial={motionOn ? { opacity: 0 } : false}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                >
                  <motion.div
                    className="export-card confirm-card"
                    variants={dialogVariants}
                    initial={motionOn ? "hidden" : false}
                    animate="shown"
                    exit="out"
                  >
                    <div className="section-head">
                      <Dialog.Title>卸载 GoStudy</Dialog.Title>
                      <Dialog.Close asChild>
                        <button className="icon-button" aria-label="关闭卸载确认">
                          <X size={20} />
                        </button>
                      </Dialog.Close>
                    </div>
                    <p className="small muted">
                      接下来会启动卸载程序，删掉程序文件、开始菜单与桌面快捷方式。
                      学习进度、笔记与对话留在本机，卸载不会动它们。
                    </p>
                    <p className="small muted">
                      卸载程序开始后 GoStudy 会自己退出；中途取消也没关系，重新打开就能继续用。
                    </p>
                    {error && (
                      <div className="error-box" role="alert">
                        {error}
                      </div>
                    )}
                    <div className="confirm-actions">
                      <button className="button danger" disabled={busy} onClick={() => void uninstall()}>
                        {busy ? <LoaderCircle className="spin" size={16} /> : <Trash2 size={16} />}
                        确认卸载
                      </button>
                      <Dialog.Close asChild>
                        <button className="button secondary" disabled={busy}>
                          先不卸载
                        </button>
                      </Dialog.Close>
                    </div>
                  </motion.div>
                </motion.div>
              </Dialog.Content>
            </Dialog.Portal>
          </Dialog.Root>
        )}
      </AnimatePresence>
    </motion.section>
  );
}

/* ---------------- 笔记 ---------------- */

function Notes({
  data,
  onOpen,
  onNotify,
  motionOn,
}: {
  data: Bootstrap;
  onOpen: (id: string) => void;
  onNotify: (s: string) => void;
  motionOn: boolean;
}) {
  const [exportContent, setExportContent] = useState("");
  const notes = data.lessons.filter((lesson) => data.progress.notes?.[lesson.id]?.trim());
  const exportNotes = () => {
    const content =
      "# 我的 GoStudy 学习笔记\n\n" +
      notes
        .map((lesson) => `## ${lesson.title}\n\n${data.progress.notes[lesson.id]}\n`)
        .join("\n");
    if (navigator.userAgent.includes("GoStudyAndroid")) {
      setExportContent(content);
      return;
    }
    const url = URL.createObjectURL(new Blob([content], { type: "text/markdown;charset=utf-8" }));
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "GoStudy-学习笔记.md";
    anchor.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
    onNotify("学习笔记下载已发起。");
  };
  return (
    <>
      <AnimatePresence>
        {exportContent && (
          <Dialog.Root open onOpenChange={(open) => !open && setExportContent("")} modal>
            <Dialog.Portal forceMount>
              <Dialog.Content asChild forceMount aria-label="导出学习笔记">
                <motion.div
                  className="export-dialog"
                  initial={motionOn ? { opacity: 0 } : false}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                >
                  <motion.div className="export-card" variants={dialogVariants} initial={motionOn ? "hidden" : false} animate="shown" exit="out">
                    <div className="section-head">
                      <Dialog.Title>导出学习笔记</Dialog.Title>
                      <Dialog.Close asChild>
                        <button className="icon-button" aria-label="关闭导出">
                          <X size={20} />
                        </button>
                      </Dialog.Close>
                    </div>
                    <p className="small muted">长按文本可以全选并复制，也可以使用下方按钮。</p>
                    <textarea
                      readOnly
                      value={exportContent}
                      aria-label="导出的 Markdown 笔记"
                      onFocus={(event) => event.target.select()}
                    />
                    <button
                      className="button primary"
                      onClick={async () => {
                        try {
                          await navigator.clipboard.writeText(exportContent);
                          onNotify("学习笔记已复制。");
                        } catch {
                          onNotify("当前环境不允许剪贴板访问，请长按文本手动复制。");
                        }
                      }}
                    >
                      复制笔记 <FileText size={15} />
                    </button>
                  </motion.div>
                </motion.div>
              </Dialog.Content>
            </Dialog.Portal>
          </Dialog.Root>
        )}
      </AnimatePresence>
      <div className="notes-heading">
        <PageHeading
          label="我的笔记"
          title="每一课留下的想法，都在这里。"
          description="回看时会发现，自己的理解已经连成了一条线。"
        />
        <button className="button secondary" disabled={!notes.length} onClick={exportNotes}>
          <ArrowDownToLine size={16} />
          导出笔记
        </button>
      </div>
      {notes.length ? (
        <div className="notes-grid">
          {notes.map((lesson, index) => (
            <motion.article className="note-card" key={lesson.id} {...rise(motionOn, index)}>
              <span className="eyebrow">
                {data.courses.find((c) => c.id === lesson.courseId)?.title}
              </span>
              <h2>{lesson.title}</h2>
              <div className="markdown">
                <ReactMarkdown>{data.progress.notes[lesson.id]}</ReactMarkdown>
              </div>
              <button className="text-button" onClick={() => onOpen(lesson.id)}>
                回到这一课 <ArrowRight size={15} />
              </button>
            </motion.article>
          ))}
        </div>
      ) : (
        <motion.div className="empty-state" {...rise(motionOn, 0)}>
          <div className="empty-illustration">
            <FileText size={34} strokeWidth={1.4} />
            <Sprout size={20} />
          </div>
          <h2>第一条笔记，等你写下。</h2>
          <p>
            打开任意课程，在「我的笔记」里
            <br />
            记下一个想法、一个问题，或一次恍然大悟。
          </p>
          <button
            className="button primary"
            onClick={() => onOpen(data.progress.lastLesson || data.lessons[0].id)}
          >
            去学一点新的 <ArrowRight size={16} />
          </button>
        </motion.div>
      )}
      <Footer />
    </>
  );
}

const rootElement = document.getElementById("root");
if (rootElement) createRoot(rootElement).render(<App />);
