export interface Course {
  id: string;
  title: string;
  description: string;
  level: string;
  duration: string;
  lessonIds: string[];
}
export interface Lesson {
  id: string;
  courseId: string;
  title: string;
  description: string;
  duration: number;
  concept: string;
  example: string;
  starter: string;
  task: string;
  expectedOutput: string;
  hints: string[];
  quiz: {
    question: string;
    options: string[];
    answer: number;
    explanation: string;
  };
  takeaways: string[];
  source: string;
}
export interface Progress {
  completed: string[];
  drafts: Record<string, string>;
  notes: Record<string, string>;
  lastLesson: string;
}
export interface Settings {
  baseUrl: string;
  model: string;
  hasKey: boolean;
  /** 关掉应用窗口之后的行为；引擎在窗口已经关掉时要用它，所以存在后端状态里。 */
  closeAction: CloseAction;
}
/** ask 每次问；background 后台静默运行；quit 直接退出程序。 */
export type CloseAction = "ask" | "background" | "quit";
export interface Bootstrap {
  courses: Course[];
  lessons: Lesson[];
  progress: Progress;
  settings: Settings;
  runtime: {
    platform: string;
    engine: string;
    version: string;
    content?: string;
    /** 这个平台能在窗口关掉之后弹一次询问框（目前只有 Windows）。 */
    closePrompt?: boolean;
    /** 程序目录里有卸载程序，可以在设置里发起卸载。 */
    canUninstall?: boolean;
  };
  conversations: ConversationList;
}
export interface CheckResult {
  kind: string;
  name?: string;
  label: string;
  hint?: string;
  passed: boolean;
}
export interface RunResult {
  output: string;
  error?: string;
  passed: boolean;
  durationMs: number;
  /** 输出之外的结构检查结果；课程没有配置检查时为空。 */
  checks?: CheckResult[];
}
export interface Message {
  role: "user" | "assistant";
  content: string;
  /** 服务端记录的秒级时间戳；刚发出、还没写回去的消息没有这个字段。 */
  at?: number;
}
export interface ConversationSummary {
  id: string;
  title: string;
  lessonId: string;
  updatedAt: number;
  messageCount: number;
  preview: string;
}
export interface ConversationList {
  activeId: string;
  items: ConversationSummary[];
}
export interface Conversation extends ConversationSummary {
  messages: Message[];
}
export interface ConversationPayload extends ConversationList {
  conversation?: Conversation;
}
export interface UpdateOffer {
  version: string;
  released?: string;
  notes?: string;
  size: number;
  asset: string;
}
export interface ContentOffer {
  version: string;
  size: number;
  notes?: string;
  current?: string;
}
export interface UpdateStatus {
  current: string;
  content?: string;
  app?: UpdateOffer;
  contentUpdate?: ContentOffer;
  checked: number;
  skipped?: boolean;
  error?: string;
  warning?: string;
}
export type DownloadStage = "idle" | "downloading" | "verifying" | "ready" | "installing" | "failed";
export interface DownloadState {
  state: DownloadStage;
  received: number;
  total: number;
  asset?: string;
  error?: string;
  /** 这次下载的是程序、安装包还是课程内容。 */
  kind?: "app" | "installer" | "content";
  /** 拿到包之后该做什么：桌面重启、Android 交给系统安装器。 */
  mode?: "restart" | "android" | "content" | "installer";
}

/* ---------------- 事故演练场 ---------------- */

export interface LabResult {
  caseId: string;
  scenario: string;
  date: string;
  title: string;
  passed: boolean;
  best: number;
  attempts: number;
  hints: number;
  updatedAt: number;
}

export interface LabProgress {
  results?: Record<string, LabResult>;
  lastCase?: string;
  solved: number;
  attempts: number;
}

export interface LabCaseSummary {
  id: string;
  scenario: string;
  scenarioId: string;
  date: string;
  title: string;
  subtitle?: string;
  level: string;
  minutes: number;
  tags: string[];
  alert?: string;
  /** 今日事故 / 历史存档 / 场景库 */
  mode: string;
  solved: boolean;
  best?: number;
  attempts?: number;
  hints?: number;
  isToday?: boolean;
}

export interface LabBootstrap {
  company: string;
  tagline: string;
  today: LabCaseSummary;
  archive: LabCaseSummary[];
  catalog: LabCaseSummary[];
  progress: LabProgress;
  modelReady: boolean;
  scenarioCount: number;
  packVersion: number;
  warnings?: string[];
}

export interface LabService {
  name: string;
  role: string;
  status: "healthy" | "degraded" | "critical";
  replicas?: string;
  cpu?: string;
  memory?: string;
  note?: string;
}

export interface LabMetric {
  service: string;
  name: string;
  unit?: string;
  alert?: number;
  points: number[];
}

export interface LabOption {
  label: string;
  detail?: string;
}

export interface LabCase {
  id: string;
  scenario: string;
  company: string;
  date: string;
  clock: string;
  title: string;
  subtitle?: string;
  brief: string;
  mission: string;
  business?: string;
  level: string;
  minutes: number;
  tags: string[];
  alerts: string[];
  timeline?: string[];
  services: LabService[];
  metrics: LabMetric[];
  presets: string[];
  rootCauses: LabOption[];
  fixes: LabOption[];
  hints: string[];
  hintCount: number;
  result?: LabResult;
  history?: { command: string }[];
}

