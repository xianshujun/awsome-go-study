import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { motion } from "motion/react";
import ReactMarkdown from "react-markdown";
import {
  Activity,
  AlertTriangle,
  BrainCircuit,
  Check,
  CheckCircle2,
  ChevronRight,
  Clock3,
  Flag,
  Gauge,
  Lightbulb,
  LoaderCircle,
  Play,
  RotateCcw,
  Send,
  Server,
  ShieldAlert,
  Sparkles,
  Terminal,
  X,
  XCircle,
} from "lucide-react";
import { api } from "./api";
import type {
  LabBootstrap,
  LabCase,
  LabCaseSummary,
  LabChatResponse,
  LabChatStep,
  LabHintResult,
  LabResult,
  LabRunResult,
  LabService,
  LabSubmitResult,
  Settings,
} from "./api";
import { rise } from "./motion";

type Line = { kind: "command" | "output" | "error" | "note"; text: string };
type ChatLine = { role: "user" | "assistant"; content: string; steps?: LabChatStep[]; error?: boolean };

const statusMeta: Record<LabService["status"], { label: string; className: string }> = {
  healthy: { label: "正常", className: "ok" },
  degraded: { label: "降级", className: "warn" },
  critical: { label: "严重", className: "bad" },
};

function LabMarkdown({ children }: { children: string }) {
  return (
    <ReactMarkdown
      components={{
        a: ({ href, children: label }) => (
          <a href={href} target="_blank" rel="noopener noreferrer">
            {label}
          </a>
        ),
      }}
    >
      {children}
    </ReactMarkdown>
  );
}

function formatMetric(value: number): string {
  if (Math.abs(value) >= 10000) {
    const scaled = value / 10000;
    return `${Number.isInteger(scaled) ? scaled : scaled.toFixed(1)} 万`;
  }
  if (Number.isInteger(value)) return String(value);
  if (Math.abs(value) >= 10) return value.toFixed(1);
  return value.toFixed(2);
}

function Sparkline({ points, alert }: { points: number[]; alert?: number }) {
  const width = 128;
  const height = 36;
  const padding = 3;
  const minimum = Math.min(...points);
  const maximum = Math.max(...points);
  const span = maximum - minimum || 1;
  const coordinates = points.map((point, index) => {
    const x = padding + (index / Math.max(1, points.length - 1)) * (width - padding * 2);
    const y = height - padding - ((point - minimum) / span) * (height - padding * 2);
    return `${x.toFixed(1)},${y.toFixed(1)}`;
  });
  const last = points[points.length - 1];
  const over = Boolean(alert && alert > 0 && last > alert);
  const [lastX, lastY] = coordinates[coordinates.length - 1].split(",");
  return (
    <svg className="lab-spark" viewBox={`0 0 ${width} ${height}`} role="img" aria-label="指标趋势">
      <polyline
        points={coordinates.join(" ")}
        fill="none"
        stroke={over ? "var(--red)" : "var(--primary)"}
        strokeWidth="1.8"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
      <circle cx={lastX} cy={lastY} r="2.6" fill={over ? "var(--red)" : "var(--primary)"} />
    </svg>
  );
}

function welcome(caseView: LabCase): Line {
  return {
    kind: "note",
    text: `${caseView.company} · ${caseView.business || "线上事故演练"}\n演练时钟 ${caseView.clock}\n告警已经加载。输入 help 查看命令，或点下面的快捷命令开始排查。`,
  };
}

