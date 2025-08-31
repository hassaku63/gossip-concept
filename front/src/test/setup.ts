import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

// 各テスト後にクリーンアップ
afterEach(() => {
  cleanup();
});

// グローバルモックの設定
(globalThis as any).ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// react-force-graph-2dのモック
vi.mock('react-force-graph-2d', () => ({
  default: vi.fn(() => null),
}));