export interface LabRunResult {
  command: string;
  output: string;
  at: number;
}

export interface LabSubmitResult {
  passed: boolean;
  score: number;
  best: number;
  rootCauseCorrect: boolean;
  fixCorrect: boolean;
  correctRootCause: number;
  correctFix: number;
  evidenceHit: number;
  evidenceTotal: number;
  hints: number;
  message: string;
  debrief: string;
  result: LabResult;
  progress: LabProgress;
}

export interface LabHintResult {
  hint: string;
  index: number;
  hintsUsed: number;
  total: number;
}

export interface LabChatStep {
  tool: string;
  args?: string;
  summary?: string;
}

export interface LabChatResponse {
  reply: string;
  steps?: LabChatStep[];
}

const incomingToken = new URLSearchParams(window.location.hash.slice(1)).get(
  "token",
);
if (incomingToken) {
  sessionStorage.setItem("gostudy-token", incomingToken);
  window.history.replaceState(
    null,
    "",
    window.location.pathname + window.location.search,
  );
}

export async function api<T>(
  path: string,
  body?: unknown,
  signal?: AbortSignal,
): Promise<T> {
  const response = await fetch(`/api/${path}`, {
    method: body === undefined ? "GET" : "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${sessionStorage.getItem("gostudy-token") ?? ""}`,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
  });
  const result = await response
    .json()
    .catch(() => ({ error: "本地引擎返回了无法读取的响应。" }));
  if (!response.ok)
    throw new Error(result.error || `请求失败 · ${response.status}`);
  return result as T;
}

/**
 * 告诉引擎"页面要走了"。引擎据此决定退出、留在后台还是问一次用户。
 *
 * 关闭窗口时浏览器只给页面最后一点时间，所以用 sendBeacon：它由浏览器接管发送，
 * 页面销毁之后也会送出去；它带不了自定义请求头，凭据只好放在查询串里，
 * 这个端点在后端也是唯一允许查询串凭据的地方。
 */
export function reportWindowClosed() {
  const token = sessionStorage.getItem("gostudy-token");
  if (!token) return;
  const url = `/api/window-closed?token=${encodeURIComponent(token)}`;
  try {
    if (typeof navigator.sendBeacon === "function" && navigator.sendBeacon(url)) return;
  } catch {
    // 少数内嵌浏览器会禁用 sendBeacon，交给下面的 fetch 兜底。
  }
  try {
    void fetch(url, { method: "POST", keepalive: true });
  } catch {
    // 实在发不出去也没关系：引擎侧还留着别的收尾办法（比如用户从设置里退出）。
  }
}

export interface ChatBody {
  lessonId: string;
  code?: string;
  message: string;
  history?: { role: string; content: string }[];
  /** 最近一次运行结果，让导师能直接看到报错。 */
  lastRun?: { output?: string; error?: string; passed: boolean };
}

/**
 * 向 AI 导师提问。优先走 SSE：服务端逐段返回，页面边收边显示；
 * 服务端（或反向代理）返回普通 JSON 时自动退化成一次性回复，接口保持兼容。
 */
export async function chatStream(
  body: ChatBody,
  onDelta: (delta: string) => void,
  signal?: AbortSignal,
): Promise<string> {
  const response = await fetch("/api/chat", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${sessionStorage.getItem("gostudy-token") ?? ""}`,
    },
    body: JSON.stringify({ ...body, stream: true }),
    signal,
  });
  if (!response.ok) {
    const result = await response.json().catch(() => ({ error: `请求失败 · ${response.status}` }));
    throw new Error(result.error || `请求失败 · ${response.status}`);
  }
  const type = response.headers.get("content-type") ?? "";
  if (!type.includes("text/event-stream") || !response.body) {
    const result = (await response.json()) as { reply?: string; error?: string };
    if (result.error) throw new Error(result.error);
    if (!result.reply) throw new Error("AI 服务返回了无效回复");
    onDelta(result.reply);
    return result.reply;
  }
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let received = "";
  let finalReply = "";
  const handle = (data: string) => {
    if (!data || data === "[DONE]") return;
    let payload: { delta?: string; error?: string; done?: boolean; reply?: string };
    try {
      payload = JSON.parse(data);
    } catch {
      return;
    }
    if (payload.error) throw new Error(payload.error);
    if (payload.delta) {
      received += payload.delta;
      onDelta(payload.delta);
    }
    if (payload.done && typeof payload.reply === "string") {
      finalReply = payload.reply;
    }
  };
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const events = buffer.split("\n\n");
      buffer = events.pop() ?? "";
      for (const event of events) {
        for (const line of event.split("\n")) {
          if (!line.startsWith("data:")) continue;
          handle(line.slice(5).trim());
        }
      }
    }
    // 流末尾可能还剩下一个没有空行结尾的事件。
    for (const line of buffer.split("\n")) {
      if (line.startsWith("data:")) handle(line.slice(5).trim());
    }
  } catch (err) {
    await reader.cancel().catch(() => undefined);
    throw err;
  } finally {
    reader.releaseLock();
  }
  if (finalReply) return finalReply;
  if (!received.trim()) throw new Error("AI 服务没有返回内容");
  return received;
}
