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
}
export interface Bootstrap {
  courses: Course[];
  lessons: Lesson[];
  progress: Progress;
  settings: Settings;
  runtime: { platform: string; engine: string; version: string; content?: string };
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
