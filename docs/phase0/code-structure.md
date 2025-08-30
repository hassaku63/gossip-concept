# Phase 0 コード構造設計

## ファイル分割戦略

### 単一責任の原則に基づいた分割

```
phase0/
├── main.go           # 26行: エントリーポイント、設定読み込み
├── node.go           # 45行: Node構造体、基本メソッド  
├── gossip.go         # 35行: ゴシップロジック、ランダム選択
├── http_server.go    # 55行: HTTPサーバー、API処理
├── config.json       # 15行: 設定ファイル
└── go.mod            # 3行: Go modules設定
```

**総行数**: 約180行（コメント含む）

## 各ファイルの詳細設計

### main.go (エントリーポイント)
```go
package main

import (
    "encoding/json"
    "flag"
    "log" 
    "os"
)

type Config struct {
    Nodes map[string]NodeConfig `json:"nodes"`
    InitialValue string         `json:"initial_value"`
}

type NodeConfig struct {
    Address string   `json:"address"`
    Peers   []string `json:"peers"`
}

func main() {
    // コマンドライン引数でノードIDを指定
    nodeID := flag.String("node", "", "Node ID (node-A, node-B, node-C)")
    flag.Parse()
    
    if *nodeID == "" {
        log.Fatal("Please specify -node flag")
    }
    
    // 設定ファイル読み込み
    config := loadConfig("config.json")
    
    // ノード作成・起動
    node := NewNode(*nodeID, config)
    log.Printf("Starting node %s on %s", node.ID, node.Address)
    
    // HTTPサーバー起動（ブロッキング）
    startHTTPServer(node)
}

func loadConfig(filename string) *Config {
    // JSON設定ファイル読み込み処理
}
```

### node.go (Node構造体)
```go
package main

import (
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

// thread-safeな値の取得
func (n *Node) GetValue() string {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return n.Value
}

// thread-safeな値の更新  
func (n *Node) SetValue(value string) {
    n.mu.Lock()
    defer n.mu.Unlock()
    if n.Value != value {
        log.Printf("[%s] Value updated: '%s' -> '%s'", n.ID, n.Value, value)
        n.Value = value
        n.LastSeen = time.Now().Unix()
    }
}

// ステータス情報取得
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

### gossip.go (ゴシップロジック)
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "math/rand" 
    "net/http"
    "time"
)

type GossipMessage struct {
    From      string `json:"from"`
    Value     string `json:"value"`
    Timestamp int64  `json:"timestamp"`
}

// ランダムなピア選択（ゴシップの本質）
func (n *Node) selectRandomPeer() string {
    n.mu.RLock()
    defer n.mu.RUnlock()
    
    if len(n.Peers) == 0 {
        return ""
    }
    
    // Go標準のrand使用（シードは自動設定される）
    index := rand.Intn(len(n.Peers))
    return n.Peers[index]
}

// ゴシップ送信実行
func (n *Node) SendGossip() (string, error) {
    target := n.selectRandomPeer()
    if target == "" {
        return "", fmt.Errorf("no peers available")
    }
    
    // メッセージ作成
    message := GossipMessage{
        From:      n.ID,
        Value:     n.GetValue(),
        Timestamp: time.Now().Unix(),
    }
    
    // HTTP送信
    err := n.sendHTTPMessage(target, message)
    if err != nil {
        return target, fmt.Errorf("failed to send to %s: %v", target, err)
    }
    
    log.Printf("[%s] Sent gossip to %s: value='%s'", n.ID, target, message.Value)
    return target, nil
}

// HTTP経由でメッセージ送信
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

// ゴシップメッセージ受信処理
func (n *Node) HandleGossipMessage(msg GossipMessage) {
    log.Printf("[%s] Received gossip from %s: value='%s'", 
               n.ID, msg.From, msg.Value)
    
    // 値を更新（内部でdiff確認）
    n.SetValue(msg.Value)
}
```

### http_server.go (HTTPサーバー)
```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
)

func startHTTPServer(node *Node) {
    mux := http.NewServeMux()
    
    // ゴシップメッセージ受信エンドポイント
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
        
        // ゴシップ処理
        node.HandleGossipMessage(msg)
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "received"})
    })
    
    // 手動ゴシップトリガー
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
    
    // ノード状態確認
    mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        status := node.GetStatus()
        json.NewEncoder(w).Encode(status)
    })
    
    // サーバー起動
    log.Fatal(http.ListenAndServe(node.Address, mux))
}
```

### config.json (設定ファイル)
```json
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
```

### go.mod (依存管理)
```go
module gossip-phase0

go 1.24
```

## 設計上の特徴

### 1. 最小限の依存
- 標準ライブラリのみ使用
- 外部パッケージなし（シンプル性重視）

### 2. 並行安全性
- `sync.RWMutex`でNode状態を保護
- HTTPサーバーは自然にgoroutineで並行処理

### 3. エラーハンドリング
- Phase 0では最小限（ログ出力中心）
- ネットワーク系エラーは上位に伝播

### 4. 観察性
- 重要な処理をすべてログ出力
- `/status`エンドポイントで状態確認可能

### 5. 拡張性
- インターフェースは使わず具体型（簡潔性重視）
- 後のPhaseでリファクタリング前提

この構造により、**半日で実装・動作確認**が可能な最小限のゴシップ実装を実現。
