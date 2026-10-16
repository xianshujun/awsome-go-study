import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { App } from './main';
import type { Bootstrap, Conversation, ConversationSummary, Lesson, Progress, RunResult, Settings } from './api';

// 使用同一 onChange 契约，测试用户编辑而不是 CodeMirror 的布局细节。
vi.mock('@uiw/react-codemirror', () => ({
  default: ({ value, onChange, editable = true, 'aria-label': label }: {
    value: string; onChange: (value: string) => void; editable?: boolean; 'aria-label'?: string;
  }) => <textarea aria-label={label} value={value} readOnly={!editable} onChange={event => onChange(event.target.value)} />,
}));

const solution = 'package main\n\nimport "fmt"\n\nfunc main() { fmt.Println("Hello, Go!") }';
const secondDraft = 'package main\n\nfunc main() { println(42) }';
const lessons: Lesson[] = [
  {
    id: 'hello', courseId: 'basics', title: '第一行 Go 代码', description: '认识程序入口与输出。', duration: 5,
    concept: '## 程序的入口\n\nmain 函数是程序开始运行的地方。', example: solution,
    starter: 'package main\n\nfunc main() {}', task: '输出一行问候语。', expectedOutput: 'Hello, Go!\n',
    hints: ['使用 fmt.Println。'], takeaways: ['main 是入口。'],
    quiz: { question: '程序从哪里开始执行？', options: ['任意函数', 'main 函数', '最后一行'], answer: 1, explanation: 'Go 从 main 函数开始执行。' },
    source: 'go-basic/go-grammar.md · 第二章 变量与常量',
  },
  {
    id: 'variables', courseId: 'basics', title: '变量与自己的表达', description: '保存一个整数。', duration: 8,
    concept: '变量为值命名。', example: secondDraft, starter: 'package main\n\nfunc main() {}',
    task: '输出整数 42。', expectedOutput: '42\n', hints: ['可以用短变量声明。'], takeaways: ['变量有类型。'],
    quiz: { question: '短变量声明使用哪个符号？', options: ['=', ':='], answer: 1, explanation: ':= 声明并赋值。' },
    source: 'go-basic/go-grammar.md · 第五章 控制流',
  },
];

function bootstrap(progress: Partial<Progress> = {}): Bootstrap {
  return {
    courses: [{ id: 'basics', title: 'Go 语言基础', description: '从可运行的小程序开始。', level: '入门', duration: '13 分钟', lessonIds: lessons.map(lesson => lesson.id) }],
    lessons,
    progress: { completed: [], drafts: {}, notes: {}, lastLesson: '', ...progress },
    settings: { baseUrl: 'https://api.example.com/v1', model: 'study-model', hasKey: false, closeAction: 'ask' as const },
    runtime: { platform: 'linux/amd64', engine: '受限解释器', version: '0.7.0' },
    conversations: { activeId: '', items: [] },
  };
}

type RequestBody = {
  lessonId?: string; code?: string; note?: string; complete?: boolean; quizAnswer?: number;
  baseUrl?: string; model?: string; apiKey?: string; clearKey?: boolean; message?: string;
  action?: string; id?: string; role?: string; content?: string; title?: string;
  history?: { role: string; content: string }[];
};
type RecordedRequest = { path: string; body?: RequestBody; authorization: string | null };
type StoredChat = { id: string; title: string; lessonId: string; updatedAt: number; messages: { role: 'user' | 'assistant'; content: string; at: number }[] };
let state: Bootstrap;
let requests: RecordedRequest[];
let runResponse: () => Promise<Response>;
let chats: StoredChat[];
let activeChat: string;
let clock: number;

function chatSummary(chat: StoredChat): ConversationSummary {
  return {
    id: chat.id, title: chat.title, lessonId: chat.lessonId, updatedAt: chat.updatedAt,
    messageCount: chat.messages.length, preview: chat.messages.at(-1)?.content.slice(0, 60) ?? '',
  };
}

type UpdateBody = { action?: string; auto?: boolean; prefer?: string };
let updateStatus: Record<string, unknown>;
let updateProgress: Record<string, unknown>;
let updateRequests: UpdateBody[];
let progressCalls: number;

