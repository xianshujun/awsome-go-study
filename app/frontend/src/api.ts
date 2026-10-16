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
export interface RunResult {
  output: string;
  error?: string;
  passed: boolean;
  durationMs: number;
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
export type DownloadStage = "idle" | "downloading" | "verifying" | "ready" | "failed";
export interface DownloadState {
  state: DownloadStage;
  received: number;
  total: number;
  asset?: string;
  error?: string;
  /** 拿到包之后该做什么：桌面重启、Android 交给系统安装器。 */
  mode?: "restart" | "android" | "content" | "installer";
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
