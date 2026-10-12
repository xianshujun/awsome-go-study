import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { App } from './main';
import type { Bootstrap, Lesson, Progress, RunResult, Settings } from './api';

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
    settings: { baseUrl: 'https://api.example.com/v1', model: 'study-model', hasKey: false },
    runtime: { platform: 'linux/amd64', engine: '受限解释器' },
  };
}

type RequestBody = {
  lessonId?: string; code?: string; note?: string; complete?: boolean; quizAnswer?: number;
  baseUrl?: string; model?: string; apiKey?: string; clearKey?: boolean; message?: string;
};
type RecordedRequest = { path: string; body?: RequestBody; authorization: string | null };
let state: Bootstrap;
let requests: RecordedRequest[];
let runResponse: () => Promise<Response>;

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
  sessionStorage.setItem('gostudy-token', 'test-session-token');
  runResponse = async () => json({ output: 'Hello, Go!\n', passed: true, durationMs: 3 } satisfies RunResult);
  vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
    const path = String(input).replace('/api/', '');
    const body = init?.body ? JSON.parse(String(init.body)) as RequestBody : undefined;
    requests.push({ path, body, authorization: new Headers(init?.headers).get('Authorization') });
    if (path === 'bootstrap') return json(state);
    if (path === 'run') return runResponse();
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
      } satisfies Settings;
      return json(state.settings);
    }
    if (path === 'chat') return json({ reply: '连接成功，可以开始学习。' });
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
});