/** 更新接口的最小实现：检查、下载、安装各返回一个可断言的状态。 */
function updatesAPI(body: RequestBody): unknown {
  updateRequests.push(body as UpdateBody);
  switch (body.action) {
    case 'check':
      return { ...updateStatus, skipped: false, checked: 1000 };
    case 'download':
      updateProgress = { state: 'downloading', received: 1024, total: 4096, asset: 'GoStudy-linux-amd64' };
      return updateProgress;
    case 'progress':
      // 轮询一次之后就当作下载完，模拟"下好并校验通过"。
      if (++progressCalls > 1) {
        updateProgress = { state: 'ready', received: 4096, total: 4096, asset: 'GoStudy-linux-amd64', mode: 'restart' };
      }
      return updateProgress;
    case 'install':
      return { ...updateProgress, state: 'ready', mode: 'restart' };
    case 'content':
      return { version: 'content-2', content: 'content-2', warning: '' };
    default:
      return { error: '未知的更新操作' };
  }
}

/** 服务端对话接口的最小实现：够验证"内容真的写回去了"就行。 */
function conversationsAPI(body: RequestBody): Response {
  const detail = (chat?: StoredChat): Record<string, unknown> => ({
    activeId: activeChat,
    items: chats.map(chatSummary),
    ...(chat ? { conversation: { ...chatSummary(chat), messages: chat.messages } satisfies Conversation } : {}),
  });
  const find = (id?: string) => chats.find(chat => chat.id === id);
  switch (body.action) {
    case 'list':
      return json(detail());
    case 'create': {
      const chat: StoredChat = { id: `chat-${chats.length + 1}`, title: '', lessonId: body.lessonId ?? '', updatedAt: ++clock, messages: [] };
      chats.unshift(chat);
      activeChat = chat.id;
      return json(detail(chat));
    }
    case 'get': {
      const chat = find(body.id);
      if (!chat) return json({ error: '对话不存在' }, 400);
      activeChat = chat.id;
      return json(detail(chat));
    }
    case 'append': {
      const chat = find(body.id);
      if (!chat) return json({ error: '对话不存在' }, 400);
      chat.messages.push({ role: body.role as 'user' | 'assistant', content: body.content ?? '', at: ++clock });
      if (!chat.title && body.role === 'user') chat.title = body.content ?? '';
      chat.updatedAt = ++clock;
      activeChat = chat.id;
      return json(detail(chat));
    }
    case 'rename': {
      const chat = find(body.id);
      if (!chat) return json({ error: '对话不存在' }, 400);
      chat.title = body.title ?? chat.title;
      return json(detail(chat));
    }
    case 'delete': {
      const index = chats.findIndex(chat => chat.id === body.id);
      if (index < 0) return json({ error: '对话不存在' }, 400);
      chats.splice(index, 1);
      if (activeChat === body.id) activeChat = chats[0]?.id ?? '';
      return json(detail(find(activeChat)));
    }
    default:
      return json({ error: '未知的对话操作' }, 400);
  }
}

function json(value: unknown, status = 200) {
  return new Response(JSON.stringify(value), { status, headers: { 'Content-Type': 'application/json' } });
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(done => { resolve = done; });
  return { promise, resolve };
}

beforeEach(() => {
  state = bootstrap();
  requests = [];
  chats = [];
  activeChat = '';
  clock = 1000;
  updateRequests = [];
  progressCalls = 0;
  updateProgress = { state: 'idle', received: 0, total: 0 };
  updateStatus = { current: '0.7.0', content: '' };
  sessionStorage.setItem('gostudy-token', 'test-session-token');
  runResponse = async () => json({ output: 'Hello, Go!\n', passed: true, durationMs: 3 } satisfies RunResult);
  vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
    const path = String(input).replace('/api/', '');
    const body = init?.body ? JSON.parse(String(init.body)) as RequestBody : undefined;
    requests.push({ path, body, authorization: new Headers(init?.headers).get('Authorization') });
    if (path === 'bootstrap') return json({ ...state, conversations: { activeId: activeChat, items: chats.map(chatSummary) } });
    if (path === 'run') return runResponse();
    if (path === 'conversations' && body) return conversationsAPI(body);
    if (path === 'progress' && body?.lessonId) {
      const id = body.lessonId;
      state.progress = {
        ...state.progress,
        lastLesson: id,
        drafts: { ...state.progress.drafts, [id]: body.code ?? '' },
        notes: { ...state.progress.notes, [id]: body.note ?? '' },
        completed: body.complete ? [...new Set([...state.progress.completed, id])] : state.progress.completed,
      };
      return json(state.progress);
    }
    if (path === 'settings' && body) {
      // 后端仅返回 hasKey，不把提交的密钥带回浏览器。
      state.settings = {
        baseUrl: body.baseUrl ?? state.settings.baseUrl,
        model: body.model ?? state.settings.model,
        hasKey: body.clearKey ? false : Boolean(body.apiKey || state.settings.hasKey),
        closeAction: state.settings.closeAction,
      } satisfies Settings;
      return json(state.settings);
    }
    if (path === 'close-behavior' && body) {
      state.settings = { ...state.settings, closeAction: body.action as Settings['closeAction'] };
      return json(state.settings);
    }
    if (path === 'uninstall') return json({ status: 'uninstalling' });
    // 页面关闭时的报信：正常走 sendBeacon，个别浏览器退回 fetch。
    if (path.startsWith('window-closed')) return json({ status: 'noted' });
    if (path === 'chat') return json({ reply: '连接成功，可以开始学习。' });
    if (path === 'update' && body) return json(updatesAPI(body));
    throw new Error(`未预期的请求：${path}`);
  });
});

