import '@testing-library/jest-dom/vitest';
import { act, cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

// jsdom 没有布局引擎，这些滚动动作不影响被测交互。
Object.defineProperty(window, 'scrollTo', { configurable: true, value: vi.fn() });
Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: vi.fn() });

afterEach(async () => {
  // 卸载时会提交未保存草稿，先等这些微任务完成再还原 fetch。
  await act(async () => { cleanup(); });
  sessionStorage.clear();
  localStorage.clear();
  vi.restoreAllMocks();
});
