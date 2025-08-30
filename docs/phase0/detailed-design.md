# Phase 0: Proof of Concept - 詳細設計

## 目標の再確認
**半日で**「ゴシップの本質（ランダム選択による確率的伝搬）」を体験できる最小実装を作る

## アーキテクチャ設計

### システム構成
```
    Node-0 (localhost:18000) ←→ Node-1 (localhost:18001)
         ↓                           ↓
    Node-2 (localhost:18002) ←→ Node-3 (localhost:18003)
         ↓                           ↓
    Node-4 (localhost:18004) ←→ Node-5 (localhost:18005)
         ↓                           ↓
    Node-6 (localhost:18006) ←→ Node-7 (localhost:18007)
         ↓                           ↓
    Node-8 (localhost:18008) ←→ Node-9 (localhost:18009)
```

**10ノードすべてが相互に通信可能** (フルメッシュ)
**ランダム性の観察に最適な規模**

### データ構造

#### Node構造体
```go
type Node struct {
    ID       string   // "node-0", "node-1", ... "node-9"  
    Address  string   // "localhost:18000" ~ "localhost:18009"
    Value    string   // 共有する状態値
    Peers    []string // 他9ノードのアドレスリスト
    LastSeen int64    // 最後にゴシップした時刻（Unix timestamp）
}
```

#### メッセージ構造体
```go
type GossipMessage struct {
    From      string `json:"from"`       // 送信元ノードID
    Value     string `json:"value"`      // 伝搬する状態値
    Timestamp int64  `json:"timestamp"`  // 送信時刻
}
```

## コア機能の実装

### 1. ランダム選択ロジック
```go
func (n *Node) selectRandomPeer() string {
    if len(n.Peers) == 0 {
        return ""
    }
    
    // Go標準のmath/randを使用
    index := rand.Intn(len(n.Peers))
    return n.Peers[index]
}
```

### 2. 状態送信機能
```go
func (n *Node) sendGossip() error {
    // ランダムにピアを選択
    targetAddr := n.selectRandomPeer()
    if targetAddr == "" {
        return fmt.Errorf("no peers available")
    }
    
    // メッセージを作成
    message := GossipMessage{
        From:      n.ID,
        Value:     n.Value,
        Timestamp: time.Now().Unix(),
    }
    
    // HTTP POSTで送信
    return n.sendHTTPMessage(targetAddr, message)
}
```

### 3. メッセージ受信処理
```go
func (n *Node) handleGossipMessage(msg GossipMessage) {
    log.Printf("[%s] Received gossip from %s: value='%s'", 
               n.ID, msg.From, msg.Value)
    
    // 受信した値で自分の状態を更新
    if n.Value != msg.Value {
        log.Printf("[%s] Updating value: '%s' -> '%s'", 
                   n.ID, n.Value, msg.Value)
        n.Value = msg.Value
        n.LastSeen = time.Now().Unix()
    }
}
```

## HTTP通信設計

### APIエンドポイント

#### POST /gossip - ゴシップメッセージ受信
```
Content-Type: application/json

{
  "from": "node-A",
  "value": "hello-world", 
  "timestamp": 1640995200
}
```

#### POST /trigger - 手動ゴシップトリガー
```
# パラメータなし、空のPOST
# レスポンス: {"status": "sent", "target": "localhost:8002"}
```

#### GET /status - ノード状態確認
```
{
  "id": "node-A",
  "value": "hello-world",
  "peers": ["localhost:8002", "localhost:8003"],
  "last_seen": 1640995200
}
```

## ファイル構成

```
gossip_concept/phase0/
├── main.go           # エントリーポイント
├── node.go           # Node構造体と基本機能
├── gossip.go         # ゴシップロジック
├── http_server.go    # HTTPサーバー
├── config.json       # 設定ファイル
└── README.md         # 実行手順
```

## 設定ファイル (config.json)

```json
{
  "nodes": {
    "node-0": {
      "address": "localhost:18000",
      "peers": ["localhost:18001", "localhost:18002", "...", "localhost:18009"]
    },
    "node-1": {
      "address": "localhost:18001", 
      "peers": ["localhost:18000", "localhost:18002", "...", "localhost:18009"]
    },
    // ... node-2 ~ node-8 ...
    "node-9": {
      "address": "localhost:18009",
      "peers": ["localhost:18000", "localhost:18001", "...", "localhost:18008"]
    }
  },
  "initial_value": "initial-state"
}
```

**注**: 実際の設定ファイルでは各ノードが他の9ノードすべてをpeersに含む

## 実行シナリオ

### シナリオ1: 基本的な状態伝搬
1. **起動**: 10ノードすべて起動、初期値は `"initial-state"`
2. **状態変更**: Node-0の値を `"new-value"` に変更
3. **ゴシップ実行**: Node-0で手動トリガー実行
4. **観察**: どのノードに送信されたかログで確認（1/9の確率でランダム）
5. **追跡**: 複数回実行して全ノードが `"new-value"` になることを確認

### シナリオ2: ランダム性の観察  
1. Node-0で100回連続でゴシップ実行
2. ログで送信先の分散を確認（理想：各ノードに約11回ずつ送信）
3. 統計的な偏りがないことを確認

### シナリオ3: 収束パターンの確認
1. Node-0で値を変更してゴシップ
2. 伝搬パスの追跡（例: 0→3→7→2→...）
3. 全10ノードが同じ値になるまでの手数を測定
4. 理論値との比較（10ノードなら平均20-30回程度で収束）

## 成功基準

### 機能要件
- [ ] 10ノードが正常に起動できる
- [ ] 手動でゴシップを実行できる  
- [ ] 9つの送信先から均等にランダム選択される
- [ ] 受信したノードが値を更新する
- [ ] 最終的に全10ノードが同じ値になる

### 観察要件
- [ ] 送信先の分布が統計的に均等（100回で各ノード8-14回程度）
- [ ] 状態の変化がリアルタイムで見える
- [ ] 伝搬パスが追跡できる（0→3→7→...のような経路）
- [ ] 収束までの手数が理論値に近い（20-30回程度）

### パフォーマンス要件（Phase 0では無視）
- 速度、効率性、リソース使用量は考慮しない
- 障害処理やエラーハンドリングも最小限

## Riak Core実装との対応

| Riak Core機能 | Phase 0実装 | 実装方法 |
|---------------|-------------|----------|
| `random_gossip/1` | `sendGossip()` | HTTP POST |
| `riak_core_ring:random_other_active_node/1` | `selectRandomPeer()` | `rand.Intn()` |
| `{reconcile_ring, RingIn}` | `handleGossipMessage()` | JSON decode |
| gen_server状態管理 | `Node.Value` | 単純な文字列 |

## 学習ポイント

この実装で体験できるゴシップの本質：

1. **確率的伝搬**: 完璧でない通信でも最終的に収束
2. **スケーラビリティ**: 各ノードは全体を知らなくても良い
3. **障害耐性**: 1つのノードが止まっても他が動き続ける（Phase 0では未検証）
4. **自己修復性**: 時間が経てば状態が一致する

実装後は「なぜこれが分散システムで重要なのか」を実感できるはず。
