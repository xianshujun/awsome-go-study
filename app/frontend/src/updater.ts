import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "./api";
import type { DownloadState, UpdateStatus } from "./api";

/**
 * 检查结果缓存在本机：自动检查有 24 小时节流，刷新页面后不该"什么都看不到了"。
 * 缓存连同当时的程序版本一起存，升级之后旧结论自然作废。
 */
const CACHE_KEY = "gostudy-update-status";

/** self 是"这份结论是对哪个程序版本做的"，取自检查结果里的 current。 */
type Cached = { self: string; status: UpdateStatus };

function readCache(version: string, content: string): UpdateStatus | undefined {
  try {
    const raw = localStorage.getItem(CACHE_KEY);
    if (!raw) return undefined;
    const cached = JSON.parse(raw) as Cached;
    if (cached.self !== version) return undefined;
    const status = cached.status;
    // 已经应用过的内容包不再是"待更新"。
    if (status?.contentUpdate && content && status.contentUpdate.version === content) {
      return { ...status, contentUpdate: undefined };
    }
    return status;
  } catch {
    return undefined;
  }
}

function writeCache(version: string, status: UpdateStatus) {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({ self: version, status } satisfies Cached));
  } catch {
    // 存不下的浏览器（隐私模式）里就只在本次会话有效。
  }
}

/**
 * 版本与课程内容的更新状态。
 *
 * 网络与落盘都在本地引擎里：这个 hook 只负责触发、轮询进度和把结果摊给界面看。
 * 自动检查默认开着，但服务端按 24 小时节流，所以每次启动也不会反复联网。
 */
export function useUpdater({
  autoUpdate,
  version,
  content = "",
}: {
  autoUpdate: boolean;
  version: string;
  content?: string;
}) {
  const [status, setStatus] = useState<UpdateStatus>();

  // 程序版本要等 bootstrap 回来才知道，所以缓存不是在初始化时读，
  // 而是等版本号到位后接回来：刷新页面、自动检查被节流时，界面照样显示上次的结论。
  useEffect(() => {
    if (!version || status) return;
    const cached = readCache(version, content);
    if (cached) setStatus(cached);
  }, [version, content, status]);
  const [checking, setChecking] = useState(false);
  const [download, setDownload] = useState<DownloadState>({ state: "idle", received: 0, total: 0 });
  const [error, setError] = useState("");
  const [applying, setApplying] = useState<"app" | "content" | "">("");
  const started = useRef(false);

  // 国内访客优先 Gitee；和落地页用同一条判断规则。
  const prefer = useCallback(() => {
    const china = /^zh/i.test(navigator.language) && /^(Asia|PRC)\//.test(
      Intl.DateTimeFormat().resolvedOptions().timeZone ?? "",
    );
    return china ? "cn" : "global";
  }, []);

  const check = useCallback(
    async (manual: boolean) => {
      setChecking(true);
      if (manual) setError("");
      try {
        const result = await api<UpdateStatus>("update", {
          action: "check",
          auto: !manual,
          prefer: prefer(),
        });
        if (!result.skipped) {
          setStatus(result);
          // 用结果里的 current 作为缓存版本号：它一定和这次检查对应得上。
          writeCache(result.current || version, result);
        }
        if (manual && result.error) setError(result.error);
      } catch (err) {
        if (manual) setError((err as Error).message);
      } finally {
        setChecking(false);
      }
    },
    [prefer, version],
  );

  const downloadUpdate = useCallback(async () => {
    setError("");
    setApplying("app");
    try {
      const state = await api<DownloadState>("update", { action: "download", prefer: prefer() });
      setDownload(state);
    } catch (err) {
      setError((err as Error).message);
      setApplying("");
    }
  }, [prefer]);

  const updateContent = useCallback(async () => {
    setError("");
    setApplying("content");
    try {
      await api<{ version: string; warning?: string }>("update", { action: "content", prefer: prefer() });
      // 课程换了，重新取一遍 bootstrap，页面上的课表才是新的。
      reloadPage();
    } catch (err) {
      setError((err as Error).message);
      setApplying("");
    }
  }, [prefer]);

  const install = useCallback(async () => {
    setError("");
    setApplying("app");
    try {
      const state = await api<DownloadState>("update", { action: "install" });
      setDownload(state);
      if (state.mode === "android") {
        // 交给 Android 外壳：它按约定去 <filesDir>/updates/GoStudy.apk 取包并拉起系统安装器。
        window.location.href = "gostudy://install";
      }
    } catch (err) {
      setError((err as Error).message);
      setApplying("");
    }
  }, []);

  // 下载、校验和安装中都要轮询：安装器还要再下 6 MB 的 Setup，
  // 后台失败也必须回到界面上，不能停在“更新包已就绪”。
  useEffect(() => {
    if (
      download.state !== "downloading" &&
      download.state !== "verifying" &&
      download.state !== "installing"
    ) {
      return;
    }
    let alive = true;
    const timer = window.setInterval(async () => {
      try {
        const state = await api<DownloadState>("update", { action: "progress" });
        if (!alive) return;
        setDownload(state);
        if (state.state === "failed") {
          setError(state.error || "更新失败，请重试");
          setApplying("");
        }
        if (state.state === "ready") setApplying("");
      } catch {
        // 轮询失败就等下一次；真正的失败会体现在状态里。
      }
    }, 400);
    return () => {
      alive = false;
      window.clearInterval(timer);
    };
  }, [download.state]);

  useEffect(() => {
    if (!autoUpdate || started.current) return;
    started.current = true;
    void check(false);
  }, [autoUpdate, check]);

  return {
    status,
    checking,
    download,
    error,
    applying,
    /** 有可用的程序更新（用于顶栏小红点）。 */
    hasApp: Boolean(status?.app),
    hasContent: Boolean(status?.contentUpdate),
    check,
    downloadUpdate,
    install,
    updateContent,
  };
}

export type Updater = ReturnType<typeof useUpdater>;

/** 个别内嵌浏览器会拦掉 reload；拦掉就下次打开时再取新内容。 */
function reloadPage() {
  try {
    window.location.reload();
  } catch {
    // 忽略：内容已经落在本机，重新打开页面即可看到。
  }
}
