# Gossip Protocol Go Implementation Plan

## 目標
Riak CoreのErlang実装をリファレンスとして、ゴシッププロトコルのコア機能のみを抽出し、Go言語で実装する。

## Riak Core実装の分析結果

### 抽出すべきコア機能

#### 1. 基本ゴシップメカニズム
- **Random Gossip**: ランダムな他ノードへの状態送信
- **Recursive Gossip**: バイナリツリー構造での効率的な状態伝播
- **Rate Limiting**: トークンバケットによる送信制御

#### 2. 必須コンポーネント
- **Node Discovery**: 他ノードのリスト管理
- **State Management**: 共有する状態の管理
- **Message Transport**: ノード間通信
- **Timer Management**: 定期実行制御

#### 3. 削除する Riak Core 固有要素
- Ring状態管理
- VNode関連処理
- Claimant機能
- ETS/Mochiglobal最適化
- OTP Supervisor階層

## Go実装アーキテクチャ設計

### パッケージ構成
```
gossip_concept/
├── cmd/
│   └── demo/           # デモアプリケーション
├── pkg/
│   ├── gossip/         # ゴシッププロトコル本体
│   ├── node/           # ノード管理
│   ├── state/          # 状態管理
│   └── transport/      # 通信レイヤー
├── internal/
│   └── timer/          # タイマー管理
└── examples/           # 使用例
```

### 主要インターフェース設計

#### State Interface
```go
type State interface {
    // 状態のバージョン比較
    Compare(other State) (newer, older, conflict bool)

    // 状態のマージ
    Merge(other State) State

    // シリアライゼーション
    Marshal() ([]byte, error)
    Unmarshal(data []byte) error

    // 状態の識別子
    Version() uint64
}
```

#### Node Interface
```go
type Node interface {
    ID() string
    Address() string
    IsActive() bool
}
```

#### Transport Interface
```go
type Transport interface {
    Send(target Node, message []byte) error
    Listen(handler MessageHandler) error
    Close() error
}
```

#### Gossip Engine
```go
type Engine struct {
    nodeID     string
    nodes      []Node
    state      State
    transport  Transport
    rateLimiter *RateLimiter
    config     Config
}
```

### コア機能実装計画

#### 1. Random Gossip (Phase 1)
**実装内容**:
- 定期タイマーで他ノードをランダム選択
- 現在の状態を送信
- 受信時の状態マージ

**Erlang対応箇所**:
- `random_gossip/1`
- `handle_cast(gossip_ring, State)`

#### 2. Rate Limiting (Phase 1)
**実装内容**:
- Token Bucket アルゴリズム
- 設定可能な制限値

**Erlang対応箇所**:
- `schedule_next_reset/0`
- `handle_info(reset_tokens, State)`

#### 3. Recursive Gossip (Phase 2)
**実装内容**:
- ノードリストからバイナリツリー構築
- 子ノードへの並列送信

**Erlang対応箇所**:
- `recursive_gossip/1`
- `riak_core_util:build_tree/3`

#### 4. Node Management (Phase 2)
**実装内容**:
- アクティブノードリスト管理
- ノードの生死監視

**Erlang対応箇所**:
- `riak_core_ring:active_members/1`
- `riak_core_node_watcher`の一部

## 実装フェーズ

### Phase 0: Proof of Concept (最小限の状態伝搬)
**目標**: ゴシップの本質（ランダム選択）を含む最小実装

**成果物**:
- 単一の文字列値を共有
- **3ノード固定構成**（ランダム性確保のため）
- 手動トリガーでのランダム送信
- ログ出力で伝搬パターン確認

**期間**: 半日

**実装内容**:
```go
// 最小限の構造体
type SimpleGossip struct {
    nodeID string
    value  string
    peers  []string // 他の2ノード固定
}

// ランダムな1ノードに送信（ゴシップの本質）
func (g *SimpleGossip) SendToRandomPeer()
```

**重要**: 2ノードではランダム性がないため、ゴシップの特徴を体験できない

### Phase 0.5: 自動化とスケーラビリティ
**目標**: 定期実行と任意ノード数対応

**成果物**:
- タイマーによる自動送信
- 5-10ノード対応
- JSON設定ファイル（ノードリスト）
- 収束パターンの観察

**期間**: 半日

### Phase 1: Basic Gossip Infrastructure  
**目標**: 実用的なゴシップの基礎を作る

**成果物**:
- key-value状態管理
- HTTP/JSONベースの通信
- 基本的なレート制限
- 任意ノード数対応

**期間**: 1-2日

### Phase 2: Efficient Gossip
**目標**: 効率的な伝播メカニズムを追加

**成果物**:
- Recursive gossip実装
- ノード管理機能
- より複雑な状態のサポート
- パフォーマンステスト

**期間**: 3-4日

### Phase 3: Production-Ready Features
**目標**: 実用的な機能を追加

**成果物**:
- 設定の外部化
- メトリクス収集
- エラーハンドリング強化
- ドキュメント整備

**期間**: 2-3日

## 技術選択

### 通信プロトコル
**Phase 1**: HTTP/JSON (簡単な実装・デバッグ)
**Phase 2**: gRPC (効率性・型安全性)

### 状態管理
**Phase 1**: `map[string]interface{}` (柔軟性)
**Phase 2**: カスタム構造体 (性能・型安全性)

### 設定管理
- YAML設定ファイル
- 環境変数での上書き対応

### ログ・メトリクス
- `slog` (標準ライブラリ)
- Prometheus メトリクス対応

## デモアプリケーション

### シナリオ1: 分散カウンター
複数ノードで共有カウンターを管理し、ゴシップで値を同期

### シナリオ2: 分散設定管理
設定値の変更を全ノードに伝播

### シナリオ3: ノード状態監視
各ノードのヘルス状態をクラスター全体で共有

## テスト戦略

### 単体テスト
- State merge ロジック
- Rate limiter
- Binary tree construction

### 統合テスト
- 多ノード環境でのゴシップ伝播
- ネットワーク分断からの復旧
- 高負荷時の動作

### ベンチマーク
- メッセージ伝播速度
- メモリ使用量
- CPU使用率

## 成功指標

### 機能面
- [ ] 3ノードクラスターでの状態同期
- [ ] ノード追加/削除への対応
- [ ] ネットワーク分断からの回復

### 性能面
- [ ] 100ノード環境での動作
- [ ] 1秒以内での状態伝播（小さな変更）
- [ ] メモリ使用量 < 100MB/ノード

### 保守性面
- [ ] 設定可能なパラメータ
- [ ] 詳細なログ出力
- [ ] メトリクス収集API

## 参考実装パターン

### Erlang → Go 変換例

**Erlang**:
```erlang
random_gossip(Ring) ->
    case riak_core_ring:random_other_active_node(Ring) of
        no_node -> ok;
        RandomNode -> send_ring(node(), RandomNode)
    end.
```

**Go**:
```go
func (e *Engine) RandomGossip() error {
    node := e.selectRandomNode()
    if node == nil {
        return nil
    }
    return e.sendState(node, e.state)
}
```

この実装計画により、Riak Coreのゴシップメカニズムの本質を理解しながら、Goでの実用的な実装を段階的に構築できる。
