# Phase 0 アーキテクチャ設計

## 実装方針

**このフロントエンド実装は、バックエンドAPIサーバーを使用せず、ブラウザ上で完結する純粋なクライアントサイドアプリケーションです。** 親ディレクトリのGo実装の概念を参考にしながら、JavaScriptで独立したシミュレーションを実行します。

### Go実装の主要概念

| Go実装 | フロントエンド実装 | 説明 |
|--------|-------------------|------|
| `Node` struct | `Node` class | ノードの状態管理 |
| `GossipMessage` | `Message` interface | ゴシップメッセージ |
| `selectRandomPeer()` | `selectRandomNode()` | ランダムピア選択 |
| `SendGossip()` | `propagate()` | ゴシップ送信 |
| `HandleGossipMessage()` | `receiveMessage()` | メッセージ受信処理 |
| Round-based execution | Timer-based simulation | ラウンドベースの実行 |

## システムアーキテクチャ

### クライアントサイド完結型アーキテクチャ

```
ブラウザ環境のみで動作（APIサーバー不要）
┌─────────────────────────────────────────────────┐
│                   App Component                  │
│  ┌───────────────────────────────────────────┐  │
│  │       SimulationContext (React Context)   │  │
│  │  - nodes: Node[] (メモリ上で管理)         │  │
│  │  - edges: Edge[] (計算で生成)             │  │
│  │  - currentRound: number                   │  │
│  │  - isRunning: boolean                     │  │
│  └───────────────────────────────────────────┘  │
│                        │                         │
│   ┌────────────────────┼────────────────────┐   │
│   │                    │                    │   │
│   ▼                    ▼                    ▼   │
│ ┌──────────┐  ┌─────────────────┐  ┌────────┐  │
│ │Control   │  │GraphVisualizer  │  │Stats   │  │
│ │Panel     │  │                 │  │Panel   │  │
│ └──────────┘  └─────────────────┘  └────────┘  │
└─────────────────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │ GossipProtocol       │
              │ (純粋なJSクラス)     │
              │ - nodes: Node[]      │
              │ - executeRound()     │
              │ - 全てメモリ上で実行 │
              └──────────────────────┘
```

## データモデル

### Node（ノード）

```typescript
interface NodeData {
  id: string;           // "node-0", "node-1", ...
  state: NodeState;     // 'Red' | 'Green' | 'Blue'
  position: {           // グラフ上の位置
    x: number;
    y: number;
  };
  peers: string[];      // 接続可能なノードのID
  lastUpdated: number;  // 最終更新時刻（Round数）
  value?: string;       // オプション：Go実装のValueに相当
}
```

### Message（メッセージ）

```typescript
interface GossipMessage {
  from: string;         // 送信元ノードID
  to: string;          // 送信先ノードID
  state: NodeState;    // 伝播する状態
  round: number;       // 送信されたRound
  timestamp: number;   // 送信時刻
}
```

### SimulationState（シミュレーション状態）

```typescript
interface SimulationState {
  nodes: NodeData[];
  edges: EdgeData[];
  messages: GossipMessage[];  // 現在飛んでいるメッセージ
  currentRound: number;
  isRunning: boolean;
  config: SimulationConfig;
  stats: SimulationStats;
}
```

## ゴシッププロトコルの実装（クライアントサイド）

すべての処理はブラウザのJavaScriptランタイムで実行され、外部APIとの通信は行いません。

### 1. ランダムピア選択アルゴリズム

```typescript
class GossipProtocol {
  // ブラウザ内でランダム選択を実行
  selectRandomPeer(node: NodeData): string | null {
    if (node.peers.length === 0) return null;
    const index = Math.floor(Math.random() * node.peers.length);
    return node.peers[index];
  }
}
```

### 2. メッセージ伝播ロジック（仮想的な通信）

```typescript
class GossipProtocol {
  // メッセージオブジェクトをメモリ上で作成
  propagate(sourceNode: NodeData, targetNodeId: string): GossipMessage {
    return {
      from: sourceNode.id,
      to: targetNodeId,
      state: sourceNode.state,
      round: this.currentRound,
      timestamp: Date.now()
    };
  }
  
  // メモリ上でノードの状態を直接更新
  receiveMessage(targetNode: NodeData, message: GossipMessage): void {
    // 状態の更新（感染モデル）
    if (targetNode.state !== message.state) {
      targetNode.state = message.state;
      targetNode.lastUpdated = message.round;
    }
  }
}
```

