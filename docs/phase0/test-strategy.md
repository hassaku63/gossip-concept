# Phase 0 テスト戦略

## テストの目的と範囲

### 主要目的
1. **ゴシップの本質が実装されているか確認** - ランダム選択による確率的伝搬
2. **基本動作の正常性確認** - HTTP通信、状態管理、JSON処理
3. **観察可能性の検証** - ログ出力、API応答の正確性

### 検証しない項目（Phase 0では除外）
- 性能・スループット
- 大規模ノードでの動作
- 障害耐性・ネットワーク分断
- セキュリティ

## 動作検証要件

### 必須要件 (MUST)
| ID | 要件 | 検証方法 |
|----|------|----------|
| R1 | 3ノードが正常起動できる | プロセス起動確認 |
| R2 | ランダム選択が機能する | 統計的検証（100回実行） |
| R3 | HTTP通信で状態送受信できる | APIテスト |
| R4 | 受信ノードで状態が更新される | 状態変化確認 |
| R5 | 全APIが正常応答する | エンドポイント疎通確認 |

### 推奨要件 (SHOULD) 
| ID | 要件 | 検証方法 |
|----|------|----------|
| R6 | ログ出力が適切である | ログ解析 |
| R7 | JSON形式が正しい | スキーマ検証 |
| R8 | 並行アクセスで競合しない | 基本的な並行テスト |

## テスト階層

### Layer 1: 単体テスト (Unit Tests)

#### 1.1 Node構造体テスト (`node_test.go`)

**テストケース**:
```go
func TestNode_GetSetValue(t *testing.T) {
    // thread-safe性の検証
    node := &Node{ID: "test", Value: "initial"}
    
    // 並行読み書きテスト
    go func() { node.SetValue("update1") }()
    go func() { node.SetValue("update2") }() 
    
    // データ競合がないこと確認
    finalValue := node.GetValue()
    assert.NotEmpty(t, finalValue)
}

func TestNode_GetStatus(t *testing.T) {
    // 状態取得の正確性
    node := NewNode("node-A", testConfig)
    status := node.GetStatus()
    
    assert.Equal(t, "node-A", status["id"])
    assert.Contains(t, status, "value")
    assert.Contains(t, status, "peers")
}
```

#### 1.2 ランダム選択テスト (`gossip_test.go`)

**最重要テスト**:
```go
func TestSelectRandomPeer_Distribution(t *testing.T) {
    node := &Node{
        Peers: []string{"peer1", "peer2", "peer3"},
    }
    
    counts := make(map[string]int)
    trials := 1000
    
    // 1000回実行して分布を確認
    for i := 0; i < trials; i++ {
        peer := node.selectRandomPeer()
        counts[peer]++
    }
    
    // 各ピアが200-500回選ばれること（統計的に妥当）
    for peer, count := range counts {
        assert.Greater(t, count, 200, "Peer %s selected too few times", peer)
        assert.Less(t, count, 500, "Peer %s selected too many times", peer)
    }
}

func TestSelectRandomPeer_EmptyPeers(t *testing.T) {
    node := &Node{Peers: []string{}}
    peer := node.selectRandomPeer()
    assert.Empty(t, peer)
}
```

#### 1.3 JSON処理テスト

```go
func TestGossipMessage_JSON(t *testing.T) {
    msg := GossipMessage{
        From:      "node-A",
        Value:     "test-value",
        Timestamp: 1640995200,
    }
    
    // Marshal/Unmarshal往復テスト
    data, err := json.Marshal(msg)
    assert.NoError(t, err)
    
    var decoded GossipMessage
    err = json.Unmarshal(data, &decoded)
    assert.NoError(t, err)
    assert.Equal(t, msg, decoded)
}
```

### Layer 2: 統合テスト (Integration Tests)

#### 2.1 HTTP通信テスト (`integration_test.go`)

