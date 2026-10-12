import { useCallback, useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import CodeMirror from "@uiw/react-codemirror";
import { go } from "@codemirror/lang-go";
import ReactMarkdown from "react-markdown";
import {
  ArrowDownToLine,
  ArrowRight,
  BookOpen,
  Check,
  CheckCheck,
  ChevronDown,
  ChevronRight,
  CircleHelp,
  Clock3,
  Code2,
  Compass,
  FileText,
  Flag,
  Layers3,
  LoaderCircle,
  MessageCircle,
  Play,
  Plus,
  Power,
  RotateCcw,
  Send,
  Settings2,
  ShieldCheck,
  Sparkles,
  Sprout,
  Terminal,
  WifiOff,
  X,
} from "lucide-react";
import { api } from "./api";
import type {
  Bootstrap,
  Course,
  Lesson,
  Message,
  Progress,
  RunResult,
  Settings,
} from "./api";
import "./styles.css";

type View = "home" | "courses" | "learn" | "notes" | "settings";
const icons = [Sprout, Layers3, Code2, Flag];
const editorExtensions = [go()];
const navigation: { id: View; label: string; icon: typeof Compass }[] = [
  { id: "home", label: "学习空间", icon: Compass },
  { id: "courses", label: "探索课程", icon: BookOpen },
  { id: "notes", label: "我的笔记", icon: FileText },
];

export function App() {
  const [data, setData] = useState<Bootstrap>();
  const [loadingError, setLoadingError] = useState("");
  const [view, setView] = useState<View>("home");
  const [lessonId, setLessonId] = useState("");
  const [assistantOpen, setAssistantOpen] = useState(false);
  const [assistantCode, setAssistantCode] = useState<string>();
  const [toast, setToast] = useState("");
  const [quit, setQuit] = useState(false);
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
  useEffect(() => {
    if (toast) {
      const timer = setTimeout(() => setToast(""), 4500);
      return () => clearTimeout(timer);
    }
  }, [toast]);
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
  if (quit)
    return (
      <div className="boot-screen">
        <Logo />
        <h1>GoStudy 已退出</h1>
        <p>本地学习引擎已经停止，现在可以关闭这个窗口。下次打开应用会重新启动引擎。</p>
      </div>
    );
  if (!data)
    return (
      <div className="boot-screen">
        <Logo />
        <h1>
          {loadingError ? "还差一步，连接本地引擎" : "正在打开你的学习空间"}
        </h1>
        <p>{loadingError || "准备课程、练习和属于你的进度。"}</p>
        {loadingError && (
          <>
            <p className="muted">
              请通过 GoStudy 启动时给出的完整链接打开，链接中包含本次运行的访问凭证。
            </p>
            <button className="button primary" onClick={load}>
              重新连接 <RotateCcw size={16} />
            </button>
          </>
        )}
        {!loadingError && <LoaderCircle className="spin" size={24} />}
      </div>
    );

  const completed = data.progress.completed || [];
  const percent = Math.round((completed.length / data.lessons.length) * 100) || 0;
  const next =
    data.lessons.find(
      (item) => item.id === data.progress.lastLesson && !completed.includes(item.id),
    ) ||
    data.lessons.find((item) => !completed.includes(item.id)) ||
    data.lessons[0];
  const currentLesson =
    data.lessons.find((item) => item.id === lessonId) || data.lessons[0];
  const currentCourse =
    data.courses.find((item) => item.id === currentLesson.courseId) || data.courses[0];
  const noteCount = Object.values(data.progress.notes || {}).filter(Boolean).length;
  const navigate = (id: View) => {
    setView(id);
    window.scrollTo(0, 0);
  };

  return (
    <div className="app">
      <header className="topbar">
        <div className="topbar-inner">
          <button
            className="brand"
            aria-label="GoStudy 首页"
            onClick={() => navigate("home")}
          >
            <Logo />
            <span>GoStudy</span>
          </button>
          <nav className="main-nav" aria-label="主导航">
            {navigation.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                className={`nav-item ${view === id ? "active" : ""}`}
                onClick={() => navigate(id)}
              >
                <Icon size={16} strokeWidth={2} />
                {label}
                {id === "notes" && <span className="nav-count">{noteCount}</span>}
              </button>
            ))}
          </nav>
          <div className="topbar-actions">
            <span className="offline-badge">
              <WifiOff size={13} />
              离线可用
            </span>
            <button className="assistant-trigger" onClick={() => setAssistantOpen(true)}>
              <Sparkles size={15} />
              AI 导师
            </button>
            <button
              className="icon-button settings-link"
              aria-label="偏好与 AI 设置"
              title="偏好与 AI 设置"
              onClick={() => navigate("settings")}
            >
              <Settings2 size={19} />
            </button>
            <div className="profile-avatar" title="本地学习者">
              G
            </div>
          </div>
        </div>
      </header>
      <main className="page">
        {view === "home" && (
          <div className="home">
            <section className="hero">
              <div className="hero-text">
                <span className="pill">
                  <Sparkles size={14} />
                  动手，才是真正的开始
                </span>
                <h1>
                  {completed.length ? (
                    <>
                      保持好奇，<span className="hl">继续</span>向前。
                    </>
                  ) : (
                    <>
                      从 Hello，到<span className="hl">新世界</span>。
                    </>
                  )}
                </h1>
                <p className="hero-lead">
                  一个概念，一次实践，一点新的理解。在阅读、修改、运行与追问之间，把 Go
                  变成自己的语言。
                </p>
                <div className="hero-actions">
                  <button
                    className="button primary large"
                    onClick={() => openLesson(next.id)}
                  >
                    {completed.length ? "继续我的学习" : "开始第一课"}
                    <ArrowRight size={18} />
                  </button>
                  <button
                    className="button secondary large"
                    onClick={() => navigate("courses")}
                  >
                    探索课程
                  </button>
                </div>
                <span className="hero-footnote">
                  <Clock3 size={14} />
                  每课约 8–16 分钟 <span>·</span> 无需安装 Go 环境 <span>·</span>{" "}
                  进度保存在本机
                </span>
              </div>
              <HeroArt />
            </section>
            <section className="stats-row">
              <div className="stat-card">
                <span className="stat-label">
                  <Compass size={16} />
                  学习进度
                </span>
                <strong className="stat-value">
                  {percent}
                  <em>%</em>
                </strong>
                <span className="stat-detail">
                  已完成 {completed.length} / {data.lessons.length} 个练习
                </span>
                <div className="progress-track">
                  <span style={{ width: `${percent}%` }} />
                </div>
              </div>
              <div className="stat-card">
                <span className="stat-label">
                  <Layers3 size={16} />
                  已掌握的章节
                </span>
                <strong className="stat-value">
                  {data.courses.filter((c) => c.lessonIds.every((id) => completed.includes(id))).length}
                  <em>/ {data.courses.length}</em>
                </strong>
                <span className="stat-detail">把知识连成自己的地图</span>
              </div>
              <div className="stat-card">
                <span className="stat-label">
                  <Code2 size={16} />
                  学习方式
                </span>
                <strong className="stat-value stat-phrase">
                  Learn by <em>doing</em>
                </strong>
                <span className="stat-detail">有引导的练习，即时反馈</span>
              </div>
            </section>
            <section className="section">
              <div className="section-head">
                <div>
                  <span className="eyebrow">YOUR LEARNING PATH</span>
                  <h2>循序渐进，也自由探索</h2>
                </div>
                <button className="text-button" onClick={() => navigate("courses")}>
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
                    onOpen={openLesson}
                  />
                ))}
              </div>
            </section>
            <section className="cta-note">
              <span className="cta-icon">
                <Sparkles size={20} />
              </span>
              <div>
                <strong>卡住的时候，不必独自面对。</strong>
                <p>让 AI 导师给你一个提示，或者换个角度解释。答案之外，理解更重要。</p>
              </div>
              <button className="text-button" onClick={() => setAssistantOpen(true)}>
                认识你的导师 <ArrowRight size={16} />
              </button>
            </section>
            <Footer />
          </div>
        )}
        {view === "courses" && (
          <div className="courses-page">
            <PageTitle
              eyebrow="A PATH TO UNDERSTANDING"
              title="探索 Go 的每一种可能。"
              description="从语言基础到工程思维。所有课程都可以自由开始，进度自动保存在本机。"
            />
            <div className="catalog">
              {data.courses.map((course, index) => {
                const Icon = icons[index % icons.length];
                const done = course.lessonIds.filter((id) => completed.includes(id)).length;
                return (
                  <section className="catalog-course" key={course.id}>
                    <div className="catalog-heading">
                      <div className={`course-icon tone-${index}`}>
                        <Icon size={24} />
                      </div>
                      <div>
                        <span className="eyebrow">
                          CHAPTER {String(index + 1).padStart(2, "0")} · {course.level}
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
                          <button
                            key={id}
                            className="lesson-row"
                            onClick={() => openLesson(id)}
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
                          </button>
                        );
                      })}
                    </div>
                  </section>
                );
              })}
            </div>
            <Footer />
          </div>
        )}
        {view === "learn" && (
          <LearningWorkspace
            key={currentLesson.id}
            lesson={currentLesson}
            course={currentCourse}
            data={data}
            onCodeChange={setAssistantCode}
            onProgress={updateProgress}
            onNext={openLesson}
            onCourses={() => navigate("courses")}
            onAI={() => setAssistantOpen(true)}
            onNotify={notify}
          />
        )}
        {view === "notes" && <Notes data={data} onOpen={openLesson} onNotify={notify} />}
        {view === "settings" && (
          <Preferences
            settings={data.settings}
            onSave={updateSettings}
            onNotify={notify}
            runtime={data.runtime}
            onQuit={quitApp}
          />
        )}
      </main>
      {assistantOpen && (
        <Assistant
          key={view === "learn" ? currentLesson.id : next.id}
          lesson={view === "learn" ? currentLesson : next}
          code={
            view === "learn"
              ? assistantCode ?? data.progress.drafts?.[currentLesson.id]
              : data.progress.drafts?.[next.id]
          }
          settings={data.settings}
          onClose={() => setAssistantOpen(false)}
          onSettings={() => {
            navigate("settings");
            setAssistantOpen(false);
          }}
        />
      )}
      {toast && (
        <div className="toast" role="status">
          <CheckCheck size={17} />
          {toast}
          <button className="icon-button" aria-label="关闭通知" onClick={() => setToast("")}>
            <X size={14} />
          </button>
        </div>
      )}
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

