export interface Course {
  id: string;
  title: string;
  description: string;
  level: string;
  duration: string;
  lessonIds: string[];
  accent: string;
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
  runtime: { platform: string; engine: string };
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