async function startFirstLesson(user: ReturnType<typeof userEvent.setup>) {
  render(<App />);
  await user.click(await screen.findByRole('button', { name: '开始第一课' }));
  await user.click(screen.getByRole('tab', { name: '动手挑战' }));
}

async function replaceCode(user: ReturnType<typeof userEvent.setup>, code: string) {
  const editor = screen.getByRole('textbox', { name: 'Go 代码编辑器' });
  await user.clear(editor);
  // paste 保留 Go 的花括号与换行，避免被 user-event 当作按键描述。
  await user.click(editor);
  await user.paste(code);
}

/** 无密钥的本机模型：不需要联网就能把导师面板解锁到可提问状态。 */
async function useLocalModel(user: ReturnType<typeof userEvent.setup>) {
  render(<App />);
  await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
  const address = screen.getByRole('textbox', { name: /^API 地址/ });
  await user.clear(address);
  await user.type(address, 'http://127.0.0.1:11434/v1');
  const model = screen.getByRole('textbox', { name: /^模型名称/ });
  await user.clear(model);
  await user.type(model, 'local-study-model');
  await user.click(screen.getByRole('button', { name: '保存配置' }));
  await screen.findByText('模型配置已保存到本机。');
}

/** 从设置页回到第一课，并打开导师浮窗。 */
async function openTutor(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'GoStudy 首页' }));
  await user.click(screen.getByRole('button', { name: '开始第一课' }));
  await user.click(screen.getByRole('button', { name: '和导师一起想' }));
  return screen.getByRole('dialog', { name: 'AI 导师' });
}

/** 面板标题栏也会显示同一句标题，断言消息时只看对话区。 */
function chatText() {
  return document.querySelector('.chat-messages')?.textContent ?? '';
}

async function ask(user: ReturnType<typeof userEvent.setup>, dialog: HTMLElement, question: string) {
  await user.type(within(dialog).getByRole('textbox', { name: '向 AI 导师提问' }), question);
  await user.click(within(dialog).getByRole('button', { name: '发送消息' }));
  await within(dialog).findByText('连接成功，可以开始学习。');
}