```go
func TestHTTPGossip_EndToEnd(t *testing.T) {
    // テスト用の2ノード起動
    nodeA := startTestNode("node-A", ":18001")
    nodeB := startTestNode("node-B", ":18002") 
    defer nodeA.Close()
    defer nodeB.Close()
    
    // Node Aの値を変更
    setValueURL := "http://localhost:18001/set?value=test-gossip"
    resp, err := http.Post(setValueURL, "", nil)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // ゴシップ実行
    triggerURL := "http://localhost:18001/trigger"
    resp, err = http.Post(triggerURL, "", nil)
    assert.NoError(t, err)
    
    // Node Bの状態確認
    time.Sleep(100 * time.Millisecond) // 処理待ち
    statusResp, err := http.Get("http://localhost:18002/status")
    assert.NoError(t, err)
    
    var status map[string]interface{}
    json.NewDecoder(statusResp.Body).Decode(&status)
    assert.Equal(t, "test-gossip", status["value"])
}
```

#### 2.2 3ノード相互通信テスト

```go
func TestThreeNodeGossip(t *testing.T) {
    // 3ノード起動
    nodes := startThreeTestNodes()
    defer closeTestNodes(nodes)
    
    // Node-A で値変更
    setValue(nodes["A"], "propagated-value")
    
    // 複数回ゴシップ実行（確率的伝搬のため）
    for i := 0; i < 10; i++ {
        triggerGossip(nodes["A"])
        time.Sleep(50 * time.Millisecond)
    }
    
    // 最終的に全ノードが同じ値になることを確認
    eventually(t, func() bool {
        valueA := getValue(nodes["A"])
        valueB := getValue(nodes["B"])
        valueC := getValue(nodes["C"])
        return valueA == valueB && valueB == valueC && valueC == "propagated-value"
    }, 5*time.Second)
}
```

### Layer 3: システムテスト (System Tests)

#### 3.1 手動検証スクリプト (`manual_test.sh`)

```bash
#!/bin/bash
# Phase 0 手動テストスクリプト

set -e

echo "=== Phase 0 Manual Test Suite ==="

# 1. 起動テスト
echo "1. Starting 3 nodes..."
./gossip-phase0 -node node-A > nodeA.log 2>&1 &
PID_A=$!
sleep 1

./gossip-phase0 -node node-B > nodeB.log 2>&1 &
PID_B=$!
sleep 1

./gossip-phase0 -node node-C > nodeC.log 2>&1 &
PID_C=$!
sleep 2

echo "Nodes started: A($PID_A), B($PID_B), C($PID_C)"

# 2. ヘルスチェック
echo "2. Health check..."
curl -s localhost:8001/status | jq .id
curl -s localhost:8002/status | jq .id  
curl -s localhost:8003/status | jq .id

# 3. ランダム性テスト
echo "3. Random selection test..."
for i in {1..10}; do
    echo "Round $i:"
    curl -s -X POST localhost:8001/trigger | jq .target
    sleep 0.5
done

# 4. 状態伝搬テスト
echo "4. State propagation test..."
curl -s -X POST "localhost:8001/set?value=propagated-$(date +%s)"
sleep 1

# 複数回ゴシップ実行
for i in {1..5}; do
    curl -s -X POST localhost:8001/trigger > /dev/null
    sleep 0.5
done

# 最終状態確認
echo "Final state check:"
curl -s localhost:8001/status | jq .value
curl -s localhost:8002/status | jq .value
curl -s localhost:8003/status | jq .value

# クリーンアップ
kill $PID_A $PID_B $PID_C
echo "Test completed."
```

#### 3.2 自動化テストスクリプト (`auto_test.sh`)