function Footer() {
  return (
    <footer>
      <span className="footer-brand">
        <Logo />
        GoStudy
      </span>
      <span className="footer-motto">为理解而学，为创造而写。</span>
      <span className="footer-local">
        <ShieldCheck size={13} />
        Local first. Always yours.
      </span>
    </footer>
  );
}

function PageTitle({
  eyebrow,
  title,
  description,
}: {
  eyebrow: string;
  title: string;
  description: string;
}) {
  return (
    <div className="page-heading">
      <span className="eyebrow">{eyebrow}</span>
      <h1>{title}</h1>
      <p>{description}</p>
    </div>
  );
}

function HeroArt() {
  return (
    <div className="hero-art" aria-hidden="true">
      <svg viewBox="0 0 520 400" fill="none">
        <defs>
          <pattern id="hexgrid" width="60" height="52" patternUnits="userSpaceOnUse">
            <path
              d="M30 0 60 15v26L30 52 0 41V15z"
              stroke="#dfe5f2"
              strokeWidth="1"
              fill="none"
            />
          </pattern>
        </defs>
        <rect x="24" y="18" width="472" height="364" rx="12" fill="url(#hexgrid)" opacity="0.85" />
        <g opacity="0.95">
          <rect
            x="70"
            y="66"
            width="26"
            height="26"
            rx="6"
            fill="#f69605"
            transform="rotate(-14 83 79)"
          />
          <path d="M440 74l20 12-20 12z" fill="#604fcd" />
          <circle cx="86" cy="318" r="7" fill="#0fa3b1" />
          <rect
            x="420"
            y="300"
            width="22"
            height="22"
            rx="6"
            fill="#2e57e8"
            transform="rotate(18 431 311)"
          />
          <path d="M392 60h18M401 51v18" stroke="#130b43" strokeWidth="2.4" strokeLinecap="round" />
          <circle cx="120" cy="150" r="3.4" fill="#130b43" />
          <rect
            x="416"
            y="196"
            width="14"
            height="14"
            rx="4"
            fill="#f69605"
            transform="rotate(24 423 203)"
          />
        </g>
        <g transform="translate(150 96)">
          <rect width="228" height="176" rx="12" fill="#ffffff" stroke="#d9dfee" />
          <path d="M0 40h228" stroke="#e7ebf6" />
          <circle cx="22" cy="20" r="4" fill="#604fcd" />
          <circle cx="38" cy="20" r="4" fill="#2e57e8" />
          <circle cx="54" cy="20" r="4" fill="#0fa3b1" />
          <rect x="22" y="62" width="34" height="9" rx="4.5" fill="#604fcd" />
          <rect x="64" y="62" width="72" height="9" rx="4.5" fill="#dfe5f2" />
          <rect x="22" y="86" width="118" height="9" rx="4.5" fill="#e7ebf6" />
          <rect x="22" y="110" width="86" height="9" rx="4.5" fill="#e7ebf6" />
          <rect x="22" y="134" width="140" height="9" rx="4.5" fill="#f69605" opacity="0.85" />
        </g>
        <g transform="translate(352 232)">
          <path
            d="M52 0 104 30v60L52 120 0 90V30z"
            fill="#f2f0fc"
            stroke="#604fcd"
            strokeWidth="2.5"
            strokeLinejoin="round"
          />
          <path
            d="m34 60 13 13 24-27"
            stroke="#604fcd"
            strokeWidth="6"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </g>
        <g transform="translate(96 226)">
          <path
            d="M38 0 76 22v44L38 88 0 66V22z"
            fill="#ffffff"
            stroke="#0fa3b1"
            strokeWidth="2.5"
            strokeLinejoin="round"
          />
          <path d="M20 40h36M38 22v36" stroke="#0fa3b1" strokeWidth="4" strokeLinecap="round" />
        </g>
        <path
          d="M268 320c22 10 44 10 66 0"
          stroke="#c9d2e8"
          strokeWidth="3"
          strokeLinecap="round"
          strokeDasharray="2 9"
        />
      </svg>
    </div>
  );
}

