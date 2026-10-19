import { useCallback, useRef, useState } from "react";
import { api, chatStream } from "./api";
import type {
  ConversationList,
  ConversationPayload,
  ConversationSummary,
  Message,
  RunResult,
} from "./api";

/** 与服务端校验保持一致：单条消息最多 4096 字节。 */
export const MESSAGE_BYTES = 4096;
/** 随问题一起发出去的历史条数与单条上限，防止一次把整段对话塞进上下文。 */
const HISTORY_LIMIT = 8;
const HISTORY_BYTES = 1800;

/** 按字节截断，避免把多字节字符截成半个。 */
export function boundedText(text: string, bytes: number) {
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

export type TutorContext = { lessonId: string; code?: string; lastRun?: RunResult };

/**
 * AI 导师的对话状态。放在 App 层而不是面板里：
 * 面板只是它的一个视图，关掉面板不影响请求、保存和恢复，
 * 「关掉再打开」时也不会连问题带回答一起丢。
 */
export function useTutor() {
  const [items, setItems] = useState<ConversationSummary[]>([]);
  const [activeId, setActiveId] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  /** 流式回复的临时正文；收到 done 后由正式消息替换。 */
  const [draft, setDraft] = useState("");
  const [sending, setSending] = useState(false);
  const [opening, setOpening] = useState(false);
  const [error, setError] = useState("");
  const abort = useRef<AbortController | undefined>(undefined);
  const hydrated = useRef(false);
  /** 已经写进本机记录的条数，用来判断重试时是否需要重新补记用户消息。 */
  const stored = useRef(0);

  const applyList = useCallback((payload: ConversationList) => {
    setItems(payload.items ?? []);
    setActiveId(payload.activeId ?? "");
  }, []);

  /** 写入成功后以服务端的副本为准：本地视图、条数与磁盘上的内容始终一致。 */
  const applySaved = useCallback(
    (payload: ConversationPayload) => {
      applyList(payload);
      if (!payload.conversation) return;
      stored.current = payload.conversation.messages.length;
      setMessages(payload.conversation.messages);
    },
    [applyList],
  );

  /** 首次拿到 bootstrap 时恢复上次的对话；之后由用户操作驱动，不再覆盖。 */
  const hydrate = useCallback(
    (list?: ConversationList) => {
      if (!list || hydrated.current) return;
      hydrated.current = true;
      setItems(list.items ?? []);
      setActiveId(list.activeId ?? "");
      if (!list.activeId) return;
      setOpening(true);
      api<ConversationPayload>("conversations", { action: "get", id: list.activeId })
        .then((payload) => {
          applyList(payload);
          stored.current = payload.conversation?.messages.length ?? 0;
          setMessages(payload.conversation?.messages ?? []);
        })
        .catch((err: Error) => setError(err.message))
        .finally(() => setOpening(false));
    },
    [applyList],
  );

  const open = useCallback(
    async (id: string) => {
      if (!id) return;
      setOpening(true);
      setError("");
      try {
        const payload = await api<ConversationPayload>("conversations", { action: "get", id });
        applyList(payload);
        stored.current = payload.conversation?.messages.length ?? 0;
        setMessages(payload.conversation?.messages ?? []);
      } catch (err) {
        setError((err as Error).message);
      } finally {
        setOpening(false);
      }
    },
    [applyList],
  );

  const create = useCallback(
    async (lessonId: string) => {
      setError("");
      try {
        const payload = await api<ConversationPayload>("conversations", { action: "create", lessonId });
        applyList(payload);
        stored.current = 0;
        setMessages([]);
        return payload.conversation?.id ?? "";
      } catch (err) {
        setError((err as Error).message);
        return "";
      }
    },
    [applyList],
  );

  const rename = useCallback(
    async (id: string, title: string) => {
      const clean = title.replace(/\s+/g, " ").trim();
      if (!clean) return;
      setItems((current) =>
        current.map((item) => (item.id === id ? { ...item, title: clean } : item)),
      );
      try {
        applyList(await api<ConversationPayload>("conversations", { action: "rename", id, title: clean }));
      } catch (err) {
        setError((err as Error).message);
      }
    },
    [applyList],
  );

  const remove = useCallback(
    async (id: string) => {
      setError("");
      try {
        const payload = await api<ConversationPayload>("conversations", { action: "delete", id });
        const switched = payload.activeId !== activeId;
        applyList(payload);
        // 删掉的是当前这段时，服务端已经把最近用过的一段一起带回来了。
        if (switched) {
          stored.current = payload.conversation?.messages.length ?? 0;
          setMessages(payload.conversation?.messages ?? []);
        }
      } catch (err) {
        setError((err as Error).message);
      }
    },
    [activeId, applyList],
  );

  /**
   * 提问：先落盘用户消息，再请求模型，最后把回复也写回去。
   * 只有主动提问才会联系模型，历史随请求一起发。
   */
  const ask = useCallback(
    async (
      message: string,
      context: TutorContext,
      options: { record: boolean; history?: Message[] },
    ) => {
      const text = message.trim();
      if (!text || sending) return;
      setError("");
      setSending(true);
      const history = (options.history ?? messages)
        .slice(-HISTORY_LIMIT)
        .map((item) => ({ role: item.role, content: boundedText(item.content, HISTORY_BYTES) }));
      const controller = new AbortController();
      abort.current = controller;
      setDraft("");
      if (options.record) setMessages((current) => [...current, { role: "user", content: text }]);
      try {
        let id = activeId;
        if (!id) {
          id = await create(context.lessonId);
          if (!id) return;
        }
        if (options.record) {
          const saved = await api<ConversationPayload>(
            "conversations",
            { action: "append", id, lessonId: context.lessonId, role: "user", content: boundedText(text, MESSAGE_BYTES) },
            controller.signal,
          );
          applySaved(saved);
        }
        const reply = await chatStream(
          {
            lessonId: context.lessonId,
            code: context.code ?? "",
            message: boundedText(text, MESSAGE_BYTES),
            history,
            lastRun: context.lastRun
              ? {
                  output: boundedText(context.lastRun.output ?? "", 16 * 1024),
                  error: boundedText(context.lastRun.error ?? "", 4096),
                  passed: context.lastRun.passed,
                }
              : undefined,
          },
          (delta) => setDraft((current) => current + delta),
          controller.signal,
        );
        setDraft("");
        setMessages((current) => [...current, { role: "assistant", content: reply }]);
        try {
          applySaved(
            await api<ConversationPayload>(
              "conversations",
              { action: "append", id, role: "assistant", content: reply },
              controller.signal,
            ),
          );
        } catch (err) {
          if ((err as Error).name !== "AbortError") setError("回复已收到，但没能写进本机对话记录。");
        }
      } catch (err) {
        setDraft("");
        if ((err as Error).name !== "AbortError") setError((err as Error).message);
      } finally {
        setSending(false);
        abort.current = undefined;
      }
    },
    [activeId, applySaved, create, messages, sending],
  );

  /** 用户主动停止：中止当前请求，已经收到的半段回复不再写进记录。 */
  const cancel = useCallback(() => {
    abort.current?.abort();
  }, []);

  const send = useCallback(
    (text: string, context: TutorContext) => ask(text, context, { record: true }),
    [ask],
  );

  /**
   * 请求失败后重试最后一条提问。已经落盘的那条不再重复记录，
   * 只重新问一次；还没落盘（例如新建对话就失败了）的先补记录再问。
   */
  const retry = useCallback(
    (context: TutorContext) => {
      const index = messages.map((item) => item.role).lastIndexOf("user");
      if (index < 0) return;
      const record = index >= stored.current;
      const history = messages.slice(0, index);
      setMessages(messages.slice(0, record ? index : index + 1));
      return ask(messages[index].content, context, { record, history });
    },
    [ask, messages],
  );

  return {
    items,
    activeId,
    active: items.find((item) => item.id === activeId),
    messages,
    draft,
    sending,
    opening,
    error,
    send,
    cancel,
    retry,
    create,
    open,
    rename,
    remove,
    hydrate,
    clearError: () => setError(""),
  };
}