```bash
#!/bin/bash
# 自動化された統合テスト

FAIL_COUNT=0

test_basic_functionality() {
    echo "Testing basic functionality..."
    
    # ノード起動
    start_test_nodes
    
    # 基本API確認
    if ! curl -s localhost:8001/status | jq -e '.id == "node-A"' > /dev/null; then
        echo "FAIL: Node A status check"
        ((FAIL_COUNT++))
    fi
    
    # ゴシップ実行
    if ! curl -s -X POST localhost:8001/trigger | jq -e '.status == "sent"' > /dev/null; then
        echo "FAIL: Gossip trigger"
        ((FAIL_COUNT++))
    fi
    
    cleanup_test_nodes
}

test_randomness() {
    echo "Testing randomness..."
    
    start_test_nodes
    
    declare -A targets
    for i in {1..50}; do
        target=$(curl -s -X POST localhost:8001/trigger | jq -r .target)
        targets[$target]=$((${targets[$target]}+1))
    done
    
    # 各ターゲットが最低10回は選ばれることを確認
    for target in "localhost:8002" "localhost:8003"; do
        if [[ ${targets[$target]} -lt 10 ]]; then
            echo "FAIL: Target $target selected only ${targets[$target]} times"
            ((FAIL_COUNT++))
        fi
    done
    
    cleanup_test_nodes
}

test_state_propagation() {
    echo "Testing state propagation..."
    
    start_test_nodes
    
    # 値設定
    test_value="test-$(date +%s)"
    curl -s -X POST "localhost:8001/set?value=$test_value"
    
    # 伝搬実行
    for i in {1..20}; do
        curl -s -X POST localhost:8001/trigger > /dev/null
        sleep 0.1
    done
    
    # 全ノードで値確認
    for port in 8001 8002 8003; do
        value=$(curl -s localhost:$port/status | jq -r .value)
        if [[ "$value" != "$test_value" ]]; then
            echo "FAIL: Node on port $port has value '$value', expected '$test_value'"
            ((FAIL_COUNT++))
        fi
    done
    
    cleanup_test_nodes
}

# テスト実行
test_basic_functionality
test_randomness  
test_state_propagation

if [[ $FAIL_COUNT -eq 0 ]]; then
    echo "All tests PASSED"
    exit 0
else
    echo "$FAIL_COUNT tests FAILED"
    exit 1
fi
```

## テスト環境

### テスト用設定 (`test_config.json`)
```json
{
  "nodes": {
    "node-A": {
      "address": "localhost:18001",
      "peers": ["localhost:18002", "localhost:18003"]
    },
    "node-B": {
      "address": "localhost:18002", 
      "peers": ["localhost:18001", "localhost:18003"]
    },
    "node-C": {
      "address": "localhost:18003",
      "peers": ["localhost:18001", "localhost:18002"]
    }
  },
  "initial_value": "test-initial"
}
```

## 成功基準

### 機能的成功基準
- [ ] **R1-R5の必須要件**すべてがパス
- [ ] **単体テスト**すべて通過（カバレッジ80%以上）
- [ ] **手動テスト**でゴシップの動作を目視確認
- [ ] **ランダム性**が統計的に検証される

### 観察的成功基準  
- [ ] ログで送信先がランダムに変わることを確認
- [ ] 状態変更→ゴシップ→伝搬の流れを追跡可能
- [ ] 最終的な収束を確認

### 定量的基準
- **ランダム性**: 100回実行で各ピアに20-60回の範囲で分散
- **伝搬成功率**: 20回のゴシップで90%以上の確率で全ノードに到達
- **API応答時間**: 全エンドポイントが100ms以内で応答

## テスト実行計画

### 開発フェーズでのテスト
1. **実装中**: 単体テストを随時実行
2. **機能完成時**: 統合テストで基本動作確認
3. **Phase 0完成時**: 手動テストでユーザー体験確認

### CI/CD（将来の拡張）
```yaml
# .github/workflows/test.yml (参考)
- name: Run Unit Tests
  run: go test ./... -v
- name: Run Integration Tests  
  run: ./auto_test.sh
- name: Verify Manual Test
  run: timeout 60s ./manual_test.sh
```

この戦略により、Phase 0の**180行の実装が確実に動作すること**を多角的に検証し、ゴシッププロトコルの本質が正しく実装されていることを確認できます。