function CourseCard({
  course,
  index,
  completed,
  onOpen,
}: {
  course: Course;
  index: number;
  completed: string[];
  onOpen: (id: string) => void;
}) {
  const Icon = icons[index % icons.length];
  const done = course.lessonIds.filter((id) => completed.includes(id)).length;
  return (
    <button
      className="course-card"
      onClick={() =>
        onOpen(course.lessonIds.find((id) => !completed.includes(id)) || course.lessonIds[0])
      }
    >
      <div className="course-card-top">
        <div className={`course-icon tone-${index}`}>
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
          {done ? `已完成 ${done} / ${course.lessonIds.length}` : "开始探索"}
          <ArrowRight size={15} />
        </span>
      </div>
      {done > 0 && (
        <div
          className="card-progress"
          style={{ width: `${(done / course.lessonIds.length) * 100}%` }}
        />
      )}
    </button>
  );
}

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
    try {
      const value = await api<RunResult>("run", { lessonId: lesson.id, code });
      setResult(value);
      setTab("task");
    } catch (err) {
      setError((err as Error).message);
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
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setCompleting(false);
    }
  };
  const index = data.lessons.findIndex((item) => item.id === lesson.id);
  const next = data.lessons[index + 1];
  const done = data.progress.completed?.includes(lesson.id);
  const completedCount = data.progress.completed?.length ?? 0;
  return (
    <div className="lesson-page">
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
            <span>{done ? <><Check size={14} />已掌握</> : "进行中"}</span>
            <span className="meta-divider" />
            <div className="lesson-progress">
              <span className="progress-chip">
                已完成 {completedCount} / {data.lessons.length} 个练习
              </span>
              <div className="progress-track slim">
                <span style={{ width: `${Math.round((completedCount / data.lessons.length) * 100)}%` }} />
              </div>
            </div>
          </div>
          {lesson.source && (
            <p className="lesson-source">
              <BookOpen size={14} />
              内容依据仓库资料：{lesson.source}
            </p>
          )}
        </div>
        <button className="button secondary" onClick={onAI}>
          <Sparkles size={16} />
          和导师一起想
        </button>
      </header>
      <div className="lesson-grid">
        <section className="lesson-card">
          <div className="tabs" role="tablist" aria-label="课程内容">
            {(
              [
                { id: "concept", name: "理解概念" },
                { id: "task", name: "动手挑战" },
                { id: "notes", name: "我的笔记" },
              ] as const
            ).map((item) => (
              <button
                key={item.id}
                role="tab"
                aria-selected={tab === item.id}
                className={tab === item.id ? "selected" : ""}
                onClick={() => setTab(item.id)}
              >
                {item.name}
              </button>
            ))}
          </div>
          <div className="lesson-body" role="tabpanel">
            {tab === "concept" && (
              <>
                <span className="eyebrow">01 / 先理解，再动手</span>
                <div className="markdown">
                  <ReactMarkdown>{lesson.concept}</ReactMarkdown>
                </div>
                <details className="reference-example">
                  <summary>
                    看看一个完整示例 <ChevronDown size={15} />
                  </summary>
                  <div className="markdown">
                    <ReactMarkdown>{"```go\n" + lesson.example + "\n```"}</ReactMarkdown>
                  </div>
                </details>
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
                <span className="eyebrow">02 / 把理解写成代码</span>
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
                  <div className="hint-box" key={i}>
                    <span>提示 {i + 1}</span>
                    {text}
                  </div>
                ))}
                <div className="quiz">
                  <span className="eyebrow">03 / 确认你真的懂了</span>
                  <h3>{lesson.quiz.question}</h3>
                  <div role="radiogroup" aria-label="理解题">
                    {lesson.quiz.options.map((option, i) => (
                      <button
                        role="radio"
                        aria-checked={answer === i}
                        key={option}
                        className={`quiz-option ${
                          answer === i ? (i === lesson.quiz.answer ? "correct" : "incorrect") : ""
                        }`}
                        onClick={() => setAnswer(i)}
                      >
                        <span className="quiz-letter">{String.fromCharCode(65 + i)}</span>
                        {option}
                        {answer === i && i === lesson.quiz.answer && <Check size={16} />}
                      </button>
                    ))}
                  </div>
                  {answer !== undefined && (
                    <p className="quiz-explanation">
                      {answer === lesson.quiz.answer ? "理解正确。" : "再想一想。"}
                      {lesson.quiz.explanation}
                    </p>
                  )}
                </div>
                <button
                  className="button primary complete-button"
                  disabled={!result?.passed || answer !== lesson.quiz.answer || completing}
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
                  <div className="completion-banner">
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
                  </div>
                )}
              </>
            )}
            {tab === "notes" && (
              <>
                <span className="eyebrow">留下一点自己的理解</span>
                <h2>写给未来的自己</h2>
                <p className="muted">用自己的话解释，比记住答案更有用。</p>
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
            <button
              className="icon-button"
              aria-label="还原初始代码"
              title="还原初始代码"
              disabled={running}
              onClick={() => {
                setCode(lesson.starter);
                setResult(undefined);
              }}
            >
              <RotateCcw size={15} />
            </button>
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
            <button className="button primary run-button" disabled={running} onClick={run}>
              {running ? (
                <LoaderCircle className="spin" size={15} />
              ) : (
                <Play size={14} fill="currentColor" />
              )}
              {running ? "运行中" : "运行代码"}
            </button>
          </div>
          <div className="console-header">
            <span>
              <Terminal size={15} />
              运行结果
            </span>
            {result && (
              <span className={result.passed ? "result-success" : "muted"}>
                {result.passed ? "✓ 挑战通过" : result.error ? "需要调整" : "继续试试"} ·{" "}
                {result.durationMs} ms
              </span>
            )}
          </div>
          <div
            className={`console ${result?.error || error ? "console-error" : ""}`}
            aria-live="polite"
          >
            {running ? (
              <span className="muted">正在本地执行你的代码…</span>
            ) : error ? (
              <pre>{error}</pre>
            ) : result ? (
              <>
                <pre>
                  {result.output}
                  {result.error && `\n${result.error}`}
                </pre>
                {!result.passed && !result.error && (
                  <p className="console-tip">运行成功。再对照左侧期望输出调整一下。</p>
                )}
              </>
            ) : (
              <div className="console-empty">
                <Terminal size={24} />
                <p>想法，等你来运行。</p>
                <span>输出和错误信息会显示在这里。</span>
              </div>
            )}
          </div>
          <div className="editor-safety">
            <ShieldCheck size={13} />
            受限 Go 运行环境 · 不开放文件与网络访问
          </div>
        </section>
      </div>
    </div>
  );
}

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

