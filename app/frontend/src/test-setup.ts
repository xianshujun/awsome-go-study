import '@testing-library/jest-dom/vitest';
import { act, cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

// jsdom 没有布局引擎，这些滚动动作不影响被测交互。
Object.defineProperty(window, 'scrollTo', { configurable: true, value: vi.fn() });
Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: vi.fn() });

// jsdom 也没有 matchMedia。统一回答"用户要求减少动效"：界面直接渲染最终状态，
// 断言可见性的元素不会卡在透明度 0；动效本身交给真实浏览器验收。
Object.defineProperty(window, 'matchMedia', {
  configurable: true,
  value: (query: string) => ({
    matches: query.includes('prefers-reduced-motion'),
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(() => false),
  }),
});

// 布局与视口观察器在 jsdom 中不存在；动效关闭后不会被真正使用，这里只防构造报错。
class NoopObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return [];
  }
}
globalThis.ResizeObserver ??= NoopObserver as unknown as typeof ResizeObserver;
globalThis.IntersectionObserver ??= NoopObserver as unknown as typeof IntersectionObserver;

afterEach(async () => {
  // 卸载时会提交未保存草稿，先等这些微任务完成再还原 fetch。
  await act(async () => { cleanup(); });
  sessionStorage.clear();
  localStorage.clear();
  vi.restoreAllMocks();
});
