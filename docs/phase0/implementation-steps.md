# Phase 0 実装手順書

## 前提条件
- Go 1.21+ がインストール済み
- curl コマンドが使用可能
- 3つのターミナルウィンドウを開けること

## Step 1: プロジェクトセットアップ (5分)

### 1.1 ディレクトリ作成
```bash
cd gossip_concept/phase0
```

### 1.2 Go module初期化
```bash
go mod init gossip-phase0
```

### 1.3 設定ファイル作成
```bash
cat > config.json << 'EOF'
{
  "nodes": {
    "node-A": {
      "address": "localhost:8001",
      "peers": ["localhost:8002", "localhost:8003"]
    },
    "node-B": {
      "address": "localhost:8002",
      "peers": ["localhost:8001", "localhost:8003"]
    },
    "node-C": {
      "address": "localhost:8003",
      "peers": ["localhost:8001", "localhost:8002"]
    }
  },
  "initial_value": "initial-state"
}
EOF
```

## Step 2: コアロジック実装 (1.5時間)

### 2.1 基本構造体実装 (20分)
`node.go` を作成:

**重要ポイント**:
- thread-safe な値管理 (`sync.RWMutex`)
- ログ出力で状態変化を可視化

```go
package main

import (
    "log"
    "sync"
    "time"
)

type Node struct {
    mu       sync.RWMutex
    ID       string
    Address  string  
    Value    string
    Peers    []string
    LastSeen int64
}

func NewNode(id string, config *Config) *Node {
    nodeConfig := config.Nodes[id]
    return &Node{
        ID:       id,
        Address:  nodeConfig.Address,
        Value:    config.InitialValue,
        Peers:    nodeConfig.Peers,
        LastSeen: 0,
    }
}

func (n *Node) GetValue() string {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return n.Value
}

func (n *Node) SetValue(value string) {
    n.mu.Lock()
    defer n.mu.Unlock()
    if n.Value != value {
        log.Printf("[%s] Value updated: '%s' -> '%s'", n.ID, n.Value, value)
        n.Value = value
        n.LastSeen = time.Now().Unix()
    }
}

func (n *Node) GetStatus() map[string]interface{} {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return map[string]interface{}{
        "id":        n.ID,
        "value":     n.Value,
        "peers":     n.Peers,
        "last_seen": n.LastSeen,
    }
}
```

### 2.2 ゴシップロジック実装 (30分)
`gossip.go` を作成:

**キーポイント**:
- `selectRandomPeer()` がゴシップの本質
- HTTP通信でシンプル実装

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "time"
)

type GossipMessage struct {
    From      string `json:"from"`
    Value     string `json:"value"`
    Timestamp int64  `json:"timestamp"`
}

// ★ ゴシップの本質：ランダム選択
func (n *Node) selectRandomPeer() string {
    n.mu.RLock()
    defer n.mu.RUnlock()
    
    if len(n.Peers) == 0 {
        return ""
    }
    
    index := rand.Intn(len(n.Peers))
    return n.Peers[index]
}

func (n *Node) SendGossip() (string, error) {
    target := n.selectRandomPeer()
    if target == "" {
        return "", fmt.Errorf("no peers available")
    }
    
    message := GossipMessage{
        From:      n.ID,
        Value:     n.GetValue(),
        Timestamp: time.Now().Unix(),
    }
    
    err := n.sendHTTPMessage(target, message)
    if err != nil {
        return target, fmt.Errorf("failed to send to %s: %v", target, err)
    }
    
    log.Printf("[%s] Sent gossip to %s: value='%s'", n.ID, target, message.Value)
    return target, nil
}

func (n *Node) sendHTTPMessage(targetAddr string, msg GossipMessage) error {
    jsonData, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    
    url := fmt.Sprintf("http://%s/gossip", targetAddr)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("HTTP error: %s", resp.Status)
    }
    
    return nil
}

func (n *Node) HandleGossipMessage(msg GossipMessage) {
    log.Printf("[%s] Received gossip from %s: value='%s'", 
               n.ID, msg.From, msg.Value)
    n.SetValue(msg.Value)
}
```

### 2.3 HTTPサーバー実装 (25分)
`http_server.go` を作成:

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
)

func startHTTPServer(node *Node) {
    mux := http.NewServeMux()
    
    // ゴシップ受信
    mux.HandleFunc("/gossip", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        var msg GossipMessage
        if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        
        node.HandleGossipMessage(msg)
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "received"})
    })
    
    // 手動トリガー
    mux.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        target, err := node.SendGossip()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        response := map[string]string{
            "status": "sent",
            "target": target,
        }
        json.NewEncoder(w).Encode(response)
    })
    
    // 状態確認
    mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        status := node.GetStatus()
        json.NewEncoder(w).Encode(status)
    })
    
    log.Fatal(http.ListenAndServe(node.Address, mux))
}
```

