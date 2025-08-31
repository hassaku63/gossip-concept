# Phase 0 テスト戦略

## TDDアプローチ

t-wadaスタイルのTDDサイクルに従い、以下の順序で実装を進めます：

```
1. Red: 失敗するテストを書く
2. Green: テストが通る最小限のコードを書く
3. Refactor: テストが通る状態を保ちながら設計を改善
```

## テスト分類

### 1. ユニットテスト

#### ドメインモデルのテスト

**Node.test.ts**
```typescript
describe('Node', () => {
  describe('初期化', () => {
    it('IDと初期状態を設定できる');
    it('デフォルトの状態はRedである');
    it('位置情報を持つ');
  });

  describe('状態変更', () => {
    it('Red→Green→Blueの順で状態を変更できる');
    it('無効な状態への変更は拒否される');
  });

  describe('通信', () => {
    it('他のノードと情報交換できる');
    it('ランダムなノードを選択できる');
  });
});
```

**GossipProtocol.test.ts**
```typescript
describe('GossipProtocol', () => {
  describe('初期化', () => {
    it('指定された数のノードを生成する');
    it('ノードの初期配置が適切である');
  });

  describe('Round実行', () => {
    it('1Round内で各ノードが通信を行う');
    it('状態が適切に伝播する');
    it('指定されたRound数で停止する');
  });

  describe('リセット', () => {
    it('すべてのノードが初期状態に戻る');
    it('Round数が0にリセットされる');
  });
});
```

### 2. コンポーネントテスト

**ControlPanel.test.tsx**
```typescript
describe('ControlPanel', () => {
  describe('表示', () => {
    it('開始/停止ボタンが表示される');
    it('リセットボタンが表示される');
    it('ノード数入力が表示される');
    it('速度調整スライダーが表示される');
    it('Round数入力が表示される');
  });

  describe('インタラクション', () => {
    it('開始ボタンクリックでシミュレーション開始');
    it('停止ボタンクリックでシミュレーション停止');
    it('リセットボタンクリックで初期状態に戻る');
    it('ノード数変更が反映される');
    it('速度変更が反映される');
  });

  describe('バリデーション', () => {
    it('ノード数が1-300の範囲内である');
    it('速度が1-100の範囲内である');
    it('Round数が正の整数である');
  });
});
```

**GraphVisualizer.test.tsx**
```typescript
describe('GraphVisualizer', () => {
  describe('レンダリング', () => {
    it('指定された数のノードが描画される');
    it('ノードの色が状態に応じて変化する');
    it('エッジが適切に描画される');
  });

  describe('アニメーション', () => {
    it('通信時にエッジがハイライトされる');
    it('状態変化時にノードの色が変わる');
  });

  describe('パフォーマンス', () => {
    it('10ノードで60fps以上を維持');
    it('300ノードで30fps以上を維持');
  });
});
```

### 3. 統合テスト

**simulation.test.ts**
```typescript
describe('Gossipシミュレーション統合テスト', () => {
  describe('エンドツーエンドシナリオ', () => {
    it('シミュレーション開始から終了まで正常動作');
    it('途中停止と再開が正常動作');
    it('リセット後に新しいシミュレーションを開始できる');
  });

  describe('大規模ノードテスト', () => {
    it('300ノードでのシミュレーションが完走する');
    it('メモリリークが発生しない');
  });
});
```

## テスト実装の具体例

### Step 1: Red（失敗するテストを書く）

```typescript
// Node.test.ts
import { describe, it, expect } from 'vitest';
import { Node } from './Node';

describe('Node', () => {
  it('IDと初期状態を設定できる', () => {
    const node = new Node('node-1');
    expect(node.id).toBe('node-1');
    expect(node.state).toBe('Red');
  });
});
```

### Step 2: Green（最小限の実装）

```typescript
// Node.ts
export class Node {
  constructor(public id: string) {
    this.state = 'Red';
  }
  
  public state: string;
}
```

### Step 3: Refactor（設計改善）

```typescript
// Node.ts
export type NodeState = 'Red' | 'Green' | 'Blue';

export interface NodePosition {
  x: number;
  y: number;
}

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

## モックとスタブ

### 外部依存のモック化

```typescript
// __mocks__/react-force-graph.ts
export const ForceGraph2D = vi.fn(() => ({
  render: vi.fn(),
  d3Force: vi.fn(),
  zoom: vi.fn(),
}));
```

### タイマーのモック

```typescript
// SimulationService.test.ts
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('SimulationService', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  
  afterEach(() => {
    vi.useRealTimers();
  });
  
  it('指定された間隔でRoundが進行する', () => {
    const service = new SimulationService();
    service.start({ speed: 5 }); // 5 rounds/sec = 200ms interval
    
    expect(service.currentRound).toBe(0);
    
    vi.advanceTimersByTime(200);
    expect(service.currentRound).toBe(1);
    
    vi.advanceTimersByTime(200);
    expect(service.currentRound).toBe(2);
  });
});
```

## カバレッジ目標

### Phase 0の目標
- ユニットテスト: 90%以上
- 統合テスト: 80%以上
- 全体: 85%以上

### 測定コマンド

```bash
# カバレッジレポート生成
npm run test:coverage

# カバレッジ閾値チェック
npm run test:coverage:check
```

## CI/CD設定

### GitHub Actions

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run type check
        run: npm run type-check
      
      - name: Run lint
        run: npm run lint
      
      - name: Run tests
        run: npm run test:coverage
      
      - name: Check coverage threshold
        run: npm run test:coverage:check
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/lcov.info
```

## テストのベストプラクティス

### 1. テストの命名規則
- 日本語で具体的な振る舞いを記述
- `〜できる`、`〜される`の形式を使用

### 2. テストの構造
- Arrange-Act-Assert（AAA）パターンを使用
- 各テストは独立して実行可能

### 3. テストデータ
- テスト用のファクトリー関数を作成
- 境界値テストを含める

### 4. 非同期テスト
- async/awaitを使用
- タイムアウトを適切に設定

## トラブルシューティング

### よくある問題と解決策

1. **タイマーテストが不安定**
   - 解決: `vi.useFakeTimers()`を使用

2. **DOM要素が見つからない**
   - 解決: `waitFor`を使用して要素の出現を待つ

3. **モックが機能しない**
   - 解決: モックファイルの配置とパスを確認

## 次のフェーズへの準備

Phase 0のテスト基盤は、将来の機能拡張でも再利用されます：
- テストヘルパー関数の整備
- カスタムマッチャーの作成
- E2Eテストフレームワークの導入検討