### 3. Round実行（setIntervalによるシミュレーション）

```typescript
class GossipProtocol {
  private intervalId: number | null = null;
  
  // ブラウザのタイマーAPIを使用してRoundを実行
  start(speed: number): void {
    const intervalMs = 1000 / speed; // speed rounds/sec
    this.intervalId = window.setInterval(() => {
      this.executeRound();
    }, intervalMs);
  }
  
  stop(): void {
    if (this.intervalId) {
      window.clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }
  
  executeRound(): void {
    const activeNodes = this.getActiveNodes();
    const messages: GossipMessage[] = [];
    
    // 各アクティブノードからランダムにゴシップ（仮想的）
    for (const node of activeNodes) {
      const targetId = this.selectRandomPeer(node);
      if (targetId) {
        const message = this.propagate(node, targetId);
        messages.push(message);
      }
    }
    
    // メッセージを受信処理（即座に反映）
    for (const message of messages) {
      const targetNode = this.findNode(message.to);
      if (targetNode) {
        this.receiveMessage(targetNode, message);
      }
    }
    
    this.currentRound++;
    this.notifySubscribers(); // UIの更新をトリガー
  }
}
```

## 可視化戦略

### 1. ノードの視覚表現

- **色**: 状態を表現（Red/Green/Blue）
- **サイズ**: 最近更新されたノードを大きく表示
- **パルス**: メッセージ送受信時にアニメーション

### 2. エッジの視覚表現

- **通常状態**: 薄いグレーの線
- **メッセージ送信中**: 送信元の色でハイライト
- **アニメーション**: メッセージが線に沿って移動

### 3. 統計情報の表示

```typescript
interface SimulationStats {
  totalRounds: number;
  convergedRounds?: number;  // 収束したRound数
  nodeStates: {
    [key in NodeState]: number;  // 各状態のノード数
  };
  messagesPerRound: number[];  // Round毎のメッセージ数
  convergenceRate: number;      // 収束率（%）
}
```

## パフォーマンス最適化

### 1. React最適化

- `React.memo`でコンポーネントの再レンダリング制御
- `useMemo`/`useCallback`で計算結果のキャッシュ
- 仮想化による大規模ノードの効率的レンダリング

### 2. グラフ描画最適化

- react-force-graphの設定調整
  - `cooldownTicks`: アニメーション停止タイミング
  - `warmupTicks`: 初期配置の計算回数
  - `d3AlphaDecay`: 力学シミュレーションの減衰率

### 3. 状態管理最適化

- イミュータブルな更新パターン
- 必要最小限の状態更新
- バッチ更新の活用

## エラーハンドリング

### 1. 入力検証

```typescript
class ValidationService {
  validateNodeCount(count: number): void {
    if (count < 1 || count > 300) {
      throw new Error('Node count must be between 1 and 300');
    }
  }
  
  validateSpeed(speed: number): void {
    if (speed < 1 || speed > 100) {
      throw new Error('Speed must be between 1 and 100 rounds/sec');
    }
  }
}
```

### 2. 実行時エラー

- メモリ不足の検知と警告
- グラフ描画エラーのフォールバック
- タイマーエラーの自動リカバリ

## テスト戦略との連携

このアーキテクチャは、test-strategy.mdで定義されたテスト戦略に対応：

1. **ユニットテスト可能**: 各クラス/関数が独立してテスト可能
2. **モック可能**: 外部依存（react-force-graph）を容易にモック化
3. **統合テスト対応**: コンポーネント間の連携をテスト

## Go実装からの学び

### 収束メトリクス

Go実装の`observe-convergence`ツールの機能を参考に：

```typescript
class ConvergenceAnalyzer {
  analyzeConvergence(rounds: number, nodeCount: number): string {
    const efficiency = rounds / nodeCount;
    
    if (efficiency <= 2) {
      return 'Excellent - Very efficient propagation';
    } else if (efficiency <= 3) {
      return 'Good - Expected range';
    } else if (efficiency <= 6) {
      return 'Acceptable - Slightly slow';
    } else {
      return 'Poor - Unusually slow propagation';
    }
  }
}
```

## 次のフェーズへの拡張性

このアーキテクチャは以下の拡張に対応：

1. **異なるゴシップアルゴリズム**: Strategyパターンで実装
2. **ネットワークトポロジー**: ピア接続の動的変更
3. **障害シミュレーション**: ノードの一時停止/復帰
4. **メッセージロス**: 確率的なメッセージ破棄