### 2.4 エントリーポイント実装 (15分)
`main.go` を作成:

```go
package main

import (
    "encoding/json"
    "flag"
    "log"
    "os"
)

type Config struct {
    Nodes        map[string]NodeConfig `json:"nodes"`
    InitialValue string                `json:"initial_value"`
}

type NodeConfig struct {
    Address string   `json:"address"`
    Peers   []string `json:"peers"`
}

func main() {
    nodeID := flag.String("node", "", "Node ID (node-A, node-B, node-C)")
    flag.Parse()
    
    if *nodeID == "" {
        log.Fatal("Please specify -node flag")
    }
    
    config := loadConfig("config.json")
    if _, exists := config.Nodes[*nodeID]; !exists {
        log.Fatalf("Unknown node ID: %s", *nodeID)
    }
    
    node := NewNode(*nodeID, config)
    log.Printf("Starting node %s on %s", node.ID, node.Address)
    
    startHTTPServer(node)
}

func loadConfig(filename string) *Config {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatalf("Cannot open config file: %v", err)
    }
    defer file.Close()
    
    var config Config
    if err := json.NewDecoder(file).Decode(&config); err != nil {
        log.Fatalf("Cannot parse config file: %v", err)
    }
    
    return &config
}
```

## Step 3: 動作確認 (30分)

### 3.1 ビルドと起動 (10分)

```bash
# ビルド
go build .

# ターミナル1: Node A起動
./gossip-phase0 -node node-A

# ターミナル2: Node B起動  
./gossip-phase0 -node node-B

# ターミナル3: Node C起動
./gossip-phase0 -node node-C
```

**期待するログ**:
```
2024/01/15 10:00:00 Starting node node-A on localhost:8001
```

### 3.2 基本動作テスト (10分)

```bash
# 全ノードの初期状態確認
curl localhost:8001/status
curl localhost:8002/status 
curl localhost:8003/status

# 期待結果: すべて "initial-state"
```

### 3.3 ゴシップテスト (10分)

```bash
# Node Aから手動ゴシップ実行
curl -X POST localhost:8001/trigger

# ログ確認: どのノードに送信されたか
# レスポンス例: {"status":"sent","target":"localhost:8002"}

# 送信先ノードの状態確認
curl localhost:8002/status
```

**期待する動作**:
1. Node Aで `/trigger` 実行
2. ランダムに選択されたノード（B or C）にメッセージ送信
3. 送信先ノードの値が "initial-state" のまま（変更なし）

## Step 4: ゴシップの観察 (30分)

### 4.1 状態変更テスト

Node Aの値を手動で変更:
```go
// http_server.go に追加するテストエンドポイント
mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    value := r.URL.Query().Get("value")
    if value == "" {
        http.Error(w, "value parameter required", http.StatusBadRequest)
        return
    }
    
    node.SetValue(value)
    json.NewEncoder(w).Encode(map[string]string{"status": "updated", "value": value})
})
```

### 4.2 伝搬観察

```bash
# 1. Node Aの値を変更
curl -X POST "localhost:8001/set?value=hello-gossip"

# 2. ゴシップ実行
curl -X POST localhost:8001/trigger

# 3. 全ノード状態確認
curl localhost:8001/status
curl localhost:8002/status  
curl localhost:8003/status

# 4. 複数回実行してランダム性確認
for i in {1..5}; do
  echo "=== Round $i ==="
  curl -X POST localhost:8001/trigger
  sleep 1
done
```

### 4.3 成功確認チェックリスト

- [ ] 3ノードすべてが正常起動
- [ ] `/status` で状態確認できる
- [ ] `/trigger` でランダムな送信先が選ばれる
- [ ] ログで送信・受信が確認できる
- [ ] 受信ノードで値が更新される
- [ ] 複数回実行で異なるノードに送信される

## トラブルシューティング

### ポートが使用中
```bash
# プロセス確認
lsof -i :8001
lsof -i :8002
lsof -i :8003

# 強制終了
pkill gossip-phase0
```

### JSON設定エラー
```bash
# JSONの構文チェック
cat config.json | jq .
```

### HTTP接続エラー
- すべてのノードが起動しているか確認
- ファイアウォール設定確認

## 次のステップ

Phase 0完了後：
1. **観察結果の記録**: どのようなパターンで伝搬したか
2. **問題点の洗い出し**: 改善すべき点
3. **Phase 0.5への移行**: 自動化機能の追加

この手順により、**半日でゴシップの本質を体験**できる実装が完成します。