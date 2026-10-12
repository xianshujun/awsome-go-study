import { useCallback, useEffect, useState } from "react";

/** 显示偏好只影响这台设备上的观感，因此放在 localStorage，不进后端设置文件。 */
const STORAGE_KEY = "gostudy-display";

export type DisplayPrefs = {
  /** null 表示跟随系统；true/false 表示用户在本机明确指定过。 */
  reduceMotion: boolean | null;
  editorSize: number;
};

export const EDITOR_SIZE = { min: 12, max: 20, step: 1, initial: 14 } as const;

const defaults: DisplayPrefs = { reduceMotion: null, editorSize: EDITOR_SIZE.initial };

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