describe('学习空间的关键交互', () => {
  it('从首页完成挑战：提交当前代码与正确答案，并更新进度及下一课', async () => {
    const user = userEvent.setup();
    await startFirstLesson(user);
    const complete = screen.getByRole('button', { name: '完成这一课' });
    expect(complete).toBeDisabled();

    await replaceCode(user, solution);
    await user.click(screen.getByRole('button', { name: '运行代码' }));
    expect(await screen.findByText(/挑战通过/)).toBeVisible();
    expect(requests.find(request => request.path === 'run')).toMatchObject({
      body: { lessonId: 'hello', code: solution }, authorization: 'Bearer test-session-token',
    });
    expect(complete).toBeDisabled();
    await user.click(screen.getByRole('radio', { name: /任意函数/ }));
    expect(screen.getByText(/再想一想/)).toBeVisible();
    expect(complete).toBeDisabled();
    await user.click(screen.getByRole('radio', { name: /main 函数/ }));
    expect(complete).toBeEnabled();
    await user.click(complete);

    expect(await screen.findByText('已掌握，做得很好。')).toBeVisible();
    expect(requests.find(request => request.path === 'progress' && request.body?.complete)).toMatchObject({
      body: { lessonId: 'hello', code: solution, note: '', complete: true, quizAnswer: 1 },
    });
    expect(screen.getByText('已完成 1 / 2 个练习')).toBeVisible();
    await user.click(screen.getByRole('button', { name: '开始下一课' }));
    expect(screen.getByRole('heading', { level: 1, name: '变量与自己的表达' })).toBeVisible();
    expect(screen.getByRole('textbox', { name: 'Go 代码编辑器' })).toHaveValue(lessons[1].starter);
  });

  it('运行失败可以重试，但代码改动后旧的运行结果不能解锁完成', async () => {
    const user = userEvent.setup();
    await startFirstLesson(user);
    await replaceCode(user, solution);
    runResponse = async () => json({ error: '本地引擎暂时不可用' }, 503);
    await user.click(screen.getByRole('button', { name: '运行代码' }));
    expect(await screen.findByText('本地引擎暂时不可用')).toBeVisible();
    expect(screen.getByRole('button', { name: '运行代码' })).toBeEnabled();

    const pending = deferred<Response>();
    runResponse = () => pending.promise;
    await user.click(screen.getByRole('button', { name: '运行代码' }));
    expect(screen.getByRole('button', { name: '运行中' })).toBeDisabled();
    expect(screen.getByRole('textbox', { name: 'Go 代码编辑器' })).toHaveAttribute('readonly');
    expect(screen.getByRole('button', { name: '还原初始代码' })).toBeDisabled();
    pending.resolve(json({ output: 'Hello, Go!\n', passed: true, durationMs: 3 } satisfies RunResult));
    await screen.findByRole('button', { name: '运行代码' });
    await user.click(screen.getByRole('radio', { name: /main 函数/ }));
    expect(screen.getByRole('button', { name: '完成这一课' })).toBeEnabled();
    await replaceCode(user, 'package main\n\nfunc main() { thisDoesNotCompile }');
    expect(screen.getByRole('button', { name: '完成这一课' })).toBeDisabled();
    expect(requests.filter(request => request.path === 'run')).toHaveLength(2);
  });

  it('未配置 AI 引导到设置，保存后只保留配置状态且支持测试和移除密钥', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: /AI 导师/ }));
    const dialog = screen.getByRole('dialog', { name: 'AI 导师' });
    expect(within(dialog).getByRole('textbox', { name: '向 AI 导师提问' })).toBeDisabled();
    expect(within(dialog).getByRole('button', { name: '发送消息' })).toBeDisabled();
    await user.click(within(dialog).getByRole('button', { name: '配置 AI 导师' }));
    // 抽屉带退场动画，等它真正卸载再断言关闭。
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(screen.getByRole('button', { name: '测试已保存的连接' })).toBeDisabled();

    const key = screen.getByLabelText(/^API Key/);
    await user.type(key, 'sk-private-test-key');
    await user.click(screen.getByRole('button', { name: '保存配置' }));
    expect(await screen.findByText('模型配置已保存到本机。')).toBeVisible();
    expect(requests.find(request => request.path === 'settings')).toMatchObject({
      body: { apiKey: 'sk-private-test-key', baseUrl: 'https://api.example.com/v1', model: 'study-model', clearKey: false },
    });
    expect(key).toHaveValue('');
    expect(state.settings).not.toHaveProperty('apiKey');
    expect(JSON.stringify({ ...localStorage })).not.toContain('sk-private-test-key');
    expect(JSON.stringify({ ...sessionStorage })).not.toContain('sk-private-test-key');
    expect(screen.getByText('已配置', { exact: true })).toBeVisible();
    await user.click(screen.getByRole('button', { name: '测试已保存的连接' }));
    expect(await screen.findByText('连接成功，可以开始学习。')).toBeVisible();
    expect(requests.find(request => request.path === 'chat')?.body).not.toHaveProperty('apiKey');
    await user.click(screen.getByRole('button', { name: '移除 Key' }));
    expect(await screen.findByText('本机保存的 API Key 已移除。')).toBeVisible();
    expect(requests.filter(request => request.path === 'settings').at(-1)?.body).toMatchObject({ clearKey: true });
    expect(screen.getByRole('button', { name: '测试已保存的连接' })).toBeDisabled();
  });

  it('无密钥的本地模型保存后可以测试连接并携带当前代码提问', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    const address = screen.getByRole('textbox', { name: /^API 地址/ });
    await user.clear(address);
    await user.type(address, 'http://127.0.0.1:11434/v1');
    const model = screen.getByRole('textbox', { name: /^模型名称/ });
    await user.clear(model);
    await user.type(model, 'local-study-model');
    await user.click(screen.getByRole('button', { name: '保存配置' }));
    await screen.findByText('模型配置已保存到本机。');
    expect(state.settings.hasKey).toBe(false);
    expect(requests.find(request => request.path === 'settings')?.body).not.toHaveProperty('apiKey');
    expect(screen.getByRole('button', { name: '测试已保存的连接' })).toBeEnabled();
    await user.click(screen.getByRole('button', { name: '测试已保存的连接' }));
    await screen.findByText('连接成功，可以开始学习。');

    await user.click(screen.getByRole('button', { name: 'GoStudy 首页' }));
    await user.click(screen.getByRole('button', { name: '开始第一课' }));
    await replaceCode(user, solution);
    await user.click(screen.getByRole('button', { name: '和导师一起想' }));
    const dialog = screen.getByRole('dialog', { name: 'AI 导师' });
    const question = within(dialog).getByRole('textbox', { name: '向 AI 导师提问' });
    expect(question).toBeEnabled();
    await user.type(question, '为什么要使用 main？');
    await user.click(within(dialog).getByRole('button', { name: '发送消息' }));
    expect(await within(dialog).findByText('连接成功，可以开始学习。')).toBeVisible();
    expect(requests.filter(request => request.path === 'chat').at(-1)?.body).toMatchObject({
      lessonId: 'hello', code: solution, message: '为什么要使用 main？', history: [],
    });
  });

  it('离开课程时保留刚编辑的代码和笔记，笔记列表可返回对应课程', async () => {
    const user = userEvent.setup();
    await startFirstLesson(user);
    await replaceCode(user, solution);
    await user.click(screen.getByRole('tab', { name: '我的笔记' }));
    const note = '入口是 main；输出使用 fmt.Println。';
    await user.click(screen.getByRole('textbox', { name: '课程笔记' }));
    await user.paste(note);
    // 不等待自动保存提示，模拟用户写完后立即切换页面。
    await user.click(screen.getByRole('button', { name: /^我的笔记/ }));
    expect(await screen.findByText(note)).toBeVisible();
    await waitFor(() => expect(requests.find(request => request.path === 'progress' && request.body?.note === note)?.body).toMatchObject({
      lessonId: 'hello', code: solution, note,
    }), { timeout: 2000 });
    await user.click(screen.getByRole('button', { name: '回到这一课' }));
    expect(screen.getByRole('textbox', { name: 'Go 代码编辑器' })).toHaveValue(solution);
    await user.click(screen.getByRole('tab', { name: '我的笔记' }));
    expect(screen.getByRole('textbox', { name: '课程笔记' })).toHaveValue(note);
  });

  it('从 bootstrap 恢复完成进度、上次课程、代码草稿与笔记', async () => {
    const user = userEvent.setup();
    const savedNote = '短变量声明用于声明并赋值。';
    state = bootstrap({ completed: ['hello'], lastLesson: 'variables', drafts: { variables: secondDraft }, notes: { variables: savedNote } });
    render(<App />);
    expect(await screen.findByText('已完成 1 / 2 个练习')).toBeVisible();
    await user.click(screen.getByRole('button', { name: '继续我的学习' }));
    expect(screen.getByRole('heading', { level: 1, name: '变量与自己的表达' })).toBeVisible();
    expect(screen.getByRole('textbox', { name: 'Go 代码编辑器' })).toHaveValue(secondDraft);
    await user.click(screen.getByRole('tab', { name: '我的笔记' }));
    expect(screen.getByRole('textbox', { name: '课程笔记' })).toHaveValue(savedNote);
    await user.click(screen.getByRole('button', { name: /^我的笔记/ }));
    expect(screen.getByRole('heading', { name: '变量与自己的表达' })).toBeVisible();
    expect(screen.getByText(savedNote)).toBeVisible();
    expect(screen.getByRole('button', { name: '导出笔记' })).toBeEnabled();
    expect(requests.filter(request => request.path === 'progress')).toHaveLength(0);
    // Android 内嵌浏览器通过可复制的预览导出，桌面浏览器走下载。
    vi.spyOn(navigator, 'userAgent', 'get').mockReturnValue('GoStudyAndroid');
    await user.click(screen.getByRole('button', { name: '导出笔记' }));
    const exported = within(screen.getByRole('dialog', { name: '导出学习笔记' }));
    expect(exported.getByRole('textbox', { name: '导出的 Markdown 笔记' })).toHaveValue(
      `# 我的 GoStudy 学习笔记\n\n## 变量与自己的表达\n\n${savedNote}\n`,
    );
    await user.click(exported.getByRole('button', { name: '复制笔记' }));
    expect(await screen.findByText('学习笔记已复制。')).toBeVisible();
    expect(await navigator.clipboard.readText()).toContain(savedNote);
  });

  it('导师浮窗不挡页面：面板开着也能改代码、跑程序，收起后再打开问题和回答都还在', async () => {
    const user = userEvent.setup();
    await useLocalModel(user);
    await openTutor(user);
    await user.click(screen.getByRole('tab', { name: '动手挑战' }));

    // 面板不是模态对话框：没有遮罩，页面其余部分照常可用。
    expect(document.querySelector('.assistant-scrim')).toBeNull();
    expect(document.querySelector('.assistant-overlay')).toBeNull();
    await replaceCode(user, solution);
    await user.click(screen.getByRole('button', { name: '运行代码' }));
    expect(await screen.findByText(/挑战通过/)).toBeVisible();
    expect(screen.getByRole('dialog', { name: 'AI 导师' })).toBeVisible();

    const dialog = screen.getByRole('dialog', { name: 'AI 导师' });
    await ask(user, dialog, '为什么要用 main？');
    // 一条提问先写进本机记录，再去问模型，回答也写回去。
    expect(requests.filter(request => request.path === 'conversations').map(request => request.body?.action)).toEqual([
      'create', 'append', 'append',
    ]);
    expect(requests.filter(request => request.path === 'conversations')[1]?.body).toMatchObject({
      lessonId: 'hello', role: 'user', content: '为什么要用 main？',
    });
    expect(chats.flatMap(chat => chat.messages.map(message => `${message.role}:${message.content}`))).toEqual([
      'user:为什么要用 main？', 'assistant:连接成功，可以开始学习。',
    ]);

    // 收起面板只是收起视图，再打开时问题和回答都在，草稿问题也不会丢。
    await user.click(within(dialog).getByRole('button', { name: '关闭 AI 导师' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: '和导师一起想' }));
    const reopened = screen.getByRole('dialog', { name: 'AI 导师' });
    expect(chatText()).toContain('为什么要用 main？');
    expect(chatText()).toContain('连接成功，可以开始学习。');
    expect(within(reopened).getByRole('textbox', { name: '向 AI 导师提问' })).toHaveValue('');
  });

  it('重启应用后恢复上次的对话，不重问也能接着聊', async () => {
    const user = userEvent.setup();
    chats = [{
      id: 'chat-restored', title: '值语义和指针语义', lessonId: 'hello', updatedAt: 1000,
      messages: [
        { role: 'user', content: '值语义是什么？', at: 1000 },
        { role: 'assistant', content: '先看赋值之后能不能改。', at: 1001 },
      ],
    }];
    activeChat = 'chat-restored';
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '开始第一课' }));
    await user.click(screen.getByRole('button', { name: '和导师一起想' }));
    const dialog = screen.getByRole('dialog', { name: 'AI 导师' });
    // bootstrap 只带摘要，打开面板时才按 id 取回正文。
    expect(await within(dialog).findByText('值语义是什么？')).toBeVisible();
    expect(within(dialog).getByText('先看赋值之后能不能改。')).toBeVisible();
    expect(requests.filter(request => request.path === 'conversations')).toMatchObject([
      { body: { action: 'get', id: 'chat-restored' } },
    ]);
  });

  it('设置页能检查、下载并安装新版本，进度按轮询推进', async () => {
    updateStatus = {
      current: '0.7.0', content: '', checked: 1000,
      app: { version: '0.8.0', released: '2026-10-20', notes: '修好了下载与校验流程', size: 12582912, asset: 'GoStudy-linux-amd64' },
    };
    const user = userEvent.setup();
    render(<App />);
    // 顶栏设置按钮带小红点，提示有新版本
    const gear = await screen.findByRole('button', { name: '偏好与 AI 设置' });
    expect(gear).toHaveAttribute('title', '有新版本可以更新');
    await user.click(gear);

    // 挂载时的自动检查带 auto 标记，手动检查不带
    expect(updateRequests.some(request => request.action === 'check' && request.auto === true)).toBe(true);
    expect(await screen.findByText('有新版本 0.8.0')).toBeVisible();
    expect(screen.getByText('新版本 0.8.0 · 2026-10-20')).toBeVisible();
    expect(screen.getByText('修好了下载与校验流程')).toBeVisible();
    expect(screen.getByText('12.0 MB')).toBeVisible();

    await user.click(screen.getByRole('button', { name: /检查更新/ }));
    await waitFor(() => expect(updateRequests.some(request => request.action === 'check' && request.auto === false)).toBe(true));

    await user.click(screen.getByRole('button', { name: /下载并安装 0.8.0/ }));
    expect(updateRequests.some(request => request.action === 'download' && request.prefer !== undefined)).toBe(true);
    // 轮询把状态推进到"已就绪"，再点一次就发起安装
    await user.click(await screen.findByRole('button', { name: '立即重启并更新' }, { timeout: 3000 }));
    expect(updateRequests.some(request => request.action === 'install')).toBe(true);
  });

  it('关掉自动检查后不再自动联网，手动检查仍然可用', async () => {
    localStorage.setItem('gostudy-display', JSON.stringify({ reduceMotion: null, editorSize: 14, autoUpdate: false }));
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    expect(updateRequests.filter(request => request.action === 'check')).toHaveLength(0);
    await user.click(screen.getByRole('button', { name: /检查更新/ }));
    await waitFor(() => expect(updateRequests.some(request => request.action === 'check')).toBe(true));
    const toggle = screen.getByRole('switch', { name: '每天自动检查一次' });
    expect(toggle).toHaveAttribute('aria-checked', 'false');
  });

  it('课程内容可以单独更新，不必等新版本', async () => {
    updateStatus = {
      current: '0.7.0', content: 'content-1', checked: 1000,
      contentUpdate: { version: 'content-2', size: 2048, current: 'content-1' },
    };
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    expect(await screen.findByText('课程内容有更新')).toBeVisible();
    expect(screen.getByText(/当前内容是 content-1，新的是 content-2/)).toBeVisible();
    await user.click(screen.getByRole('button', { name: '更新课程内容' }));
    await waitFor(() => expect(updateRequests.some(request => request.action === 'content')).toBe(true));
  });

  it('对话历史：新建、切换、重命名与删除都在面板里完成', async () => {    const user = userEvent.setup();
    await useLocalModel(user);
    const dialog = await openTutor(user);
    await ask(user, dialog, '第一段对话');

    // 历史列表按最近使用排序，标题取第一条提问（按钮名字以标题开头，后面跟课程与条数）。
    await user.click(within(dialog).getByRole('button', { name: '对话历史' }));
    expect(within(dialog).getByRole('button', { name: /^第一段对话/ })).toBeVisible();

    // 新建：面板回到欢迎态，历史里多出一段。
    await user.click(within(dialog).getByRole('button', { name: '新建对话' }));
    expect(within(dialog).getByText('先自己跑一遍，再来问为什么。')).toBeVisible();
    await ask(user, dialog, '第二段对话');
    await user.click(within(dialog).getByRole('button', { name: '对话历史' }));
    expect(within(dialog).getByRole('button', { name: /^第一段对话/ })).toBeVisible();
    expect(within(dialog).getByRole('button', { name: /^第二段对话/ })).toBeVisible();

    // 切换：点回第一段，问题与回答都回来。
    await user.click(within(dialog).getByRole('button', { name: /^第一段对话/ }));
    await waitFor(() => expect(chatText()).toContain('第一段对话'));
    expect(chatText()).toContain('连接成功，可以开始学习。');

    // 重命名
    await user.click(within(dialog).getByRole('button', { name: '对话历史' }));
    await user.click(within(dialog).getByRole('button', { name: '重命名 第一段对话' }));
    const name = within(dialog).getByRole('textbox', { name: '对话名称' });
    await user.clear(name);
    await user.type(name, '入口与变量');
    await user.click(within(dialog).getByRole('button', { name: '保存' }));
    expect(await within(dialog).findByRole('button', { name: /^入口与变量/ })).toBeVisible();
    expect(requests.filter(request => request.path === 'conversations').at(-1)?.body).toMatchObject({
      action: 'rename', title: '入口与变量',
    });

    // 删除要确认一次，删完落到剩下的一段。
    await user.click(within(dialog).getByRole('button', { name: '删除 第二段对话' }));
    await user.click(within(dialog).getByRole('button', { name: '确认删除' }));
    await waitFor(() => expect(within(dialog).queryByRole('button', { name: /^第二段对话/ })).not.toBeInTheDocument());
    expect(within(dialog).getByRole('button', { name: /^入口与变量/ })).toBeVisible();
    expect(chats.map(chat => chat.title)).toEqual(['入口与变量']);
  });

  it('Windows 上可以选关闭窗口后的行为，选择写回引擎', async () => {
    state.runtime = { platform: 'windows', engine: '受限解释器', version: '0.7.0', closePrompt: true };
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    const group = screen.getByRole('radiogroup', { name: '关闭应用窗口时' });
    expect(within(group).getByRole('radio', { name: '每次问我' })).toBeChecked();
    expect(screen.getByText(/窗口关掉之后弹一个询问框/)).toBeVisible();

    await user.click(within(group).getByRole('radio', { name: '后台静默运行' }));
    expect(await screen.findByText('关闭方式已保存。')).toBeVisible();
    expect(requests.find(request => request.path === 'close-behavior')).toMatchObject({
      body: { action: 'background' }, authorization: 'Bearer test-session-token',
    });
    expect(within(group).getByRole('radio', { name: '后台静默运行' })).toBeChecked();
    // 关窗口之后进程留着，说明文字要跟上。
    expect(screen.getByText(/下次打开 GoStudy 直接连回它/)).toBeVisible();
  });

  it('非 Windows 平台不提供关闭方式选项，按直接退出说明', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    expect(screen.queryByRole('radiogroup', { name: '关闭应用窗口时' })).not.toBeInTheDocument();
    expect(screen.getByText(/这个平台关掉窗口之后会直接结束进程/)).toBeVisible();
    expect(requests.every(request => request.path !== 'close-behavior')).toBe(true);
  });

  it('卸载要先确认一次，确认后进入卸载提示页', async () => {
    state.runtime = { platform: 'windows', engine: '受限解释器', version: '0.7.0', closePrompt: true, canUninstall: true };
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    await user.click(screen.getByRole('button', { name: '卸载' }));

    const dialog = await screen.findByRole('dialog', { name: '卸载 GoStudy' });
    expect(within(dialog).getByText(/卸载不会动它们/)).toBeVisible();
    // 先关掉再打开一次，确认没有意外发起卸载。
    await user.click(within(dialog).getByRole('button', { name: '先不卸载' }));
    await waitFor(() => expect(screen.queryByRole('dialog', { name: '卸载 GoStudy' })).not.toBeInTheDocument());
    expect(requests.every(request => request.path !== 'uninstall')).toBe(true);

    await user.click(screen.getByRole('button', { name: '卸载' }));
    await user.click(await within(await screen.findByRole('dialog', { name: '卸载 GoStudy' })).findByRole('button', { name: '确认卸载' }));
    await waitFor(() => expect(requests.some(request => request.path === 'uninstall')).toBe(true));
    expect(await screen.findByText('正在卸载 GoStudy')).toBeVisible();
  });

  it('免安装版没有卸载入口', async () => {
    state.runtime = { platform: 'windows', engine: '受限解释器', version: '0.7.0', closePrompt: true };
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole('button', { name: '偏好与 AI 设置' }));
    expect(screen.queryByRole('button', { name: '卸载' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '退出 GoStudy' })).toBeVisible();
  });

  it('关窗口时用 sendBeacon 报信，带不出 request 头的老浏览器退回 fetch', async () => {
    const beacon = vi.fn().mockReturnValue(true);
    Object.defineProperty(navigator, 'sendBeacon', { configurable: true, value: beacon });
    try {
      render(<App />);
      await screen.findByRole('button', { name: '开始第一课' });
      window.dispatchEvent(new Event('pagehide'));
      expect(beacon).toHaveBeenCalledTimes(1);
      expect(String(beacon.mock.calls[0][0])).toBe('/api/window-closed?token=test-session-token');

      // sendBeacon 用不了时（个别内嵌浏览器）走 keepalive fetch。
      beacon.mockReturnValue(false);
      window.dispatchEvent(new Event('pagehide'));
      await waitFor(() => expect(requests.some(request => request.path.startsWith('window-closed'))).toBe(true));
    } finally {
      Reflect.deleteProperty(navigator, 'sendBeacon');
    }
  });
});
