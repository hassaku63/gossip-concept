# Phase 0 セットアップガイド

## プロジェクト初期化

### 1. Vite + React + TypeScriptプロジェクトの作成

```bash
# プロジェクトの初期化
npm create vite@latest . -- --template react-ts

# 依存関係のインストール
npm install

# 追加の依存関係
npm install react-force-graph-2d
npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event
npm install -D @vitest/ui @vitest/coverage-v8
npm install -D eslint-plugin-testing-library eslint-plugin-jest-dom
```

### 2. プロジェクト構造の作成

```bash
# ディレクトリ構造の作成
mkdir -p src/models
mkdir -p src/components/{ControlPanel,GraphVisualizer,App}
mkdir -p src/services
mkdir -p src/types
mkdir -p src/utils
mkdir -p src/hooks
mkdir -p tests/integration
mkdir -p docs/phase0
```

## 設定ファイル

### vite.config.ts

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'src/test/',
        '*.config.*',
        'src/main.tsx',
        'src/vite-env.d.ts',
      ],
      thresholds: {
        branches: 80,
        functions: 80,
        lines: 80,
        statements: 80,
      },
    },
  },
});
```

### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "types": ["vitest/globals", "@testing-library/jest-dom"]
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

### .eslintrc.json

```json
{
  "root": true,
  "env": { "browser": true, "es2020": true },
  "extends": [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended",
    "plugin:testing-library/react",
    "plugin:jest-dom/recommended"
  ],
  "ignorePatterns": ["dist", ".eslintrc.cjs"],
  "parser": "@typescript-eslint/parser",
  "plugins": ["react-refresh", "testing-library", "jest-dom"],
  "rules": {
    "react-refresh/only-export-components": [
      "warn",
      { "allowConstantExport": true }
    ],
    "@typescript-eslint/no-unused-vars": [
      "error",
      { "argsIgnorePattern": "^_" }
    ]
  }
}
```

### .prettierrc

```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2,
  "useTabs": false,
  "arrowParens": "always",
  "endOfLine": "lf"
}
```

### src/test/setup.ts

```typescript
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// 各テスト後にクリーンアップ
afterEach(() => {
  cleanup();
});

// グローバルモックの設定
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// react-force-graphのモック
vi.mock('react-force-graph-2d', () => ({
  default: vi.fn(() => null),
}));
```

## package.jsonスクリプト

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "lint:fix": "eslint . --ext ts,tsx --fix",
    "preview": "vite preview",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage",
    "test:coverage:check": "vitest run --coverage --coverage.thresholds.branches=80",
    "test:watch": "vitest watch",
    "type-check": "tsc --noEmit",
    "format": "prettier --write .",
    "format:check": "prettier --check ."
  }
}
```

## 初期ファイルテンプレート

### src/types/index.ts

```typescript
export type NodeState = 'Red' | 'Green' | 'Blue';

export interface NodePosition {
  x: number;
  y: number;
}

export interface NodeData {
  id: string;
  state: NodeState;
  position: NodePosition;
}

export interface EdgeData {
  source: string;
  target: string;
  active?: boolean;
}

export interface SimulationConfig {
  nodeCount: number;
  maxRounds: number;
  speed: number; // rounds per second
}

export interface SimulationState {
  isRunning: boolean;
  currentRound: number;
  nodes: NodeData[];
  edges: EdgeData[];
}
```

### src/models/Node.ts

```typescript
import { NodeState, NodePosition } from '../types';

export class Node {
  private _state: NodeState;
  private _position: NodePosition;
  
  constructor(
    public readonly id: string,
    initialState: NodeState = 'Red',
    position?: NodePosition
  ) {
    this._state = initialState;
    this._position = position || { x: 0, y: 0 };
  }
  
  get state(): NodeState {
    return this._state;
  }
  
  get position(): NodePosition {
    return { ...this._position };
  }
  
  setState(newState: NodeState): void {
    if (!this.isValidState(newState)) {
      throw new Error(`Invalid state: ${newState}`);
    }
    this._state = newState;
  }
  
  private isValidState(state: string): state is NodeState {
    return ['Red', 'Green', 'Blue'].includes(state);
  }
}
```

### src/models/Node.test.ts

```typescript
import { describe, it, expect } from 'vitest';
import { Node } from './Node';

describe('Node', () => {
  describe('初期化', () => {
    it('IDと初期状態を設定できる', () => {
      const node = new Node('node-1');
      expect(node.id).toBe('node-1');
      expect(node.state).toBe('Red');
    });
    
    it('カスタム初期状態を設定できる', () => {
      const node = new Node('node-1', 'Green');
      expect(node.state).toBe('Green');
    });
    
    it('位置情報を持つ', () => {
      const node = new Node('node-1', 'Red', { x: 10, y: 20 });
      expect(node.position).toEqual({ x: 10, y: 20 });
    });
  });
  
  describe('状態変更', () => {
    it('有効な状態に変更できる', () => {
      const node = new Node('node-1');
      node.setState('Green');
      expect(node.state).toBe('Green');
      
      node.setState('Blue');
      expect(node.state).toBe('Blue');
    });
    
    it('無効な状態への変更は拒否される', () => {
      const node = new Node('node-1');
      expect(() => {
        node.setState('Yellow' as any);
      }).toThrow('Invalid state: Yellow');
    });
  });
});
```

## 開発フロー

### 1. TDDサイクルの実行

```bash
# テストをウォッチモードで実行
npm run test:watch

# 1. Redフェーズ: 失敗するテストを書く
# 2. Greenフェーズ: テストが通る最小限のコードを書く
# 3. Refactorフェーズ: コードを改善する

# カバレッジ確認
npm run test:coverage
```

### 2. 型チェックとリント

```bash
# TypeScriptの型チェック
npm run type-check

# ESLintの実行
npm run lint

# 自動修正
npm run lint:fix
```

### 3. フォーマット

```bash
# コードフォーマット
npm run format

# フォーマットチェック
npm run format:check
```

## Git設定

### .gitignore

```
# Logs
logs
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*
pnpm-debug.log*
lerna-debug.log*

node_modules
dist
dist-ssr
*.local
coverage

# Editor directories and files
.vscode/*
!.vscode/extensions.json
.idea
.DS_Store
*.suo
*.ntvs*
*.njsproj
*.sln
*.sw?
```

### pre-commitフック設定

```bash
# husky と lint-staged のインストール
npm install -D husky lint-staged

# husky の初期化
npx husky-init && npm install

# pre-commit フックの設定
npx husky set .husky/pre-commit "npx lint-staged"
```

### lint-staged設定（package.json）

```json
{
  "lint-staged": {
    "*.{ts,tsx}": [
      "eslint --fix",
      "prettier --write",
      "vitest related --run"
    ],
    "*.{json,md,css}": [
      "prettier --write"
    ]
  }
}
```

## VS Code推奨設定

### .vscode/settings.json

```json
{
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "typescript.tsdk": "node_modules/typescript/lib",
  "testing.automaticallyOpenPeekView": "never"
}
```

### .vscode/extensions.json

```json
{
  "recommendations": [
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "vitest.explorer",
    "ms-vscode.vscode-typescript-next"
  ]
}
```

## 次のステップ

セットアップ完了後の実装順序：

1. **基本モデルの実装**
   - Nodeクラスの完成
   - GossipProtocolクラスの実装

2. **UIコンポーネントの実装**
   - ControlPanelコンポーネント
   - GraphVisualizerコンポーネント

3. **シミュレーションロジック**
   - SimulationServiceの実装
   - 状態管理の実装

4. **統合とテスト**
   - Appコンポーネントでの統合
   - E2Eテストの追加