export function LabPage({
  settings,
  motionOn,
  onOpenSettings,
  onNotify,
}: {
  settings: Settings;
  motionOn: boolean;
  onOpenSettings: () => void;
  onNotify: (text: string) => void;
}) {
  const [boot, setBoot] = useState<LabBootstrap>();
  const [bootError, setBootError] = useState("");
  const [caseView, setCaseView] = useState<LabCase>();
  const [caseError, setCaseError] = useState("");
  const [opening, setOpening] = useState(false);
  const [lines, setLines] = useState<Line[]>([]);
  const [input, setInput] = useState("");
  const [history, setHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [running, setRunning] = useState(false);
  const [hints, setHints] = useState<string[]>([]);
  const [hintBusy, setHintBusy] = useState(false);
  const [root, setRoot] = useState("");
  const [fix, setFix] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [verdict, setVerdict] = useState<LabSubmitResult>();
  const [result, setResult] = useState<LabResult>();
  const [aiOpen, setAiOpen] = useState(false);
  const [messages, setMessages] = useState<ChatLine[]>([]);
  const [aiInput, setAiInput] = useState("");
  const [aiBusy, setAiBusy] = useState(false);
  const request = useRef(0);
  const terminalEnd = useRef<HTMLDivElement>(null);
  const chatEnd = useRef<HTMLDivElement>(null);
  const verdictRef = useRef<HTMLElement>(null);

  const modelReady = Boolean(settings.baseUrl && settings.model);

  const openSummary = useCallback(async (summary: LabCaseSummary) => {
    const token = ++request.current;
    setOpening(true);
    setCaseError("");
    try {
      const view = await api<LabCase>("lab", {
        action: "case",
        scenario: summary.scenario,
        date: summary.date,
      });
      if (token !== request.current) return;
      setCaseView(view);
      setHints(view.hints ?? []);
      setResult(view.result);
      setVerdict(undefined);
      setRoot("");
      setFix("");
      setMessages([]);
      setLines([welcome(view)]);
      setOpening(false);
    } catch (err) {
      if (token !== request.current) return;
      setCaseError((err as Error).message);
      setOpening(false);
    }
  }, []);

  const refreshBoot = useCallback(async () => {
    try {
      const value = await api<LabBootstrap>("lab", { action: "bootstrap" });
      setBoot(value);
    } catch {
      // 首页数据刷新失败不影响当前演练。
    }
  }, []);

  useEffect(() => {
    let alive = true;
    api<LabBootstrap>("lab", { action: "bootstrap" })
      .then((value) => {
        if (!alive) return;
        setBoot(value);
        void openSummary(value.today);
      })
      .catch((err) => {
        if (alive) setBootError((err as Error).message);
      });
    return () => {
      alive = false;
    };
  }, [openSummary]);

  useEffect(() => {
    // 只滚终端内部，别把整页带到页面底部（手机上打开演练会直接跳到指标区）。
    const body = terminalEnd.current?.closest<HTMLElement>(".lab-terminal-body");
    if (body) body.scrollTop = body.scrollHeight;
  }, [lines, running]);

  useEffect(() => {
    const log = chatEnd.current?.closest<HTMLElement>(".lab-ai-log");
    if (log) log.scrollTop = log.scrollHeight;
  }, [messages, aiBusy]);

  useEffect(() => {
    if (verdict) verdictRef.current?.scrollIntoView?.({ behavior: "smooth", block: "start" });
  }, [verdict]);

  const runCommand = useCallback(
    async (raw?: string) => {
      const command = (raw ?? input).trim();
      if (!command || !caseView || running) return;
      setInput("");
      setHistory((current) => [...current.filter((item) => item !== command), command].slice(-30));
      setHistoryIndex(-1);
      setLines((current) => [...current, { kind: "command", text: command }]);
      setRunning(true);
      try {
        const value = await api<LabRunResult>("lab", {
          action: "run",
          caseId: caseView.id,
          command,
        });
        setLines((current) => [...current, { kind: "output", text: value.output }]);
      } catch (err) {
        setLines((current) => [...current, { kind: "error", text: (err as Error).message }]);
      } finally {
        setRunning(false);
      }
    },
    [caseView, input, running],
  );

  const revealHint = async () => {
    if (!caseView || hintBusy) return;
    setHintBusy(true);
    try {
      const value = await api<LabHintResult>("lab", { action: "hint", caseId: caseView.id });
      setHints((current) => [...current, value.hint]);
    } catch (err) {
      setCaseError((err as Error).message);
    } finally {
      setHintBusy(false);
    }
  };

  const submit = async () => {
    if (!caseView || submitting) return;
    if (root === "" || fix === "") {
      setCaseError("请先选择根因与处置方案，再提交诊断。");
      return;
    }
    setSubmitting(true);
    setCaseError("");
    try {
      const value = await api<LabSubmitResult>("lab", {
        action: "submit",
        caseId: caseView.id,
        rootCause: Number(root),
        fix: Number(fix),
      });
      setVerdict(value);
      setResult(value.result);
      if (value.passed) onNotify(`事故复盘通过 · ${value.score} 分`);
      void refreshBoot();
    } catch (err) {
      setCaseError((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  };

  const reset = async () => {
    if (!caseView) return;
    try {
      await api("lab", { action: "reset", caseId: caseView.id });
      setLines([welcome(caseView)]);
      setHints([]);
      setRoot("");
      setFix("");
      setVerdict(undefined);
      setCaseError("");
    } catch (err) {
      setCaseError((err as Error).message);
    }
  };

  const ask = async (question: string) => {
    const text = question.trim();
    if (!text || !caseView || aiBusy) return;
    if (!modelReady) {
      setMessages((current) => [
        ...current,
        { role: "user", content: text },
        {
          role: "assistant",
          error: true,
          content: "还没有配置 AI 服务。先在「偏好与 AI 设置」里填写基础地址和模型，再回来问我。",
        },
      ]);
      setAiInput("");
      return;
    }
    const historyPayload = messages
      .filter((item) => !item.error)
      .slice(-10)
      .map((item) => ({ role: item.role, content: item.content }));
    setMessages((current) => [...current, { role: "user", content: text }]);
    setAiInput("");
    setAiBusy(true);
    try {
      const value = await api<LabChatResponse>("lab-chat", {
        caseId: caseView.id,
        message: text,
        history: historyPayload,
      });
      setMessages((current) => [
        ...current,
        { role: "assistant", content: value.reply, steps: value.steps },
      ]);
    } catch (err) {
      setMessages((current) => [
        ...current,
        { role: "assistant", error: true, content: `查证失败：${(err as Error).message}` },
      ]);
    } finally {
      setAiBusy(false);
    }
  };

  const solved = boot ? Object.values(boot.progress.results ?? {}).filter((item) => item.passed).length : 0;
  const attempts = boot ? Object.values(boot.progress.results ?? {}).reduce((sum, item) => sum + item.attempts, 0) : 0;

  const days = useMemo(() => {
    if (!boot) return [];
    return [boot.today, ...boot.archive];
  }, [boot]);

  if (bootError) {
    return (
      <div className="lab-page">
        <div className="lab-loading">
          <AlertTriangle size={26} />
          <strong>演练环境没有拉起来</strong>
          <span>{bootError}</span>
          <button className="button secondary" onClick={() => window.location.reload()}>
            重新加载
          </button>
        </div>
      </div>
    );
  }

  if (!boot || !caseView) {
    if (caseError) {
      return (
        <div className="lab-page">
          <div className="lab-loading">
            <AlertTriangle size={26} />
            <strong>这场事故没有打开</strong>
            <span>{caseError}</span>
            <button className="button secondary" onClick={() => window.location.reload()}>
              重新加载
            </button>
          </div>
        </div>
      );
    }
    return (
      <div className="lab-page">
        <div className="lab-loading">
          <LoaderCircle size={26} className="spin" />
          <strong>正在拉起演练环境</strong>
          <span>准备今日事故、服务拓扑、日志与指标。</span>
        </div>
      </div>
    );
  }

  return (
    <motion.div className="lab-page" {...(motionOn ? { initial: { opacity: 0, y: 10 }, animate: { opacity: 1, y: 0 } } : {})}>
      <header className="lab-top">
        <div>
          <span className="eyebrow">高级模式 · 干中学</span>
          <h1>线上事故演练场</h1>
          <p>{boot.tagline}</p>
        </div>
        <div className="lab-stats">
          <div>
            <strong>{solved}</strong>
            <span>已复盘</span>
          </div>
          <div>
            <strong>{boot.scenarioCount}</strong>
            <span>个场景</span>
          </div>
          <div>
            <strong>{attempts}</strong>
            <span>次提交</span>
          </div>
        </div>
      </header>

      <div className="lab-days" aria-label="演练日期">
        {days.map((day) => {
          const active = day.scenario === caseView.scenario && day.date === caseView.date;
          return (
            <button
              key={day.id}
              className={`lab-day ${active ? "active" : ""} ${day.solved ? "solved" : ""}`}
              onClick={() => void openSummary(day)}
              disabled={opening}
            >
              <span>{day.isToday ? "今日" : day.date.slice(5)}</span>
              <strong>{day.title}</strong>
              {day.best ? <em>{day.best} 分</em> : null}
            </button>
          );
        })}
      </div>

      <motion.section className="lab-hero" {...rise(motionOn, 0)}>
        <div className="lab-hero-main">
          <div className="lab-badges">
            <span className="lab-badge">{caseView.date}</span>
            <span className="lab-badge">{caseView.level}</span>
            <span className="lab-badge">
              <Clock3 size={12} /> {caseView.minutes} 分钟
            </span>
            {caseView.tags.map((tag) => (
              <span className="lab-tag" key={tag}>
                {tag}
              </span>
            ))}
          </div>
          <h2>{caseView.title}</h2>
          {caseView.subtitle && <p className="lab-subtitle">{caseView.subtitle}</p>}
          <p className="lab-brief">{caseView.brief}</p>
          <div className="lab-mission">
            <Flag size={16} />
            <strong>本轮任务</strong>
            <span>{caseView.mission}</span>
          </div>
          <div className="lab-hero-actions">
            <button className={`button ${aiOpen ? "secondary" : "primary"}`} onClick={() => setAiOpen((open) => !open)}>
              <BrainCircuit size={17} />
              {aiOpen ? "收起 AI 值班助手" : "让 AI 陪我查"}
            </button>
            <button className="button secondary" onClick={() => void runCommand("trace --slow")} disabled={running}>
              <Play size={16} />
              先看最慢链路
            </button>
          </div>
        </div>
        <aside className="lab-alert-box">
          <h3>
            <ShieldAlert size={17} />
            正在告警
          </h3>
          <ul>
            {caseView.alerts.map((alert) => (
              <li key={alert}>
                <span className="lab-alert-dot" />
                {alert}
              </li>
            ))}
          </ul>
          <div className="lab-alert-guide">
            <strong>值班顺序</strong>
            <ol>
              <li>先把告警和变更时间线对齐</li>
              <li>看服务状态，找出最先异常的那一跳</li>
              <li>用日志、指标、链路交叉验证，不靠猜</li>
              <li>先止血，再修根因；变更前评估影响面</li>
            </ol>
          </div>
          {caseView.timeline && caseView.timeline.length > 0 && (
            <details className="lab-timeline">
              <summary>变更与告警时间线</summary>
              <ol>
                {caseView.timeline.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ol>
            </details>
          )}
          <div className="lab-alert-foot">
            {caseView.services.length} 个服务 · {caseView.metrics.length} 组指标 · {caseView.hintCount} 条提示
          </div>
        </aside>
      </motion.section>

      <div className="lab-layout">
        <div className="lab-main">
          <motion.section className="lab-terminal" {...rise(motionOn, 1)}>
            <div className="lab-terminal-head">
              <span className="lab-terminal-title">
                <Terminal size={16} />
                演练终端
              </span>
              <span className="lab-terminal-note">只读 · 模拟环境</span>
              <button
                className="text-button"
                onClick={() => setLines([welcome(caseView)])}
                disabled={running}
              >
                清屏
              </button>
            </div>
            <div className="lab-terminal-body" aria-live="polite">
              {lines.map((line, index) =>
                line.kind === "command" ? (
                  <div className="lab-line command" key={`${index}-${line.text}`}>
                    <span className="lab-prompt">sre@orange-mall:~$</span>
                    <code>{line.text}</code>
                  </div>
                ) : (
                  <pre className={`lab-output ${line.kind}`} key={`${index}-${line.text}`}>
                    {line.text}
                  </pre>
                ),
              )}
              {running && (
                <div className="lab-line command">
                  <LoaderCircle size={14} className="spin" />
                  <span className="lab-terminal-note">正在查询演练环境…</span>
                </div>
              )}
              <div ref={terminalEnd} />
            </div>
            <div className="lab-presets">
              {caseView.presets.map((preset) => (
                <button key={preset} className="quick-chip" onClick={() => void runCommand(preset)} disabled={running}>
                  {preset}
                </button>
              ))}
            </div>
            <form
              className="lab-input-row"
              onSubmit={(event) => {
                event.preventDefault();
                void runCommand();
              }}
            >
              <span className="lab-prompt">sre@orange-mall:~$</span>
              <input
                value={input}
                onChange={(event) => setInput(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "ArrowUp") {
                    event.preventDefault();
                    if (!history.length) return;
                    const next = Math.min(history.length - 1, historyIndex + 1);
                    setHistoryIndex(next);
                    setInput(history[history.length - 1 - next]);
                  }
                  if (event.key === "ArrowDown") {
                    event.preventDefault();
                    if (historyIndex <= 0) {
                      setHistoryIndex(-1);
                      setInput("");
                      return;
                    }
                    const next = historyIndex - 1;
                    setHistoryIndex(next);
                    setInput(history[history.length - 1 - next]);
                  }
                }}
                placeholder="输入命令，例如 logs order-api --level=ERROR --since=30m"
                aria-label="演练命令"
                spellCheck={false}
                autoComplete="off"
                disabled={running}
              />
              <button className="button primary" type="submit" disabled={running || !input.trim()}>
                执行
              </button>
            </form>
          </motion.section>

          <motion.section className="lab-card" {...rise(motionOn, 2)}>
            <div className="lab-card-head">
              <Server size={18} />
              <h3>服务状态</h3>
              <span className="lab-card-note">入口视角的聚合，不代表依赖没问题</span>
            </div>
            <div className="lab-service-grid">
              {caseView.services.map((service) => {
                const meta = statusMeta[service.status];
                return (
                  <div className={`lab-service ${meta.className}`} key={service.name}>
                    <div className="lab-service-name">
                      <span className="lab-status-dot" />
                      <strong>{service.name}</strong>
                      <span>{meta.label}</span>
                    </div>
                    <p>{service.role}</p>
                    <div className="lab-service-meta">
                      <span>副本 {service.replicas || "-"}</span>
                      <span>CPU {service.cpu || "-"}</span>
                      <span>内存 {service.memory || "-"}</span>
                    </div>
                    {service.note && <span className="lab-service-note">{service.note}</span>}
                  </div>
                );
              })}
            </div>
          </motion.section>

          <motion.section className="lab-card" {...rise(motionOn, 3)}>
            <div className="lab-card-head">
              <Activity size={18} />
              <h3>关键指标</h3>
              <span className="lab-card-note">最近一小时的采样，越靠右越接近告警时刻</span>
            </div>
            <div className="lab-metric-grid">
              {caseView.metrics.map((metric) => {
                const last = metric.points[metric.points.length - 1];
                const over = Boolean(metric.alert && metric.alert > 0 && last > metric.alert);
                return (
                  <div className={`lab-metric ${over ? "over" : ""}`} key={`${metric.service}-${metric.name}`}>
                    <div className="lab-metric-head">
                      <span>{metric.service}</span>
                      <strong>{formatMetric(last)}</strong>
                      <em>{metric.unit || ""}</em>
                    </div>
                    <p>{metric.name}</p>
                    <Sparkline points={metric.points} alert={metric.alert} />
                    <div className="lab-metric-foot">
                      <span>最低 {formatMetric(Math.min(...metric.points))}</span>
                      <span>最高 {formatMetric(Math.max(...metric.points))}</span>
                      {metric.alert ? <span className={over ? "over" : ""}>阈值 {formatMetric(metric.alert)}</span> : null}
                    </div>
                  </div>
                );
              })}
            </div>
          </motion.section>
        </div>

        <aside className="lab-side">
          <motion.section className="lab-card lab-diagnosis" {...rise(motionOn, 1)}>
            <div className="lab-card-head">
              <Gauge size={18} />
              <h3>提交诊断</h3>
              {result?.best ? <span className="lab-best">历史最好 {result.best} 分</span> : null}
            </div>
            <label className="lab-field">
              <span>根因</span>
              <select value={root} onChange={(event) => setRoot(event.target.value)}>
                <option value="">选择最可能的根因</option>
                {caseView.rootCauses.map((option, index) => (
                  <option value={String(index)} key={option.label}>
                    {index + 1}. {option.label}
                  </option>
                ))}
              </select>
            </label>
            {root !== "" && caseView.rootCauses[Number(root)]?.detail && (
              <p className="lab-option-detail">{caseView.rootCauses[Number(root)].detail}</p>
            )}
            <label className="lab-field">
              <span>处置</span>
              <select value={fix} onChange={(event) => setFix(event.target.value)}>
                <option value="">选择处置方案</option>
                {caseView.fixes.map((option, index) => (
                  <option value={String(index)} key={option.label}>
                    {index + 1}. {option.label}
                  </option>
                ))}
              </select>
            </label>
            {fix !== "" && caseView.fixes[Number(fix)]?.detail && (
              <p className="lab-option-detail">{caseView.fixes[Number(fix)].detail}</p>
            )}
            <div className="lab-diagnosis-actions">
              <button className="button primary" onClick={() => void submit()} disabled={submitting}>
                {submitting ? <LoaderCircle size={16} className="spin" /> : <Check size={16} />}
                {submitting ? "正在评分…" : "提交诊断"}
              </button>
              <button className="text-button" onClick={() => void reset()} disabled={submitting}>
                <RotateCcw size={15} />
                重新开始
              </button>
            </div>
            <p className="lab-diagnosis-note">
              终端里查过的关键证据会计入评分；提示用一条扣 5 分，最多扣 15 分。
            </p>
          </motion.section>

          <motion.section className="lab-card lab-hints" {...rise(motionOn, 2)}>
            <div className="lab-card-head">
              <Lightbulb size={18} />
              <h3>提示</h3>
              <span className="lab-card-note">
                {hints.length} / {caseView.hintCount}
              </span>
            </div>
            {hints.length === 0 && <p className="lab-hints-empty">卡住时再点，不急着一次看完。</p>}
            <ul>
              {hints.map((hint) => (
                <li key={hint}>{hint}</li>
              ))}
            </ul>
            <button
              className="button secondary"
              onClick={() => void revealHint()}
              disabled={hintBusy || hints.length >= caseView.hintCount}
            >
              <Lightbulb size={15} />
              {hintBusy ? "正在取提示…" : hints.length >= caseView.hintCount ? "提示已用完" : "再给一条提示"}
            </button>
          </motion.section>

          <motion.section className="lab-card lab-catalog-card" {...rise(motionOn, 3)}>
            <div className="lab-card-head">
              <Sparkles size={18} />
              <h3>场景库</h3>
              <span className="lab-card-note">今日日期下自由练习</span>
            </div>
            <div className="lab-catalog-list">
              {boot.catalog.map((item) => (
                <button
                  key={item.id}
                  className={`lab-catalog-item ${item.scenario === caseView.scenario ? "active" : ""}`}
                  onClick={() => void openSummary(item)}
                  disabled={opening}
                >
                  <span className={`lab-level level-${item.level === "高级" ? "high" : item.level === "进阶" ? "mid" : "base"}`}>
                    {item.level}
                  </span>
                  <strong>{item.title}</strong>
                  <span className="lab-catalog-meta">
                    {item.solved ? <CheckCircle2 size={13} /> : <ChevronRight size={13} />}
                    {item.solved ? `已复盘 · ${item.best} 分` : `${item.minutes} 分钟`}
                  </span>
                </button>
              ))}
            </div>
          </motion.section>

          {caseError && (
            <div className="error-box" role="alert">
              {caseError}
            </div>
          )}
        </aside>
      </div>

      {verdict && (
        <motion.section
          ref={verdictRef}
          className={`lab-verdict ${verdict.passed ? "pass" : "fail"}`}
          {...rise(motionOn, 0)}
        >
          <div className="lab-verdict-head">
            {verdict.passed ? <CheckCircle2 size={22} /> : <XCircle size={22} />}
            <strong>{verdict.passed ? "这次排查通过" : "诊断还没对上"}</strong>
            <span className="lab-verdict-score">{verdict.score} 分</span>
            <span className="lab-verdict-best">历史最好 {verdict.best} 分</span>
          </div>
          <p className="lab-verdict-message">{verdict.message}</p>
          <div className="lab-verdict-facts">
            <span className={verdict.rootCauseCorrect ? "ok" : "bad"}>
              根因 {verdict.rootCauseCorrect ? "正确" : "偏了"}
            </span>
            <span className={verdict.fixCorrect ? "ok" : "bad"}>
              处置 {verdict.fixCorrect ? "正确" : "有隐患"}
            </span>
            <span>
              关键证据 {verdict.evidenceHit} / {verdict.evidenceTotal}
            </span>
            <span>使用提示 {verdict.hints} 条</span>
          </div>
          {!verdict.rootCauseCorrect && (
            <p className="lab-correct-answer">
              正确根因：<strong>{caseView.rootCauses[verdict.correctRootCause]?.label}</strong>
            </p>
          )}
          {!verdict.fixCorrect && (
            <p className="lab-correct-answer">
              推荐处置：<strong>{caseView.fixes[verdict.correctFix]?.label}</strong>
            </p>
          )}
          <div className="markdown lab-debrief">
            <LabMarkdown>{verdict.debrief}</LabMarkdown>
          </div>
        </motion.section>
      )}

      {!aiOpen && (
        <button className="lab-ai-fab" onClick={() => setAiOpen(true)}>
          <BrainCircuit size={18} />
          AI 值班助手
        </button>
      )}

      <aside className={`lab-ai ${aiOpen ? "open" : ""}`} aria-hidden={!aiOpen} inert={!aiOpen}>
        <header className="lab-ai-head">
          <span>
            <BrainCircuit size={17} />
            AI 值班助手
          </span>
          <button className="icon-button" aria-label="收起 AI 值班助手" onClick={() => setAiOpen(false)}>
            <X size={16} />
          </button>
        </header>
        <div className="lab-ai-context">
          已接入本场演练的日志、指标、链路与中间件命令；它只能查证，不能执行变更。
        </div>
        <div className="lab-ai-log">
          {messages.length === 0 && (
            <div className="lab-ai-empty">
              <p>不知道从哪查起？让它先给你一个排查顺序。</p>
              {!modelReady && (
                <button className="button secondary" onClick={onOpenSettings}>
                  去配置 AI 服务
                </button>
              )}
            </div>
          )}
          {messages.map((message, index) => (
            <div className={`lab-ai-message ${message.role} ${message.error ? "error" : ""}`} key={`${index}-${message.content.slice(0, 16)}`}>
              <span className="lab-ai-role">{message.role === "user" ? "你" : "AI 值班助手"}</span>
              {message.steps && message.steps.length > 0 && (
                <div className="lab-ai-steps">
                  {message.steps.map((step, stepIndex) => (
                    <span key={`${step.tool}-${stepIndex}`} title={step.args}>
                      {step.tool}
                    </span>
                  ))}
                </div>
              )}
              <div className="markdown">
                <LabMarkdown>{message.content}</LabMarkdown>
              </div>
            </div>
          ))}
          {aiBusy && (
            <div className="lab-ai-message assistant">
              <span className="lab-ai-role">AI 值班助手</span>
              <div className="lab-ai-thinking">
                <LoaderCircle size={14} className="spin" />
                正在查日志和指标…
              </div>
            </div>
          )}
          <div ref={chatEnd} />
        </div>
        <div className="lab-ai-suggestions">
          {["先帮我梳理排查顺序", "最近 30 分钟有哪些 ERROR？", "这条链路最慢在哪一层？"].map((suggestion) => (
            <button
              key={suggestion}
              className="quick-chip"
              onClick={() => void ask(suggestion)}
              disabled={aiBusy}
            >
              {suggestion}
            </button>
          ))}
        </div>
        <form
          className="lab-ai-input"
          onSubmit={(event) => {
            event.preventDefault();
            void ask(aiInput);
          }}
        >
          <input
            value={aiInput}
            onChange={(event) => setAiInput(event.target.value)}
            placeholder={modelReady ? "问它：支付超时，先查什么？" : "先配置 AI 服务"}
            aria-label="向 AI 值班助手提问"
            disabled={aiBusy}
          />
          <button className="button primary" type="submit" disabled={aiBusy || !aiInput.trim()}>
            <Send size={15} />
          </button>
        </form>
        {!modelReady && (
          <div className="lab-ai-config">
            <span>课程和演练本身完全离线；AI 需要你自己配置服务。</span>
            <button className="text-button" onClick={onOpenSettings}>
              打开设置
            </button>
          </div>
        )}
      </aside>
    </motion.div>
  );
}