function boundedText(text: string, bytes: number) {
  const encoder = new TextEncoder();
  let result = "";
  let size = 0;
  for (const character of text) {
    size += encoder.encode(character).length;
    if (size > bytes) break;
    result += character;
  }
  return result;
}

function Assistant({
  lesson,
  code,
  settings,
  onClose,
  onSettings,
}: {
  lesson: Lesson;
  code?: string;
  settings: Settings;
  onClose: () => void;
  onSettings: () => void;
}) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const abort = useRef<AbortController | undefined>(undefined);
  const bottom = useRef<HTMLDivElement>(null);
  useEffect(() => {
    bottom.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, sending]);
  useEffect(() => () => abort.current?.abort(), []);
  const send = async (message = input) => {
    if (!message.trim() || sending || !modelReady(settings)) return;
    setInput("");
    setError("");
    setSending(true);
    const history = messages
      .slice(-8)
      .map((item) => ({ ...item, content: boundedText(item.content, 1800) }));
    setMessages((current) => [...current, { role: "user", content: message }]);
    abort.current = new AbortController();
    try {
      const response = await api<{ reply: string }>(
        "chat",
        {
          lessonId: lesson.id,
          code: code ?? lesson.starter,
          message: boundedText(message, 4096),
          history,
        },
        abort.current.signal,
      );
      setMessages((current) => [...current, { role: "assistant", content: response.reply }]);
    } catch (err) {
      if ((err as Error).name !== "AbortError") setError((err as Error).message);
    } finally {
      setSending(false);
    }
  };
  return (
    <div className="assistant-overlay">
      <button className="assistant-scrim" aria-label="关闭 AI 导师" onClick={onClose} />
      <aside className="assistant-panel" role="dialog" aria-modal="true" aria-label="AI 导师">
        <div className="assistant-header">
          <div className="ai-avatar">
            <Sparkles size={20} />
          </div>
          <div>
            <h2>一起想，慢慢懂。</h2>
            <span>{modelReady(settings) ? settings.model : "你的个人 AI 导师"}</span>
          </div>
          <button className="icon-button" aria-label="关闭 AI 导师" onClick={onClose}>
            <X size={20} />
          </button>
        </div>
        <div className="assistant-context">
          <BookOpen size={14} />
          当前上下文<span>{lesson.title}</span>
        </div>
        <div className="chat-messages">
          {messages.length === 0 && (
            <div className="assistant-welcome">
              <span className="welcome-mark">
                <Sparkles size={26} />
              </span>
              <h3>理解，往往始于一个好问题。</h3>
              <p>
                我可以解释概念、给你一点线索，
                <br />
                也可以陪你一起读懂错误信息。
              </p>
              {!modelReady(settings) ? (
                <div className="ai-setup">
                  <span>先连接你选择的模型</span>
                  <p>使用自己的 API Key。课程可以离线学习，AI 对话需要连接模型服务。</p>
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
                  ].map((text) => (
                    <button key={text} onClick={() => send(text)}>
                      {text}
                      <Plus size={14} />
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
          {messages.map((message, i) => (
            <div key={i} className={`chat-message ${message.role}`}>
              <span className="message-label">
                {message.role === "user" ? "你" : "AI 导师"}
              </span>
              <div className="markdown">
                <ReactMarkdown>{message.content}</ReactMarkdown>
              </div>
            </div>
          ))}
          {sending && (
            <div className="thinking">
              <LoaderCircle size={15} className="spin" />
              正在换个角度想一想…
            </div>
          )}
          {error && (
            <div className="error-box" role="alert">
              {error}
              <button
                className="text-button"
                onClick={() => send(messages.filter((m) => m.role === "user").at(-1)?.content)}
              >
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
            void send();
          }}
        >
          <textarea
            aria-label="向 AI 导师提问"
            placeholder={modelReady(settings) ? "把你的问题说出来…" : "配置模型后，开始你的第一次对话"}
            value={input}
            onChange={(event) => setInput(boundedText(event.target.value, 4096))}
            disabled={!modelReady(settings)}
            maxLength={2000}
            onKeyDown={(event) => {
              if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing) {
                event.preventDefault();
                void send();
              }
            }}
          />
          <div>
            <span>Enter 发送 · Shift + Enter 换行</span>
            <button
              className="send-button"
              type="submit"
              disabled={sending || !input.trim() || !modelReady(settings)}
              aria-label="发送消息"
            >
              <Send size={17} />
            </button>
          </div>
        </form>
        <p className="ai-disclaimer">
          <ShieldCheck size={12} />
          仅发送当前课题、代码、课程笔记与本轮对话。AI 回复请运行验证。
        </p>
      </aside>
    </div>
  );
}

function Preferences({
  settings,
  onSave,
  onNotify,
  runtime,
  onQuit,
}: {
  settings: Settings;
  onSave: (s: Settings) => void;
  onNotify: (s: string) => void;
  runtime: Bootstrap["runtime"];
  onQuit: () => void;
}) {
  const [baseUrl, setBaseUrl] = useState(settings.baseUrl || "https://api.openai.com/v1");
  const [model, setModel] = useState(settings.model || "gpt-4o-mini");
  const [key, setKey] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState("");
  const desktop = !/android|ios/i.test(runtime.platform);
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
    <div className="settings-page">
      <PageTitle
        eyebrow="MAKE IT YOURS"
        title="你的空间，你的选择。"
        description="学习留在本地，模型由你决定。让工具适应你的节奏。"
      />
      <section className="settings-card">
        <div className="settings-card-heading">
          <div className="course-icon tone-0">
            <Sparkles size={22} />
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
            <div className="test-result">
              <Check size={16} />
              <span>{testResult}</span>
            </div>
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
      </section>
      <section className="settings-card">
        <div className="settings-card-heading">
          <div className="course-icon tone-2">
            <ShieldCheck size={22} />
          </div>
          <div>
            <h2>本地优先，安心探索</h2>
            <p>没有账号，没有学习数据上传。你的进度属于你。</p>
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
        </div>
        {desktop && (
          <div className="quit-row">
            <div>
              <strong>退出 GoStudy</strong>
              <p>停止本机学习引擎。关闭浏览器窗口不会自动退出应用。</p>
            </div>
            <button className="button secondary" onClick={() => void onQuit()}>
              <Power size={16} />
              退出 GoStudy
            </button>
          </div>
        )}
      </section>
      <Footer />
    </div>
  );
}

function Notes({
  data,
  onOpen,
  onNotify,
}: {
  data: Bootstrap;
  onOpen: (id: string) => void;
  onNotify: (s: string) => void;
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
    <div className="notes-page">
      {exportContent && (
        <div className="export-dialog" role="dialog" aria-modal="true" aria-label="导出学习笔记">
          <div className="export-card">
            <div className="section-head">
              <h2>带走你的理解</h2>
              <button
                className="icon-button"
                aria-label="关闭导出"
                onClick={() => setExportContent("")}
              >
                <X size={20} />
              </button>
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
          </div>
        </div>
      )}
      <div className="notes-heading">
        <PageTitle
          eyebrow="THOUGHTS WORTH KEEPING"
          title="把知识，写成自己的理解。"
          description="每一课留下的思考，都在这里。回看时，你会发现自己走了很远。"
        />
        <button className="button secondary" disabled={!notes.length} onClick={exportNotes}>
          <ArrowDownToLine size={16} />
          导出笔记
        </button>
      </div>
      {notes.length ? (
        <div className="notes-grid">
          {notes.map((lesson) => (
            <article className="note-card" key={lesson.id}>
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
            </article>
          ))}
        </div>
      ) : (
        <div className="empty-state">
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
        </div>
      )}
      <Footer />
    </div>
  );
}

const rootElement = document.getElementById("root");
if (rootElement) createRoot(rootElement).render(<App />);
