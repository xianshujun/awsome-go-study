import { useCallback, useEffect, useState } from "react";

/** 显示偏好只影响这台设备上的观感，因此放在 localStorage，不进后端设置文件。 */
const STORAGE_KEY = "gostudy-display";

export type DisplayPrefs = {
  /** null 表示跟随系统；true/false 表示用户在本机明确指定过。 */
  reduceMotion: boolean | null;
  editorSize: number;
  /** 每天最多检查一次新版本；用户可以在设置里关掉。 */
  autoUpdate: boolean;
};

export const EDITOR_SIZE = { min: 12, max: 20, step: 1, initial: 14 } as const;

const defaults: DisplayPrefs = { reduceMotion: null, editorSize: EDITOR_SIZE.initial, autoUpdate: true };

function clampSize(value: number) {
  if (!Number.isFinite(value)) return EDITOR_SIZE.initial;
  return Math.min(EDITOR_SIZE.max, Math.max(EDITOR_SIZE.min, Math.round(value)));
}

export function readPrefs(): DisplayPrefs {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return defaults;
    const parsed = JSON.parse(raw) as Partial<DisplayPrefs>;
    return {
      reduceMotion:
        typeof parsed.reduceMotion === "boolean" ? parsed.reduceMotion : null,
      editorSize: clampSize(Number(parsed.editorSize ?? EDITOR_SIZE.initial)),
      // 老版本没有这个字段，默认开启。
      autoUpdate: parsed.autoUpdate !== false,
    };
  } catch {
    // 隐私模式或写入被拦截时退回默认值，不影响学习功能。
    return defaults;
  }
}

export function useDisplayPrefs() {
  const [prefs, setPrefs] = useState<DisplayPrefs>(readPrefs);
  const update = useCallback((patch: Partial<DisplayPrefs>) => {
    setPrefs((current) => {
      const next = { ...current, ...patch };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      } catch {
        // 存不下就只影响本次会话。
      }
      return next;
    });
  }, []);
  useEffect(() => {
    const root = document.documentElement;
    if (prefs.reduceMotion === null) root.removeAttribute("data-reduce-motion");
    else root.dataset.reduceMotion = prefs.reduceMotion ? "on" : "off";
    root.style.setProperty("--editor-size", `${prefs.editorSize}px`);
  }, [prefs.reduceMotion, prefs.editorSize]);
  return { prefs, update };
}

/** 用户是否明确关掉了动效；用于选择 motion 的 reducedMotion 策略。 */
export function motionPolicy(reduceMotion: boolean | null) {
  return reduceMotion ? "always" : "user";
}

/* ---------------- AI 导师浮窗的位置与尺寸 ---------------- */

/** 面板摆在哪里同样只是这台设备上的事，所以放在 localStorage，不进后端状态文件。 */
const ASSISTANT_KEY = "gostudy-assistant-window";

export const ASSISTANT_WINDOW = {
  minWidth: 330,
  maxWidth: 760,
  minHeight: 300,
  defaultWidth: 400,
  defaultHeight: 560,
  /** 浮窗离窗口边缘的空隙，也是拖动时不允许越过的边界。 */
  margin: 14,
} as const;

/**
 * 面板默认贴着顶栏下方的右侧停靠：页面内容会给它让出位置，
 * 所以默认状态下它不会压住编辑器、运行按钮或测验。
 * 拖走之后就变成普通浮窗，位置记在本机。
 */
export type AssistantWindow = {
  docked: boolean;
  x: number;
  y: number;
  width: number;
  height: number;
};
type Viewport = { width: number; height: number };

function fitWidth(width: number, view: Viewport) {
  const limitWidth = Math.max(ASSISTANT_WINDOW.minWidth, view.width - ASSISTANT_WINDOW.margin * 2);
  return Math.min(Math.max(width, ASSISTANT_WINDOW.minWidth), Math.min(ASSISTANT_WINDOW.maxWidth, limitWidth));
}

export function dockedAssistantWindow(width: number, view: Viewport, topbar: number): AssistantWindow {
  const top = topbar + ASSISTANT_WINDOW.margin;
  const fitted = fitWidth(width, view);
  return {
    docked: true,
    width: fitted,
    height: Math.max(ASSISTANT_WINDOW.minHeight, view.height - top - ASSISTANT_WINDOW.margin),
    x: view.width - fitted - ASSISTANT_WINDOW.margin,
    y: top,
  };
}

/** 把浮窗按视口夹回可见范围：换窗口大小、换显示器后都不会丢到屏幕外。 */
export function clampAssistantWindow(win: AssistantWindow, view: Viewport): AssistantWindow {
  const limitHeight = Math.max(ASSISTANT_WINDOW.minHeight, view.height - ASSISTANT_WINDOW.margin * 2);
  const width = fitWidth(win.width, view);
  const height = Math.min(Math.max(win.height, ASSISTANT_WINDOW.minHeight), limitHeight);
  return {
    docked: win.docked,
    width,
    height,
    x: Math.min(Math.max(win.x, ASSISTANT_WINDOW.margin), Math.max(ASSISTANT_WINDOW.margin, view.width - width - ASSISTANT_WINDOW.margin)),
    y: Math.min(Math.max(win.y, ASSISTANT_WINDOW.margin), Math.max(ASSISTANT_WINDOW.margin, view.height - height - ASSISTANT_WINDOW.margin)),
  };
}

export function defaultAssistantWindow(view: Viewport, topbar: number): AssistantWindow {
  return dockedAssistantWindow(ASSISTANT_WINDOW.defaultWidth, view, topbar);
}

export function readAssistantWindow(view: Viewport, topbar: number): AssistantWindow {
  try {
    const raw = localStorage.getItem(ASSISTANT_KEY);
    if (!raw) return defaultAssistantWindow(view, topbar);
    const parsed = JSON.parse(raw) as Partial<AssistantWindow>;
    const width = typeof parsed.width === "number" && Number.isFinite(parsed.width) ? parsed.width : ASSISTANT_WINDOW.defaultWidth;
    if (parsed.docked !== false) return dockedAssistantWindow(width, view, topbar);
    const numbers = [parsed.x, parsed.y, parsed.height];
    if (numbers.some((value) => typeof value !== "number" || !Number.isFinite(value))) {
      return defaultAssistantWindow(view, topbar);
    }
    return clampAssistantWindow({ docked: false, x: parsed.x!, y: parsed.y!, width, height: parsed.height! }, view);
  } catch {
    return defaultAssistantWindow(view, topbar);
  }
}

export function saveAssistantWindow(win: AssistantWindow) {
  try {
    localStorage.setItem(ASSISTANT_KEY, JSON.stringify(win));
  } catch {
    // 存不下就只在本次会话里生效。
  }
}
