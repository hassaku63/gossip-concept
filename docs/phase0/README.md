# Phase 0: Gossip Protocol Proof of Concept

## 概要
Riak CoreのErlang実装を参考に、ゴシッププロトコルの本質（ランダム選択による確率的伝搬）を体験できる最小実装。

**実装規模**: 約240行（Go）  
**実装状況**: ✅ 完了
**ノード数**: 10ノード（ランダム性の観察に最適）
**起動方式**: 1つのコマンドで全10ノードを並行起動
**管理機能**: 独立ポート（17999）で管理サービス提供

## 学習目標
1. **ランダム選択**によるピア選択の仕組み
2. **確率的伝搬**による最終的な状態収束
3. **統計的分散**の観察と理解

## ファイル構成

### 設計ドキュメント
- `detailed-design.md` - システム設計書
- `code-structure.md` - コード構造設計
- `implementation-steps.md` - 実装手順書
- `test-strategy.md` - テスト戦略
- `10-node-update-summary.md` - 10ノード変更の詳細
- `experiment-guide.md` - 実験手順書

### 実装ファイル（完了）
- `../../main.go` - エントリーポイント（10ノード並行起動）
- `../../node.go` - Node構造体と基本機能
- `../../gossip.go` - ゴシップロジック（ランダム選択）
- `../../http_server.go` - HTTPサーバー
- `../../admin_server.go` - 管理サービス（新規追加）

### 実験用スクリプト
- `../../start-cluster.sh` - クラスター起動ヘルパー
- `../../observe-randomness.sh` - ランダム性観察
- `../../observe-convergence.sh` - 収束性観察

### 旧設定ファイル（統合により不要）
- ~~`../../generate-config.py`~~ - 削除済み
- ~~`../../config-10nodes.json`~~ - 削除済み

## クイックスタート

### 1. ビルドと起動（設定ファイル不要）
```bash
cd gossip_concept/
go build .
./gossip-concept
```

**期待される出力**:
```
2025/08/30 21:57:34 Starting 10 nodes...
2025/08/30 21:57:34 All 10 nodes started successfully
2025/08/30 21:57:34 Use curl to interact with nodes:
2025/08/30 21:57:34   Status:  curl localhost:18000/status
2025/08/30 21:57:34   Gossip:  curl -X POST localhost:18000/trigger
2025/08/30 21:57:34   Set:     curl -X POST 'localhost:18000/set?value=hello'
```

### コマンドラインオプション
```bash
# デフォルト: 10ノード、ポート18000-18009
./gossip-concept

# 5ノードで起動
./gossip-concept --nodes=5

# 異なるポート範囲（19000-19009）
./gossip-concept --base-port=19000

# ヘルプ表示
./gossip-concept --help
```

### 2. 動作確認
```bash
# 全ノードの状態確認
for port in {18000..18009}; do 
  curl -s "localhost:$port/status" | jq -r '.id + ": " + .value'
done

# ゴシップ実行
curl -X POST localhost:18000/trigger

# 値変更
curl -X POST "localhost:18000/set?value=hello"
```

## 実験シナリオ

**詳細な実験手順**: `experiment-guide.md` を参照

### 実験1: ランダム性の観察
```bash
# スクリプトを使用（推奨）
cd gossip_concept/
./observe-randomness.sh 100

# または手動実行
echo "Testing 20 random gossip selections:"
for i in {1..20}; do 
  echo -n "Round $i: "
  curl -s -X POST localhost:18000/trigger | jq -r '.target'
  sleep 0.1
done
```

### 実験2: 状態伝搬と収束性
```bash
# スクリプトを使用（推奨）
cd gossip_concept/
./observe-convergence.sh

# または手動実行
test_value="experiment-$(date +%s)"
curl -s -X POST "localhost:18000/set?value=$test_value"
for round in {1..20}; do
  curl -s -X POST localhost:18000/trigger
  sleep 0.2
done
```

## API仕様

### ノードAPI（ポート18000-18009）

#### GET /status
現在のノード状態を取得
```json
{
  "id": "node-0",
  "value": "current-value",
  "peers": ["localhost:18001", ...],
  "last_seen": 1640995200
}
```

#### POST /trigger
手動でゴシップを実行
```json
{
  "status": "sent",
  "target": "localhost:18003"
}
```

#### POST /set?value=xxx
ノードの値を変更
```json
{
  "status": "updated",
  "value": "xxx"
}
```

#### POST /gossip
ゴシップメッセージ受信（内部API）

### 管理API（ポート17999）

#### GET /cluster
クラスター設定情報
```json
{
  "node_count": 10,
  "base_port": 18000,
  "admin_port": 17999,
  "topology": "full-mesh",
  "started_at": 1756559586
}
```

#### GET /nodes
全ノード情報
```json
[
  {
    "id": "node-0",
    "port": 18000,
    "address": "localhost:18000",
    "value": "initial-state",
    "peer_count": 9,
    "last_seen": 0
  },
  ...
]
```

#### GET /health
ヘルスチェック
```json
{
  "all_healthy": true,
  "nodes": [
    {"id": "node-0", "port": 18000, "healthy": true},
    ...
  ]
}
```

## 重要な実装箇所

### ゴシップの本質: ランダム選択
```go
// gossip.go
func (n *Node) selectRandomPeer() string {
    if len(n.Peers) == 0 {
        return ""
    }
    index := rand.Intn(len(n.Peers))
    return n.Peers[index]
}
```

### 状態の同期
```go
// gossip.go
func (n *Node) HandleGossipMessage(msg GossipMessage) {
    log.Printf("[%s] Received gossip from %s: value='%s'", 
               n.ID, msg.From, msg.Value)
    n.SetValue(msg.Value)
}
```

## トラブルシューティング

### ポート使用中エラー
```bash
# 既存プロセスを終了
pkill -f gossip-concept

# ポート確認
lsof -i :18000-18009
```

### JSON解析エラー
```bash
# HTTPステータスを確認
curl -i localhost:18000/status
```

### ノード間通信エラー
```bash
# 対象ノードの起動状況を確認
netstat -tulpn | grep 180
```

## 学習成果

このPhase 0実装により体験できること：

1. **ランダム性**: N-1つの送信先から統計的に均等に選択される様子
2. **確率的伝搬**: 完全でない通信でも最終的に収束する仕組み
3. **スケーラビリティ**: 各ノードが全体を知らなくても自律的に動作
4. **観察可能性**: HTTPAPIで伝搬過程をリアルタイム追跡
5. **障害耐性**: 一部のノードが失敗しても全体システムは継続動作
6. **設定の簡潔性**: フラグ2つだけで任意サイズのクラスター起動

## 成功基準

- **ランダム性**: 100回中、各ノードに8-14回の範囲で分散
- **収束性**: 30回以内のゴシップで全ノード更新
- **レスポンス**: 全APIが100ms以内で応答

## 次のステップ

- **Phase 0.5**: 自動タイマー機能の追加
- **Phase 1**: レート制限とkey-value状態管理
- **独自実験**: アルゴリズムの改良や性能測定

この実験により、Riak Coreの`riak_core_gossip.erl`で実装されているゴシッププロトコルの本質を実体験で理解できます。